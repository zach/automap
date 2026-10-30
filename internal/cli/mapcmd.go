package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/zach/automap/internal/mapmodel"
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
		r.Summary = "No map: give the public map's directory."
		r.fail("NO_MAP", "Pass --map <dir> or set AUTOMAP_MAP to the public map's directory.")
		return nil, exitUsage
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
		r.Entries = append(r.Entries, entry{"kind": "library", "id": id, "version": v.Version, "variants": variants})
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
