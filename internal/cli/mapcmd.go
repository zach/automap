package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/zach/automap/internal/composition"
	"github.com/zach/automap/internal/mapmodel"
	"github.com/zach/automap/internal/release"
	"github.com/zach/automap/schemas"
)

// loadMap loads the public map from --map or AUTOMAP_MAP, and the private
// maps the manifest in the current directory names, if there is one. On
// failure it records why in r and returns a nil map and the exit status.
func loadMap(a *args, r *result) (*mapmodel.Map, int) {
	dir := a.value("map")
	if dir == "" {
		dir = os.Getenv("AUTOMAP_MAP")
	}
	if dir == "" {
		var ok bool
		if dir, ok = releaseDir(r); !ok {
			return nil, exitUsage
		}
	}
	privates, ok := manifestPrivateMaps(r)
	if !ok {
		return nil, exitUsage
	}
	m, problems := mapmodel.Load(dir, privates)
	for _, p := range problems {
		r.fail("MAP_INVALID("+p.File+")", p.Reason)
	}
	if len(problems) == 0 {
		checkPrivateMaps(m, privates, r)
	}
	if len(r.Failures) > 0 {
		r.Summary = fmt.Sprintf("The map is invalid: %d %s.", len(r.Failures), plural(len(r.Failures), "problem", "problems"))
		return nil, exitUsage
	}
	return m, exitOK
}

// releaseDir finds the map release to read when no map directory is
// given: the one the lock in the current directory records, or else the
// newest, fetched into the cache if it isn't there (ADR-0053, ADR-0038).
func releaseDir(r *result) (string, bool) {
	var lock struct {
		MapRelease string `json:"map_release"`
	}
	tag := ""
	if raw, err := os.ReadFile(composition.LockFile); err == nil && json.Unmarshal(raw, &lock) == nil {
		tag = lock.MapRelease
	}
	var dir string
	var err error
	switch {
	case release.Tag.MatchString(tag):
		dir, err = release.Open(tag)
	case tag != "":
		r.Summary = "No map: the lock was resolved against a map directory."
		r.fail("NO_MAP", "The lock records map release "+tag+", which isn't a published release: pass --map <dir> or set AUTOMAP_MAP to that map's directory.")
		return "", false
	default:
		tag, dir, err = release.Newest()
	}
	var invalid *release.InvalidError
	switch {
	case errors.As(err, &invalid):
		r.Summary = "Map release " + invalid.Tag + " is invalid."
		r.fail("MAP_RELEASE_INVALID("+invalid.Tag+")", invalid.Reason)
		return "", false
	case err != nil:
		r.Summary = "No map: no map release could be fetched or found in the cache."
		r.fail("NO_MAP", err.Error()+". Pass --map <dir> or set AUTOMAP_MAP to the public map's directory, or fetch a release with the network on.")
		return "", false
	}
	return dir, true
}

// manifestPrivateMaps reads the private maps the manifest names, in
// layering order. A repo without a manifest has none.
func manifestPrivateMaps(r *result) ([]mapmodel.PrivateMap, bool) {
	raw, err := os.ReadFile(manifestFile)
	if os.IsNotExist(err) {
		return nil, true
	}
	var doc any
	if err == nil {
		err = json.Unmarshal(raw, &doc)
	}
	if err != nil {
		r.Summary = "The manifest can't be read."
		r.fail("MANIFEST_INVALID", err.Error())
		return nil, false
	}
	if v := schemas.Validate("manifest", doc); v != nil {
		r.Summary = "The manifest doesn't match its schema."
		for _, msg := range v {
			r.fail("MANIFEST_INVALID", "manifest: "+msg)
		}
		return nil, false
	}
	var m struct {
		PrivateMaps []struct {
			ID        string   `json:"id"`
			Location  string   `json:"location"`
			Platforms []string `json:"platforms"`
		} `json:"private_maps"`
	}
	json.Unmarshal(raw, &m)
	var out []mapmodel.PrivateMap
	for _, pm := range m.PrivateMaps {
		out = append(out, mapmodel.PrivateMap{ID: pm.ID, Location: pm.Location, Platforms: pm.Platforms})
	}
	return out, true
}

// checkPrivateMaps checks the manifest's private maps against their map.json.
func checkPrivateMaps(m *mapmodel.Map, privates []mapmodel.PrivateMap, r *result) {
	for i, pm := range privates {
		meta := m.Maps[i+1]
		if meta.ID != pm.ID {
			r.fail("MANIFEST_INVALID", fmt.Sprintf("private map %s is at %s, whose map.json has id %s", pm.ID, pm.Location, meta.ID))
			continue
		}
		a, b := slices.Clone(pm.Platforms), slices.Clone(meta.Platforms)
		slices.Sort(a)
		slices.Sort(b)
		if !slices.Equal(a, b) {
			r.fail("MANIFEST_INVALID", fmt.Sprintf("private map %s: platforms differ from its map.json", pm.ID))
		}
	}
}

