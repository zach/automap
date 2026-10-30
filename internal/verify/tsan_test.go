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
