package main_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zach/automap/internal/clitest"
)

// commitRepo makes the repo a git repository with everything committed, as
// a studio's repo holding its private map is.
func commitRepo(t *testing.T, repo *clitest.Repo) {
	t.Helper()
	for _, args := range [][]string{{"init", "-q"}, {"add", "-A"}, {"commit", "-q", "-m", "acme"}} {
		cmd := exec.Command("git", append([]string{"-C", repo.Dir}, args...)...)
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=acme", "GIT_AUTHOR_EMAIL=acme@example.invalid",
			"GIT_COMMITTER_NAME=acme", "GIT_COMMITTER_EMAIL=acme@example.invalid")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

// useAcmePrivateMap lists Acme's private map in the manifest and chooses
// its patched stub-ui, as the agent does in the spec's journey 3.
func useAcmePrivateMap(t *testing.T, repo *clitest.Repo) {
	t.Helper()
	repo.EditJSON(t, "automap.manifest.json", set("private_maps", []any{map[string]any{"id": "acme", "location": "automap-private"}}))
	repo.EditJSON(t, "automap.manifest.json", set("variants/stub-ui", "acme-patched"))
	commitRepo(t, repo)
}

// acmeStubsRepo is Acme's repo with stub sources, on a release of the
// fixture map served by s (in mapDir), where the testbed has been started
// and then upgraded in place onto the private map and its patched stub-ui.
func acmeStubsRepo(t *testing.T, s *releaseServer, mapDir string) *clitest.Repo {
	t.Helper()
	s.publish(t, "map-2026.10.01", mapDir, nil)
	repo := clitest.NewRepo(t, "acme-stubs")
	repo.Setenv("AUTOMAP_RELEASES", s.URL)
	repo.Setenv("AUTOMAP_EVIDENCE_URL", s.URL+"/main")
	repo.Setenv("AUTOMAP_CACHE", t.TempDir())
	repo.Automap(t, "start", "--with", "stub-window", "--with", "stub-ui", "--no-verify", "--json").ExpectExit(t, 0)
	useAcmePrivateMap(t, repo)
	repo.Automap(t, "upgrade", "stub-ui", "--no-verify", "--json").ExpectExit(t, 0)
	return repo
}

type lockFile struct {
	PrivateMaps map[string]struct{ Commit string } `json:"private_maps"`
	Pins        map[string]struct {
		Variant     string
		Path        string
		ContentHash string `json:"content_hash"`
		Registry    *struct{ Kind, ID, Commit string }
	}
}

func readLock(t *testing.T, repo *clitest.Repo) lockFile {
	t.Helper()
	var lock lockFile
	json.Unmarshal([]byte(clitest.Read(t, repo, "automap.lock.json")), &lock)
	return lock
}

func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return strings.TrimSpace(string(out))
}

