package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zach/automap/internal/composition"
	"github.com/zach/automap/internal/mapmodel"
)

// PredicateType is the evidence predicate's type (SPEC §6.1).
const PredicateType = "https://github.com/zach/automap/evidence/v1"

// Result is one rule's result, as evidence records it.
type Result struct {
	Rule        string `json:"rule"`
	Library     string `json:"library,omitempty"`
	Status      string `json:"status"`
	FailureCode string `json:"failure_code,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

// Predicate is the evidence predicate. The fields a composition's evidence
// shares with its edges' are carried over as they are.
type Predicate struct {
	SubjectKind     string          `json:"subject_kind"`
	SubjectDetail   any             `json:"subject_detail"`
	Platform        string          `json:"platform"`
	Level           string          `json:"level"`
	ProducedAt      string          `json:"produced_at"`
	MapRelease      string          `json:"map_release"`
	Conditions      json.RawMessage `json:"conditions"`
	Results         []Result        `json:"results"`
	SetAsideReports json.RawMessage `json:"set_aside_reports,omitempty"`
	Verifier        Verifier        `json:"verifier"`
}

type Verifier struct {
	Kind   string `json:"kind"`
	ID     string `json:"id"`
	Signed bool   `json:"signed,omitempty"`
}

type Subject struct {
	Name   string            `json:"name"`
	Digest map[string]string `json:"digest"`
}

// Statement is an in-toto v1 Statement carrying the evidence predicate.
type Statement struct {
	Type          string    `json:"_type"`
	Subject       []Subject `json:"subject"`
	PredicateType string    `json:"predicateType"`
	Predicate     Predicate `json:"predicate"`
}

// Uncaught is a seeded violation its probe didn't catch, by rule and the
// library a library rule's seed is about.
type Uncaught struct{ Rule, Library string }

// guidance is the reason an uncaught seed's rule gives in published
// evidence: until its probe catches its seed, the rule is guidance and
// doesn't count toward the level (ADR-0071).
const guidance = "Its probe didn't catch its seeded violation in this run, so the rule counts as guidance, not toward the level, until the probe is fixed (ADR-0071)."

// EdgeEvidence builds an edge's evidence from the evidence of the
// composition it was verified in: the edge's subject and pins, the results
// of its rules, with any rule whose probe missed its seed recorded as
// guidance, and the level those results reach. It returns nil when the
// composition reports no result for the edge's rules. It is a pure
// function.
func EdgeEvidence(m *mapmodel.Map, e Edge, comp Statement, lock composition.Lock, lockFile []byte, uncaught []Uncaught, verifierID string) *Statement {
	p := comp.Predicate
	out := &Statement{Type: comp.Type, PredicateType: PredicateType, Predicate: Predicate{
		SubjectKind: e.Kind, Platform: p.Platform, ProducedAt: p.ProducedAt, MapRelease: p.MapRelease,
		Conditions: p.Conditions, SetAsideReports: p.SetAsideReports,
		Verifier: Verifier{Kind: "automap-ci", ID: verifierID, Signed: true},
	}}
	switch e.Kind {
	case "binding":
		contract := m.Recipes[e.ID].Subject.Contract
		out.Subject = []Subject{subject(m, lock, e.Picks[0])}
		out.Predicate.SubjectDetail = map[string]any{"pin": pin(lock, e.Picks[0]), "contract": contract, "contract_version": lock.Contracts[contract]}
	case "pairing":
		out.Subject = []Subject{subject(m, lock, e.Picks[0]), subject(m, lock, e.Picks[1])}
		out.Predicate.SubjectDetail = map[string]any{"pins": []map[string]string{pin(lock, e.Picks[0]), pin(lock, e.Picks[1])}}
	case "composition":
		sum := sha256.Sum256(lockFile)
		out.Subject = []Subject{{Name: composition.LockFile, Digest: map[string]string{"sha256": hex.EncodeToString(sum[:])}}}
		out.Predicate.SubjectDetail = map[string]any{"lock_hash": lock.LockHash}
	}
	for _, r := range p.Results {
		if !edgeRule(m, e, r) {
			continue
		}
		for _, u := range uncaught {
			if u.Rule == r.Rule && u.Library == r.Library {
				r = Result{Rule: r.Rule, Library: r.Library, Status: "not-applicable", Reason: guidance}
			}
		}
		out.Predicate.Results = append(out.Predicate.Results, r)
	}
	if len(out.Predicate.Results) == 0 {
		return nil
	}
	out.Predicate.Level = p.Level
	if p.Level == "smoke" || p.Level == "probed" {
		// An edge reaches probed on its own rules, whatever other rules in
		// the composition did.
		out.Predicate.Level = "probed"
		for _, r := range out.Predicate.Results {
			if r.Status == "fail" || r.Status == "unverifiable" {
				out.Predicate.Level = "smoke"
			}
		}
	}
	if e.Kind != "composition" {
		out.Predicate.SetAsideReports = nil
	}
	return out
}

// edgeRule reports whether a composition's result belongs in an edge's
// evidence: a binding's contract's rules about its library or the whole
// composition, a pairing's own rules, and every rule for a composition.
func edgeRule(m *mapmodel.Map, e Edge, r Result) bool {
	switch e.Kind {
	case "binding":
		return strings.HasPrefix(r.Rule, m.Recipes[e.ID].Subject.Contract+".") && (r.Library == "" || r.Library == e.Picks[0].Library)
	case "pairing":
		return strings.HasPrefix(r.Rule, e.ID+".")
	}
	return true
}

func pin(lock composition.Lock, p Pick) map[string]string {
	lp := lock.Pins[p.Library]
	out := map[string]string{"library": p.Library, "version": lp.Version, "variant": lp.Variant, "content_hash": lp.ContentHash}
	if lp.Commit != "" {
		out["commit"] = lp.Commit
	}
	return out
}

// subject names a library's source at its pin, as a package URL where
// the source is on GitHub, with the pin's commit and content hash.
func subject(m *mapmodel.Map, lock composition.Lock, p Pick) Subject {
	lp := lock.Pins[p.Library]
	name := p.Library + "@" + lp.Version
	if l := m.Libraries[p.Library]; l != nil && l.Source != nil {
		if repo, ok := strings.CutPrefix(l.Source.URL, "https://github.com/"); ok {
			parts := strings.SplitN(repo, "/", 3)
			tag := lp.Version
			if v := version(m, p); v != nil && v.Pin != nil && v.Pin.Tag != "" {
				tag = v.Pin.Tag
			}
			if len(parts) >= 2 {
				name = fmt.Sprintf("pkg:github/%s/%s@%s", parts[0], parts[1], tag)
			}
		}
	}
	digest := map[string]string{}
	if lp.Commit != "" {
		digest["gitCommit"] = lp.Commit
	}
	if algo, sum, ok := strings.Cut(lp.ContentHash, ":"); ok {
		digest[algo] = sum
	}
	return Subject{Name: name, Digest: digest}
}
