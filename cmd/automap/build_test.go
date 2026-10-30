package main_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/zach/automap/internal/clitest"
	"github.com/zach/automap/internal/verify"
)

// hostPlatform is the platform verify runs on here.
var hostPlatform = verify.HostPlatform()

// requireBuilds skips a test that builds C++ through vcpkg unless
// VCPKG_ROOT is set. CI sets AUTOMAP_TEST_BUILDS=1, so there they fail
// instead of skipping.
func requireBuilds(t *testing.T) {
	t.Helper()
	if os.Getenv("VCPKG_ROOT") != "" {
		return
	}
	if os.Getenv("AUTOMAP_TEST_BUILDS") == "1" {
		t.Fatal("AUTOMAP_TEST_BUILDS=1 but VCPKG_ROOT isn't set")
	}
	t.Skip("builds through vcpkg: set VCPKG_ROOT to run")
}

var (
	registryOnce   sync.Once
	registryURL    string
	registryCommit string
	registryErr    error
)

// fixtureRegistry creates, once per test run, a git registry from the
// fixture map's registry directory: its ports on a registry branch, with a
// versions database computed from their git trees (SPEC §5.5).
func fixtureRegistry(t *testing.T) (string, string) {
	t.Helper()
	registryOnce.Do(func() {
		dir, err := os.MkdirTemp("", "automap-registry-")
		if err != nil {
			registryErr = err
			return
		}
		git := func(args ...string) string {
			if registryErr != nil {
				return ""
			}
			cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
			cmd.Env = append(os.Environ(),
				"GIT_AUTHOR_NAME=automap", "GIT_AUTHOR_EMAIL=automap@example.invalid", "GIT_AUTHOR_DATE=2026-01-01T00:00:00Z",
				"GIT_COMMITTER_NAME=automap", "GIT_COMMITTER_EMAIL=automap@example.invalid", "GIT_COMMITTER_DATE=2026-01-01T00:00:00Z")
			out, err := cmd.CombinedOutput()
			if err != nil {
				registryErr = err
				return string(out)
			}
			return strings.TrimSpace(string(out))
		}
		repo := clitest.NewRepo(t, "no-manifest")
		repo.Copy(t, "testdata/map/registry/ports", "ports")
		if err := os.Rename(filepath.Join(repo.Dir, "ports"), filepath.Join(dir, "ports")); err != nil {
			registryErr = err
			return
		}
		git("init", "-q", "-b", "registry")
		git("add", "ports")
		git("commit", "-q", "-m", "ports")
		baseline := map[string]any{}
		entries, _ := os.ReadDir(filepath.Join(dir, "ports"))
		for _, e := range entries {
			port := e.Name()
			var manifest struct{ Version string }
			raw, _ := os.ReadFile(filepath.Join(dir, "ports", port, "vcpkg.json"))
			json.Unmarshal(raw, &manifest)
			tree := git("rev-parse", "HEAD:ports/"+port)
			versions, _ := json.Marshal(map[string]any{"versions": []any{map[string]any{"version": manifest.Version, "git-tree": tree}}})
			os.MkdirAll(filepath.Join(dir, "versions", port[:1]+"-"), 0o755)
			os.WriteFile(filepath.Join(dir, "versions", port[:1]+"-", port+".json"), versions, 0o644)
			baseline[port] = map[string]any{"baseline": manifest.Version, "port-version": 0}
		}
		raw, _ := json.Marshal(map[string]any{"default": baseline})
		os.WriteFile(filepath.Join(dir, "versions", "baseline.json"), raw, 0o644)
		git("add", "versions")
		git("commit", "-q", "-m", "versions")
		registryURL, registryCommit = "file://"+filepath.ToSlash(dir), git("rev-parse", "HEAD")
	})
	if registryErr != nil {
		t.Fatalf("creating the fixture registry: %v", registryErr)
	}
	return registryURL, registryCommit
}

// buildableRepo is a fixture repo whose map points at the fixture registry.
func buildableRepo(t *testing.T) *clitest.Repo {
	t.Helper()
	requireBuilds(t)
	url, commit := fixtureRegistry(t)
	repo := fixtureRepo(t)
	repo.EditJSON(t, "map/map.json", set("build_backend/vcpkg/automap_registry", map[string]any{"url": url, "commit": commit}))
	return repo
}

