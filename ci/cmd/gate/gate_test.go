package main

import (
	"testing"
	"time"
)

var now = time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)

func TestTheCooldownGate(t *testing.T) {
	old := Version{Library: "jolt", Version: "5.7.0", Released: "2026-10-20"}
	fresh := Version{Library: "jolt", Version: "5.7.0", Released: "2026-10-29"}
	security := Version{Library: "jolt", Version: "5.6.1", Released: "2026-10-30", SecurityFix: true}
	for _, c := range []struct {
		name string
		in   Change
		want string
	}{
		{"old enough with evidence", Change{Versions: []Version{old}, EvidenceOK: true, ChecksOK: true}, "merge"},
		{"inside the cooldown", Change{Versions: []Version{fresh}, EvidenceOK: true, ChecksOK: true}, "wait"},
		{"a security release skips the wait", Change{Versions: []Version{security}, EvidenceOK: true, ChecksOK: true}, "merge"},
		{"never the evidence", Change{Versions: []Version{security}, EvidenceOK: false, ChecksOK: true}, "wait"},
		{"failing checks", Change{Versions: []Version{old}, EvidenceOK: true, ChecksOK: false}, "wait"},
		{"a new library", Change{Added: []string{"map/libraries/flecs.json"}, EvidenceOK: true, ChecksOK: true}, "human"},
		{"a new slot", Change{Added: []string{"map/slots/ecs.json"}, EvidenceOK: true, ChecksOK: true}, "human"},
		{"a contract major version", Change{Added: []string{"map/contracts/jobs/2.0.0.json"}, Majors: map[string]int{"jobs": 1}, EvidenceOK: true, ChecksOK: true}, "human"},
		{"a contract minor version", Change{Added: []string{"map/contracts/jobs/1.1.0.json"}, Majors: map[string]int{"jobs": 1}, Versions: []Version{old}, EvidenceOK: true, ChecksOK: true}, "merge"},
		{"nothing to curate", Change{EvidenceOK: true, ChecksOK: true}, "wait"},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got, why := Decide(c.in, now); got != c.want {
				t.Errorf("%s (%s), want %s", got, why, c.want)
			}
		})
	}
}
