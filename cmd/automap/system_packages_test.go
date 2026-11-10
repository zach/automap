package main_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/zach/automap/internal/clitest"
)

// packagesRepo is a started testbed whose stub-window needs system
// packages on this platform, on a machine whose dpkg-query knows only
// those installed.
func packagesRepo(t *testing.T, installed ...string) *clitest.Repo {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("system packages are checked through dpkg-query, on Linux")
	}
	repo := fixtureRepo(t)
	repo.EditJSON(t, "map/libraries/stub-window.json", set("versions/0/facts/platform/system_packages", map[string]any{
		hostPlatform: map[string]any{"apt": []any{"libstub-dev", "stub-protocols"}, "apt_headless": []any{"stub-software-driver"}},
	}))
	repo.Automap(t, startTestbed...).ExpectExit(t, 0)
	// A dpkg-query that lists the installed packages, as the real one does
	// for the format the CLI asks for; one package provides another.
	bin := filepath.Join(repo.Dir, ".fake-bin")
	script := "#!/bin/sh\n"
	for _, p := range installed {
		name, provides, _ := strings.Cut(p, "=")
		script += "printf 'ii \\t" + name + "\\t" + provides + "\\n'\n"
	}
	repo.WriteFile(t, ".fake-bin/dpkg-query", script)
	os.Chmod(filepath.Join(bin, "dpkg-query"), 0o755)
	repo.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return repo
}

// verify names a system package a library needs and the machine lacks,
// before building: the alternative is a port build that fails minutes in.
func TestVerifyNamesAMissingSystemPackage(t *testing.T) {
	repo := packagesRepo(t, "libstub-dev", "stub-software-driver")

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	run.ExpectExit(t, 1)
	run.ExpectFailure(t, "SYSTEM_PACKAGE_MISSING(stub-protocols)", "sudo apt-get install -y stub-protocols")
	if len(run.Result.Failures) != 1 || run.Result.Level != "none" {
		t.Errorf("failures %v at level %q, want the one missing package at level none: nothing was built", run.Result.Failures, run.Result.Level)
	}
	if _, err := os.Stat(filepath.Join(repo.Dir, "build", "automap", "build-"+hostPlatform+".log")); err == nil {
		t.Error("verify built before reporting the missing package")
	}
}

// A package another installed package provides counts as installed.
func TestAProvidedSystemPackageCounts(t *testing.T) {
	repo := packagesRepo(t, "libstub-dev", "stub-protocols-ng=stub-protocols (= 2.0), other", "stub-software-driver")
	repo.Setenv("VCPKG_ROOT", "")
	repo.Setenv("AUTOMAP_VCPKG_TOOL_URL", "http://127.0.0.1:1")

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	for _, f := range run.Result.Failures {
		if strings.HasPrefix(f.Code, "SYSTEM_PACKAGE_MISSING") {
			t.Errorf("%s: %s", f.Code, f.Message)
		}
	}
}

// A package only a headless run needs, such as a software GPU driver, is a
// warning: a machine with a real driver doesn't need it.
func TestAMissingHeadlessPackageIsAWarning(t *testing.T) {
	repo := packagesRepo(t, "libstub-dev", "stub-protocols")
	repo.Setenv("VCPKG_ROOT", "")
	repo.Setenv("AUTOMAP_VCPKG_TOOL_URL", "http://127.0.0.1:1")

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	var warned bool
	for _, w := range run.Result.Warnings {
		warned = warned || w.Code == "SYSTEM_PACKAGE_MISSING(stub-software-driver)"
	}
	if !warned {
		t.Errorf("no SYSTEM_PACKAGE_MISSING(stub-software-driver) warning in %+v", run.Result.Warnings)
	}
	for _, f := range run.Result.Failures {
		if strings.HasPrefix(f.Code, "SYSTEM_PACKAGE_MISSING") {
			t.Errorf("a headless package failed verify: %s", f.Code)
		}
	}
}

// show library carries the packages, for a workflow to install.
func TestShowLibraryCarriesItsSystemPackages(t *testing.T) {
	run := clitest.Automap(t, "no-manifest", "show", "library", "sdl3", "--map", filepath.Join(clitest.Root(t), "map"), "--json")

	run.ExpectExit(t, 0)
	apt, _ := run.Result.Entries[0].Get("definition", "versions", 0, "facts", "platform", "system_packages", "linux-x64-clang", "apt").([]any)
	if len(apt) < 10 {
		t.Errorf("sdl3's apt packages on linux-x64-clang are %v, want its X11, Wayland and autotools packages", apt)
	}
}
