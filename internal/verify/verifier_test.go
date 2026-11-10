package verify_test

import (
	"testing"

	"github.com/zach/automap/internal/verify"
)

// Evidence from CI is signed when the job has an identity token, the map
// is released, and every private map the lock pins has a commit behind
// it: content with no commit can't be read back.
func TestEvidenceIsSignedOnlyWhenEverythingPinnedCanBeReadBack(t *testing.T) {
	ci := map[string]string{"GITHUB_ACTIONS": "true", "GITHUB_WORKFLOW_REF": "acme/game/.github/workflows/automap.yml@refs/heads/main",
		"GITHUB_SERVER_URL": "https://github.com", "ACTIONS_ID_TOKEN_REQUEST_URL": "u", "ACTIONS_ID_TOKEN_REQUEST_TOKEN": "t"}
	getenv := func(k string) string { return ci[k] }
	for _, c := range []struct {
		release   string
		committed bool
		signed    bool
	}{
		{"map-2026.11.07", true, true},
		{"map-2026.11.07", false, false},
		{"unreleased", true, false},
	} {
		if got := verify.Verifier(c.release, c.committed, getenv)["signed"]; got != c.signed {
			t.Errorf("release %s, private maps committed %v: signed %v, want %v", c.release, c.committed, got, c.signed)
		}
	}
	if local := verify.Verifier("map-2026.11.07", true, func(string) string { return "" }); local["kind"] != "developer-local" || local["signed"] != false {
		t.Errorf("off CI the verifier is %v, want developer-local, unsigned", local)
	}
}
