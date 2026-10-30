package main_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/zach/automap/internal/clitest"
)

var startWindow = []string{"start", "--with", "stub-window", "--map", "map", "--no-verify", "--json"}

// windowRepo is a fixture repo where start has written a composition with
// only stub-window.
func windowRepo(t *testing.T) *clitest.Repo {
	t.Helper()
	repo := fixtureRepo(t)
	repo.Automap(t, startWindow...).ExpectExit(t, 0)
	return repo
}

func TestExtendAddsALibraryAndThePairingItCompletes(t *testing.T) {
	run := windowRepo(t).Automap(t, "extend", "stub-ui", "--map", "map", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	for _, f := range []string{
		"cmake/automap/stub-ui.cmake", "glue/stub-ui/stub_ui_setup.cpp", "glue/stub-ui/stub_ui_allocation.cpp",
		"glue/stub-ui--stub-window/stub_ui_stub_window.cpp", "glue/automap/startup.cpp", "automap.lock.json", "AGENTS.md",
	} {
		if !slices.Contains(run.Result.FilesWritten, f) {
			t.Errorf("files_written lacks %s", f)
		}
	}
	if slices.Contains(run.Result.FilesWritten, "glue/stub-window/stub_window_setup.cpp") {
		t.Error("extend rewrote stub-window's setup glue, which was already there")
	}
	run.ExpectGolden(t, "extend-stub-ui")
}

func TestExtendListsTheNewSlotsInAGENTSMD(t *testing.T) {
	run := windowRepo(t).Automap(t, "extend", "stub-ui", "--map", "map", "--no-verify", "--json")

	if section := run.ReadFile(t, "AGENTS.md"); !strings.Contains(section, "tools-ui = stub-ui") {
		t.Errorf("AGENTS.md doesn't list tools-ui = stub-ui:\n%s", section)
	}
}

func TestExtendLeavesTheDevelopersGlueAlone(t *testing.T) {
	repo := windowRepo(t)
	setup := "glue/stub-window/stub_window_setup.cpp"
	adapted := clitest.Read(t, repo, setup) + "// adapted by the developer\n"
	repo.WriteFile(t, setup, adapted)

	run := repo.Automap(t, "extend", "stub-ui", "--map", "map", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	if got := run.ReadFile(t, setup); got != adapted {
		t.Error("extend overwrote the developer's adapted glue")
	}
}

func TestExtendWarnsAboutACopyleftLibrary(t *testing.T) {
	repo := windowRepo(t)
	repo.EditJSON(t, "map/libraries/stub-ui.json", set("license", map[string]any{
		"spdx": "GPL-3.0-only", "class": "copyleft", "warning": "Linking stub-ui makes the game GPL-3.0.",
	}))

	run := repo.Automap(t, "extend", "stub-ui", "--map", "map", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	found := false
	for _, w := range run.Result.Warnings {
		found = found || (w.Code == "LICENSE_WARNING(stub-ui)" && w.Message == "Linking stub-ui makes the game GPL-3.0.")
	}
	if !found {
		t.Errorf("warnings %+v lack LICENSE_WARNING(stub-ui) with the map's reason", run.Result.Warnings)
	}
}

func TestExtendRejectsATakenSlot(t *testing.T) {
	repo := windowRepo(t)
	repo.EditJSON(t, "map/libraries/stub-ui.json", push("slots", "windowing"))

	run := repo.Automap(t, "extend", "stub-ui", "--map", "map", "--no-verify", "--json")

	run.ExpectExit(t, 2)
	run.ExpectFailure(t, "SLOT_TAKEN(windowing)", "stub-window already fills windowing")
	run.ExpectUnchanged(t)
}

func TestExtendWithAnExplicitSlot(t *testing.T) {
	repo := windowRepo(t)
	repo.EditJSON(t, "map/libraries/stub-ui.json", push("slots", "windowing"))

	run := repo.Automap(t, "extend", "stub-ui", "--slot", "tools-ui", "--map", "map", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	if !strings.Contains(run.Result.Summary, "Added stub-ui 1.0.0 (default) to tools-ui") {
		t.Errorf("summary %q", run.Result.Summary)
	}
}

func TestExtendEndsByVerifying(t *testing.T) {
	repo := buildableRepo(t)
	repo.Automap(t, startWindow...).ExpectExit(t, 0)

	run := repo.Automap(t, "extend", "stub-ui", "--map", "map", "--json")

	run.ExpectExit(t, 0)
	if run.Result.Level != "probed" {
		run.LogDetails(t)
		t.Fatalf("level %q, want probed\n%s", run.Result.Level, run.Stdout)
	}
	checkEvidence(t, run, "probed")
}

func TestExtendingResolvesTheSameLockAsStartingWithBoth(t *testing.T) {
	extended := windowRepo(t).Automap(t, "extend", "stub-ui", "--map", "map", "--no-verify", "--json")
	started := fixtureRepo(t).Automap(t, startTestbed...)

	if a, b := extended.ReadFile(t, "automap.lock.json"), started.ReadFile(t, "automap.lock.json"); a != b {
		t.Errorf("locks differ:\nextended %s\nstarted  %s", a, b)
	}
	if a, b := extended.ReadFile(t, "glue/automap/startup.cpp"), started.ReadFile(t, "glue/automap/startup.cpp"); a != b {
		t.Errorf("startups differ:\nextended\n%s\nstarted\n%s", a, b)
	}
}