func TestVerifyReachesProbed(t *testing.T) {
	repo := buildableRepo(t)
	repo.Automap(t, startTestbed...).ExpectExit(t, 0)

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	run.ExpectExit(t, 0)
	if run.Result.Level != "probed" || run.Result.EvidenceFile != "evidence/"+hostPlatform+".json" {
		t.Fatalf("level %q, evidence %q; want probed, evidence/%s.json\n%s", run.Result.Level, run.Result.EvidenceFile, hostPlatform, run.Stdout)
	}
	evidence := checkEvidence(t, run, "probed")
	if env := evidence.Predicate.Conditions.Environment; env["STUB_WINDOW_HEADLESS"] != "1" {
		t.Errorf("run conditions' environment %v, want STUB_WINDOW_HEADLESS=1 from stub-window's platform facts", env)
	}
}

func TestVerifyWithWarmCachesTakesUnderAMinute(t *testing.T) {
	repo := buildableRepo(t)
	repo.Automap(t, startTestbed...).ExpectExit(t, 0)
	repo.Automap(t, "verify", "--map", "map", "--json").ExpectExit(t, 0)

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	run.ExpectExit(t, 0)
	if run.Duration.Seconds() >= 60 {
		t.Errorf("verify with warm caches took %s", run.Duration)
	}
}

func TestABuildFailureReportsTheBuildLog(t *testing.T) {
	repo := buildableRepo(t)
	repo.Automap(t, startTestbed...).ExpectExit(t, 0)
	repo.WriteFile(t, "glue/stub-ui/stub_ui_setup.cpp", "this is not C++\n")

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	run.ExpectExit(t, 1)
	run.ExpectFailure(t, "BUILD_FAILED", "build log")
	if run.Result.Level != "none" {
		t.Errorf("level %q, want none", run.Result.Level)
	}
	if f := run.Result.Failures[0].DetailsFile; !strings.Contains(run.ReadFile(t, f), "stub_ui_setup.cpp") {
		t.Errorf("details file %s doesn't show the failing source", f)
	}
}

func TestACrashOnStartupFailsTheSmokeRun(t *testing.T) {
	repo := buildableRepo(t)
	repo.Automap(t, startTestbed...).ExpectExit(t, 0)
	setup := clitest.Read(t, repo, "glue/stub-ui/stub_ui_setup.cpp")
	repo.WriteFile(t, "glue/stub-ui/stub_ui_setup.cpp", strings.Replace(setup, "if (!stub_ui_create_context()) return false;", "return false;", 1))

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	run.ExpectExit(t, 1)
	run.ExpectFailure(t, "SMOKE_FAILED", "INIT_FAILED(am_stub_ui_setup_init)")
	if run.Result.Level != "builds" {
		t.Errorf("level %q, want builds", run.Result.Level)
	}
}

func TestStartEndsByVerifying(t *testing.T) {
	repo := buildableRepo(t)

	run := repo.Automap(t, "start", "--with", "stub-window", "--with", "stub-ui", "--map", "map", "--json")

	run.ExpectExit(t, 0)
	if run.Result.Level != "probed" || run.Result.EvidenceFile != "evidence/"+hostPlatform+".json" {
		t.Errorf("start reports level %q and evidence %q; want probed and evidence/%s.json", run.Result.Level, run.Result.EvidenceFile, hostPlatform)
	}
	checkEvidence(t, run, "probed")
	if !strings.Contains(run.Result.Summary, "probed on "+hostPlatform) {
		t.Errorf("summary %q doesn't report the level", run.Result.Summary)
	}
}

func TestVerifyWithoutVcpkgReportsABuildFailureAndEvidence(t *testing.T) {
	repo := startedRepo(t)
	repo.Setenv("VCPKG_ROOT", "")

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	run.ExpectExit(t, 1)
	run.ExpectFailure(t, "BUILD_FAILED", "VCPKG_ROOT")
	checkEvidence(t, run, "none")
}

