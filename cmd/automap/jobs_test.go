package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/zach/automap/internal/clitest"
)

// requireTSan skips a test that needs clang's ThreadSanitizer runtime
// where it isn't installed; CI installs it, and fails instead of skipping.
func requireTSan(t *testing.T) {
	t.Helper()
	requireBuilds(t)
	if hostPlatform == "windows-x64-msvc" {
		t.Skip("ThreadSanitizer doesn't run on " + hostPlatform)
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "t.cpp")
	os.WriteFile(src, []byte("int main() { return 0; }\n"), 0o644)
	if out, err := exec.Command("clang++", "-fsanitize=thread", src, "-o", filepath.Join(dir, "t")).CombinedOutput(); err != nil {
		if os.Getenv("AUTOMAP_TEST_BUILDS") == "1" {
			t.Fatalf("clang's ThreadSanitizer runtime is missing: %s", out)
		}
		t.Skip("needs clang's ThreadSanitizer runtime")
	}
}

// physicsRepo is a fixture repo with stub-window, extended with
// stub-physics, whose jobs facts bring in the jobs contract.
func physicsRepo(t *testing.T, repo *clitest.Repo) *clitest.Repo {
	t.Helper()
	repo.Automap(t, startWindow...).ExpectExit(t, 0)
	repo.Automap(t, "extend", "stub-physics", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)
	return repo
}

