package main

import (
	"slices"
	"testing"
	"time"
)

func TestTagPatternsComeFromThePin(t *testing.T) {
	for _, c := range []struct{ tag, version, upstream, want string }{
		{"release-3.4.18", "3.4.18", "release-3.4.20", "3.4.20"},
		{"v1.93.0-docking", "1.93.0", "v1.94.0-docking", "1.94.0"},
		{"v1.93.0-docking", "1.93.0", "v1.94.0", ""},
		{"v5.6.0", "5.6.0", "v5.7.0", "5.7.0"},
		{"v5.6.0", "5.6.0", "v5.7.0-rc1", ""},
		{"v5.6.0", "5.6.0", "nightly", ""},
	} {
		if got := versionOf(c.tag, c.version, c.upstream); got != c.want {
			t.Errorf("pin %s (%s), upstream %s: version %q, want %q", c.tag, c.version, c.upstream, got, c.want)
		}
	}
}

func TestOnlyVersionsNewerThanTheMapsNewestAreNew(t *testing.T) {
	got := newer("5.6.0", []string{"5.5.0", "5.6.0", "5.6.1", "5.10.0", "6.0.0"})
	if want := []string{"5.6.1", "5.10.0", "6.0.0"}; !slices.Equal(got, want) {
		t.Errorf("new versions %v, want %v", got, want)
	}
}

func TestReleaseNotesThatNameASecurityFix(t *testing.T) {
	for notes, want := range map[string]bool{
		"Fixes CVE-2026-1234 in the parser.": true,
		"Security: fix a heap overflow.":     true,
		"New features and bug fixes.":        false,
	} {
		if got := securityFix(notes); got != want {
			t.Errorf("securityFix(%q) = %v, want %v", notes, got, want)
		}
	}
}

func TestFreshnessCountsLibrariesBehindTheWindow(t *testing.T) {
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	f := freshness([]string{"sdl3", "jolt", "glm"}, []Release{
		{Library: "jolt", Version: "5.7.0", Released: "2026-10-25"}, // 7 days ago: still within the window
		{Library: "glm", Version: "1.0.4", Released: "2026-09-01"},  // long behind
		{Library: "glm", Version: "1.0.5", Released: "2026-10-30"},
	}, now)
	if f.Libraries != 3 || f.Fresh != 2 || len(f.Behind) != 1 || f.Behind[0] != "glm" {
		t.Errorf("freshness %+v, want 2 of 3 with glm behind", f)
	}
}
