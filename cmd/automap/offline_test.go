package main_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/zach/automap/internal/clitest"
)

// disconnect takes the repo's later runs off the network: every proxy
// variable points nowhere, so HTTP clients, git and vcpkg can't reach a
// host, and vcpkg may download nothing it doesn't already have.
func disconnect(repo *clitest.Repo) {
	for _, k := range []string{"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "http_proxy", "https_proxy", "all_proxy"} {
		repo.Setenv(k, "http://127.0.0.1:1")
	}
	repo.Setenv("NO_PROXY", "")
	repo.Setenv("no_proxy", "")
	repo.Setenv("X_VCPKG_ASSET_SOURCES", "x-block-origin")
}

// withoutCaches gives the repo's later runs empty caches of their own:
// Automap's, and vcpkg's downloads, registries and binaries.
func withoutCaches(t *testing.T, repo *clitest.Repo) {
	t.Helper()
	repo.Setenv("AUTOMAP_CACHE", t.TempDir())
	repo.Setenv("VCPKG_DOWNLOADS", t.TempDir())
	repo.Setenv("X_VCPKG_REGISTRIES_CACHE", t.TempDir())
	repo.Setenv("VCPKG_BINARY_SOURCES", "clear")
}

// clone copies the repo's files, as a fresh clone of it would have them:
// everything but build output.
func clone(t *testing.T, repo *clitest.Repo) *clitest.Repo {
	t.Helper()
	c := clitest.NewRepo(t, "no-manifest")
	err := filepath.WalkDir(repo.Dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(repo.Dir, p)
		if d.IsDir() {
			if rel == "build" || rel == ".automap-test-cache" {
				return filepath.SkipDir
			}
			return os.MkdirAll(filepath.Join(c.Dir, rel), 0o755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(c.Dir, rel), data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// buildableRelease publishes the fixture map, pointing at the fixture
// registry, as a release with stub-ui's published evidence, and returns
// a repo that reads it.
func buildableRelease(t *testing.T) (*releaseServer, *clitest.Repo, string) {
	t.Helper()
	requireBuilds(t)
	url, commit := fixtureRegistry(t)
	dir := fixtureMapDir(t)
	file := filepath.Join(dir, "map.json")
	var doc map[string]any
	raw, _ := os.ReadFile(file)
	json.Unmarshal(raw, &doc)
	doc["build_backend"].(map[string]any)["vcpkg"].(map[string]any)["automap_registry"] = map[string]any{"url": url, "commit": commit}
	raw, _ = json.MarshalIndent(doc, "", "  ")
	os.WriteFile(file, raw, 0o644)
	s := newReleaseServer(t)
	s.publish(t, "map-2026.10.01", dir, nil)
	for p, e := range stubUIEvidence() {
		s.evidence[p] = []byte(e)
	}
	cache := t.TempDir()
	repo := releaseRepo(t, s, cache)
	repo.Automap(t, "start", "--with", "stub-window", "--with", "stub-ui", "--no-verify", "--json").ExpectExit(t, 0)
	return s, repo, cache
}

const brokenPairing = "EVIDENCE_BROKEN(stub-ui--stub-window,linux-x64-clang)"

func expectWarning(t *testing.T, run *clitest.Run, code string) {
	t.Helper()
	for _, w := range run.Result.Warnings {
		if w.Code == code {
			return
		}
	}
	t.Errorf("no %s warning in %s", code, run.Stdout)
}

func expectProbed(t *testing.T, run *clitest.Run) {
	t.Helper()
	if run.Exit != 0 || run.Result.Level != "probed" {
		run.LogDetails(t)
		t.Fatalf("exit %d, level %q; want 0, probed\n%s", run.Exit, run.Result.Level, run.Stdout)
	}
}

func TestVerifyWorksOfflineAfterFetch(t *testing.T) {
	s, repo, _ := buildableRelease(t)

	fetch := repo.Automap(t, "fetch", "--json")

	fetch.ExpectExit(t, 0)
	expectWarning(t, fetch, brokenPairing)
	s.Close()
	disconnect(repo)
	run := repo.Automap(t, "verify", "--json")
	expectProbed(t, run)
	// The published evidence comes from fetch's copy.
	expectWarning(t, run, brokenPairing)
}

func TestAVendoredCompositionVerifiesFromAFreshCloneOffline(t *testing.T) {
	s, repo, _ := buildableRelease(t)

	run := repo.Automap(t, "vendor", "--json")

	run.ExpectExit(t, 0)
	for _, f := range []string{
		"vendor/automap/ports/stub-window/portfile.cmake",
		"vendor/automap/ports/stub-ui/portfile.cmake",
		"vendor/automap/maps/map-2026.10.01/map.json",
		"vendor/automap/evidence/pairing/stub-ui--stub-window/stub-ui-1.0.0-default+stub-window-1.0.0-default/linux-x64-clang.json",
	} {
		if _, err := os.Stat(filepath.Join(repo.Dir, filepath.FromSlash(f))); err != nil {
			t.Errorf("vendor didn't write %s", f)
		}
	}
	var config struct {
		OverlayPorts []string `json:"overlay-ports"`
	}
	json.Unmarshal([]byte(clitest.Read(t, repo, "vcpkg-configuration.json")), &config)
	if got := strings.Join(config.OverlayPorts, ","); got != "vendor/automap/ports/stub-ui,vendor/automap/ports/stub-window" {
		t.Errorf("vcpkg-configuration.json's overlay ports are %s, want the vendored ports", got)
	}

	s.Close()
	fresh := clone(t, repo)
	fresh.Setenv("AUTOMAP_RELEASES", s.URL)
	fresh.Setenv("AUTOMAP_EVIDENCE_URL", s.URL+"/main")
	withoutCaches(t, fresh)
	disconnect(fresh)
	verify := fresh.Automap(t, "verify", "--json")
	expectProbed(t, verify)
	expectWarning(t, verify, brokenPairing)
}

func TestWithoutVcpkgRootFetchBootstrapsTheMapsVcpkgTool(t *testing.T) {
	s, repo, cache := buildableRelease(t)
	repo.Setenv("VCPKG_ROOT", "")
	repo.Setenv("AUTOMAP_VCPKG_TOOL_URL", "https://github.com/microsoft/vcpkg-tool/releases/download")

	repo.Automap(t, "fetch", "--json").ExpectExit(t, 0)

	tool := "vcpkg"
	if runtime.GOOS == "windows" {
		tool += ".exe"
	}
	// The fixture map names the vcpkg tool release 2026-09-26.
	if _, err := os.Stat(filepath.Join(cache, "vcpkg", "2026-09-26", tool)); err != nil {
		t.Fatalf("fetch didn't bootstrap vcpkg into the cache: %v", err)
	}
	s.Close()
	disconnect(repo)
	expectProbed(t, repo.Automap(t, "verify", "--json"))
}

func TestFetchAndVendorNeedAComposition(t *testing.T) {
	for _, cmd := range []string{"fetch", "vendor"} {
		run := clitest.Automap(t, "no-manifest", cmd, "--json")
		run.ExpectExit(t, 2)
		run.ExpectFailure(t, "NO_MANIFEST", "")
	}
}

// The testbed with SDL from vcpkg's curated registry and Jolt's double
// precision from the Automap registry, both git registries the build must
// not reach offline.
func TestTheTestbedVerifiesOfflineAfterFetch(t *testing.T) {
	requirePublicMapBuilds(t)
	repo := testbedRepo(t)
	repo.Setenv("VCPKG_BINARY_SOURCES", "clear") // build from the fetched sources
	repo.Automap(t, "start", "--with", "sdl3", "--with", "jolt:double-precision", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)
	if run := repo.Automap(t, "fetch", "--map", "map", "--json"); run.Exit != 0 {
		run.LogDetails(t)
		t.Fatalf("fetch: exit %d\n%s", run.Exit, run.Stdout)
	}

	disconnect(repo)
	run := repo.Automap(t, "verify", "--map", "map", "--json")

	if run.Exit != 0 || run.Result.Level != "probed" {
		logVcpkgErrors(t)
	}
	expectProbed(t, run)
}

func TestAVendoredTestbedVerifiesFromAFreshCloneOffline(t *testing.T) {
	requirePublicMapBuilds(t)
	repo := testbedRepo(t)
	repo.Automap(t, "start", "--with", "sdl3", "--with", "dear-imgui:docking", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)
	if run := repo.Automap(t, "vendor", "--map", "map", "--json"); run.Exit != 0 {
		run.LogDetails(t)
		t.Fatalf("vendor: exit %d\n%s", run.Exit, run.Stdout)
	}

	fresh := clone(t, repo)
	withoutCaches(t, fresh)
	disconnect(fresh)
	run := fresh.Automap(t, "verify", "--map", "map", "--json")

	if run.Exit != 0 || run.Result.Level != "probed" {
		logVcpkgErrors(t)
	}
	expectProbed(t, run)
}
