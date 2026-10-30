package main_test

import (
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