// The lock pins a private map at the commit that last changed it, a private
// variant at that map, and an in-repo library by path and content hash
// (SPEC §4.2).
func TestPrivateContentIsPinned(t *testing.T) {
	repo := acmeStubsRepo(t, newReleaseServer(t), fixtureMapDir(t))

	run := repo.Automap(t, "extend", "acme-tools", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	lock := readLock(t, repo)
	commit := gitOutput(t, repo.Dir, "log", "-1", "--format=%H", "--", "automap-private")
	if lock.PrivateMaps["acme"].Commit != commit {
		t.Errorf("private map acme pinned at %q, want its last commit %s", lock.PrivateMaps["acme"].Commit, commit)
	}
	ui := lock.Pins["stub-ui"]
	if ui.Variant != "acme-patched" || ui.Registry == nil || ui.Registry.Kind != "private" || ui.Registry.ID != "acme" || ui.Registry.Commit != commit {
		t.Errorf("stub-ui pinned as %+v, want acme-patched from private map acme at %s", ui, commit)
	}
	tools := lock.Pins["acme-tools"]
	if tools.Path != "engine/acme-tools" || !strings.HasPrefix(tools.ContentHash, "sha256:") || tools.Registry != nil {
		t.Errorf("acme-tools pinned as %+v, want path engine/acme-tools with a sha256 content hash", tools)
	}
}

// vcpkg builds the private variant from the private map's overlay port,
// and CMake adds the in-repo library's directory.
func TestPrivateContentJoinsTheBuild(t *testing.T) {
	repo := acmeStubsRepo(t, newReleaseServer(t), fixtureMapDir(t))

	repo.Automap(t, "extend", "acme-tools", "--no-verify", "--json").ExpectExit(t, 0)

	var config struct {
		OverlayPorts []string `json:"overlay-ports"`
	}
	json.Unmarshal([]byte(clitest.Read(t, repo, "vcpkg-configuration.json")), &config)
	if len(config.OverlayPorts) != 1 || config.OverlayPorts[0] != "automap-private/ports" {
		t.Errorf("vcpkg-configuration.json overlay ports %v, want the private map's automap-private/ports", config.OverlayPorts)
	}
	cmake := clitest.Read(t, repo, "cmake/automap/automap.cmake")
	if !strings.Contains(cmake, "add_subdirectory(${AUTOMAP_ROOT}/engine/acme-tools") {
		t.Errorf("automap.cmake doesn't add engine/acme-tools:\n%s", cmake)
	}
}

// Changing an in-repo library's source after the lock leaves the lock
// pinning other code: verify refuses until the composition is re-resolved.
func TestVerifyRefusesAnInRepoLibraryThatChangedSinceTheLock(t *testing.T) {
	repo := acmeStubsRepo(t, newReleaseServer(t), fixtureMapDir(t))
	repo.Automap(t, "extend", "acme-tools", "--no-verify", "--json").ExpectExit(t, 0)
	os.WriteFile(filepath.Join(repo.Dir, "engine/acme-tools/acme_tools.cpp"), []byte("// changed\n"), 0o644)

	run := repo.Automap(t, "verify", "--json")

	run.ExpectExit(t, 2)
	run.ExpectFailure(t, "LOCK_MISMATCH(acme-tools)", "engine/acme-tools")
	before := readLock(t, repo).Pins["acme-tools"].ContentHash
	repo.Automap(t, "upgrade", "acme-tools", "--no-verify", "--json").ExpectExit(t, 0)
	if after := readLock(t, repo).Pins["acme-tools"].ContentHash; after == before {
		t.Errorf("upgrade acme-tools kept content hash %s for the changed source", before)
	}
}

// The CLI never sends private-map content over the network (ADR-0047):
// published evidence for an edge with private pieces is read from the
// private map, so no request names them.
func TestNoRequestCarriesPrivateContent(t *testing.T) {
	s := newReleaseServer(t)
	repo := acmeStubsRepo(t, s, fixtureMapDir(t))

	repo.Automap(t, "extend", "acme-tools", "--no-verify", "--json").ExpectExit(t, 0)
	repo.Automap(t, "show", "library", "stub-ui", "--json").ExpectExit(t, 0)
	repo.Automap(t, "find", "--slot", "scripting", "--json").ExpectExit(t, 0)

	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.requests) == 0 {
		t.Fatal("no requests reached the release server")
	}
	for _, r := range s.requests {
		if strings.Contains(r, "acme") || strings.Contains(r, "scripting") {
			t.Errorf("request %s carries private content", r)
		}
	}
}

// The spec's Acme example end to end with stub sources: the in-repo library
// builds with add_subdirectory and the patched stub-ui from the private
// map's overlay port, and the composition reaches probed.
func TestAcmeStubsBuildWithPrivateContent(t *testing.T) {
	requireBuilds(t)
	url, commit := fixtureRegistry(t)
	mapDir := fixtureMapDir(t)
	(&clitest.Repo{Dir: filepath.Dir(mapDir)}).EditJSON(t, "map/map.json", set("build_backend/vcpkg/automap_registry", map[string]any{"url": url, "commit": commit}))
	repo := acmeStubsRepo(t, newReleaseServer(t), mapDir)

	run := repo.Automap(t, "extend", "acme-tools", "--json")

	if run.Exit != 0 || run.Result.Level != "probed" {
		run.LogDetails(t)
		t.Fatalf("exit %d, level %q; want 0, probed\n%s", run.Exit, run.Result.Level, run.Stdout)
	}
	headers, _ := filepath.Glob(filepath.Join(repo.Dir, "build/automap", hostPlatform, "vcpkg_installed/*/include/stub_ui.h"))
	if len(headers) != 1 {
		t.Fatalf("installed stub_ui.h: %v, want one", headers)
	}
	if raw, _ := os.ReadFile(headers[0]); !strings.Contains(string(raw), "STUB_UI_ACME_PATCHED") {
		t.Error("stub-ui wasn't built from the private map's overlay port")
	}
}
