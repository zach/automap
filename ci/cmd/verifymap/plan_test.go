package main

import (
	"fmt"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/zach/automap/internal/clitest"
	"github.com/zach/automap/internal/mapmodel"
)

var now = time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)

var testbed = []Reference{{ID: "testbed", With: []string{"stub-window", "stub-ui"}}}

func fixtureMap(t *testing.T) *mapmodel.Map {
	t.Helper()
	repo := clitest.NewRepo(t, "no-manifest")
	repo.FixtureMap(t, "map")
	m, problems := mapmodel.Load(filepath.Join(repo.Dir, "map"), nil)
	if len(problems) > 0 {
		t.Fatalf("fixture map: %v", problems)
	}
	return m
}

func names(edges []Edge) []string {
	var out []string
	for _, e := range edges {
		out = append(out, fmt.Sprintf("%s %s %v", e.Kind, e.ID, e.Picks))
	}
	slices.Sort(out)
	return out
}

func TestEveryDeclaredEdgeAtEachVariant(t *testing.T) {
	got := names(Edges(fixtureMap(t), testbed))
	want := []string{
		"binding stub-physics--jobs [{stub-physics 1.0.0 default}]",
		"binding stub-physics--math-types [{stub-physics 1.0.0 default}]",
		"binding stub-ui--allocation [{stub-ui 1.0.0 default}]",
		"binding stub-ui--diagnostics [{stub-ui 1.0.0 default}]",
		"binding stub-window--allocation [{stub-window 1.0.0 default}]",
		"binding stub-window--allocation [{stub-window 1.0.0 wide-events}]",
		"binding stub-window--diagnostics [{stub-window 1.0.0 default}]",
		"binding stub-window--diagnostics [{stub-window 1.0.0 wide-events}]",
		"composition testbed [{stub-window 1.0.0 default} {stub-ui 1.0.0 default}]",
		"pairing stub-ui--stub-window [{stub-ui 1.0.0 default} {stub-window 1.0.0 default}]",
		"pairing stub-ui--stub-window [{stub-ui 1.0.0 default} {stub-window 1.0.0 wide-events}]",
	}
	if !slices.Equal(got, want) {
		t.Errorf("edges:\n%s\nwant:\n%s", fmt.Sprint(got), fmt.Sprint(want))
	}
}

func planned(jobs []Job) []string {
	var out []string
	for _, j := range jobs {
		for _, e := range j.Edges {
			out = append(out, fmt.Sprintf("%s %s %v on %s", e.Kind, e.ID, e.Picks, j.Platform))
		}
	}
	slices.Sort(out)
	return slices.Compact(out)
}

func TestEdgesWithoutEvidenceAreVerifiedOnEveryPlatform(t *testing.T) {
	m := fixtureMap(t)
	edges := Edges(m, testbed)
	platforms := []string{"linux-x64-clang", "windows-x64-msvc"}

	jobs := Plan(m, edges, platforms, func(Edge, string) *Published { return nil }, touches(nil), now)

	if got := len(planned(jobs)); got != len(edges)*len(platforms) {
		t.Errorf("planned %d edge runs, want %d", got, len(edges)*len(platforms))
	}
	// One job per composition and platform: stub-window's two bindings
	// share one, and the pairing shares the testbed's.
	var compositions []string
	for _, j := range jobs {
		if j.Platform == "linux-x64-clang" {
			compositions = append(compositions, j.Name)
		}
	}
	slices.Sort(compositions)
	want := []string{
		"stub-physics@1.0.0:default on linux-x64-clang",
		"stub-ui@1.0.0:default on linux-x64-clang",
		"stub-ui@1.0.0:default stub-window@1.0.0:default on linux-x64-clang",
		"stub-ui@1.0.0:default stub-window@1.0.0:wide-events on linux-x64-clang",
		"stub-window@1.0.0:default on linux-x64-clang",
		"stub-window@1.0.0:wide-events on linux-x64-clang",
	}
	if !slices.Equal(compositions, want) {
		t.Errorf("jobs on linux:\n%v\nwant:\n%v", compositions, want)
	}
}

func TestChangingOneRecipeReverifiesExactlyTheEdgesThatDependOnIt(t *testing.T) {
	m := fixtureMap(t)
	fresh := func(Edge, string) *Published {
		return &Published{ProducedAt: now.Add(-24 * time.Hour), Commit: "c0ffee"}
	}

	jobs := Plan(m, Edges(m, testbed), []string{"linux-x64-clang"}, fresh, touches([]string{"map/recipes/binding/stub-ui--allocation/stub_ui_allocation.cpp"}), now)

	want := []string{
		"binding stub-ui--allocation [{stub-ui 1.0.0 default}] on linux-x64-clang",
		"composition testbed [{stub-window 1.0.0 default} {stub-ui 1.0.0 default}] on linux-x64-clang",
	}
	if got := planned(jobs); !slices.Equal(got, want) {
		t.Errorf("planned:\n%v\nwant:\n%v", got, want)
	}
}

func TestEvidenceNearingStaleIsReverified(t *testing.T) {
	m := fixtureMap(t)
	old := func(e Edge, _ string) *Published {
		age := 24 * time.Hour
		if e.ID == "stub-ui--diagnostics" {
			age = MaxAge + time.Hour
		}
		return &Published{ProducedAt: now.Add(-age), Commit: "c0ffee"}
	}

	jobs := Plan(m, Edges(m, testbed), []string{"linux-x64-clang"}, old, touches(nil), now)

	if got, want := planned(jobs), []string{"binding stub-ui--diagnostics [{stub-ui 1.0.0 default}] on linux-x64-clang"}; !slices.Equal(got, want) {
		t.Errorf("planned %v, want %v", got, want)
	}
}

func TestALibraryIsVerifiedOnlyWhereItBuilds(t *testing.T) {
	m := fixtureMap(t)
	m.Libraries["stub-physics"].Versions[0].Facts["platform"]["targets"] = []any{"linux-x64-clang"}

	jobs := Plan(m, Edges(m, nil), []string{"windows-x64-msvc"}, func(Edge, string) *Published { return nil }, touches(nil), now)

	for _, j := range jobs {
		for _, p := range j.Picks {
			if p.Library == "stub-physics" {
				t.Errorf("job %s verifies stub-physics on a platform it doesn't target", j.Name)
			}
		}
	}
}
