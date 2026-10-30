package main_test

import (
	"slices"
	"testing"

	"github.com/zach/automap/internal/clitest"
)

// fixtureRepo is a repo with the fixture map copied to map/.
func fixtureRepo(t *testing.T) *clitest.Repo {
	t.Helper()
	repo := clitest.NewRepo(t, "no-manifest")
	repo.FixtureMap(t, "map")
	return repo
}

func TestShowLibraryPrintsItsFacts(t *testing.T) {
	run := fixtureRepo(t).Automap(t, "show", "library", "stub-window", "--map", "map", "--json")

	run.ExpectExit(t, 0)
	if len(run.Result.Entries) != 1 {
		t.Fatalf("entries %v, want one", run.Result.Entries)
	}
	e := run.Result.Entries[0]
	if e["kind"] != "library" || e["id"] != "stub-window" {
		t.Errorf("entry is %v %v, want library stub-window", e["kind"], e["id"])
	}
	if got := e.Get("definition", "versions", 0, "facts", "allocation", "hook_style"); got != "function-pointer-globals" {
		t.Errorf("allocation hook_style %v, want function-pointer-globals", got)
	}
	run.ExpectGolden(t, "show-library")
}

func TestShowSlotContractAndRecipe(t *testing.T) {
	for _, c := range []struct{ kind, id, golden string }{
		{"slot", "windowing", "show-slot"},
		{"contract", "allocation", "show-contract"},
		{"recipe", "stub-ui--stub-window", "show-recipe"},
	} {
		t.Run(c.kind, func(t *testing.T) {
			run := fixtureRepo(t).Automap(t, "show", c.kind, c.id, "--map", "map", "--json")

			run.ExpectExit(t, 0)
			if len(run.Result.Entries) != 1 || run.Result.Entries[0]["kind"] != c.kind || run.Result.Entries[0]["id"] != c.id {
				t.Errorf("entries %v, want %s %s", run.Result.Entries, c.kind, c.id)
			}
			run.ExpectGolden(t, c.golden)
		})
	}
}

func TestShowSlotNamesTheLibrariesThatFillIt(t *testing.T) {
	run := fixtureRepo(t).Automap(t, "show", "slot", "windowing", "--map", "map", "--json")

	if got := run.Result.Entries[0].Get("libraries", 0); got != "stub-window" {
		t.Errorf("libraries[0] %v, want stub-window", got)
	}
}

func TestShowUnknownEntry(t *testing.T) {
	run := fixtureRepo(t).Automap(t, "show", "library", "jolt", "--map", "map", "--json")

	run.ExpectExit(t, 2)
	run.ExpectFailure(t, "UNKNOWN_ENTRY(library jolt)", "jolt")
}

func TestFindBySlot(t *testing.T) {
	run := fixtureRepo(t).Automap(t, "find", "--slot", "tools-ui", "--map", "map", "--json")

	run.ExpectExit(t, 0)
	if len(run.Result.Entries) != 1 || run.Result.Entries[0]["id"] != "stub-ui" || run.Result.Entries[0]["version"] != "1.0.0" {
		t.Errorf("entries %v, want stub-ui 1.0.0", run.Result.Entries)
	}
	run.ExpectGolden(t, "find-slot")
}

func TestFindByQuery(t *testing.T) {
	run := fixtureRepo(t).Automap(t, "find", "window", "--map", "map", "--json")

	run.ExpectExit(t, 0)
	if len(run.Result.Entries) != 1 || run.Result.Entries[0]["id"] != "stub-window" {
		t.Errorf("entries %v, want stub-window", run.Result.Entries)
	}
}

func TestMapFromEnvironment(t *testing.T) {
	repo := fixtureRepo(t)
	repo.Setenv("AUTOMAP_MAP", "map")
	run := repo.Automap(t, "show", "slot", "windowing", "--json")

	run.ExpectExit(t, 0)
	run.ExpectGolden(t, "show-slot")
}

func TestNoMap(t *testing.T) {
	run := clitest.Automap(t, "no-manifest", "find", "--json")

	run.ExpectExit(t, 2)
	run.ExpectFailure(t, "NO_MAP", "--map")
}

// The spec's examples: a public map and two private maps layered in the
// manifest's order.
func acmeRepo(t *testing.T) *clitest.Repo {
	t.Helper()
	repo := clitest.NewRepo(t, "acme-engine")
	for _, m := range []string{"map", "private-map", "private-map-nebula"} {
		repo.Copy(t, "docs/spec/examples/"+m, m)
	}
	return repo
}

func TestPrivateMapsLayerInManifestOrder(t *testing.T) {
	run := acmeRepo(t).Automap(t, "show", "library", "sdl3", "--map", "map", "--json")

	run.ExpectExit(t, 0)
	e := run.Result.Entries[0]
	var variants []string
	for _, v := range e.Get("definition", "versions", 0, "variants").([]any) {
		variants = append(variants, v.(map[string]any)["name"].(string))
	}
	if !slices.Equal(variants, []string{"default", "acme-patched", "nebula"}) {
		t.Errorf("variants %v, want default, then acme's acme-patched, then acme-nebula's nebula", variants)
	}
	run.ExpectGolden(t, "show-library-layered")
}

func TestNDAMapAddsItsPlatformToTargets(t *testing.T) {
	run := acmeRepo(t).Automap(t, "show", "library", "jolt", "--map", "map", "--json")

	run.ExpectExit(t, 0)
	targets := run.Result.Entries[0].Get("definition", "versions", 0, "facts", "platform", "targets").([]any)
	if targets[len(targets)-1] != "nebula-arm64-clang" {
		t.Errorf("targets %v, want nebula-arm64-clang added by the NDA map", targets)
	}
}

func TestFindSeesPrivateLibraries(t *testing.T) {
	run := acmeRepo(t).Automap(t, "find", "--slot", "jobs", "--map", "map", "--json")

	run.ExpectExit(t, 0)
	var ids []string
	for _, e := range run.Result.Entries {
		ids = append(ids, e["id"].(string))
	}
	if !slices.Equal(ids, []string{"acme-jobs"}) {
		t.Errorf("libraries filling jobs %v, want acme-jobs", ids)
	}
}
