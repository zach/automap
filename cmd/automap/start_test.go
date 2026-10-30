package main_test

import (
	"encoding/json"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/zach/automap/internal/composition"
)

var startTestbed = []string{"start", "--with", "stub-window", "--with", "stub-ui", "--map", "map", "--no-verify", "--json"}

func TestStartWritesATestbedComposition(t *testing.T) {
	run := fixtureRepo(t).Automap(t, startTestbed...)

	run.ExpectExit(t, 0)
	for _, f := range []string{
		"automap.manifest.json", "automap.lock.json", "vcpkg.json", "vcpkg-configuration.json",
		"CMakeLists.txt", "main.cpp",
		"cmake/automap/automap.cmake", "cmake/automap/stub-window.cmake", "cmake/automap/stub-ui.cmake",
		"glue/automap/alloc.h", "glue/automap/diag.h", "glue/automap/probe.h", "glue/automap/lifecycle.h",
		"glue/automap/choices.h", "glue/automap/context.h", "glue/automap/startup.h", "glue/automap/startup.cpp",
		"glue/stub-window/stub_window_allocation.cpp", "glue/stub-window/stub_window_setup.cpp",
		"glue/stub-ui/stub_ui_setup.cpp", "glue/stub-ui--stub-window/stub_ui_stub_window.cpp",
	} {
		if !slices.Contains(run.Result.FilesWritten, f) {
			t.Errorf("files_written lacks %s", f)
		}
	}
	run.ExpectGolden(t, "start-testbed")
}

func TestStartRecordsTheTestbedExecutableAndAnUnreleasedMap(t *testing.T) {
	run := fixtureRepo(t).Automap(t, startTestbed...)

	var man composition.Manifest
	if err := json.Unmarshal([]byte(run.ReadFile(t, "automap.manifest.json")), &man); err != nil {
		t.Fatal(err)
	}
	if man.Executable.Target != "testbed" || !slices.Equal(man.Executable.Sources, []string{"main.cpp"}) {
		t.Errorf("executable %+v, want target testbed with sources main.cpp", man.Executable)
	}
	if man.MapRelease != "unreleased" {
		t.Errorf("map_release %q, want unreleased for a map that isn't a git checkout", man.MapRelease)
	}
	if !slices.Equal(man.Platforms, composition.V1Platforms) {
		t.Errorf("platforms %v, want the three v1 platforms", man.Platforms)
	}
}

func TestStartWritesACanonicalLockWithACorrectHash(t *testing.T) {
	run := fixtureRepo(t).Automap(t, startTestbed...)

	if _, err := composition.CheckLock([]byte(run.ReadFile(t, "automap.lock.json"))); err != nil {
		t.Error(err)
	}
}

