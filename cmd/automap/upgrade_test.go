package main_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zach/automap/internal/clitest"
)

const allocationGlue = "map/recipes/binding/stub-ui--allocation/stub_ui_allocation.cpp"

// releaseMap is a copy of the fixture map, as a map/ directory in a repo
// of its own, for a test to change before publishing it as a release.
func releaseMap(t *testing.T, from string) *clitest.Repo {
	t.Helper()
	repo := clitest.NewRepo(t, "no-manifest")
	if err := os.CopyFS(filepath.Join(repo.Dir, "map"), os.DirFS(from)); err != nil {
		t.Fatal(err)
	}
	return repo
}

// newerStubUI adds stub-ui 1.1.0 to the map, the newest version, and makes
// its recipes apply to it.
func newerStubUI(t *testing.T, m *clitest.Repo, lib string) {
	t.Helper()
	m.EditJSON(t, "map/libraries/"+lib+".json", func(doc any) any {
		versions := doc.(map[string]any)["versions"].([]any)
		raw, _ := json.Marshal(versions[0])
		var next map[string]any
		json.Unmarshal(raw, &next)
		next["version"], next["released"] = "1.1.0", "2026-10-20"
		doc.(map[string]any)["versions"] = append(versions, next)
		return doc
	})
	matches, _ := filepath.Glob(filepath.Join(m.Dir, "map", "recipes", "*", "*.json"))
	for _, f := range matches {
		rel, _ := filepath.Rel(m.Dir, f)
		m.EditJSON(t, filepath.ToSlash(rel), func(doc any) any {
			if applies, ok := doc.(map[string]any)["applies_to"].(map[string]any)[lib].(map[string]any); ok {
				applies["versions"] = append(applies["versions"].([]any), "1.1.0")
			}
			return doc
		})
	}
}

// upgradeRepo is a testbed started from release map-2026.09.01 of the
// fixture map, with release map-2026.10.01, made by change, published
// after it.
func upgradeRepo(t *testing.T, base string, change func(*clitest.Repo)) *clitest.Repo {
	t.Helper()
	s := newReleaseServer(t)
	old := releaseMap(t, base)
	s.publish(t, "map-2026.09.01", filepath.Join(old.Dir, "map"), nil)
	repo := releaseRepo(t, s, t.TempDir())
	repo.Automap(t, "start", "--with", "stub-window", "--with", "stub-ui", "--no-verify", "--json").ExpectExit(t, 0)
	next := releaseMap(t, base)
	change(next)
	s.publish(t, "map-2026.10.01", filepath.Join(next.Dir, "map"), nil)
	return repo
}

// fixtureMapRepo is the fixture map in a repo's map/ directory, the source
// of the releases.
func fixtureMapRepo(t *testing.T) string {
	t.Helper()
	repo := clitest.NewRepo(t, "no-manifest")
	repo.FixtureMap(t, "map")
	return filepath.Join(repo.Dir, "map")
}

// newAllocationGlue changes stub-ui's allocation binding's reference glue
// in the newer release.
func newAllocationGlue(m *clitest.Repo, t *testing.T) {
	m.WriteFile(t, allocationGlue, clitest.Read(t, m, allocationGlue)+"\n// Upgraded in map-2026.10.01.\n")
}

func TestUpgradeMovesToTheNewestReleaseAndUpdatesUnchangedGlue(t *testing.T) {
	repo := upgradeRepo(t, fixtureMapRepo(t), func(m *clitest.Repo) {
		newerStubUI(t, m, "stub-ui")
		newAllocationGlue(m, t)
	})

	run := repo.Automap(t, "upgrade", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	run.ExpectGolden(t, "upgrade-clean")
}

func TestUpgradeLeavesDivergedGlueAndWritesTheNewReferenceBesideIt(t *testing.T) {
	repo := upgradeRepo(t, fixtureMapRepo(t), func(m *clitest.Repo) { newAllocationGlue(m, t) })
	repo.WriteFile(t, "glue/stub-ui/stub_ui_allocation.cpp", clitest.Read(t, repo, "glue/stub-ui/stub_ui_allocation.cpp")+"\n// The developer's change.\n")

	run := repo.Automap(t, "upgrade", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	run.ExpectGolden(t, "upgrade-diverged")
}

func TestUpgradeALibraryMovesOnlyThatLibrary(t *testing.T) {
	repo := upgradeRepo(t, fixtureMapRepo(t), func(m *clitest.Repo) {
		newerStubUI(t, m, "stub-ui")
		newerStubUI(t, m, "stub-window")
	})

	repo.Automap(t, "upgrade", "stub-ui", "--no-verify", "--json").ExpectExit(t, 0)

	var lock struct {
		MapRelease string `json:"map_release"`
		Pins       map[string]struct{ Version string }
	}
	json.Unmarshal([]byte(clitest.Read(t, repo, "automap.lock.json")), &lock)
	if lock.MapRelease != "map-2026.10.01" || lock.Pins["stub-ui"].Version != "1.1.0" || lock.Pins["stub-window"].Version != "1.0.0" {
		t.Errorf("lock %+v, want map-2026.10.01 with stub-ui 1.1.0 and stub-window still 1.0.0", lock)
	}
	// Later commands read the map held to the lock's versions.
	run := repo.Automap(t, "show", "library", "stub-window", "--json")
	if !strings.Contains(string(run.Stdout), `"version": "1.0.0"`) || strings.Contains(string(run.Stdout), `"version": "1.1.0"`) {
		t.Errorf("show reads stub-window at a version the lock doesn't pin:\n%s", run.Stdout)
	}
}

func TestUpgradeToAMapReleaseByName(t *testing.T) {
	repo := upgradeRepo(t, fixtureMapRepo(t), func(m *clitest.Repo) {})

	run := repo.Automap(t, "upgrade", "--map-release", "map-2026.09.01", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	if got := lockRelease(t, repo); got != "map-2026.09.01" {
		t.Errorf("lock records %s, want the release named", got)
	}
}

func TestUpgradeEndsWithVerifyAndNamesFailingRules(t *testing.T) {
	base := buildableRepo(t)
	repo := upgradeRepo(t, filepath.Join(base.Dir, "map"), func(m *clitest.Repo) {
		// The newer release's reference glue no longer installs stub-ui's
		// allocator hooks.
		glue := clitest.Read(t, m, allocationGlue)
		m.WriteFile(t, allocationGlue, strings.Replace(glue, "stub_ui_set_allocator(", "(void)(", 1))
	})

	run := repo.Automap(t, "upgrade", "--json")

	run.ExpectExit(t, 1)
	run.ExpectFailure(t, "ALLOC_NOT_ROUTED(stub-ui)", "")
	if hostPlatform == "linux-x64-clang" { // the result names the platform
		// The lock names the registry the test builds, whose commit
		// differs between machines.
		run.ExpectGoldenResult(t, "upgrade-failing-probe")
	}
	if !strings.Contains(clitest.Read(t, repo, "glue/stub-ui/stub_ui_allocation.cpp"), "(void)(&allocator)") {
		t.Error("upgrade didn't apply the new release's reference glue")
	}
}
