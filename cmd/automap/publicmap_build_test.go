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
	var patches int
	filepath.WalkDir(filepath.Join(clitest.Root(t), "map"), func(p string, d os.DirEntry, err error) error {
		if err == nil && strings.Contains(p, string(filepath.Separator)+"seeded"+string(filepath.Separator)) {
			patches++
		}
		return nil
	})
	if covered := len(testbedSeeds) + len(joltSeeds) + len(enkitsSeeds) + len(mimallocSeeds); patches != covered {
		t.Errorf("the public map carries %d seed patches but the testbed tests cover %d", patches, covered)
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

// The seeds Jolt adds when it extends the testbed, with their failure
// codes. A composition rule's seed is named by the recipe's library.
var joltSeeds = []struct{ seed, code string }{
	{"allocation.hooks-before-first-allocation:jolt", "ALLOC_HOOK_LATE(jolt)"},
	{"allocation.routed-through-facade:jolt", "ALLOC_NOT_ROUTED(jolt)"},
	{"diagnostics.routed-through-facade:jolt", "DIAG_NOT_ROUTED(jolt)"},
	{"jobs.routed-through-interface:jolt", "JOBS_NOT_ROUTED(jolt)"},
	{"jobs.tsan-clean", "TSAN_REPORT(data race jobs.cpp"},
	{"lifecycle.init-order:jolt", "before jolt:Trace+AssertFailed)"},
	{"build-flags.abi-defines-match:jolt", "ABI_DEFINE_MISMATCH(jolt)"},
	{"math-types.precision-agrees", "PRECISION_MISMATCH(jolt)"},
}

func TestTheTestbedExtendsWithJoltProbed(t *testing.T) {
	requirePublicMapBuilds(t)
	tsan := hostPlatform != "windows-x64-msvc" // the jobs contract's ThreadSanitizer rule doesn't run on Windows
	if tsan {
		requireTSan(t)
	}
	repo := testbedRepo(t)
	if run := repo.Automap(t, "start", "--with", "sdl3", "--with", "dear-imgui:docking", "--map", "map", "--no-verify", "--json"); run.Exit != 0 {
		t.Fatalf("start: exit %d\n%s", run.Exit, run.Stdout)
	}

	run := repo.Automap(t, "extend", "jolt", "--map", "map", "--json")

	if run.Exit != 0 || run.Result.Level != "probed" {
		run.LogDetails(t)
		logVcpkgErrors(t)
		t.Fatalf("exit %d, level %q; want 0, probed\n%s", run.Exit, run.Result.Level, run.Stdout)
	}
	e := checkEvidence(t, run, "probed")
	want := map[string]bool{"build-flags.abi-defines-match": false, "jobs.routed-through-interface": false}
	if tsan {
		want["jobs.tsan-clean"] = false
	}
	for _, r := range e.Predicate.Results {
		if r.Rule == "jobs.tsan-clean" && !tsan && r.Status != "not-applicable" {
			t.Errorf("jobs.tsan-clean %+v on %s, want not-applicable", r, hostPlatform)
		}
		if _, ok := want[r.Rule]; ok && r.Status == "pass" {
			want[r.Rule] = true
		}
		// verify builds Debug, where Jolt's assertions are compiled in, so
		// its assert hook is checked and the result carries no condition.
		if r.Rule == "diagnostics.routed-through-facade" && r.Library == "jolt" && (r.Status != "pass" || r.Reason != "") {
			t.Errorf("Jolt's diagnostics result %+v, want a pass that checked the assert hook", r)
		}
	}
	for rule, passed := range want {
		if !passed {
			t.Errorf("%s didn't pass in the Jolt composition", rule)
		}
	}

	for _, c := range joltSeeds {
		if c.seed == "jobs.tsan-clean" && !tsan {
			continue
		}
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

// The seeds enkiTS adds when it fills the jobs slot of testbed + Jolt.
var enkitsSeeds = []struct{ seed, code string }{
	{"allocation.hooks-before-first-allocation:enkits", "ALLOC_HOOK_LATE(enkits)"},
	{"allocation.routed-through-facade:enkits", "ALLOC_NOT_ROUTED(enkits)"},
	{"diagnostics.routed-through-facade:enkits", "DIAG_NOT_ROUTED(enkits)"},
	{"jobs.routed-through-interface:enkits", "JOBS_NOT_ROUTED(enkits)"},
}

func TestTheTestbedExtendsWithJoltAndEnkiTSProbed(t *testing.T) {
	requirePublicMapBuilds(t)
	tsan := hostPlatform != "windows-x64-msvc"
	if tsan {
		requireTSan(t)
	}
	repo := testbedRepo(t)
	repo.Automap(t, "start", "--with", "sdl3", "--with", "dear-imgui:docking", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)
	repo.Automap(t, "extend", "jolt", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)

	run := repo.Automap(t, "extend", "enkits", "--map", "map", "--json")

	if run.Exit != 0 || run.Result.Level != "probed" {
		run.LogDetails(t)
		logVcpkgErrors(t)
		t.Fatalf("exit %d, level %q; want 0, probed\n%s", run.Exit, run.Result.Level, run.Stdout)
	}
	e := checkEvidence(t, run, "probed")
	routed := map[string]bool{}
	for _, r := range e.Predicate.Results {
		if r.Rule == "jobs.routed-through-interface" && r.Status == "pass" {
			routed[r.Library] = true
		}
		if r.Rule == "jobs.tsan-clean" && tsan && r.Status != "pass" {
			t.Errorf("jobs.tsan-clean %+v with enkiTS scheduling Jolt's jobs, want pass", r)
		}
	}
	if !routed["jolt"] || !routed["enkits"] {
		t.Errorf("jobs routed for %v, want Jolt's jobs through the interface and the interface's onto enkiTS", routed)
	}

	for _, c := range enkitsSeeds {
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

// The seeds mimalloc adds when it fills the allocator slot.
var mimallocSeeds = []struct{ seed, code string }{
	{"allocation.routed-through-facade:mimalloc", "ALLOC_NOT_ROUTED(mimalloc)"},
	{"diagnostics.routed-through-facade:mimalloc", "DIAG_NOT_ROUTED(mimalloc)"},
}

func TestTheTestbedExtendsWithJoltAndMimallocProbed(t *testing.T) {
	requirePublicMapBuilds(t)
	tsan := hostPlatform != "windows-x64-msvc"
	if tsan {
		requireTSan(t)
	}
	repo := testbedRepo(t)
	repo.Automap(t, "start", "--with", "sdl3", "--with", "dear-imgui:docking", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)
	repo.Automap(t, "extend", "jolt", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)

	run := repo.Automap(t, "extend", "mimalloc", "--map", "map", "--json")

	if run.Exit != 0 || run.Result.Level != "probed" {
		run.LogDetails(t)
		logVcpkgErrors(t)
		t.Fatalf("exit %d, level %q; want 0, probed\n%s", run.Exit, run.Result.Level, run.Stdout)
	}
	e := checkEvidence(t, run, "probed")
	routed := map[string]bool{}
	for _, r := range e.Predicate.Results {
		if r.Rule == "allocation.routed-through-facade" && r.Status == "pass" {
			routed[r.Library] = true
		}
		if r.Rule == "jobs.tsan-clean" && tsan && r.Status != "pass" {
			t.Errorf("jobs.tsan-clean %+v with mimalloc serving the facade, want pass", r)
		}
	}
	for _, lib := range []string{"sdl3", "dear-imgui", "jolt", "mimalloc"} {
		if !routed[lib] {
			t.Errorf("allocations routed for %v, want every library's through the facade and the facade's onto mimalloc", routed)
			break
		}
	}

	for _, c := range mimallocSeeds {
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
