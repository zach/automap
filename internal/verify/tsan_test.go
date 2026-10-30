package verify_test

import (
	"os"
	"strings"
	"testing"

	"github.com/zach/automap/internal/verify"
)

// The samples are real ThreadSanitizer output (GCC 13, Linux x64): a race
// in the executable's own code, and a race wholly inside an uninstrumented
// shared library, libnoise.so, through the memset interceptor.
func sample(t *testing.T, name string) string {
	t.Helper()
	raw, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestARaceInTheExecutableCounts(t *testing.T) {
	reports := verify.ParseTSan(sample(t, "tsan-glue-race.txt"), "tsan")

	if len(reports) == 0 {
		t.Fatal("no reports")
	}
	for _, r := range reports {
		if !r.Counts {
			t.Errorf("report %+v doesn't count, but its frames are in the executable", r)
		}
	}
	if !strings.HasPrefix(reports[0].Summary, "data race tsan.cpp:3") {
		t.Errorf("summary %q, want it to start data race tsan.cpp:3", reports[0].Summary)
	}
}

func TestARaceWhollyInASharedLibraryIsSetAside(t *testing.T) {
	reports := verify.ParseTSan(sample(t, "tsan-shared-library-race.txt"), "nmain")

	if len(reports) != 1 {
		t.Fatalf("%d reports, want 1", len(reports))
	}
	r := reports[0]
	if r.Counts {
		t.Errorf("report %+v counts, but every access frame is in libnoise.so or the sanitizer", r)
	}
	if !strings.Contains(strings.Join(r.Modules, ","), "libnoise.so") {
		t.Errorf("modules %v lack libnoise.so", r.Modules)
	}
}

func TestOutputWithoutReports(t *testing.T) {
	if got := verify.ParseTSan("frame 1\nframe 2\n", "testbed"); len(got) != 0 {
		t.Errorf("%d reports in clean output", len(got))
	}
}

// Clang links its ThreadSanitizer runtime into the executable without
// debug information, so an interceptor's frame is in the executable but has
// no source location; the thread that created the racing thread was main,
// in the executable, but creation stacks don't count. The sample has the
// frame layout of clang 21's output in Automap's CI.
func TestClangsRuntimeFramesAndThreadCreationDontCount(t *testing.T) {
	reports := verify.ParseTSan(sample(t, "tsan-clang-runtime-frames.txt"), "testbed")

	if len(reports) != 1 {
		t.Fatalf("%d reports, want 1", len(reports))
	}
	if reports[0].Counts {
		t.Errorf("report %+v counts, but its only executable frames are the sanitizer's and the thread creation's", reports[0])
	}
	if reports[0].Summary != "data race (testbed+0x638e1) in memset" {
		t.Errorf("summary %q", reports[0].Summary)
	}
}

func TestTSanTriplet(t *testing.T) {
	name, contents := verify.TSanTriplet("linux-x64-clang")
	if name != "x64-linux-automap-tsan" {
		t.Errorf("triplet %q, want x64-linux-automap-tsan", name)
	}
	for _, want := range []string{
		"set(VCPKG_TARGET_ARCHITECTURE x64)",
		"set(VCPKG_CMAKE_SYSTEM_NAME Linux)",
		"set(VCPKG_LIBRARY_LINKAGE static)",
		`set(VCPKG_C_FLAGS "-fsanitize=thread")`,
		`set(VCPKG_CXX_FLAGS "-fsanitize=thread")`,
		`set(VCPKG_LINKER_FLAGS "-fsanitize=thread")`,
	} {
		if !strings.Contains(contents, want) {
			t.Errorf("triplet lacks %s:\n%s", want, contents)
		}
	}
	if name, contents := verify.TSanTriplet("macos-arm64-appleclang"); name != "arm64-osx-automap-tsan" || !strings.Contains(contents, "set(VCPKG_CMAKE_SYSTEM_NAME Darwin)") {
		t.Errorf("macOS triplet %q:\n%s", name, contents)
	}
}
