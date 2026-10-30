package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/zach/automap/internal/mapmodel"
)

// MaxAge is how old an edge's evidence may get before a scheduled run
// re-verifies it, leaving a week's margin before it goes stale at 30 days
// (ADR-0031, ADR-0033).
const MaxAge = 23 * 24 * time.Hour

// Pick is one library in a composition Automap's CI verifies: its version
// and variant.
type Pick struct {
	Library string `json:"library"`
	Version string `json:"version"`
	Variant string `json:"variant"`
}

// Edge is one thing Automap's CI publishes evidence for (ADR-0033): a
// binding, a pairing with a recipe, or a reference composition, at exact
// versions and variants.
type Edge struct {
	Kind  string `json:"kind"` // binding, pairing or composition
	ID    string `json:"id"`   // the recipe id, or the reference composition's
	Picks []Pick `json:"picks"`
}

// Dir is where the edge's evidence lives in the map, without the platform
// file: evidence/<kind>/<id>/<pins>, where a composition's pins are its
// lock hash, known only once it has been composed.
func (e Edge) Dir(lockHash string) string {
	var pins []string
	for _, p := range e.Picks {
		pins = append(pins, p.Library+"-"+p.Version+"-"+p.Variant)
	}
	name := strings.Join(pins, "+")
	if e.Kind == "composition" {
		name = strings.TrimPrefix(lockHash, "sha256:")[:16]
	}
	return path.Join("evidence", e.Kind, e.ID, name)
}

// Reference is one of the reference compositions (SPEC §8.6), from
// ci/compositions.json.
type Reference struct {
	ID   string   `json:"id"`
	With []string `json:"with"` // library[:variant], as automap start takes them
}

// Job is one composition verified on one platform, and the edges whose
// evidence it publishes.
type Job struct {
	Name     string `json:"name"`
	Platform string `json:"platform"`
	Picks    []Pick `json:"picks"`
	Edges    []Edge `json:"edges"`
}

// Edges lists every edge the map declares: each binding and pairing recipe
// at the latest and previous versions it applies to and each variant, and
// each reference composition at the libraries' newest versions and the
// variants it names.
func Edges(m *mapmodel.Map, refs []Reference) []Edge {
	var out []Edge
	ids := make([]string, 0, len(m.Recipes))
	for id := range m.Recipes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		r := m.Recipes[id]
		switch r.Kind {
		case "binding":
			for _, p := range picks(m, r, r.Subject.Library) {
				out = append(out, Edge{Kind: "binding", ID: id, Picks: []Pick{p}})
			}
		case "pairing":
			libs := r.Subject.Libraries
			if len(libs) != 2 {
				continue
			}
			for _, a := range picks(m, r, libs[0]) {
				for _, b := range picks(m, r, libs[1]) {
					out = append(out, Edge{Kind: "pairing", ID: id, Picks: []Pick{a, b}})
				}
			}
		}
	}
	for _, ref := range refs {
		e := Edge{Kind: "composition", ID: ref.ID}
		for _, w := range ref.With {
			lib, variant, _ := strings.Cut(w, ":")
			l := m.Libraries[lib]
			if l == nil {
				continue
			}
			v := l.Newest()
			if variant == "" {
				variant = v.Variants[0].Name
			}
			e.Picks = append(e.Picks, Pick{Library: lib, Version: v.Version, Variant: variant})
		}
		out = append(out, e)
	}
	return out
}

// picks lists the versions and variants of a library a recipe applies to,
// among the library's latest and previous versions.
func picks(m *mapmodel.Map, r *mapmodel.Recipe, lib string) []Pick {
	l := m.Libraries[lib]
	if l == nil {
		return nil
	}
	versions := slices.Clone(l.Versions)
	sort.SliceStable(versions, func(i, j int) bool { return versions[i].Released > versions[j].Released })
	if len(versions) > 2 {
		versions = versions[:2]
	}
	applies := r.AppliesTo[lib]
	var out []Pick
	for _, v := range versions {
		if len(applies.Versions) > 0 && !slices.Contains(applies.Versions, v.Version) {
			continue
		}
		for _, variant := range v.Variants {
			if len(applies.Variants) > 0 && !slices.Contains(applies.Variants, variant.Name) {
				continue
			}
			out = append(out, Pick{Library: lib, Version: v.Version, Variant: variant.Name})
		}
	}
	return out
}