func TestExtendingWithAJobsLibraryWritesTheJobsInterface(t *testing.T) {
	repo := fixtureRepo(t)
	repo.Automap(t, startWindow...).ExpectExit(t, 0)

	run := repo.Automap(t, "extend", "stub-physics", "--map", "map", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	for _, f := range []string{"glue/automap/jobs.h", "glue/automap/jobs.cpp", "glue/stub-physics/stub_physics_jobs.cpp"} {
		if !slices.Contains(run.Result.FilesWritten, f) {
			t.Errorf("files_written lacks %s", f)
		}
	}
	header := run.ReadFile(t, "glue/automap/choices.h")
	for _, want := range []string{"#define AUTOMAP_JOBS_SCHEDULER_REFERENCE_THREAD_POOL 1\n", "#define AUTOMAP_JOBS_WORKERS 3\n"} {
		if !strings.Contains(header, want) {
			t.Errorf("choices.h lacks %q", want)
		}
	}
	startup := run.ReadFile(t, "glue/automap/startup.cpp")
	if a, b := strings.Index(startup, "am_jobs_setup_init(ctx)"), strings.Index(startup, "am_bind_stub_physics_jobs(ctx)"); a < 0 || b < 0 || a > b {
		t.Errorf("the startup doesn't start the scheduler before stub-physics's jobs binding:\n%s", startup)
	}
	if !strings.Contains(startup, "am_jobs_setup_shutdown(ctx)") {
		t.Error("the startup doesn't stop the scheduler")
	}
}

func TestTheJobsCompositionIsProbedWithACleanThreadSanitizerRun(t *testing.T) {
	requireTSan(t)
	repo := physicsRepo(t, buildableRepo(t))

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	if run.Exit != 0 || run.Result.Level != "probed" {
		run.LogDetails(t)
		t.Fatalf("exit %d, level %q; want probed\n%s", run.Exit, run.Result.Level, run.Stdout)
	}
	e := checkEvidence(t, run, "probed")
	for _, r := range e.Predicate.Results {
		if r.Rule == "jobs.tsan-clean" && r.Status != "pass" {
			t.Errorf("jobs.tsan-clean %s, want pass", r.Status)
		}
	}
	if !strings.Contains(run.ReadFile(t, "evidence/"+hostPlatform+".json"), `"thread": "separate Debug build with -fsanitize=thread`) {
		t.Error("the evidence's run conditions don't record the ThreadSanitizer run")
	}
}

// The seeds the stub-physics composition adds to the fixture's.
var physicsSeeds = []struct{ seed, code string }{
	{"jobs.routed-through-interface:stub-physics", "JOBS_NOT_ROUTED(stub-physics)"},
	{"jobs.tsan-clean", "TSAN_REPORT(data race jobs.cpp"},
	{"math-types.precision-agrees", "PRECISION_MISMATCH(stub-physics)"},
}

func TestEveryJobsSeedIsCaught(t *testing.T) {
	requireTSan(t)
	repo := physicsRepo(t, buildableRepo(t))
	for _, c := range physicsSeeds {
		t.Run(c.seed, func(t *testing.T) {
			t.Parallel()
			run := repo.Automap(t, "verify", "--seed", c.seed, "--map", "map", "--json")

			if run.Exit != 0 || !strings.Contains(run.Result.Summary, c.code) {
				run.LogDetails(t)
				t.Errorf("exit %d, summary %q; want the probe to report %s", run.Exit, run.Result.Summary, c.code)
			}
		})
	}
}

func TestARaceInsideAnUninstrumentedSharedLibraryIsSetAside(t *testing.T) {
	requireTSan(t)
	if runtime.GOOS != "linux" {
		t.Skip("the uninstrumented library is a Linux shared object")
	}
	repo := physicsRepo(t, buildableRepo(t))
	// noise is a shared library built without ThreadSanitizer, like a GPU
	// driver: its threads race inside it, through the memset interceptor.
	repo.WriteFile(t, "noise/noise.c", `#include <pthread.h>
#include <string.h>
static char buf[4096];
static void* touch(void* a) { (void)a; for (int i = 0; i < 2000; ++i) memset(buf, i, sizeof buf); return 0; }
void noise_race(void) { pthread_t a, b; pthread_create(&a, 0, touch, 0); pthread_create(&b, 0, touch, 0); pthread_join(a, 0); pthread_join(b, 0); }
`)
	repo.WriteFile(t, "CMakeLists.txt", clitest.Read(t, repo, "CMakeLists.txt")+`
add_library(noise SHARED noise/noise.c)
target_compile_options(noise PRIVATE -fno-sanitize=thread)
target_link_libraries(testbed PRIVATE noise)
`)
	main := clitest.Read(t, repo, "main.cpp")
	main = strings.Replace(main, "int main(", "extern \"C\" void noise_race();\n\nint main(", 1)
	main = strings.Replace(main, "    am::shutdown(ctx);", "    noise_race();\n    am::shutdown(ctx);", 1)
	repo.WriteFile(t, "main.cpp", main)

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	if run.Exit != 0 || run.Result.Level != "probed" {
		run.LogDetails(t)
		t.Fatalf("exit %d, level %q; want probed, with the driver-like race set aside\n%s", run.Exit, run.Result.Level, run.Stdout)
	}
	evidence := run.ReadFile(t, "evidence/"+hostPlatform+".json")
	if !strings.Contains(evidence, `"set_aside_reports"`) || !strings.Contains(evidence, "libnoise.so") {
		t.Errorf("the evidence doesn't record the set-aside report in libnoise.so:\n%s", evidence)
	}
}

func TestThreadSanitizerIsNotApplicableOffItsPlatforms(t *testing.T) {
	repo := buildableRepo(t)
	// The rule runs only on another platform than this one.
	elsewhere := "macos-arm64-appleclang"
	if hostPlatform == elsewhere {
		elsewhere = "linux-x64-clang"
	}
	repo.EditJSON(t, "map/contracts/jobs/1.0.0.json", set("rules/1/platforms", []any{elsewhere}))
	physicsRepo(t, repo)

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	if run.Exit != 0 || run.Result.Level != "probed" {
		run.LogDetails(t)
		t.Fatalf("exit %d, level %q; want probed\n%s", run.Exit, run.Result.Level, run.Stdout)
	}
	e := checkEvidence(t, run, "probed")
	for _, r := range e.Predicate.Results {
		if r.Rule == "jobs.tsan-clean" && (r.Status != "not-applicable" || !strings.Contains(r.Reason, elsewhere)) {
			t.Errorf("jobs.tsan-clean %+v, want not-applicable on %s, naming the platform it runs on", r, hostPlatform)
		}
	}
}
