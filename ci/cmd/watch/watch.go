package main

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var plainVersion = regexp.MustCompile(`^\d+(\.\d+)*$`)

// versionOf is the version an upstream tag names, by the pattern of the
// map's own pin: the pin's tag with its version swapped out, such as
// release-{v} or v{v}-docking. A tag that doesn't fit, or names a
// pre-release, names none. It is a pure function.
func versionOf(pinTag, pinVersion, tag string) string {
	i := strings.Index(pinTag, pinVersion)
	if i < 0 {
		return ""
	}
	prefix, suffix := pinTag[:i], pinTag[i+len(pinVersion):]
	if !strings.HasPrefix(tag, prefix) || !strings.HasSuffix(tag, suffix) || len(tag) <= len(prefix)+len(suffix) {
		return ""
	}
	v := tag[len(prefix) : len(tag)-len(suffix)]
	if !plainVersion.MatchString(v) {
		return ""
	}
	return v
}

// newer lists the versions later than newest, in version order. It is a
// pure function.
func newer(newest string, versions []string) []string {
	var out []string
	for _, v := range versions {
		if compare(v, newest) > 0 {
			out = append(out, v)
		}
	}
	sort.Slice(out, func(i, j int) bool { return compare(out[i], out[j]) < 0 })
	return out
}

// compare orders dotted versions numerically.
func compare(a, b string) int {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) || i < len(bs); i++ {
		var x, y int
		if i < len(as) {
			x, _ = strconv.Atoi(as[i])
		}
		if i < len(bs) {
			y, _ = strconv.Atoi(bs[i])
		}
		if x != y {
			return x - y
		}
	}
	return 0
}

var security = regexp.MustCompile(`(?i)\bCVE-\d{4}-\d+|\bsecurity\b|\bvulnerabilit`)

// securityFix reports whether release notes name a security fix, which
// lets the release skip the cooldown, never the evidence (ADR-0037).
func securityFix(notes string) bool { return security.MatchString(notes) }