type evidenceDoc struct {
	Subject []struct {
		Name   string
		Digest map[string]string
	}
	Predicate struct {
		SubjectKind   string `json:"subject_kind"`
		SubjectDetail struct {
			LockHash string `json:"lock_hash"`
		} `json:"subject_detail"`
		Level      string
		MapRelease string `json:"map_release"`
		Conditions struct {
			BuildType   string `json:"build_type"`
			RunnerImage string `json:"runner_image"`
			Compiler    string
			Environment map[string]string
		}
		Results []struct {
			Rule, Library, Status, Reason string
		}
		Verifier struct {
			Kind, ID string
			Signed   bool
		}
	}
}

// checkEvidence checks the run's evidence file: schema-valid, about the
// lock (its hash and the file's digest), with run conditions, unsigned.
func checkEvidence(t *testing.T, run *clitest.Run, level string) evidenceDoc {
	t.Helper()
	raw := run.ReadFile(t, "evidence/"+hostPlatform+".json")
	clitest.ExpectSchema(t, "evidence", raw)
	var e evidenceDoc
	json.Unmarshal([]byte(raw), &e)
	lock := run.ReadFile(t, "automap.lock.json")
	sum := sha256.Sum256([]byte(lock))
	var l struct {
		LockHash string `json:"lock_hash"`
	}
	json.Unmarshal([]byte(lock), &l)
	p := e.Predicate
	switch {
	case p.SubjectKind != "composition" || p.SubjectDetail.LockHash != l.LockHash:
		t.Errorf("evidence is about %s %s, want the composition with lock_hash %s", p.SubjectKind, p.SubjectDetail.LockHash, l.LockHash)
	case len(e.Subject) != 1 || e.Subject[0].Name != "automap.lock.json" || e.Subject[0].Digest["sha256"] != hex.EncodeToString(sum[:]):
		t.Errorf("evidence subject %+v, want automap.lock.json with the lock file's sha256", e.Subject)
	case p.Level != level:
		t.Errorf("evidence level %q, want %q", p.Level, level)
	case p.Verifier.Kind != "developer-local" || p.Verifier.ID != "local" || p.Verifier.Signed:
		t.Errorf("verifier %+v, want developer-local, id local, unsigned", p.Verifier)
	case p.Conditions.BuildType == "" || p.Conditions.Compiler == "":
		t.Errorf("run conditions %+v lack the build type or compiler", p.Conditions)
	case os.Getenv("ImageOS") != "" && !strings.Contains(p.Conditions.RunnerImage, os.Getenv("ImageOS")):
		t.Errorf("run conditions' runner image %q, want the hosted runner's %s", p.Conditions.RunnerImage, os.Getenv("ImageOS"))
	case p.MapRelease != "unreleased":
		t.Errorf("map_release %q, want unreleased", p.MapRelease)
	}
	if len(p.Results) == 0 {
		t.Error("evidence has no results")
	}
	for _, r := range p.Results {
		switch {
		case level == "probed" && r.Status != "pass" && r.Status != "not-applicable":
			t.Errorf("result %+v at level probed, want pass or not-applicable", r)
		case level != "probed" && level != "smoke" && (r.Status != "unverifiable" || r.Reason == ""):
			t.Errorf("result %+v: with no probes run, every probed rule is unverifiable with a reason", r)
		case (r.Status == "unverifiable" || r.Status == "not-applicable") && r.Reason == "":
			t.Errorf("result %+v has no reason", r)
		}
	}
	return e
}

// startedRepo is a fixture repo where start has written the stub testbed,
// without verifying it.
func startedRepo(t *testing.T) *clitest.Repo {
	t.Helper()
	repo := fixtureRepo(t)
	repo.Automap(t, startTestbed...).ExpectExit(t, 0)
	return repo
}

func TestVerifyRejectsAManifestThatDisagreesWithTheMap(t *testing.T) {
	repo := startedRepo(t)
	repo.EditJSON(t, "automap.manifest.json", set("variants/stub-ui", "docking"))

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	run.ExpectExit(t, 2)
	run.ExpectFailure(t, "LOCK_MISMATCH(stub-ui)", "manifest asks for docking")
}

func TestVerifyRejectsAManifestNamingALibraryTheMapLacks(t *testing.T) {
	repo := startedRepo(t)
	repo.EditJSON(t, "automap.manifest.json", set("slots/physics-3d", "jolt"))

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	run.ExpectExit(t, 2)
	run.ExpectFailure(t, "MANIFEST_INVALID", "the map has no library jolt")
}
