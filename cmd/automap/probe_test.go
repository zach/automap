package main_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The fixture map's seeded violations, each with the failure code its
// probe must report (ADR-0071).
var fixtureSeeds = []struct{ seed, code string }{
	{"allocation.hooks-before-first-allocation:stub-window", "ALLOC_HOOK_LATE(stub-window)"},
	{"allocation.hooks-before-first-allocation:stub-ui", "ALLOC_HOOK_LATE(stub-ui)"},
	{"allocation.routed-through-facade:stub-window", "ALLOC_NOT_ROUTED(stub-window)"},
	{"allocation.routed-through-facade:stub-ui", "ALLOC_NOT_ROUTED(stub-ui)"},
	{"diagnostics.routed-through-facade:stub-window", "DIAG_NOT_ROUTED(stub-window)"},
	{"diagnostics.routed-through-facade:stub-ui", "DIAG_NOT_ROUTED(stub-ui)"},
	{"build-flags.language-settings-agree", "LANG_SETTINGS_MISMATCH(cxx_standard)"},
	{"build-flags.abi-defines-match", "ABI_DEFINE_MISMATCH(STUB_UI_WIDE_IDS)"},
	{"stub-ui--stub-window.input-forwarded", "INPUT_NOT_FORWARDED(stub-ui--stub-window)"},
	{"lifecycle.init-order", "INIT_ORDER(stub-ui--stub-window:route_input before stub-ui--stub-window:attach)"},
	{"lifecycle.shutdown-reverse", "SHUTDOWN_ORDER(stub-ui--stub-window:detach before stub-ui--stub-window:unroute_input)"},
	{"platform.targets-supported", "PLATFORM_UNSUPPORTED(stub-ui,linux-arm64-gcc)"},
}

func TestEverySeededViolationIsCaught(t *testing.T) {
	repo := buildableRepo(t)
	repo.Automap(t, startTestbed...).ExpectExit(t, 0)
	for _, c := range fixtureSeeds {
		t.Run(c.seed, func(t *testing.T) {
			t.Parallel()
			run := repo.Automap(t, "verify", "--seed", c.seed, "--map", "map", "--json")

			run.ExpectExit(t, 0)
			if !strings.Contains(run.Result.Summary, c.code) {
				t.Errorf("summary %q, want the probe to report %s\n%s", run.Result.Summary, c.code, run.Stdout)
			}
			if run.Result.EvidenceFile != "" {
				t.Errorf("a seed run wrote evidence %s", run.Result.EvidenceFile)
			}
		})
	}
	if _, err := os.Stat(filepath.Join(repo.Dir, "evidence")); err == nil {
		t.Error("seed runs wrote an evidence directory")
	}
}

func TestEverySeedInTheFixtureMapIsTested(t *testing.T) {
	repo := startedRepo(t)
	run := repo.Automap(t, "verify", "--seed", "no.such-rule", "--map", "map", "--json")
	run.ExpectExit(t, 2)
	// The fixture map's seeds, found the way verify finds them: by the
	// patch files its recipes and contracts carry.
	var patches int
	filepath.WalkDir(filepath.Join(repo.Dir, "map"), func(p string, d os.DirEntry, err error) error {
		if err == nil && strings.Contains(p, string(filepath.Separator)+"seeded"+string(filepath.Separator)) {
			patches++
		}
		return nil
	})
	if patches != len(fixtureSeeds) {
		t.Errorf("the fixture map carries %d seed patches but the test covers %d", patches, len(fixtureSeeds))
	}
}

func TestAmbiguousSeedNeedsALibrary(t *testing.T) {
	run := startedRepo(t).Automap(t, "verify", "--seed", "allocation.routed-through-facade", "--map", "map", "--json")

	run.ExpectExit(t, 2)
	run.ExpectFailure(t, "USAGE", "allocation.routed-through-facade:stub-ui")
}

func TestALifecycleViolationFailsVerify(t *testing.T) {
	repo := buildableRepo(t)
	repo.Automap(t, startTestbed...).ExpectExit(t, 0)
	glue := "glue/stub-ui--stub-window/stub_ui_stub_window.cpp"
	src := strings.Replace(strings.Replace(readRepo(t, repo.Dir, glue),
		"    stub_ui_disable_input();\n    am::lifecycle::shutdown_step(\"stub-ui--stub-window:unroute_input\");\n", "", 1),
		"    am::lifecycle::shutdown_step(\"stub-ui--stub-window:detach\");\n",
		"    am::lifecycle::shutdown_step(\"stub-ui--stub-window:detach\");\n    stub_ui_disable_input();\n    am::lifecycle::shutdown_step(\"stub-ui--stub-window:unroute_input\");\n", 1)
	repo.WriteFile(t, glue, src)

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	run.ExpectExit(t, 1)
	run.ExpectFailure(t, "SHUTDOWN_ORDER(stub-ui--stub-window:detach before stub-ui--stub-window:unroute_input)", "pairing requires")
	if run.Result.Level != "smoke" {
		t.Errorf("level %q, want smoke", run.Result.Level)
	}
}

func TestAPlatformLimitedRuleIsNotApplicableElsewhere(t *testing.T) {
	repo := buildableRepo(t)
	repo.EditJSON(t, "map/recipes/pairing/stub-ui--stub-window.json", set("rules/0/platforms", []any{"macos-arm64-appleclang"}))
	repo.Automap(t, startTestbed...).ExpectExit(t, 0)

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	run.ExpectExit(t, 0)
	if run.Result.Level != "probed" {
		t.Errorf("level %q, want probed", run.Result.Level)
	}
	e := checkEvidence(t, run, "probed")
	for _, r := range e.Predicate.Results {
		if r.Rule == "stub-ui--stub-window.input-forwarded" && (r.Status != "not-applicable" || !strings.Contains(r.Reason, "macos-arm64-appleclang")) {
			t.Errorf("result %+v, want not-applicable on linux, naming the platform it runs on", r)
		}
	}
}

func TestAnUnverifiableRuleBlocksProbed(t *testing.T) {
	repo := buildableRepo(t)
	repo.Automap(t, startTestbed...).ExpectExit(t, 0)
	probe := "glue/stub-ui/stub_ui_diagnostics_probe.cpp"
	repo.WriteFile(t, probe, strings.Replace(readRepo(t, repo.Dir, probe), "const am::probe::Registration registration", "[[maybe_unused]] const am::probe::Registration* unregistered = nullptr; //", 1))

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	run.ExpectExit(t, 0)
	if run.Result.Level != "smoke" {
		t.Errorf("level %q, want smoke: an unverifiable rule blocks probed", run.Result.Level)
	}
	found := false
	for _, w := range run.Result.Warnings {
		found = found || w.Code == "UNVERIFIABLE(diagnostics.routed-through-facade,stub-ui)"
	}
	if !found {
		t.Errorf("warnings %+v lack UNVERIFIABLE(diagnostics.routed-through-facade,stub-ui)", run.Result.Warnings)
	}
}

func readRepo(t *testing.T, dir, path string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(path)))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
