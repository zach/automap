package main_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/zach/automap/internal/clitest"
)

// buildWithoutAutomap configures and builds the repo as a plain CMake and
// vcpkg project, never running the CLI, and runs its executable for a smoke
// test's frames with env. It reports failure with the output.
func buildWithoutAutomap(t *testing.T, repo *clitest.Repo, target string, env ...string) {
	t.Helper()
	build := filepath.Join(repo.Dir, "build", "plain")
	configure := []string{"-S", repo.Dir, "-B", build, "-G", "Ninja", "-DCMAKE_BUILD_TYPE=Debug",
		"-DCMAKE_TOOLCHAIN_FILE=" + filepath.Join(os.Getenv("VCPKG_ROOT"), "scripts", "buildsystems", "vcpkg.cmake")}
	switch runtime.GOOS {
	case "linux":
		configure = append(configure, "-DCMAKE_C_COMPILER=clang", "-DCMAKE_CXX_COMPILER=clang++")
	case "windows":
		configure = append(configure, "-DCMAKE_C_COMPILER=cl", "-DCMAKE_CXX_COMPILER=cl")
	}
	for _, args := range [][]string{configure, {"--build", build, "--target", target}} {
		if out, err := exec.Command("cmake", args...).CombinedOutput(); err != nil {
			logVcpkgErrors(t)
			t.Fatalf("cmake %s: %v\n%s", args[0], err, tail(out))
		}
	}
	exe := filepath.Join(build, target)
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	cmd := exec.Command(exe, "--frames", "120")
	cmd.Env = append(os.Environ(), env...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("the smoke run failed: %v\n%s", err, tail(out))
	}
}

func tail(out []byte) string {
	lines := strings.Split(string(out), "\n")
	if len(lines) > 60 {
		lines = lines[len(lines)-60:]
	}
	return strings.Join(lines, "\n")
}

func TestAnEjectedCompositionBuildsWithoutAutomap(t *testing.T) {
	repo := buildableRepo(t)
	repo.Automap(t, startTestbed...).ExpectExit(t, 0)
	before := map[string][]byte{}
	filepath.WalkDir(repo.Dir, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			rel, _ := filepath.Rel(repo.Dir, p)
			before[filepath.ToSlash(rel)], _ = os.ReadFile(p)
		}
		return nil
	})

	run := repo.Automap(t, "eject", "--map", "map", "--json")

	run.ExpectExit(t, 0)
	for _, f := range []string{"automap.manifest.json", "automap.lock.json"} {
		if _, err := os.Stat(filepath.Join(repo.Dir, f)); err == nil {
			t.Errorf("eject left %s", f)
		}
	}
	// The Automap registry's ports are in the repo, and the vcpkg
	// configuration names no Automap registry.
	for _, port := range []string{"stub-ui", "stub-window"} {
		if _, err := os.Stat(filepath.Join(repo.Dir, "ports", port, "portfile.cmake")); err != nil {
			t.Errorf("eject didn't vendor the %s port: %v", port, err)
		}
	}
	var config struct {
		Registries   []map[string]any `json:"registries"`
		OverlayPorts []string         `json:"overlay-ports"`
	}
	json.Unmarshal([]byte(clitest.Read(t, repo, "vcpkg-configuration.json")), &config)
	if len(config.Registries) != 0 || strings.Join(config.OverlayPorts, ",") != "ports/stub-ui,ports/stub-window" {
		t.Errorf("vcpkg-configuration.json has registries %v and overlay ports %v; want none, and the vendored ports", config.Registries, config.OverlayPorts)
	}
	// Glue, shared interfaces and build files stay as they were.
	for p, data := range before {
		switch {
		case p == "automap.manifest.json", p == "automap.lock.json", p == "vcpkg-configuration.json",
			strings.HasPrefix(p, "map/"), strings.HasPrefix(p, ".automap-test-cache/"):
			continue
		}
		after, err := os.ReadFile(filepath.Join(repo.Dir, filepath.FromSlash(p)))
		if err != nil || !bytes.Equal(after, data) {
			t.Errorf("eject changed %s", p)
		}
	}

	// Automap is gone: the map and its registry with it.
	os.RemoveAll(filepath.Join(repo.Dir, "map"))
	buildWithoutAutomap(t, repo, "testbed", "STUB_WINDOW_HEADLESS=1")
}

func TestEjectRefusesAHandEditedLock(t *testing.T) {
	repo := fixtureRepo(t)
	repo.Automap(t, startTestbed...).ExpectExit(t, 0)
	repo.EditJSON(t, "automap.lock.json", set("pins/stub-ui/version", "9.9.9"))

	run := repo.Automap(t, "eject", "--map", "map", "--json")

	run.ExpectExit(t, 3)
	run.ExpectFailure(t, "LOCK_EDITED", "")
	if _, err := os.Stat(filepath.Join(repo.Dir, "automap.manifest.json")); err != nil {
		t.Error("eject removed the manifest it refused to act on")
	}
}

func TestEjectNeedsAComposition(t *testing.T) {
	run := clitest.Automap(t, "no-manifest", "eject", "--json")

	run.ExpectExit(t, 2)
	run.ExpectFailure(t, "NO_MANIFEST", "")
}

// The reference compositions, and Jolt's double precision from the
// Automap registry, eject to plain CMake and vcpkg projects that build and
// pass their smoke test without Automap.
func TestTheReferenceCompositionsBuildWithoutAutomapOnceEjected(t *testing.T) {
	requirePublicMapBuilds(t)
	headless := []string{}
	if runtime.GOOS == "linux" {
		headless = append(headless, "SDL_VIDEO_DRIVER=offscreen")
	}
	for name, with := range map[string][]string{
		"testbed":             {"sdl3", "dear-imgui:docking"},
		"testbed-jolt":        {"sdl3", "dear-imgui:docking", "jolt"},
		"testbed-jolt-double": {"sdl3", "dear-imgui:docking", "jolt:double-precision"},
	} {
		t.Run(name, func(t *testing.T) {
			repo := testbedRepo(t)
			args := []string{"start", "--map", "map", "--no-verify", "--json"}
			for _, w := range with {
				args = append(args, "--with", w)
			}
			repo.Automap(t, args...).ExpectExit(t, 0)
			if strings.Contains(strings.Join(with, ","), "double-precision") {
				repo.EditJSON(t, "automap.manifest.json", set("choices/math-types/position_precision", "double"))
			}

			repo.Automap(t, "eject", "--map", "map", "--json").ExpectExit(t, 0)

			os.RemoveAll(filepath.Join(repo.Dir, "map"))
			buildWithoutAutomap(t, repo, "testbed", headless...)
		})
	}
}
