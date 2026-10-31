package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Cooldown is how old an upstream release must be before it enters the map
// (ADR-0037).
const Cooldown = 7 * 24 * time.Hour

// Version is a library version a map change adds.
type Version struct {
	Library     string `json:"library"`
	Version     string `json:"version"`
	Released    string `json:"released"` // YYYY-MM-DD
	SecurityFix bool   `json:"security_fix"`
}

// Change is what the gate knows about a curation change: the files it adds,
// the contracts' current major versions on main, the library versions it
// adds, whether Automap's CI verified it at the probed level on every
// platform, and whether its other checks passed.
type Change struct {
	Added      []string
	Majors     map[string]int
	Versions   []Version
	EvidenceOK bool
	ChecksOK   bool
}

var contractVersion = regexp.MustCompile(`^map/contracts/([^/]+)/(\d+)\.\d+\.\d+\.json$`)

// Decide is the cooldown gate (ADR-0037, ADR-0051, ADR-0069): a change that
// adds a slot, a library or a contract major version waits for a human;
// otherwise it merges once every version it adds is at least Cooldown old
// (a security release skips the wait) and its evidence is known-good on
// every platform, never without the evidence. It returns merge, wait or
// human, and why. It is a pure function.
func Decide(c Change, now time.Time) (action, reason string) {
	for _, f := range c.Added {
		switch {
		case strings.HasPrefix(f, "map/slots/"):
			return "human", "it adds a slot, " + f + ": what the map covers is the product owner's call"
		case strings.HasPrefix(f, "map/libraries/"):
			return "human", "it adds a library, " + f + ": what the map covers is the product owner's call"
		}
		if m := contractVersion.FindStringSubmatch(f); m != nil {
			if major, _ := strconv.Atoi(m[2]); major > c.Majors[m[1]] {
				return "human", "it adds a major version of contract " + m[1] + ": the product owner's call"
			}
		}
	}
	if len(c.Versions) == 0 {
		return "wait", "it adds no library version"
	}
	if !c.ChecksOK {
		return "wait", "its checks haven't passed"
	}
	if !c.EvidenceOK {
		return "wait", "Automap's CI hasn't verified it at the probed level on every platform"
	}
	for _, v := range c.Versions {
		released, err := time.Parse("2006-01-02", v.Released)
		if err != nil {
			return "wait", fmt.Sprintf("%s %s has no release date", v.Library, v.Version)
		}
		if age := now.Sub(released); age < Cooldown && !v.SecurityFix {
			return "wait", fmt.Sprintf("%s %s was released %s; it can merge from %s", v.Library, v.Version, v.Released, released.Add(Cooldown).Format("2006-01-02"))
		}
	}
	return "merge", "every version is past the cooldown or a security fix, with probed evidence on every platform"
}
