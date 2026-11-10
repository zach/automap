package main_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/zach/automap/internal/clitest"
)

// repoFiles lists the repo's files, leaving out the fixture map, the
// test's cache and git.
func repoFiles(t *testing.T, repo *clitest.Repo) map[string]string {
	t.Helper()
	out := map[string]string{}
	filepath.WalkDir(repo.Dir, func(p string, d os.DirEntry, err error) error {
		rel, _ := filepath.Rel(repo.Dir, p)
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if rel == "map" || rel == ".automap-test-cache" || rel == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		raw, _ := os.ReadFile(p)
		out[rel] = string(raw)
		return nil
	})
	return out
}

// remove is the other half of extend: taking out the library extend added
// leaves the repo as it was before, file for file.
func TestRemoveUndoesExtend(t *testing.T) {
	repo := fixtureRepo(t)
	repo.Automap(t, startTestbed...).ExpectExit(t, 0)
	before := repoFiles(t, repo)
	repo.Automap(t, "extend", "stub-physics", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)

	run := repo.Automap(t, "remove", "stub-physics", "--map", "map", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	after := repoFiles(t, repo)
	var paths []string
	for p := range before {
		paths = append(paths, p)
	}
	for p := range after {
		if _, ok := before[p]; !ok {
			paths = append(paths, p)
		}
	}
	sort.Strings(paths)
	for _, p := range paths {
		switch {
		case before[p] == "":
			t.Errorf("remove left %s, which extend wrote", p)
		case after[p] == "":
			t.Errorf("remove deleted %s, which the composition had before stub-physics", p)
		case before[p] != after[p]:
			t.Errorf("%s differs from what it was before stub-physics was added:\n%s", p, after[p])
		}
	}
	if _, err := os.Stat(filepath.Join(repo.Dir, "glue", "stub-physics")); err == nil {
		t.Error("remove left the empty directory glue/stub-physics")
	}
	if !strings.HasPrefix(run.Result.Summary, "Removed stub-physics from physics-3d") {
		t.Errorf("summary %q, want it to begin: Removed stub-physics from physics-3d", run.Result.Summary)
	}
}

// Glue the developer changed isn't remove's to delete: it stays, with a
// warning that names it.
func TestRemoveLeavesGlueTheDeveloperChanged(t *testing.T) {
	repo := fixtureRepo(t)
	repo.Automap(t, startTestbed...).ExpectExit(t, 0)
	repo.Automap(t, "extend", "stub-physics", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)
	const mine = "glue/stub-physics/stub_physics_setup.cpp"
	repo.WriteFile(t, mine, clitest.Read(t, repo, mine)+"// Mine.\n")

	run := repo.Automap(t, "remove", "stub-physics", "--map", "map", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	if got := clitest.Read(t, repo, mine); !strings.HasSuffix(got, "// Mine.\n") {
		t.Errorf("remove rewrote or deleted %s", mine)
	}
	var warned bool
	for _, w := range run.Result.Warnings {
		warned = warned || w.Code == "GLUE_LEFT("+mine+")"
	}
	if !warned {
		t.Errorf("no GLUE_LEFT(%s) warning in %+v", mine, run.Result.Warnings)
	}
	if _, err := os.Stat(filepath.Join(repo.Dir, "glue/stub-physics/stub_physics_setup.h")); err == nil {
		t.Error("remove left glue/stub-physics/stub_physics_setup.h, which still matched the reference")
	}
}

func TestRemoveNeedsALibraryOfTheComposition(t *testing.T) {
	repo := fixtureRepo(t)
	repo.Automap(t, startTestbed...).ExpectExit(t, 0)

	run := repo.Automap(t, "remove", "stub-physics", "--map", "map", "--no-verify", "--json")

	run.ExpectExit(t, 2)
	run.ExpectFailure(t, "UNKNOWN_ENTRY(library stub-physics)", "isn't in the composition")
	run.ExpectUnchanged(t)
}

// A choice that named the removed library's slot goes back to the
// contract's default: with no profiler in the slot, diagnostics.profiler
// can't stay profiling-slot.
func TestRemoveResetsAChoiceThatNamedTheSlot(t *testing.T) {
	repo := testbedRepo(t)
	repo.Automap(t, "start", "--with", "sdl3", "--with", "dear-imgui:docking", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)
	repo.Automap(t, "extend", "tracy", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)
	var man struct {
		Slots   map[string]string
		Choices map[string]map[string]any
	}
	json.Unmarshal([]byte(clitest.Read(t, repo, "automap.manifest.json")), &man)
	if man.Choices["diagnostics"]["profiler"] != "profiling-slot" {
		t.Fatalf("diagnostics.profiler is %v after extend tracy, want profiling-slot", man.Choices["diagnostics"]["profiler"])
	}

	repo.Automap(t, "remove", "tracy", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)

	man.Slots, man.Choices = nil, nil
	json.Unmarshal([]byte(clitest.Read(t, repo, "automap.manifest.json")), &man)
	if man.Slots["profiling"] != "" || man.Choices["diagnostics"]["profiler"] != "none" {
		t.Errorf("after remove tracy the profiling slot holds %q and diagnostics.profiler is %v; want an empty slot and the default, none", man.Slots["profiling"], man.Choices["diagnostics"]["profiler"])
	}
}

// resolve re-resolves against the lock's own map release after the
// manifest changed: the release doesn't move, though a newer one exists.
func TestResolveKeepsTheMapRelease(t *testing.T) {
	repo := upgradeRepo(t, fixtureMapRepo(t), func(m *clitest.Repo) { newerStubUI(t, m, "stub-ui") })
	repo.EditJSON(t, "automap.manifest.json", set("variants/stub-window", "wide-events"))

	run := repo.Automap(t, "resolve", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	var lock struct {
		MapRelease string `json:"map_release"`
		Pins       map[string]struct{ Version, Variant string }
	}
	json.Unmarshal([]byte(clitest.Read(t, repo, "automap.lock.json")), &lock)
	if lock.MapRelease != "map-2026.09.01" || lock.Pins["stub-ui"].Version != "1.0.0" {
		t.Errorf("the lock is at %s with stub-ui %s; want map-2026.09.01 and 1.0.0: resolve doesn't move the release", lock.MapRelease, lock.Pins["stub-ui"].Version)
	}
	if lock.Pins["stub-window"].Variant != "wide-events" {
		t.Errorf("stub-window is pinned as %q, want the manifest's wide-events", lock.Pins["stub-window"].Variant)
	}
	if !strings.HasPrefix(run.Result.Summary, "Resolved against map-2026.09.01") {
		t.Errorf("summary %q, want it to begin: Resolved against map-2026.09.01", run.Result.Summary)
	}
}

// resolve carries a private map's changed reference glue, as upgrade does,
// without moving anything else.
func TestResolveCarriesAPrivateMapsChangedGlue(t *testing.T) {
	repo := acmeToolsRepo(t, nil)

	run := repo.Automap(t, "resolve", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	if got, want := clitest.Read(t, repo, acmeToolsFragment), clitest.Read(t, repo, acmeToolsReference); got != want {
		t.Errorf("%s is\n%s\nwant the private map's new reference\n%s", acmeToolsFragment, got, want)
	}
}

// A lock that no longer matches says which command re-resolves it.
func TestALockMismatchNamesResolve(t *testing.T) {
	repo := acmeToolsRepo(t, nil)

	run := repo.Automap(t, "extend", "stub-physics", "--no-verify", "--json")

	run.ExpectExit(t, 2)
	run.ExpectFailure(t, "LOCK_MISMATCH(private map acme)", "automap resolve")
}
