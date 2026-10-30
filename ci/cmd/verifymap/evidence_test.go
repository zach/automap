package main

import (
	"encoding/json"
	"testing"

	"github.com/zach/automap/internal/composition"
	"github.com/zach/automap/schemas"
)

// compositionRun is the evidence a testbed composition's verify run wrote:
// every rule passed but stub-window's diagnostics, which failed.
func compositionRun() Statement {
	return Statement{Type: "https://in-toto.io/Statement/v1", PredicateType: PredicateType, Predicate: Predicate{
		SubjectKind: "composition", SubjectDetail: map[string]any{"lock_hash": "sha256:" + hex64},
		Platform: "linux-x64-clang", Level: "smoke", ProducedAt: "2026-10-30T12:00:00Z", MapRelease: "git:" + hex40,
		Conditions: json.RawMessage(`{"build_type":"Debug","compiler":"clang 21","environment":{},"sanitizers":{}}`),
		Results: []Result{
			{Rule: "allocation.routed-through-facade", Library: "stub-ui", Status: "pass"},
			{Rule: "allocation.routed-through-facade", Library: "stub-window", Status: "pass"},
			{Rule: "allocation.hooks-before-first-allocation", Library: "stub-ui", Status: "pass"},
			{Rule: "diagnostics.routed-through-facade", Library: "stub-window", Status: "fail", FailureCode: "DIAG_NOT_ROUTED(stub-window)"},
			{Rule: "lifecycle.init-order", Status: "pass"},
			{Rule: "stub-ui--stub-window.input-forwarded", Status: "pass"},
		},
		Verifier: Verifier{Kind: "developer-local", ID: "local"},
	}}
}

const (
	hex40 = "1111111111111111111111111111111111111111"
	hex64 = "2222222222222222222222222222222222222222222222222222222222222222"
)

var lock = composition.Lock{
	Pins: map[string]composition.Pin{
		"stub-ui":     {Version: "1.0.0", Variant: "default", Commit: hex40, ContentHash: "sha512:" + hex64 + hex64},
		"stub-window": {Version: "1.0.0", Variant: "default", Commit: hex40, ContentHash: "sha512:" + hex64 + hex64},
	},
	Contracts: map[string]string{"allocation": "1.0.0", "diagnostics": "1.0.0"},
	LockHash:  "sha256:" + hex64,
}

func valid(t *testing.T, s *Statement) {
	t.Helper()
	raw, _ := json.Marshal(s)
	var doc any
	json.Unmarshal(raw, &doc)
	if problems := schemas.Validate("evidence", doc); len(problems) > 0 {
		t.Errorf("evidence doesn't match the schema: %v\n%s", problems, raw)
	}
}

func TestABindingsEvidenceHasItsContractsRulesForItsLibrary(t *testing.T) {
	m := fixtureMap(t)
	e := Edge{Kind: "binding", ID: "stub-ui--allocation", Picks: []Pick{{"stub-ui", "1.0.0", "default"}}}

	s := EdgeEvidence(m, e, compositionRun(), lock, nil, nil, "https://github.com/zach/automap/.github/workflows/verify-map.yml@refs/heads/main")

	valid(t, s)
	if len(s.Predicate.Results) != 2 {
		t.Errorf("results %+v, want stub-ui's two allocation rules", s.Predicate.Results)
	}
	if s.Predicate.Level != "probed" {
		t.Errorf("level %q: the binding's own rules passed, want probed although the composition reached smoke", s.Predicate.Level)
	}
	if s.Subject[0].Digest["gitCommit"] != hex40 || s.Predicate.Verifier.Kind != "automap-ci" || !s.Predicate.Verifier.Signed {
		t.Errorf("subject %+v, verifier %+v", s.Subject, s.Predicate.Verifier)
	}
}

func TestABindingWithAFailedRuleStaysAtSmoke(t *testing.T) {
	m := fixtureMap(t)
	e := Edge{Kind: "binding", ID: "stub-window--diagnostics", Picks: []Pick{{"stub-window", "1.0.0", "default"}}}

	s := EdgeEvidence(m, e, compositionRun(), lock, nil, nil, "ci")

	valid(t, s)
	if s.Predicate.Level != "smoke" {
		t.Errorf("level %q, want smoke", s.Predicate.Level)
	}
}

func TestAProbeThatMissedItsSeedIsGuidance(t *testing.T) {
	m := fixtureMap(t)
	e := Edge{Kind: "binding", ID: "stub-ui--allocation", Picks: []Pick{{"stub-ui", "1.0.0", "default"}}}

	s := EdgeEvidence(m, e, compositionRun(), lock, nil, []Uncaught{{"allocation.hooks-before-first-allocation", "stub-ui"}}, "ci")

	valid(t, s)
	for _, r := range s.Predicate.Results {
		if r.Rule == "allocation.hooks-before-first-allocation" && (r.Status != "not-applicable" || r.Reason != guidance) {
			t.Errorf("result %+v, want not-applicable as guidance", r)
		}
	}
}

func TestAPairingAndACompositionsEvidence(t *testing.T) {
	m := fixtureMap(t)
	pairing := EdgeEvidence(m, Edge{Kind: "pairing", ID: "stub-ui--stub-window", Picks: []Pick{{"stub-ui", "1.0.0", "default"}, {"stub-window", "1.0.0", "default"}}},
		compositionRun(), lock, nil, nil, "ci")
	valid(t, pairing)
	if len(pairing.Predicate.Results) != 1 || len(pairing.Subject) != 2 {
		t.Errorf("pairing results %+v and subjects %+v, want its one rule and both libraries", pairing.Predicate.Results, pairing.Subject)
	}

	comp := EdgeEvidence(m, Edge{Kind: "composition", ID: "testbed", Picks: []Pick{{"stub-window", "1.0.0", "default"}, {"stub-ui", "1.0.0", "default"}}},
		compositionRun(), lock, []byte("{}"), nil, "ci")
	valid(t, comp)
	if len(comp.Predicate.Results) != 6 || comp.Predicate.Level != "smoke" || comp.Subject[0].Name != "automap.lock.json" {
		t.Errorf("composition evidence %+v", comp)
	}
}