// Inputs lists the repository paths an edge's evidence depends on: its
// libraries' entries and library recipes, its own recipe, the contracts,
// the probe harness and testbed template, and the CLI itself. A change to
// any of them re-verifies the edge.
func Inputs(m *mapmodel.Map, e Edge) []string {
	in := []string{"map/map.json", "map/probe-harness", "map/testbed", "map/contracts", "map/slots",
		"cmd", "internal", "schemas", "go.mod", "go.sum"}
	for _, p := range e.Picks {
		in = append(in, "map/libraries/"+p.Library+".json", "map/recipes/library/"+p.Library+".json", "map/recipes/library/"+p.Library)
	}
	switch e.Kind {
	case "binding", "pairing":
		in = append(in, "map/recipes/"+e.Kind+"/"+e.ID+".json", "map/recipes/"+e.Kind+"/"+e.ID)
	case "composition":
		// A reference composition depends on every recipe among its
		// libraries.
		libs := map[string]bool{}
		for _, p := range e.Picks {
			libs[p.Library] = true
		}
		for id, r := range m.Recipes {
			among := libs[r.Subject.Library]
			if r.Kind == "pairing" {
				among = len(r.Subject.Libraries) > 0
				for _, l := range r.Subject.Libraries {
					among = among && libs[l]
				}
			}
			if among && r.Kind != "library" {
				in = append(in, "map/recipes/"+r.Kind+"/"+id+".json", "map/recipes/"+r.Kind+"/"+id)
			}
		}
		in = append(in, "ci/compositions.json")
	}
	sort.Strings(in)
	return slices.Compact(in)
}

// Published is what the map's evidence for one edge on one platform says
// about when and against which commit it was verified.
type Published struct {
	ProducedAt time.Time
	Commit     string // from map_release git:<commit>, or "" for any other release
}

// Changed reports whether any of the paths changed between a commit and
// the checkout being planned.
type Changed func(commit string, paths []string) bool

// Plan decides which edges need verifying on which platforms: those with
// no evidence there, with evidence older than MaxAge, or whose inputs
// changed since the commit their evidence was verified at. It groups them
// into one job per composition and platform. It is a pure function.
func Plan(m *mapmodel.Map, edges []Edge, platforms []string, published func(e Edge, platform string) *Published, changed Changed, now time.Time) []Job {
	jobs := map[string]*Job{}
	var order []string
	for _, platform := range platforms {
		for _, e := range edges {
			if !needs(m, e, platform, published(e, platform), changed, now) {
				continue
			}
			key := platform + " " + compositionKey(e)
			j := jobs[key]
			if j == nil {
				// Edges with the same libraries share one composition: a
				// library's bindings, or a pairing and the reference
				// composition of the same two libraries.
				j = &Job{Name: compositionKey(e) + " on " + platform, Platform: platform, Picks: e.Picks}
				jobs[key] = j
				order = append(order, key)
			}
			j.Edges = append(j.Edges, e)
		}
	}
	out := make([]Job, 0, len(order))
	for _, k := range order {
		out = append(out, *jobs[k])
	}
	return out
}

func needs(m *mapmodel.Map, e Edge, platform string, p *Published, changed Changed, now time.Time) bool {
	for _, pick := range e.Picks {
		v := version(m, pick)
		if v == nil {
			return false
		}
		targets, _ := v.Facts["platform"]["targets"].([]any)
		if !slices.ContainsFunc(targets, func(t any) bool { return t == platform }) {
			return false // the library doesn't build there
		}
	}
	switch {
	case p == nil, p.Commit == "", now.Sub(p.ProducedAt) > MaxAge:
		return true
	}
	return changed(p.Commit, Inputs(m, e))
}

func version(m *mapmodel.Map, p Pick) *mapmodel.Version {
	l := m.Libraries[p.Library]
	if l == nil {
		return nil
	}
	for i := range l.Versions {
		if l.Versions[i].Version == p.Version {
			return &l.Versions[i]
		}
	}
	return nil
}

// compositionKey names the composition an edge verifies in: its libraries at
// their versions and variants.
func compositionKey(e Edge) string {
	var parts []string
	for _, p := range e.Picks {
		parts = append(parts, p.Library+"@"+p.Version+":"+p.Variant)
	}
	sort.Strings(parts)
	return strings.Join(parts, " ")
}

// readReferences reads ci/compositions.json.
func readReferences(file string) ([]Reference, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var refs []Reference
	if err := json.Unmarshal(raw, &refs); err != nil {
		return nil, fmt.Errorf("%s: %v", file, err)
	}
	return refs, nil
}

// touches is the Changed for a list of files that changed: an input
// changed when a file is the input or lies under it.
func touches(files []string) Changed {
	return func(_ string, inputs []string) bool {
		for _, f := range files {
			for _, in := range inputs {
				if f == in || strings.HasPrefix(f, in+"/") {
					return true
				}
			}
		}
		return false
	}
}