func show(a *args, r *result) int {
	if len(a.positional) != 2 {
		r.Summary = "Usage error: show takes a kind and an id."
		r.fail("USAGE", "automap show <library|slot|contract|recipe> <id>")
		return exitUsage
	}
	kind, id := a.positional[0], a.positional[1]
	m, status := loadMap(a, r)
	if m == nil {
		return status
	}
	e := entry{"kind": kind, "id": id, "maps": m.Origins(kind, id)}
	found := true
	switch kind {
	case "library":
		lib := m.Libraries[id]
		found = lib != nil
		e["definition"] = lib
		if evidence := libraryEvidence(m, id, time.Now()); found && len(evidence) > 0 {
			e["evidence"] = evidence
			warnEvidence(r, evidence)
		}
	case "slot":
		slot := m.Slots[id]
		found = slot != nil
		e["definition"] = slot
		e["libraries"] = librariesFilling(m, id)
	case "contract":
		versions := m.Contracts[id]
		found = len(versions) > 0
		if found {
			e["definition"] = versions[len(versions)-1]
			var vs []string
			for _, c := range versions {
				vs = append(vs, c.Version)
			}
			e["versions"] = vs
		}
	case "recipe":
		recipe := m.Recipes[id]
		found = recipe != nil
		e["definition"] = recipe
	default:
		r.Summary = "Usage error: unknown kind " + kind + "."
		r.fail("USAGE", "show takes library, slot, contract or recipe")
		return exitUsage
	}
	if !found {
		r.Summary = fmt.Sprintf("No %s %s in the map.", kind, id)
		r.fail(fmt.Sprintf("UNKNOWN_ENTRY(%s %s)", kind, id), fmt.Sprintf("The map has no %s with id %s; automap find lists libraries.", kind, id))
		return exitUsage
	}
	r.OK = true
	r.Summary = fmt.Sprintf("%s %s, from %s.", kind, id, strings.Join(e["maps"].([]string), ", "))
	r.Entries = []entry{e}
	return exitOK
}

func librariesFilling(m *mapmodel.Map, slot string) []string {
	out := []string{}
	for id, lib := range m.Libraries {
		if slices.Contains(lib.Slots, slot) {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

func find(a *args, r *result) int {
	if len(a.positional) > 1 {
		r.Summary = "Usage error: find takes at most one query."
		r.fail("USAGE", "automap find [<query>] [--slot <slot>]")
		return exitUsage
	}
	m, status := loadMap(a, r)
	if m == nil {
		return status
	}
	slot := a.value("slot")
	query := ""
	if len(a.positional) == 1 {
		query = strings.ToLower(a.positional[0])
	}
	if slot != "" && m.Slots[slot] == nil {
		r.Summary = "No slot " + slot + " in the map."
		r.fail(fmt.Sprintf("UNKNOWN_ENTRY(slot %s)", slot), "The map has no slot with id "+slot+".")
		return exitUsage
	}
	var ids []string
	for id, lib := range m.Libraries {
		if slot != "" && !slices.Contains(lib.Slots, slot) {
			continue
		}
		if query != "" && !strings.Contains(id, query) && !strings.Contains(strings.ToLower(lib.Name), query) {
			continue
		}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var found []string
	for _, id := range ids {
		v := m.Libraries[id].Newest()
		var variants []string
		for _, x := range v.Variants {
			variants = append(variants, x.Name)
		}
		e := entry{"kind": "library", "id": id, "version": v.Version, "variants": variants}
		if evidence := libraryEvidence(m, id, time.Now()); len(evidence) > 0 {
			e["evidence"] = evidence
			warnEvidence(r, evidence)
		}
		r.Entries = append(r.Entries, e)
		found = append(found, id+" "+v.Version)
	}
	r.OK = true
	what := "in the map"
	switch {
	case slot != "" && query != "":
		what = fmt.Sprintf("fill %s and match %q", slot, query)
	case slot != "":
		what = "fill " + slot
	case query != "":
		what = fmt.Sprintf("match %q", query)
	}
	if len(found) == 0 {
		r.Summary = "No libraries " + what + "."
		return exitOK
	}
	r.Summary = fmt.Sprintf("%d %s %s: %s.", len(found), plural(len(found), "library", "libraries"), what, strings.Join(found, ", "))
	if len(found) == 1 && slot != "" {
		r.Summary = fmt.Sprintf("1 library fills %s: %s.", slot, found[0])
	}
	return exitOK
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
