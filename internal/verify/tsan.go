package verify

import (
	"path"
	"regexp"
	"strings"
)

// TSanRule is the rule the ThreadSanitizer run checks (ADR-0073).
const TSanRule = "jobs.tsan-clean"

// TSanReport is one ThreadSanitizer report from a run.
type TSanReport struct {
	Summary string // such as "data race jobs.cpp:42 in run"
	// Counts is whether a frame in either stack is in code the composition
	// builds, which makes the report fail the rule.
	Counts bool
	// Modules lists the modules the report's frames are in.
	Modules []string
}

var (
	tsanFrame   = regexp.MustCompile(`^\s*#\d+ (\S+)(?: (.*?))? \(([^()+]+)\+0x[0-9a-f]+\)`)
	tsanBuildID = regexp.MustCompile(`\s*\(BuildId: [0-9a-f]+\)`)
	sourceLine  = regexp.MustCompile(`:\d+(:\d+)?$`)
	tsanSummary = regexp.MustCompile(`^SUMMARY: ThreadSanitizer: (.*)$`)
	tsanSection = regexp.MustCompile(`^  \S`)
	tsanAccess  = regexp.MustCompile(`^  (Previous )?(atomic )?(Atomic )?(write|read|Write|Read) of size`)
	// sanitizerRuntime marks frames in ThreadSanitizer's own runtime, which
	// is linked into the executable but isn't the composition's code.
	sanitizerRuntime = regexp.MustCompile(`compiler-rt/|sanitizer_common|/tsan/rtl/|^__(tsan|interceptor|sanitizer)`)
)

// ParseTSan reads every report in a ThreadSanitizer run's output. Only the
// frames of the racing accesses' stacks count, not those of the stacks that
// created the threads. A frame is in code the composition builds when it is in the executable itself,
// where the composition's glue and its statically linked libraries are,
// and not in the sanitizer's runtime; frames in shared libraries, such as
// a GPU driver, are not. It is a pure function.
func ParseTSan(output, executable string) []TSanReport {
	var reports []TSanReport
	var cur *TSanReport
	access := false // whether the current stack is one of the report's access stacks
	for _, line := range strings.Split(output, "\n") {
		switch {
		case strings.Contains(line, "WARNING: ThreadSanitizer:"):
			reports = append(reports, TSanReport{})
			cur = &reports[len(reports)-1]
			access = true
		case cur == nil:
		case tsanSummary.MatchString(line):
			cur.Summary = shortenSummary(tsanSummary.FindStringSubmatch(line)[1])
			cur = nil
		case tsanSection.MatchString(line):
			// "Write of size ...", "Previous read of size ...": the two
			// stacks of the racing accesses. "Location is ...", "Thread T1
			// ... created by ...", "Mutex ...": where they live, not the race.
			access = tsanAccess.MatchString(line)
		case !access:
		default:
			m := tsanFrame.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			fn, file, module := m[1], m[2], m[3]
			if !contains(cur.Modules, module) {
				cur.Modules = append(cur.Modules, module)
			}
			if module == executable && sourceLine.MatchString(file) && !sanitizerRuntime.MatchString(file) && !sanitizerRuntime.MatchString(fn) {
				cur.Counts = true
			}
		}
	}
	return reports
}

// shortenSummary keeps a summary's file name without its directory, such as
// "data race jobs.cpp:42 in run".
func shortenSummary(s string) string {
	fields := strings.Fields(tsanBuildID.ReplaceAllString(s, ""))
	for i, f := range fields {
		if strings.Contains(f, "/") {
			fields[i] = path.Base(f)
		}
	}
	return strings.Join(fields, " ")
}

func contains(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}