func TestStartFromACleanGitCheckoutRecordsItsCommit(t *testing.T) {
	repo := fixtureRepo(t)
	git := func(args ...string) string {
		cmd := exec.Command("git", append([]string{"-C", repo.Dir + "/map", "-c", "user.name=t", "-c", "user.email=t@example.invalid"}, args...)...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "-q")
	git("add", ".")
	git("commit", "-q", "-m", "fixture map")
	commit := git("rev-parse", "HEAD")

	run := repo.Automap(t, startTestbed...)

	run.ExpectExit(t, 0)
	if want := `"map_release": "git:` + commit + `"`; !strings.Contains(run.ReadFile(t, "automap.manifest.json"), want) {
		t.Errorf("manifest lacks %s", want)
	}
	if want := `"map_release":"git:` + commit + `"`; !strings.Contains(run.ReadFile(t, "automap.lock.json"), want) {
		t.Errorf("lock lacks %s", want)
	}
}

func TestStartWritesEveryChoiceWithContractDefaults(t *testing.T) {
	run := fixtureRepo(t).Automap(t, startTestbed...)

	header := run.ReadFile(t, "glue/automap/choices.h")
	for _, want := range []string{
		"#define AUTOMAP_ALLOCATION_ALLOCATOR_SYSTEM 1\n",
		"#define AUTOMAP_DIAGNOSTICS_LOG_STDERR 1\n",
		"#define AUTOMAP_DIAGNOSTICS_PROFILER_NONE 1\n",
		"#define AUTOMAP_BUILD_FLAGS_CXX_STANDARD 20\n",
	} {
		if !strings.Contains(header, want) {
			t.Errorf("choices.h lacks %q:\n%s", want, header)
		}
	}
}

func TestGlueReadsChoicesOnlyThroughTheChoicesHeader(t *testing.T) {
	run := fixtureRepo(t).Automap(t, startTestbed...)

	macro := regexp.MustCompile(`AUTOMAP_[A-Z_]+`)
	for _, f := range run.Result.FilesWritten {
		if !strings.HasPrefix(f, "glue/") || f == "glue/automap/choices.h" {
			continue
		}
		for _, name := range macro.FindAllString(run.ReadFile(t, f), -1) {
			if name != "AUTOMAP_PROBE" && !strings.Contains(run.ReadFile(t, f), `#include "automap/choices.h"`) {
				t.Errorf("%s reads %s without including the choices header", f, name)
			}
		}
	}
}

func TestBuildFragmentsReferToLibrariesByPackageAndImportedTarget(t *testing.T) {
	run := fixtureRepo(t).Automap(t, startTestbed...)

	for lib, want := range map[string][]string{
		"stub-window": {"find_package(StubWindow CONFIG REQUIRED)", "StubWindow::StubWindow"},
		"stub-ui":     {"find_package(StubUI CONFIG REQUIRED)", "StubUI::StubUI"},
	} {
		fragment := run.ReadFile(t, "cmake/automap/"+lib+".cmake")
		for _, w := range want {
			if !strings.Contains(fragment, w) {
				t.Errorf("%s's build fragment lacks %s", lib, w)
			}
		}
	}
	for _, f := range run.Result.FilesWritten {
		if strings.HasSuffix(f, ".cmake") || f == "CMakeLists.txt" {
			if body := run.ReadFile(t, f); strings.Contains(body, "vcpkg_installed") || strings.Contains(body, "/usr/") {
				t.Errorf("%s refers to an installed library by path", f)
			}
		}
	}
}

func TestStartAssignsOnlyTheSlotsNamed(t *testing.T) {
	run := fixtureRepo(t).Automap(t, "start", "--with", "stub-window", "--slot", "windowing=stub-window", "--map", "map", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	if !strings.Contains(run.ReadFile(t, "automap.manifest.json"), "\"slots\": {\n    \"windowing\": \"stub-window\"\n  }") {
		t.Errorf("manifest slots aren't just windowing:\n%s", run.ReadFile(t, "automap.manifest.json"))
	}
}

func TestStartRejectsATakenSlot(t *testing.T) {
	run := fixtureRepo(t).Automap(t, "start", "--slot", "windowing=stub-window", "--slot", "windowing=stub-window", "--map", "map", "--json")

	run.ExpectExit(t, 2)
	run.ExpectFailure(t, "SLOT_TAKEN(windowing)", "stub-window already fills windowing")
	run.ExpectUnchanged(t)
}

func TestStartRefusesWhereACompositionExists(t *testing.T) {
	repo := fixtureRepo(t)
	repo.WriteFile(t, "automap.manifest.json", "{}\n")
	run := repo.Automap(t, startTestbed...)

	run.ExpectExit(t, 3)
	run.ExpectFailure(t, "COMPOSITION_EXISTS", "automap.manifest.json")
	run.ExpectUnchanged(t)
}

func TestStartRefusesARepoWithACMakeProject(t *testing.T) {
	repo := fixtureRepo(t)
	repo.WriteFile(t, "CMakeLists.txt", "project(game)\n")
	run := repo.Automap(t, startTestbed...)

	run.ExpectExit(t, 3)
	run.ExpectFailure(t, "CMAKE_PROJECT_EXISTS", "automap adopt")
	run.ExpectUnchanged(t)
}

func TestStartReportsAContextFieldNothingProvides(t *testing.T) {
	repo := fixtureRepo(t)
	repo.EditJSON(t, "map/recipes/pairing/stub-ui--stub-window.json", set("entry_points/0/reads", []any{"stub-window.surface"}))
	run := repo.Automap(t, startTestbed...)

	run.ExpectExit(t, 1)
	run.ExpectFailure(t, "CONTEXT_UNPROVIDED(stub-window.surface)", "am_pair_stub_ui_stub_window_init")
	run.ExpectUnchanged(t)
}

func TestEveryCommandRefusesAHandEditedLock(t *testing.T) {
	repo := fixtureRepo(t)
	repo.Automap(t, startTestbed...).ExpectExit(t, 0)
	repo.EditJSON(t, "automap.lock.json", set("pins/stub-ui/version", "1.0.1"))

	for _, args := range [][]string{
		{"verify", "--json"},
		{"find", "--map", "map", "--json"},
		{"show", "slot", "windowing", "--map", "map", "--json"},
		startTestbed,
	} {
		t.Run(args[0], func(t *testing.T) {
			run := repo.Automap(t, args...)

			run.ExpectExit(t, 3)
			run.ExpectFailure(t, "LOCK_EDITED", "automap.lock.json")
		})
	}
}
