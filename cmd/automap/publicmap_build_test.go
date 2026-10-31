package main_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
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
	patches := seedPatches(t, filepath.Join(clitest.Root(t), "map"))
	if covered := len(testbedSeeds) + len(joltSeeds) + len(enkitsSeeds) + len(mimallocSeeds) + len(diagnosticsBackendSeeds) + len(glmSeeds) + len(miniaudioSeeds) + len(assetsSeeds) + len(deterministicSeeds); patches != covered {
		t.Errorf("the public map carries %d seed patches but the testbed tests cover %d", patches, covered)
	}
}

// seedPatches counts the seed patches a map's recipes and contracts carry.
// A contract's older versions keep their seeds for compositions locked to
// them; only each contract's newest counts, and only for contracts some
// library in the map declares facts for.
func seedPatches(t *testing.T, root string) int {
	t.Helper()
	used := map[string]bool{}
	libraries, _ := filepath.Glob(filepath.Join(root, "libraries", "*.json"))
	for _, l := range libraries {
		var lib struct {
			Versions []struct{ Facts map[string]any }
		}
		raw, _ := os.ReadFile(l)
		json.Unmarshal(raw, &lib)
		for _, v := range lib.Versions {
			for c := range v.Facts {
				used[c] = true
			}
		}
	}
	var patches int
	filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		rel, _ := filepath.Rel(root, p)
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if err == nil && d.IsDir() && len(parts) == 2 && parts[0] == "contracts" && !used[parts[1]] {
			return filepath.SkipDir
		}
		if err == nil && d.IsDir() && len(parts) == 3 && parts[0] == "contracts" && parts[2] != newestContract(t, filepath.Dir(p)) {
			return filepath.SkipDir
		}
		if err == nil && slices.Contains(parts, "seeded") && !d.IsDir() {
			patches++
		}
		return nil
	})
	return patches
}

