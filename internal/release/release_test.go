package release_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/zach/automap/internal/release"
)

func TestEvidenceState(t *testing.T) {
	now := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	evidence := func(level string, age time.Duration, status string) []byte {
		return []byte(fmt.Sprintf(`{"predicate":{"level":%q,"produced_at":%q,"results":[{"rule":"a.b","status":%q}]}}`,
			level, now.Add(-age).Format(time.RFC3339), status))
	}
	for _, c := range []struct {
		name string
		raw  []byte
		want string
	}{
		{"recent and probed", evidence("probed", 24*time.Hour, "pass"), "current"},
		{"recent at smoke", evidence("smoke", 24*time.Hour, "unverifiable"), "current"},
		{"older than 30 days", evidence("probed", 31*24*time.Hour, "pass"), "stale"},
		{"a rule failed", evidence("smoke", time.Hour, "fail"), "broken"},
		{"failed long ago", evidence("smoke", 40*24*time.Hour, "fail"), "broken"},
		{"didn't run", evidence("builds", time.Hour, "unverifiable"), "broken"},
	} {
		t.Run(c.name, func(t *testing.T) {
			if _, got, err := release.State(c.raw, now); err != nil || got != c.want {
				t.Errorf("state %q (%v), want %q", got, err, c.want)
			}
		})
	}
}
