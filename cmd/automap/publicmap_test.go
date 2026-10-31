package main_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
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

// GLM fills the math slot; Jolt's and Dear ImGui's math-types bindings carry
// their adapters to it.
func TestExtendingWithGLMFillsTheMathSlot(t *testing.T) {
	repo := testbedRepo(t)
	repo.Automap(t, "start", "--with", "sdl3", "--with", "dear-imgui:docking", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)
	repo.Automap(t, "extend", "jolt", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)

	run := repo.Automap(t, "extend", "glm", "--map", "map", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	var man struct{ Slots map[string]string }
	json.Unmarshal([]byte(clitest.Read(t, repo, "automap.manifest.json")), &man)
	if man.Slots["math"] != "glm" {
		t.Errorf("slots %v, want GLM in the math slot", man.Slots)
	}
	for _, f := range []string{"glue/jolt/jolt_math_glm.h", "glue/dear-imgui/imgui_math_glm.h"} {
		if _, err := os.Stat(filepath.Join(repo.Dir, f)); err != nil {
			t.Errorf("an adapter to GLM wasn't written: %v", err)
		}
	}
}

// miniaudio fills the audio slot, with its setup glue and its allocation
// and diagnostics bindings.
func TestExtendingWithMiniaudioFillsTheAudioSlot(t *testing.T) {
	repo := testbedRepo(t)
	repo.Automap(t, "start", "--with", "sdl3", "--with", "dear-imgui:docking", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)

	run := repo.Automap(t, "extend", "miniaudio", "--map", "map", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	var man struct{ Slots map[string]string }
	json.Unmarshal([]byte(clitest.Read(t, repo, "automap.manifest.json")), &man)
	if man.Slots["audio"] != "miniaudio" {
		t.Errorf("slots %v, want miniaudio in the audio slot", man.Slots)
	}
	for _, f := range []string{"glue/miniaudio/miniaudio_setup.cpp", "glue/miniaudio/miniaudio_allocation.cpp", "glue/miniaudio/miniaudio_diagnostics.cpp"} {
		if _, err := os.Stat(filepath.Join(repo.Dir, f)); err != nil {
			t.Errorf("miniaudio's glue wasn't written: %v", err)
		}
	}
}

// fastgltf and stb_image bring the assets contract: its file-I/O interface
// and each library's glue that reads through it.
func TestExtendingWithAssetLibrariesBringsTheAssetsInterface(t *testing.T) {
	repo := testbedRepo(t)
	repo.Automap(t, "start", "--with", "sdl3", "--with", "dear-imgui:docking", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)
	repo.Automap(t, "extend", "fastgltf", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)

	run := repo.Automap(t, "extend", "stb-image", "--map", "map", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	var man struct{ Slots map[string]string }
	json.Unmarshal([]byte(clitest.Read(t, repo, "automap.manifest.json")), &man)
	if man.Slots["model-import"] != "fastgltf" || man.Slots["image-decode"] != "stb-image" {
		t.Errorf("slots %v, want fastgltf importing models and stb_image decoding images", man.Slots)
	}
	for _, f := range []string{"glue/automap/assets.h", "glue/automap/assets.cpp", "glue/fastgltf/fastgltf_assets.cpp", "glue/stb-image/stb_image_assets.cpp", "glue/stb-image/stb_image_allocation.cpp"} {
		if _, err := os.Stat(filepath.Join(repo.Dir, f)); err != nil {
			t.Errorf("the assets interface or its glue wasn't written: %v", err)
		}
	}
}

// flecs fills the ECS slot, with setup glue that owns the world and
// bindings for allocation, diagnostics and jobs through its OS API.
func TestExtendingWithFlecsFillsTheECSSlot(t *testing.T) {
	repo := testbedRepo(t)
	repo.Automap(t, "start", "--with", "sdl3", "--with", "dear-imgui:docking", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)

	run := repo.Automap(t, "extend", "flecs", "--map", "map", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	var man struct{ Slots map[string]string }
	json.Unmarshal([]byte(clitest.Read(t, repo, "automap.manifest.json")), &man)
	if man.Slots["ecs"] != "flecs" {
		t.Errorf("slots %v, want flecs in the ECS slot", man.Slots)
	}
	for _, f := range []string{"glue/flecs/flecs_setup.cpp", "glue/flecs/flecs_allocation.cpp", "glue/flecs/flecs_diagnostics.cpp", "glue/flecs/flecs_jobs.cpp"} {
		if _, err := os.Stat(filepath.Join(repo.Dir, f)); err != nil {
			t.Errorf("flecs's glue wasn't written: %v", err)
		}
	}
}

// Jolt's registry variants come from the Automap registry: the lock pins
// the registry at the public map's commit, vcpkg-configuration.json takes
// joltphysics from it, and vcpkg.json leaves the port version to it.
func TestJoltsRegistryVariantsResolveThroughTheAutomapRegistry(t *testing.T) {
	repo := testbedRepo(t)
	repo.Automap(t, "start", "--with", "sdl3", "--with", "dear-imgui:docking", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)

	run := repo.Automap(t, "extend", "jolt:deterministic", "--map", "map", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	var mapJSON struct {
		BuildBackend struct {
			Vcpkg struct {
				AutomapRegistry struct{ URL, Commit string } `json:"automap_registry"`
			}
		} `json:"build_backend"`
	}
	json.Unmarshal([]byte(clitest.Read(t, repo, "map/map.json")), &mapJSON)
	registry := mapJSON.BuildBackend.Vcpkg.AutomapRegistry
	var lock struct {
		Pins map[string]struct {
			Variant  string
			Registry struct{ Kind, Commit string }
		}
	}
	json.Unmarshal([]byte(clitest.Read(t, repo, "automap.lock.json")), &lock)
	if pin := lock.Pins["jolt"]; pin.Variant != "deterministic" || pin.Registry.Kind != "automap" || pin.Registry.Commit != registry.Commit {
		t.Errorf("jolt pinned as %+v, want deterministic from the Automap registry at %s", pin, registry.Commit)
	}
	var config struct {
		Registries []struct {
			Repository, Baseline string
			Packages             []string
		}
	}
	json.Unmarshal([]byte(clitest.Read(t, repo, "vcpkg-configuration.json")), &config)
	if len(config.Registries) != 1 || config.Registries[0].Repository != registry.URL || config.Registries[0].Baseline != registry.Commit ||
		!slices.Equal(config.Registries[0].Packages, []string{"joltphysics"}) {
		t.Errorf("vcpkg-configuration.json registries %+v, want joltphysics from %s at %s", config.Registries, registry.URL, registry.Commit)
	}
	if vcpkg := clitest.Read(t, repo, "vcpkg.json"); strings.Contains(vcpkg, `"name": "joltphysics",
      "version": "5.6.0",
      "port-version"`) {
		t.Errorf("vcpkg.json pins the curated port version for the registry's joltphysics:\n%s", vcpkg)
	}
}
