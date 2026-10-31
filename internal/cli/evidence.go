package cli

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/zach/automap/internal/composition"
	"github.com/zach/automap/internal/mapmodel"
	"github.com/zach/automap/internal/release"
)

// published is one piece of the map's evidence a command reports, with its
// state (ADR-0031).
type published struct {
	Subject  string `json:"subject"`
	Platform string `json:"platform"`
	Level    string `json:"level"`
	State    string `json:"state"`
}

// edgePin is a library at a version and variant, as evidence paths name it.
type edgePin struct{ library, version, variant string }

func (p edgePin) String() string { return p.library + "-" + p.version + "-" + p.variant }

// readPublished reads the evidence for a binding or pairing at exact pins
// on each platform, from the map the edge's recipe and variants come from:
// the topmost that adds any of them (ADR-0047). That is a map's own
// directory, or for the public map read from a release, its main branch,
// where Automap's CI adds it, or the cache when offline (ADR-0082). Missing
// evidence is left out.
func readPublished(m *mapmodel.Map, kind, id string, pins []edgePin, platforms []string, now time.Time) []published {
	pieces := [][2]string{{"recipe", id}}
	for _, p := range pins {
		pieces = append(pieces, [2]string{"variant", p.library + ":" + p.version + ":" + p.variant})
	}
	owner := m.MapFor(pieces...)
	fromRelease := owner == m.Maps[0] && release.Tag.MatchString(mapRelease(m.PublicDir()))
	var out []published
	for _, platform := range platforms {
		rel := evidencePath(kind, id, pins, platform)
		var raw []byte
		if fromRelease {
			raw, _ = release.Evidence(rel)
		} else {
			raw, _ = os.ReadFile(filepath.Join(owner.Dir, filepath.FromSlash(rel)))
		}
		if raw == nil {
			continue
		}
		level, state, err := release.State(raw, now)
		if err != nil {
			continue
		}
		out = append(out, published{Subject: id, Platform: platform, Level: level, State: state})
	}
	return out
}

// evidencePath is where a binding's or pairing's evidence at exact pins on
// a platform is in a map.
func evidencePath(kind, id string, pins []edgePin, platform string) string {
	var names []string
	for _, p := range pins {
		names = append(names, p.String())
	}
	return path.Join("evidence", kind, id, strings.Join(names, "+"), platform+".json")
}

// libraryEvidence is the published evidence for a library's bindings and
// pairings at its newest version and default variant, and the other
// library's in a pairing, on each platform it targets.
func libraryEvidence(m *mapmodel.Map, lib string, now time.Time) []published {
	l := m.Libraries[lib]
	if l == nil {
		return nil
	}
	pinOf := func(id string) edgePin {
		v := m.Libraries[id].Newest()
		return edgePin{id, v.Version, v.Variants[0].Name}
	}
	var targets []string
	if raw, ok := l.Newest().Facts["platform"]["targets"].([]any); ok {
		for _, t := range raw {
			targets = append(targets, fmt.Sprint(t))
		}
	}
	var out []published
	for _, id := range sortedRecipeIDs(m) {
		r := m.Recipes[id]
		switch {
		case r.Kind == "binding" && r.Subject.Library == lib:
			out = append(out, readPublished(m, "binding", id, []edgePin{pinOf(lib)}, targets, now)...)
		case r.Kind == "pairing" && len(r.Subject.Libraries) == 2 && (r.Subject.Libraries[0] == lib || r.Subject.Libraries[1] == lib):
			if m.Libraries[r.Subject.Libraries[0]] == nil || m.Libraries[r.Subject.Libraries[1]] == nil {
				continue
			}
			out = append(out, readPublished(m, "pairing", id, []edgePin{pinOf(r.Subject.Libraries[0]), pinOf(r.Subject.Libraries[1])}, targets, now)...)
		}
	}
	return out
}

// edge is one of a composition's bindings or pairings at the lock's exact
// pins.
type edge struct {
	kind, id string
	pins     []edgePin
}

// compositionEdges are the composition's bindings and pairings at the
// lock's exact pins.
func compositionEdges(m *mapmodel.Map, man *composition.Manifest, lock *composition.Lock) []edge {
	pinOf := func(lib string) edgePin {
		p := lock.Pins[lib]
		return edgePin{lib, p.Version, p.Variant}
	}
	var out []edge
	for _, id := range composition.Recipes(m, man) {
		r := m.Recipes[id]
		switch r.Kind {
		case "binding":
			out = append(out, edge{"binding", id, []edgePin{pinOf(r.Subject.Library)}})
		case "pairing":
			if len(r.Subject.Libraries) == 2 {
				out = append(out, edge{"pairing", id, []edgePin{pinOf(r.Subject.Libraries[0]), pinOf(r.Subject.Libraries[1])}})
			}
		}
	}
	return out
}

// compositionEvidence is the published evidence for the composition's
// bindings and pairings at the lock's exact pins, on the manifest's
// platforms.
func compositionEvidence(m *mapmodel.Map, man *composition.Manifest, lock *composition.Lock, now time.Time) []published {
	var out []published
	for _, e := range compositionEdges(m, man, lock) {
		out = append(out, readPublished(m, e.kind, e.id, e.pins, man.Platforms, now)...)
	}
	return out
}

// warnEvidence warns about every stale or broken piece of evidence.
func warnEvidence(r *result, evidence []published) {
	for _, e := range evidence {
		switch e.State {
		case "stale":
			r.warn(fmt.Sprintf("EVIDENCE_STALE(%s,%s)", e.Subject, e.Platform),
				fmt.Sprintf("The published evidence for %s on %s is more than 30 days old; it isn't known-good until Automap's CI verifies it again.", e.Subject, e.Platform))
		case "broken":
			r.warn(fmt.Sprintf("EVIDENCE_BROKEN(%s,%s)", e.Subject, e.Platform),
				fmt.Sprintf("Automap's latest run of %s on %s failed (level %s).", e.Subject, e.Platform, e.Level))
		}
	}
}

func sortedRecipeIDs(m *mapmodel.Map) []string {
	ids := make([]string, 0, len(m.Recipes))
	for id := range m.Recipes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
