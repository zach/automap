package main_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zach/automap/internal/clitest"
)

// The public map's own builds run in Automap's CI, not with the CLI's
// tests (ADR-0077): the public-map job sets AUTOMAP_TEST_PUBLIC_MAP=1 on a
// runner with vcpkg at the map's baseline, the libraries' system packages,
// and Mesa lavapipe.
func requirePublicMapBuilds(t *testing.T) {
	t.Helper()
	if os.Getenv("AUTOMAP_TEST_PUBLIC_MAP") != "1" {
		t.Skip("builds the public map's libraries: set AUTOMAP_TEST_PUBLIC_MAP=1 (Automap's CI does)")
	}
	requireBuilds(t)
}

// testbedRepo is an empty repo where the testbed has been started from the
// public map, at map/.
func testbedRepo(t *testing.T) *clitest.Repo {
	t.Helper()
	repo := clitest.NewRepo(t, "no-manifest")
	repo.Copy(t, "map", "map")
	return repo
}

// logVcpkgErrors logs the tail of every port build's error log in vcpkg's
// buildtrees, which name the cause when vcpkg install fails.
func logVcpkgErrors(t *testing.T) {
	t.Helper()
	logs, _ := filepath.Glob(filepath.Join(os.Getenv("VCPKG_ROOT"), "buildtrees", "*", "*-err.log"))
	for _, l := range logs {
		raw, err := os.ReadFile(l)
		if err != nil || len(raw) == 0 {
			continue
		}
		lines := strings.Split(string(raw), "\n")
		if len(lines) > 30 {
			lines = lines[len(lines)-30:]
		}
		t.Logf("%s:\n%s", l, strings.Join(lines, "\n"))
	}
}

// The testbed's seeded violations, each with the failure code its probe
// must report (ADR-0071).
var testbedSeeds = []struct{ seed, code string }{
	{"allocation.hooks-before-first-allocation:sdl3", "ALLOC_HOOK_LATE(sdl3)"},
	{"allocation.hooks-before-first-allocation:dear-imgui", "ALLOC_HOOK_LATE(dear-imgui)"},
	{"allocation.routed-through-facade:sdl3", "ALLOC_NOT_ROUTED(sdl3)"},
	{"allocation.routed-through-facade:dear-imgui", "ALLOC_NOT_ROUTED(dear-imgui)"},
	{"diagnostics.routed-through-facade:sdl3", "DIAG_NOT_ROUTED(sdl3)"},
	{"lifecycle.init-order", "INIT_ORDER(dear-imgui:CreateContext before sdl3:SDL_SetMemoryFunctions)"},
	{"lifecycle.shutdown-reverse", "SHUTDOWN_ORDER(dear-imgui--sdl3:ImGui_ImplSDLGPU3_Shutdown before dear-imgui--sdl3:SDL_WaitForGPUIdle)"},
	{"build-flags.language-settings-agree", "LANG_SETTINGS_MISMATCH(cxx_standard)"},
	{"build-flags.abi-defines-match", "ABI_DEFINE_MISMATCH(dear-imgui)"},
	{"dear-imgui--sdl3.events-forwarded", "EVENTS_NOT_FORWARDED(dear-imgui--sdl3)"},
	{"platform.targets-supported", "PLATFORM_UNSUPPORTED(dear-imgui,linux-arm64-gcc)"},
}

func TestEveryPublicMapSeedIsTested(t *testing.T) {
	// The jobs contract's seed needs a library with jobs facts; the
	// fixture's stub-physics composition runs it (TestEveryJobsSeedIsCaught).
	exercisedByFixture := filepath.Join("contracts", "jobs")
	var patches int
	filepath.WalkDir(filepath.Join(clitest.Root(t), "map"), func(p string, d os.DirEntry, err error) error {
		if err == nil && strings.Contains(p, string(filepath.Separator)+"seeded"+string(filepath.Separator)) && !strings.Contains(p, exercisedByFixture) {
			patches++
		}
		return nil
	})
	if patches != len(testbedSeeds) {
		t.Errorf("the public map carries %d seed patches but the testbed test covers %d", patches, len(testbedSeeds))
	}
}

func TestTheTestbedStartsProbed(t *testing.T) {
	requirePublicMapBuilds(t)
	repo := testbedRepo(t)

	run := repo.Automap(t, "start", "--with", "sdl3", "--with", "dear-imgui:docking", "--map", "map", "--json")

	if run.Exit != 0 || run.Result.Level != "probed" {
		run.LogDetails(t)
		logVcpkgErrors(t)
		t.Fatalf("exit %d, level %q; want 0, probed\n%s", run.Exit, run.Result.Level, run.Stdout)
	}
	e := checkEvidence(t, run, "probed")
	for _, r := range e.Predicate.Results {
		if r.Rule == "diagnostics.routed-through-facade" && r.Library == "dear-imgui" &&
			(r.Status != "not-applicable" || !strings.Contains(r.Reason, "no runtime log or assert hook")) {
			t.Errorf("Dear ImGui's diagnostics result %+v, want not-applicable with the binding's reason", r)
		}
	}
	if _, err := os.Stat(filepath.Join(repo.Dir, "imgui.ini")); err == nil {
		t.Error("a verify run wrote imgui.ini into the repo")
	}

	for _, c := range testbedSeeds {
		t.Run(c.seed, func(t *testing.T) {
			t.Parallel()
			run := repo.Automap(t, "verify", "--seed", c.seed, "--map", "map", "--json")

			if run.Exit != 0 || !strings.Contains(run.Result.Summary, c.code) {
				run.LogDetails(t)
				t.Errorf("exit %d, summary %q; want the probe to report %s\n%s", run.Exit, run.Result.Summary, c.code, run.Stdout)
			}
		})
	}
}
