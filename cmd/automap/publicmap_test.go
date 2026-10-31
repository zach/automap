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