// newestContract is the newest version among a contract's directories.
func newestContract(t *testing.T, dir string) string {
	t.Helper()
	entries, _ := os.ReadDir(dir)
	newest, best := "", [3]int{-1}
	for _, e := range entries {
		var v [3]int
		if !e.IsDir() {
			continue
		}
		if _, err := fmt.Sscanf(e.Name(), "%d.%d.%d", &v[0], &v[1], &v[2]); err == nil && slices.Compare(v[:], best[:]) > 0 {
			newest, best = e.Name(), v
		}
	}
	return newest
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
		// No library in the testbed exposes world positions, and no math
		// library fills the math slot.
		if r.Rule == "math-types.precision-agrees" && (r.Status != "not-applicable" || !strings.Contains(r.Reason, "world positions")) {
			t.Errorf("math-types.precision-agrees %+v in the testbed, want not-applicable: nothing exposes world positions", r)
		}
		if r.Rule == "math-types.adapters-round-trip" && (r.Status != "not-applicable" || !strings.Contains(r.Reason, "math slot")) {
			t.Errorf("math-types.adapters-round-trip %+v in the testbed, want not-applicable: the math slot is empty", r)
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
	{"handles.bits-preserved:jolt", "HANDLE_TRUNCATED(jolt)"},
	{"serialization.rebuild-key-complete:jolt", "REBUILD_KEY_INCOMPLETE(jolt)"},
	{"determinism.flags-set", "DETERMINISM_FLAGS_MISSING(jolt)"},
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

// The seeds spdlog and Tracy add as the diagnostics facade's backends.
var diagnosticsBackendSeeds = []struct{ seed, code string }{
	{"diagnostics.routed-through-facade:spdlog", "DIAG_NOT_ROUTED(spdlog)"},
	{"diagnostics.routed-through-facade:tracy", "DIAG_NOT_ROUTED(tracy)"},
	{"build-flags.abi-defines-match:tracy", "ABI_DEFINE_MISMATCH(tracy)"},
}

// spdlog and Tracy back the facade of a composition whose libraries already
// log and profile through it: enkiTS's profiler callbacks open facade zones.
func TestTheTestbedExtendsWithSpdlogAndTracyProbed(t *testing.T) {
	requirePublicMapBuilds(t)
	tsan := hostPlatform != "windows-x64-msvc"
	if tsan {
		requireTSan(t)
	}
	repo := testbedRepo(t)
	repo.Automap(t, "start", "--with", "sdl3", "--with", "dear-imgui:docking", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)
	for _, lib := range []string{"jolt", "enkits", "spdlog"} {
		repo.Automap(t, "extend", lib, "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)
	}

	run := repo.Automap(t, "extend", "tracy", "--map", "map", "--json")

	if run.Exit != 0 || run.Result.Level != "probed" {
		run.LogDetails(t)
		logVcpkgErrors(t)
		t.Fatalf("exit %d, level %q; want 0, probed\n%s", run.Exit, run.Result.Level, run.Stdout)
	}
	e := checkEvidence(t, run, "probed")
	routed := map[string]bool{}
	for _, r := range e.Predicate.Results {
		if r.Rule == "diagnostics.routed-through-facade" && r.Status == "pass" {
			routed[r.Library] = true
		}
	}
	for _, lib := range []string{"sdl3", "jolt", "enkits", "spdlog", "tracy"} {
		if !routed[lib] {
			t.Errorf("diagnostics routed for %v, want every library's through the facade and the facade's onto spdlog and Tracy", routed)
			break
		}
	}

	for _, c := range diagnosticsBackendSeeds {
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

// The seeds GLM's arrival adds: Jolt's and Dear ImGui's adapters to it.
var glmSeeds = []struct{ seed, code string }{
	{"math-types.adapters-round-trip:jolt", "ADAPTER_LOSSY(jolt)"},
	{"math-types.adapters-round-trip:dear-imgui", "ADAPTER_LOSSY(dear-imgui)"},
}

// GLM fills the math slot of a composition with Jolt and Dear ImGui, whose
// math-types bindings convert their types to GLM's and back.
func TestTheTestbedExtendsWithJoltAndGLMProbed(t *testing.T) {
	requirePublicMapBuilds(t)
	tsan := hostPlatform != "windows-x64-msvc"
	if tsan {
		requireTSan(t)
	}
	repo := testbedRepo(t)
	repo.Automap(t, "start", "--with", "sdl3", "--with", "dear-imgui:docking", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)
	repo.Automap(t, "extend", "jolt", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)

	run := repo.Automap(t, "extend", "glm", "--map", "map", "--json")

	if run.Exit != 0 || run.Result.Level != "probed" {
		run.LogDetails(t)
		logVcpkgErrors(t)
		t.Fatalf("exit %d, level %q; want 0, probed\n%s", run.Exit, run.Result.Level, run.Stdout)
	}
	e := checkEvidence(t, run, "probed")
	passed := map[string]bool{}
	for _, r := range e.Predicate.Results {
		if strings.HasPrefix(r.Rule, "math-types.") && r.Status == "pass" {
			passed[r.Rule+":"+r.Library] = true
		}
	}
	for _, want := range []string{"math-types.adapters-round-trip:jolt", "math-types.adapters-round-trip:dear-imgui", "math-types.precision-agrees:"} {
		if !passed[want] {
			t.Errorf("math-types results passed: %v; want %s", passed, want)
		}
	}

	seeds := append(slices.Clone(glmSeeds), struct{ seed, code string }{"math-types.precision-agrees", "PRECISION_MISMATCH(jolt)"})
	for _, c := range seeds {
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

// The seeds miniaudio adds to the testbed.
var miniaudioSeeds = []struct{ seed, code string }{
	{"allocation.hooks-before-first-allocation:miniaudio", "ALLOC_HOOK_LATE(miniaudio)"},
	{"allocation.routed-through-facade:miniaudio", "ALLOC_NOT_ROUTED(miniaudio)"},
	{"diagnostics.routed-through-facade:miniaudio", "DIAG_NOT_ROUTED(miniaudio)"},
}

// miniaudio joins the testbed and runs on its null backend, so the
// composition reaches probed on a runner without an audio device.
func TestTheTestbedExtendsWithMiniaudioProbed(t *testing.T) {
	requirePublicMapBuilds(t)
	tsan := hostPlatform != "windows-x64-msvc"
	if tsan {
		requireTSan(t)
	}
	repo := testbedRepo(t)
	repo.Automap(t, "start", "--with", "sdl3", "--with", "dear-imgui:docking", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)

	run := repo.Automap(t, "extend", "miniaudio", "--map", "map", "--json")

	if run.Exit != 0 || run.Result.Level != "probed" {
		run.LogDetails(t)
		logVcpkgErrors(t)
		t.Fatalf("exit %d, level %q; want 0, probed\n%s", run.Exit, run.Result.Level, run.Stdout)
	}
	e := checkEvidence(t, run, "probed")
	passed := map[string]bool{}
	for _, r := range e.Predicate.Results {
		if r.Library == "miniaudio" && r.Status == "pass" {
			passed[r.Rule] = true
		}
	}
	for _, rule := range []string{"allocation.hooks-before-first-allocation", "allocation.routed-through-facade", "diagnostics.routed-through-facade"} {
		if !passed[rule] {
			t.Errorf("miniaudio's results passed: %v; want %s", passed, rule)
		}
	}

	for _, c := range miniaudioSeeds {
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

// The seeds fastgltf and stb_image add, with the assets contract.
var assetsSeeds = []struct{ seed, code string }{
	{"assets.reads-through-interface:fastgltf", "ASSET_READ_BYPASSED(fastgltf)"},
	{"assets.output-identical:fastgltf", "ASSET_OUTPUT_DIFFERS(fastgltf)"},
	{"assets.reads-through-interface:stb-image", "ASSET_READ_BYPASSED(stb-image)"},
	{"assets.output-identical:stb-image", "ASSET_OUTPUT_DIFFERS(stb-image)"},
	{"allocation.routed-through-facade:stb-image", "ALLOC_NOT_ROUTED(stb-image)"},
}

// fastgltf imports a model and stb_image decodes an image, both reading
// files only through the assets interface.
func TestTheTestbedExtendsWithAssetLibrariesProbed(t *testing.T) {
	requirePublicMapBuilds(t)
	if hostPlatform != "windows-x64-msvc" {
		requireTSan(t)
	}
	repo := testbedRepo(t)
	repo.Automap(t, "start", "--with", "sdl3", "--with", "dear-imgui:docking", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)
	repo.Automap(t, "extend", "fastgltf", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)

	run := repo.Automap(t, "extend", "stb-image", "--map", "map", "--json")

	if run.Exit != 0 || run.Result.Level != "probed" {
		run.LogDetails(t)
		logVcpkgErrors(t)
		t.Fatalf("exit %d, level %q; want 0, probed\n%s", run.Exit, run.Result.Level, run.Stdout)
	}
	e := checkEvidence(t, run, "probed")
	passed := map[string]bool{}
	for _, r := range e.Predicate.Results {
		if r.Status == "pass" {
			passed[r.Rule+":"+r.Library] = true
		}
	}
	for _, want := range []string{"assets.reads-through-interface:fastgltf", "assets.output-identical:fastgltf",
		"assets.reads-through-interface:stb-image", "assets.output-identical:stb-image", "allocation.routed-through-facade:stb-image"} {
		if !passed[want] {
			t.Errorf("%s didn't pass", want)
		}
	}

	for _, c := range assetsSeeds {
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

// extendWithJoltVariant extends the testbed with a variant of Jolt from the
// Automap registry, with choices set for it, and verifies.
func extendWithJoltVariant(t *testing.T, variant string, choices map[string]any) (*clitest.Repo, *clitest.Run) {
	t.Helper()
	repo := testbedRepo(t)
	repo.Automap(t, "start", "--with", "sdl3", "--with", "dear-imgui:docking", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)
	repo.Automap(t, "extend", "jolt:"+variant, "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)
	for path, value := range choices {
		repo.EditJSON(t, "automap.manifest.json", set("choices/"+path, value))
	}

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	if run.Exit != 0 || run.Result.Level != "probed" {
		run.LogDetails(t)
		logVcpkgErrors(t)
		t.Fatalf("exit %d, level %q; want 0, probed\n%s", run.Exit, run.Result.Level, run.Stdout)
	}
	return repo, run
}

// Jolt with double precision world positions, from the Automap registry,
// in a composition that chooses double precision.
func TestTheTestbedWithDoublePrecisionJoltProbed(t *testing.T) {
	requirePublicMapBuilds(t)
	if hostPlatform != "windows-x64-msvc" {
		requireTSan(t)
	}
	_, run := extendWithJoltVariant(t, "double-precision", map[string]any{"math-types/position_precision": "double"})

	e := checkEvidence(t, run, "probed")
	for _, r := range e.Predicate.Results {
		if r.Library == "jolt" && r.Status != "pass" && r.Status != "not-applicable" {
			t.Errorf("Jolt's result %+v, want a pass", r)
		}
		if r.Rule == "math-types.precision-agrees" && r.Status != "pass" {
			t.Errorf("math-types.precision-agrees %+v with double precision Jolt, want a pass", r)
		}
	}
}

// The seed only the deterministic build can run.
var deterministicSeeds = []struct{ seed, code string }{
	{"determinism.replay-hash-matches:jolt", "REPLAY_MISMATCH(jolt)"},
}

// Jolt built for cross-platform determinism, from the Automap registry,
// under the determinism profile: the fixed replay gives the same hash on
// every platform.
func TestTheTestbedWithDeterministicJoltProbed(t *testing.T) {
	requirePublicMapBuilds(t)
	if hostPlatform != "windows-x64-msvc" {
		requireTSan(t)
	}
	repo, run := extendWithJoltVariant(t, "deterministic", map[string]any{"determinism/profile": "on"})

	e := checkEvidence(t, run, "probed")
	passed := map[string]bool{}
	for _, r := range e.Predicate.Results {
		if r.Status == "pass" {
			passed[r.Rule+":"+r.Library] = true
		}
	}
	for _, want := range []string{"determinism.flags-set:", "determinism.replay-hash-matches:jolt", "diagnostics.routed-through-facade:jolt"} {
		if !passed[want] {
			t.Errorf("%s didn't pass:\n%s", want, run.Stdout)
		}
	}

	for _, c := range deterministicSeeds {
		t.Run(c.seed, func(t *testing.T) {
			run := repo.Automap(t, "verify", "--seed", c.seed, "--map", "map", "--json")

			if run.Exit != 0 || !strings.Contains(run.Result.Summary, c.code) {
				run.LogDetails(t)
				t.Errorf("exit %d, summary %q; want the probe to report %s\n%s", run.Exit, run.Result.Summary, c.code, run.Stdout)
			}
		})
	}
}
