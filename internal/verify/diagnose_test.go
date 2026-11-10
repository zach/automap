package verify_test

import (
	"os"
	"strings"
	"testing"

	"github.com/zach/automap/internal/verify"
)

func readLog(t *testing.T, name string) string {
	t.Helper()
	raw, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// A build that failed because vcpkg couldn't download a source is the
// network's failure, with the URL: the agent needn't open the log.
func TestDiagnoseNamesTheDownloadThatFailed(t *testing.T) {
	d := verify.DiagnoseBuild(readLog(t, "build-download-403.log"))

	if d.URL != "https://github.com/facebook/zstd/archive/v1.5.7.tar.gz" {
		t.Errorf("URL %q, want the zstd archive vcpkg was downloading", d.URL)
	}
	if !strings.Contains(d.Excerpt, "curl operation failed with response code 403") {
		t.Errorf("excerpt %q, want vcpkg's first error", d.Excerpt)
	}
}

// Any other failed build is summed up by its first error and the lines
// that belong to it.
func TestDiagnoseQuotesTheFirstError(t *testing.T) {
	for _, c := range []struct {
		log  string
		want []string
	}{
		{"build-compile-error.log", []string{"game/app/main.cpp:42:5: error: use of undeclared identifier 'frame_count'"}},
		{"build-cmake-error.log", []string{"CMake Error at cmake/automap/stb-truetype.cmake:3 (find_package):", `Could not find a package configuration file provided by "Stb"`}},
	} {
		d := verify.DiagnoseBuild(readLog(t, c.log))
		if d.URL != "" {
			t.Errorf("%s: URL %q, want none: nothing failed to download", c.log, d.URL)
		}
		for _, want := range c.want {
			if !strings.Contains(d.Excerpt, want) {
				t.Errorf("%s: excerpt %q lacks %q", c.log, d.Excerpt, want)
			}
		}
		if n := strings.Count(d.Excerpt, "\n"); n > 5 || len(d.Excerpt) > 600 {
			t.Errorf("%s: excerpt is %d lines and %d bytes, want a few lines:\n%s", c.log, n+1, len(d.Excerpt), d.Excerpt)
		}
	}
	if d := verify.DiagnoseBuild("$ cmake -S .\n-- Configuring done\n"); d.Excerpt != "" || d.URL != "" {
		t.Errorf("a log with no error diagnosed as %+v", d)
	}
}
