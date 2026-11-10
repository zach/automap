package verify

import (
	"regexp"
	"strings"
)

// Diagnosis is what a failed build's log says went wrong.
type Diagnosis struct {
	// URL is what vcpkg failed to download, when the build failed because
	// a download did; "" otherwise.
	URL string
	// Excerpt is the log's first error and the lines that belong to it:
	// a few lines, for the failure's message.
	Excerpt string
}

var (
	// The first line of an error from a compiler, a linker, CMake, Ninja or
	// vcpkg.
	errorLine = regexp.MustCompile(`^(CMake Error\b|error:|fatal:|ninja: error:|LINK : fatal error|\S*\bld(\.lld|\.bfd|\.gold|64)?: |.*\b(fatal )?error( [A-Z]+\d+)?:)`)
	// vcpkg's download failures: its own downloader's, and git's when a
	// registry or a source repository can't be reached.
	downloadError = regexp.MustCompile(`(?i)^error: .*(download|curl|status code|WinHttp)|^fatal: unable to access|Download failed, halting portfile|Could not resolve host`)
	downloading   = regexp.MustCompile(`^(?:-- )?Downloading (https?://\S+)`)
	anyURL        = regexp.MustCompile(`https?://[^\s'"]+`)
)

// excerptLines and excerptBytes bound a diagnosis's excerpt.
const (
	excerptLines = 4
	excerptBytes = 500
)

// DiagnoseBuild reads the log of a build that failed. It is a pure
// function of the log.
func DiagnoseBuild(log string) Diagnosis {
	lines := strings.Split(strings.ReplaceAll(log, "\r\n", "\n"), "\n")
	var d Diagnosis
	lastDownload := ""
	for i, line := range lines {
		if m := downloading.FindStringSubmatch(line); m != nil {
			lastDownload = m[1]
		}
		if !errorLine.MatchString(line) {
			continue
		}
		if downloadError.MatchString(line) {
			d.URL = lastDownload
			if u := anyURL.FindString(line); u != "" && (d.URL == "" || strings.HasPrefix(line, "fatal:")) {
				d.URL = strings.TrimRight(u, ".:,")
			}
		}
		// The error, then the lines that continue it: indented ones, and
		// for CMake the message after its "at <file> (<command>):" line.
		excerpt := []string{strings.TrimSpace(line)}
		linker := strings.Contains(line, "ld: ") || strings.Contains(line, "ld.lld: ")
		for _, next := range lines[i+1:] {
			// A linker says where, then what, on lines of their own.
			continues := strings.HasPrefix(next, " ") || linker && !strings.Contains(next, "linker command failed") && !strings.HasPrefix(next, "ninja:")
			if len(excerpt) == excerptLines || strings.TrimSpace(next) == "" || !continues {
				break
			}
			excerpt = append(excerpt, strings.TrimSpace(next))
		}
		d.Excerpt = strings.Join(excerpt, "\n")
		if len(d.Excerpt) > excerptBytes {
			d.Excerpt = d.Excerpt[:excerptBytes] + "…"
		}
		break
	}
	// A download can fail after the log's first error line, or be the
	// cause the port's own error line doesn't name.
	if d.URL == "" {
		for _, line := range lines {
			if m := downloading.FindStringSubmatch(line); m != nil {
				lastDownload = m[1]
			}
			if downloadError.MatchString(line) {
				d.URL = lastDownload
				if u := anyURL.FindString(line); u != "" && d.URL == "" {
					d.URL = strings.TrimRight(u, ".:,")
				}
				if d.URL != "" {
					break
				}
			}
		}
	}
	return d
}
