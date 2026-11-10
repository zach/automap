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
	PrivateMaps map[string]struct {
		Commit      string
		ContentHash string `json:"content_hash"`
	} `json:"private_maps"`
	Pins map[string]struct {
		Variant     string
		Path        string
		ContentHash string `json:"content_hash"`
		Registry    *struct {
			Kind, ID, Commit string
			ContentHash      string `json:"content_hash"`
		}
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

// The lock pins a private map by its content, with the commit that last
// changed it beside that, a private variant at that map, and an in-repo
// library by path and content hash (SPEC §4.2).
func TestPrivateContentIsPinned(t *testing.T) {
	repo := acmeStubsRepo(t, newReleaseServer(t), fixtureMapDir(t))

	run := repo.Automap(t, "extend", "acme-tools", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	lock := readLock(t, repo)
	commit := gitOutput(t, repo.Dir, "log", "-1", "--format=%H", "--", "automap-private")
	acme := lock.PrivateMaps["acme"]
	if acme.Commit != commit || !strings.HasPrefix(acme.ContentHash, "sha256:") {
		t.Errorf("private map acme pinned as %+v, want its content hash, and its last commit %s beside it", acme, commit)
	}
	ui := lock.Pins["stub-ui"]
	if ui.Variant != "acme-patched" || ui.Registry == nil || ui.Registry.Kind != "private" || ui.Registry.ID != "acme" || ui.Registry.ContentHash != acme.ContentHash || ui.Registry.Commit != commit {
		t.Errorf("stub-ui pinned as %+v, want acme-patched from private map acme at its content hash %s", ui, acme.ContentHash)
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

// addSharedPortLibrary adds acme-fonts to Acme's private map: a second
// library from the stub-ui port, at version, with the port's fonts
// feature, filling a private text slot.
func addSharedPortLibrary(t *testing.T, repo *clitest.Repo, version string) {
	t.Helper()
	repo.WriteFile(t, "automap-private/slots/text.json", `{"id": "text", "description": "Draws text from fonts."}`)
	repo.WriteFile(t, "automap-private/libraries/acme-fonts.json", `{
  "id": "acme-fonts",
  "name": "Acme Fonts",
  "license": {"spdx": "MIT-0", "class": "permissive"},
  "contribution_policy": {"ai_generated_contributions": "not-applicable", "source": "In-house."},
  "slots": ["text"],
  "source": {"kind": "git", "url": "https://github.com/zach/automap/tree/main/testdata/map/ports/stub-ui"},
  "versions": [{
    "version": "`+version+`",
    "released": "2026-01-15",
    "pin": {"commit": "2222222222222222222222222222222222222222", "content_hash": "sha512:00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002"},
    "build_backend": {"vcpkg": {"port": "stub-ui", "registry": "automap"}},
    "cmake": {"package": "StubUI", "targets": ["StubUI::StubUI"]},
    "variants": [{"name": "default", "build_options": {"vcpkg_features": "fonts"}}],
    "facts": {
      "lifecycle": {"init_steps": [], "shutdown_steps": [], "reloadable": true},
      "build-flags": {"abi_defines": [], "self_check": null, "min_language_standard": "C99", "exceptions": "never-thrown", "rtti": "no-constraint"},
      "platform": {"targets": ["windows-x64-msvc", "linux-x64-clang", "macos-arm64-appleclang"], "headless": "", "web": false}
    }
  }]
}`)
	repo.WriteFile(t, "automap-private/recipes/library/acme-fonts.json", `{
  "kind": "library",
  "subject": {"library": "acme-fonts"},
  "applies_to": {"acme-fonts": {"versions": ["`+version+`"], "variants": ["default"]}},
  "reference_glue": [{"path": "acme-fonts/acme-fonts.cmake", "role": "build-fragment"}],
  "probes": [],
  "adaptation_notes": ["acme-fonts is the fonts header of the stub-ui port."]
}`)
	repo.WriteFile(t, "automap-private/recipes/library/acme-fonts/acme-fonts.cmake", "# acme-fonts ships in the stub-ui port, which stub-ui's fragment finds.\n")
	commitRepo(t, repo)
}

// sharedPortRepo is Acme's repo with the public stub-ui in the composition
// and acme-fonts, from the same port at version, in its private map.
func sharedPortRepo(t *testing.T, version string) *clitest.Repo {
	t.Helper()
	repo := acmeStubsRepo(t, newReleaseServer(t), fixtureMapDir(t))
	repo.EditJSON(t, "automap.manifest.json", set("variants/stub-ui", "default"))
	addSharedPortLibrary(t, repo, version)
	repo.Automap(t, "upgrade", "stub-ui", "--no-verify", "--json").ExpectExit(t, 0)
	return repo
}

// Two libraries that come from one vcpkg port share it: vcpkg.json names
// the port once, with the union of their features and one override, and
// vcpkg-configuration.json names it once among the registry's packages
// (SPEC §5.5).
func TestLibrariesThatShareAPortShareItsEntry(t *testing.T) {
	repo := sharedPortRepo(t, "1.0.0")

	run := repo.Automap(t, "extend", "acme-fonts", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	var manifest struct {
		Dependencies []struct {
			Name     string
			Features []any
		}
		Overrides []struct{ Name, Version string }
	}
	json.Unmarshal([]byte(clitest.Read(t, repo, "vcpkg.json")), &manifest)
	var deps, overrides int
	for _, d := range manifest.Dependencies {
		if d.Name == "stub-ui" {
			deps++
			if len(d.Features) != 1 || d.Features[0] != "fonts" {
				t.Errorf("stub-ui's features %v, want the union of its libraries': [fonts]", d.Features)
			}
		}
	}
	for _, o := range manifest.Overrides {
		if o.Name == "stub-ui" {
			overrides++
		}
	}
	if deps != 1 || overrides != 1 {
		t.Errorf("vcpkg.json names stub-ui in %d dependencies and %d overrides, want one of each:\n%s", deps, overrides, clitest.Read(t, repo, "vcpkg.json"))
	}
	var config struct {
		Registries []struct{ Packages []string }
	}
	json.Unmarshal([]byte(clitest.Read(t, repo, "vcpkg-configuration.json")), &config)
	for _, reg := range config.Registries {
		seen := map[string]int{}
		for _, p := range reg.Packages {
			if seen[p]++; seen[p] > 1 {
				t.Errorf("vcpkg-configuration.json names package %s twice: %v", p, reg.Packages)
			}
		}
	}
}

// Two libraries can't pin one port at two versions: vcpkg builds a port
// once.
func TestLibrariesThatPinOnePortAtTwoVersionsConflict(t *testing.T) {
	repo := sharedPortRepo(t, "2.0.0")

	run := repo.Automap(t, "extend", "acme-fonts", "--no-verify", "--json")

	run.ExpectExit(t, 2)
	run.ExpectFailure(t, "PORT_CONFLICT(stub-ui)", "acme-fonts")
	run.ExpectUnchanged(t)
}

// A lock written on Linux checks on a Windows checkout, where git has
// turned the in-repo library's line endings into CRLF: the pin hashes text
// with its line endings normalized.
func TestAnInRepoPinSurvivesLineEndingConversion(t *testing.T) {
	repo := acmeStubsRepo(t, newReleaseServer(t), fixtureMapDir(t))
	repo.Automap(t, "extend", "acme-tools", "--no-verify", "--json").ExpectExit(t, 0)
	pinned := readLock(t, repo).Pins["acme-tools"].ContentHash
	for _, f := range []string{"engine/acme-tools/CMakeLists.txt", "engine/acme-tools/acme_tools.cpp", "engine/acme-tools/include/acme_tools.h"} {
		repo.WriteFile(t, f, strings.ReplaceAll(clitest.Read(t, repo, f), "\n", "\r\n"))
	}

	run := repo.Automap(t, "extend", "stub-physics", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	if got := readLock(t, repo).Pins["acme-tools"].ContentHash; got != pinned {
		t.Errorf("acme-tools pinned at %s after its line endings changed, want %s as before", got, pinned)
	}
}

// A change to an in-repo library's content still fails the lock.
func TestAnInRepoPinFailsWhenItsSourceChanges(t *testing.T) {
	repo := acmeStubsRepo(t, newReleaseServer(t), fixtureMapDir(t))
	repo.Automap(t, "extend", "acme-tools", "--no-verify", "--json").ExpectExit(t, 0)
	repo.WriteFile(t, "engine/acme-tools/acme_tools.cpp", clitest.Read(t, repo, "engine/acme-tools/acme_tools.cpp")+"// changed\n")

	run := repo.Automap(t, "extend", "stub-physics", "--no-verify", "--json")

	run.ExpectExit(t, 2)
	run.ExpectFailure(t, "LOCK_MISMATCH(acme-tools)", "has changed since the lock pinned it")
}

const (
	acmeToolsReference = "automap-private/recipes/library/acme-tools/acme-tools.cmake"
	acmeToolsFragment  = "cmake/automap/acme-tools.cmake"
)

// acmeToolsRepo is Acme's repo with acme-tools in the composition and
// committed, and then the private recipe's build fragment changed and
// committed: what upgrade has to carry into the repo.
func acmeToolsRepo(t *testing.T, edit func(repo *clitest.Repo)) *clitest.Repo {
	t.Helper()
	repo := acmeStubsRepo(t, newReleaseServer(t), fixtureMapDir(t))
	repo.Automap(t, "extend", "acme-tools", "--no-verify", "--json").ExpectExit(t, 0)
	commitRepo(t, repo)
	if edit != nil {
		edit(repo)
	}
	repo.WriteFile(t, acmeToolsReference, clitest.Read(t, repo, acmeToolsReference)+"# Changed in the private map.\n")
	commitRepo(t, repo)
	return repo
}

// upgrade compares against each private map as the lock pinned it: a
// private recipe's changed reference reaches the repo's copy that still
// matched the old one (SPEC §5.4).
func TestUpgradeCarriesAPrivateMapsChangedGlue(t *testing.T) {
	repo := acmeToolsRepo(t, nil)

	run := repo.Automap(t, "upgrade", "acme-tools", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	if got, want := clitest.Read(t, repo, acmeToolsFragment), clitest.Read(t, repo, acmeToolsReference); got != want {
		t.Errorf("%s is\n%s\nwant the private map's new reference\n%s", acmeToolsFragment, got, want)
	}
	if !strings.Contains(run.Result.Summary, "1 file updated, 0 diverged") {
		t.Errorf("summary %q, want 1 file updated, 0 diverged", run.Result.Summary)
	}
	if commit := gitOutput(t, repo.Dir, "log", "-1", "--format=%H", "--", "automap-private"); readLock(t, repo).PrivateMaps["acme"].Commit != commit {
		t.Errorf("the lock doesn't pin private map acme at its new commit %s", commit)
	}
}

// A copy the developer changed is left alone when the private reference
// changes too, with the new reference beside it.
func TestUpgradeLeavesGlueThatDivergedFromAPrivateMap(t *testing.T) {
	repo := acmeToolsRepo(t, func(repo *clitest.Repo) {
		repo.WriteFile(t, acmeToolsFragment, clitest.Read(t, repo, acmeToolsFragment)+"# Mine.\n")
	})
	mine := clitest.Read(t, repo, acmeToolsFragment)

	run := repo.Automap(t, "upgrade", "acme-tools", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	if got := clitest.Read(t, repo, acmeToolsFragment); got != mine {
		t.Errorf("%s was rewritten:\n%s", acmeToolsFragment, got)
	}
	if got, want := clitest.Read(t, repo, acmeToolsFragment+".reference"), clitest.Read(t, repo, acmeToolsReference); got != want {
		t.Errorf("%s.reference is\n%s\nwant the private map's new reference", acmeToolsFragment, got)
	}
	var warned bool
	for _, w := range run.Result.Warnings {
		warned = warned || w.Code == "GLUE_DIVERGED("+acmeToolsFragment+")"
	}
	if !warned {
		t.Errorf("no GLUE_DIVERGED(%s) warning in %+v", acmeToolsFragment, run.Result.Warnings)
	}
}

// A pin that differs from the map is one failure, however many platforms
// the composition has.
func TestALockMismatchIsReportedOnce(t *testing.T) {
	repo := acmeToolsRepo(t, nil)

	run := repo.Automap(t, "extend", "stub-physics", "--no-verify", "--json")

	run.ExpectExit(t, 2)
	if len(run.Result.Failures) != 1 {
		t.Errorf("failures %+v, want the one changed private map, once", run.Result.Failures)
	}
	run.ExpectFailure(t, "LOCK_MISMATCH(private map acme)", "has changed since the lock pinned it")
}

// In a shallow clone, as CI checks a repo out, the commit the lock pins a
// private map at isn't there to read: upgrade compares against the map as
// it is, and still upgrades.
func TestUpgradeInAShallowCloneReadsThePrivateMapAsItIs(t *testing.T) {
	s := newReleaseServer(t)
	repo := acmeStubsRepo(t, s, fixtureMapDir(t))
	repo.Automap(t, "extend", "acme-tools", "--no-verify", "--json").ExpectExit(t, 0)
	commitRepo(t, repo)
	repo.WriteFile(t, acmeToolsReference, clitest.Read(t, repo, acmeToolsReference)+"# Changed in the private map.\n")
	commitRepo(t, repo)
	shallow := &clitest.Repo{Dir: t.TempDir()}
	if out, err := exec.Command("git", "clone", "-q", "--depth", "1", "file://"+repo.Dir, shallow.Dir).CombinedOutput(); err != nil {
		t.Fatalf("git clone: %v\n%s", err, out)
	}
	shallow.Setenv("AUTOMAP_RELEASES", s.URL)
	shallow.Setenv("AUTOMAP_EVIDENCE_URL", s.URL+"/main")
	shallow.Setenv("AUTOMAP_CACHE", t.TempDir())

	run := shallow.Automap(t, "upgrade", "acme-tools", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	if !strings.Contains(run.Result.Summary, "0 files updated, 0 diverged") {
		t.Errorf("summary %q, want 0 files updated, 0 diverged: the old reference can't be read", run.Result.Summary)
	}
}

// find searches every map layer: a private map's slot and its library are
// found by the slot's id and by its description.
func TestFindSearchesPrivateSlots(t *testing.T) {
	repo := sharedPortRepo(t, "1.0.0")
	for _, query := range []string{"text", "font"} {
		run := repo.Automap(t, "find", query, "--json")

		run.ExpectExit(t, 0)
		if len(run.Result.Entries) != 1 || run.Result.Entries[0]["id"] != "acme-fonts" {
			t.Errorf("find %s: entries %v, want acme-fonts, which fills the private text slot", query, run.Result.Entries)
		}
	}
}

// A private map needs no commit to be composed from: its content is what
// the lock pins. The result says the pin has no commit behind it.
func TestAnUncommittedPrivateMapComposesWithAWarning(t *testing.T) {
	s := newReleaseServer(t)
	s.publish(t, "map-2026.10.01", fixtureMapDir(t), nil)
	repo := clitest.NewRepo(t, "acme-stubs")
	repo.Setenv("AUTOMAP_RELEASES", s.URL)
	repo.Setenv("AUTOMAP_EVIDENCE_URL", s.URL+"/main")
	repo.Setenv("AUTOMAP_CACHE", t.TempDir())
	repo.Automap(t, "start", "--with", "stub-window", "--with", "stub-ui", "--no-verify", "--json").ExpectExit(t, 0)
	repo.EditJSON(t, "automap.manifest.json", set("private_maps", []any{map[string]any{"id": "acme", "location": "automap-private"}}))

	run := repo.Automap(t, "extend", "acme-tools", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	var warned bool
	for _, w := range run.Result.Warnings {
		warned = warned || w.Code == "PRIVATE_MAP_UNCOMMITTED(acme)"
	}
	if !warned {
		t.Errorf("no PRIVATE_MAP_UNCOMMITTED(acme) warning in %+v", run.Result.Warnings)
	}
	if acme := readLock(t, repo).PrivateMaps["acme"]; acme.Commit != "" || !strings.HasPrefix(acme.ContentHash, "sha256:") {
		t.Errorf("private map acme pinned as %+v, want its content hash and no commit", acme)
	}
}

// A change to a private map, committed or not, is a change to what the
// lock pins: commands that check the lock say so, and resolve re-pins it.
func TestAChangedPrivateMapNeedsResolve(t *testing.T) {
	repo := acmeStubsRepo(t, newReleaseServer(t), fixtureMapDir(t))
	repo.Automap(t, "extend", "acme-tools", "--no-verify", "--json").ExpectExit(t, 0)
	repo.WriteFile(t, acmeToolsReference, clitest.Read(t, repo, acmeToolsReference)+"# Changed, not committed.\n")

	run := repo.Automap(t, "extend", "stub-physics", "--no-verify", "--json")

	run.ExpectExit(t, 2)
	run.ExpectFailure(t, "LOCK_MISMATCH(private map acme)", "automap resolve")

	run = repo.Automap(t, "resolve", "--no-verify", "--json")
	run.ExpectExit(t, 0)
	repo.Automap(t, "extend", "stub-physics", "--no-verify", "--json").ExpectExit(t, 0)
}

// A lock checks the same in a shallow clone, as CI checks a repo out, as in
// a full one: a private map's content doesn't depend on git's history.
func TestALockWithAPrivateVariantChecksInAShallowClone(t *testing.T) {
	s := newReleaseServer(t)
	repo := acmeStubsRepo(t, s, fixtureMapDir(t))
	commitRepo(t, repo)
	repo.WriteFile(t, "README.md", "More.\n")
	commitRepo(t, repo)
	shallow := &clitest.Repo{Dir: t.TempDir()}
	if out, err := exec.Command("git", "clone", "-q", "--depth", "1", "file://"+repo.Dir, shallow.Dir).CombinedOutput(); err != nil {
		t.Fatalf("git clone: %v\n%s", err, out)
	}
	shallow.Setenv("AUTOMAP_RELEASES", s.URL)
	shallow.Setenv("AUTOMAP_EVIDENCE_URL", s.URL+"/main")
	shallow.Setenv("AUTOMAP_CACHE", t.TempDir())

	run := shallow.Automap(t, "extend", "stub-physics", "--no-verify", "--json")

	run.ExpectExit(t, 0)
}
