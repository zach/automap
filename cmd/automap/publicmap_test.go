package main_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zach/automap/internal/clitest"
)

// The slice-1 contracts and their v1 probed rules (SPEC §8.2).
var slice1Contracts = map[string][]string{
	"allocation":  {"allocation.hooks-before-first-allocation", "allocation.routed-through-facade"},
	"diagnostics": {"diagnostics.routed-through-facade"},
	"lifecycle":   {"lifecycle.init-order", "lifecycle.shutdown-reverse"},
	"build-flags": {"build-flags.abi-defines-match", "build-flags.language-settings-agree"},
	"platform":    {"platform.targets-supported"},
}

func TestThePublicMapsContractsHaveTheirProbedRules(t *testing.T) {
	repo := fixtureRepo(t)
	for contract, rules := range slice1Contracts {
		t.Run(contract, func(t *testing.T) {
			run := repo.Automap(t, "show", "contract", contract, "--map", "map", "--json")

			run.ExpectExit(t, 0)
			if v := run.Result.Entries[0].Get("definition", "version"); v != "1.0.0" {
				t.Errorf("version %v, want 1.0.0", v)
			}
			probed := map[string]bool{}
			for _, r := range run.Result.Entries[0].Get("definition", "rules").([]any) {
				rule := r.(map[string]any)
				if rule["kind"] == "probed" {
					probed[rule["id"].(string)] = true
				}
			}
			for _, id := range rules {
				if !probed[id] {
					t.Errorf("%s lacks probed rule %s", contract, id)
				}
			}
		})
	}
}

func TestEverySourceThePublicMapShipsIsMIT0(t *testing.T) {
	root := filepath.Join(clitest.Root(t), "map")
	filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if ext := filepath.Ext(p); ext == ".h" || ext == ".cpp" || ext == ".cmake" {
			raw, _ := os.ReadFile(p)
			if !strings.Contains(string(raw), "SPDX-License-Identifier: MIT-0") {
				rel, _ := filepath.Rel(root, p)
				t.Errorf("map/%s has no MIT-0 license identifier (ADR-0055)", filepath.ToSlash(rel))
			}
		}
		return nil
	})
}

// Filling the jobs slot switches the jobs interface to the slot's
// library: extend sets jobs.scheduler to jobs-slot, and the composition
// gets the library's implementation of the interface (SPEC §2.5).
func TestExtendingWithASchedulerFillsTheJobsSlot(t *testing.T) {
	repo := testbedRepo(t)
	repo.Automap(t, "start", "--with", "sdl3", "--with", "dear-imgui:docking", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)
	repo.Automap(t, "extend", "jolt", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)

	run := repo.Automap(t, "extend", "enkits", "--map", "map", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	var man struct {
		Slots   map[string]string
		Choices map[string]map[string]any
	}
	json.Unmarshal([]byte(clitest.Read(t, repo, "automap.manifest.json")), &man)
	if man.Slots["jobs"] != "enkits" || man.Choices["jobs"]["scheduler"] != "jobs-slot" {
		t.Errorf("slots %v, jobs choices %v; want enkits in the jobs slot and scheduler jobs-slot", man.Slots, man.Choices["jobs"])
	}
	if !strings.Contains(clitest.Read(t, repo, "glue/automap/choices.h"), "AUTOMAP_JOBS_SCHEDULER_JOBS_SLOT") {
		t.Error("the choices header doesn't select the jobs slot")
	}
	if _, err := os.Stat(filepath.Join(repo.Dir, "glue/enkits/enkits_jobs.cpp")); err != nil {
		t.Errorf("enkiTS's implementation of the jobs interface wasn't written: %v", err)
	}
}

// Filling the allocator slot backs the allocation facade with the slot's
// library: extend sets allocation.allocator to allocator-slot, and the
// composition gets the library's implementation of the facade's backend.
func TestExtendingWithAnAllocatorFillsTheAllocatorSlot(t *testing.T) {
	repo := testbedRepo(t)
	repo.Automap(t, "start", "--with", "sdl3", "--with", "dear-imgui:docking", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)

	run := repo.Automap(t, "extend", "mimalloc", "--map", "map", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	var man struct {
		Slots   map[string]string
		Choices map[string]map[string]any
	}
	json.Unmarshal([]byte(clitest.Read(t, repo, "automap.manifest.json")), &man)
	if man.Slots["allocator"] != "mimalloc" || man.Choices["allocation"]["allocator"] != "allocator-slot" {
		t.Errorf("slots %v, allocation choices %v; want mimalloc in the allocator slot and allocator allocator-slot", man.Slots, man.Choices["allocation"])
	}
	if !strings.Contains(clitest.Read(t, repo, "glue/automap/choices.h"), "AUTOMAP_ALLOCATION_ALLOCATOR_ALLOCATOR_SLOT") {
		t.Error("the choices header doesn't select the allocator slot")
	}
	if _, err := os.Stat(filepath.Join(repo.Dir, "glue/mimalloc/mimalloc_allocation.cpp")); err != nil {
		t.Errorf("mimalloc's backend for the allocation facade wasn't written: %v", err)
	}
}

// Filling the logging and profiling slots backs the diagnostics facade with
// their libraries: extend sets diagnostics.log and diagnostics.profiler, and
// the composition gets each library's backend for the facade.
func TestExtendingWithDiagnosticsBackendsFillsTheirSlots(t *testing.T) {
	repo := testbedRepo(t)
	repo.Automap(t, "start", "--with", "sdl3", "--with", "dear-imgui:docking", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)
	repo.Automap(t, "extend", "spdlog", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)

	run := repo.Automap(t, "extend", "tracy", "--map", "map", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	var man struct {
		Slots   map[string]string
		Choices map[string]map[string]any
	}
	json.Unmarshal([]byte(clitest.Read(t, repo, "automap.manifest.json")), &man)
	if man.Slots["logging"] != "spdlog" || man.Slots["profiling"] != "tracy" ||
		man.Choices["diagnostics"]["log"] != "logging-slot" || man.Choices["diagnostics"]["profiler"] != "profiling-slot" {
		t.Errorf("slots %v, diagnostics choices %v; want spdlog logging and Tracy profiling", man.Slots, man.Choices["diagnostics"])
	}
	choices := clitest.Read(t, repo, "glue/automap/choices.h")
	if !strings.Contains(choices, "AUTOMAP_DIAGNOSTICS_LOG_LOGGING_SLOT") || !strings.Contains(choices, "AUTOMAP_DIAGNOSTICS_PROFILER_PROFILING_SLOT") {
		t.Error("the choices header doesn't select the logging and profiling slots")
	}
	// The tracy port enables TRACY_ENABLE from its port version 1: the
	// override pins the port version along with the version.
	if !strings.Contains(clitest.Read(t, repo, "vcpkg.json"), `"name": "tracy",
      "version": "0.14.1",
      "port-version": 1`) {
		t.Errorf("vcpkg.json doesn't pin tracy 0.14.1 at port version 1:\n%s", clitest.Read(t, repo, "vcpkg.json"))
	}
	for _, f := range []string{"glue/spdlog/spdlog_diagnostics.cpp", "glue/tracy/tracy_diagnostics.cpp"} {
		if _, err := os.Stat(filepath.Join(repo.Dir, f)); err != nil {
			t.Errorf("a diagnostics backend wasn't written: %v", err)
		}
	}
}
