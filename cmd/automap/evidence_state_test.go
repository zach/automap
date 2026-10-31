package main_test

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/zach/automap/internal/clitest"
)

const (
	hex40 = "1111111111111111111111111111111111111111"
	hex64 = "2222222222222222222222222222222222222222222222222222222222222222"
)

// publishedEvidence is a piece of Automap CI's evidence for one of the
// fixture's stub-ui edges on linux-x64-clang, produced age ago, with one
// result.
func publishedEvidence(kind, rule, status string, age time.Duration) string {
	detail := map[string]any{"pin": map[string]string{"library": "stub-ui", "version": "1.0.0", "variant": "default", "content_hash": "sha512:" + hex64 + hex64},
		"contract": strings.Split(rule, ".")[0], "contract_version": "1.0.0"}
	if kind == "pairing" {
		detail = map[string]any{"pins": []map[string]string{
			{"library": "stub-ui", "version": "1.0.0", "variant": "default", "content_hash": "sha512:" + hex64 + hex64},
			{"library": "stub-window", "version": "1.0.0", "variant": "default", "content_hash": "sha512:" + hex64 + hex64},
		}}
	}
	level := "probed"
	result := map[string]any{"rule": rule, "status": status}
	if status == "fail" {
		level, result["failure_code"] = "smoke", "X"
	}
	raw, _ := json.MarshalIndent(map[string]any{
		"_type":         "https://in-toto.io/Statement/v1",
		"subject":       []map[string]any{{"name": "stub-ui", "digest": map[string]string{"sha256": hex64}}},
		"predicateType": "https://github.com/zach/automap/evidence/v1",
		"predicate": map[string]any{
			"subject_kind": kind, "subject_detail": detail, "platform": "linux-x64-clang", "level": level,
			"produced_at": time.Now().Add(-age).UTC().Format(time.RFC3339), "map_release": "git:" + hex40,
			"conditions": map[string]any{"build_type": "Debug", "compiler": "clang", "environment": map[string]string{}, "sanitizers": map[string]string{}},
			"results":    []any{result},
			"verifier":   map[string]any{"kind": "automap-ci", "id": "https://github.com/zach/automap/.github/workflows/verify-map.yml@refs/heads/main", "signed": true},
		},
	}, "", "  ")
	return string(raw)
}

// stubUIEvidence is current evidence for stub-ui's allocation binding,
// stale evidence for its diagnostics binding and broken evidence for its
// pairing with stub-window, by path in the map.
func stubUIEvidence() map[string]string {
	day := 24 * time.Hour
	return map[string]string{
		"evidence/binding/stub-ui--allocation/stub-ui-1.0.0-default/linux-x64-clang.json":                            publishedEvidence("binding", "allocation.routed-through-facade", "pass", day),
		"evidence/binding/stub-ui--diagnostics/stub-ui-1.0.0-default/linux-x64-clang.json":                           publishedEvidence("binding", "diagnostics.routed-through-facade", "pass", 40*day),
		"evidence/pairing/stub-ui--stub-window/stub-ui-1.0.0-default+stub-window-1.0.0-default/linux-x64-clang.json": publishedEvidence("pairing", "stub-ui--stub-window.input-forwarded", "fail", day),
	}
}

type evidenceEntry struct {
	Subject, Platform, Level, State string
}

func entryEvidence(t *testing.T, run *clitest.Run) map[string]string {
	t.Helper()
	var doc struct {
		Entries []struct {
			ID       string
			Evidence []evidenceEntry
		}
	}
	json.Unmarshal(run.Stdout, &doc)
	out := map[string]string{}
	for _, e := range doc.Entries {
		for _, ev := range e.Evidence {
			out[e.ID+" "+ev.Subject+" "+ev.Platform] = ev.Level + " " + ev.State
		}
	}
	return out
}

func expectStubUIStates(t *testing.T, run *clitest.Run) {
	t.Helper()
	got := entryEvidence(t, run)
	want := map[string]string{
		"stub-ui stub-ui--allocation linux-x64-clang":  "probed current",
		"stub-ui stub-ui--diagnostics linux-x64-clang": "probed stale",
		"stub-ui stub-ui--stub-window linux-x64-clang": "smoke broken",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s: %q, want %q (all: %v)", k, got[k], v, got)
		}
	}
	for _, code := range []string{"EVIDENCE_STALE(stub-ui--diagnostics,linux-x64-clang)", "EVIDENCE_BROKEN(stub-ui--stub-window,linux-x64-clang)"} {
		if !strings.Contains(string(run.Stdout), code) {
			t.Errorf("no %s warning:\n%s", code, run.Stdout)
		}
	}
	if strings.Contains(string(run.Stdout), "EVIDENCE_STALE(stub-ui--allocation") || strings.Contains(string(run.Stdout), "EVIDENCE_BROKEN(stub-ui--allocation") {
		t.Errorf("current evidence warned about:\n%s", run.Stdout)
	}
}

func TestShowReportsEachPieceOfEvidencesState(t *testing.T) {
	repo := fixtureRepo(t)
	for p, content := range stubUIEvidence() {
		repo.WriteFile(t, "map/"+p, content)
	}

	run := repo.Automap(t, "show", "library", "stub-ui", "--map", "map", "--json")

	run.ExpectExit(t, 0)
	expectStubUIStates(t, run)
}

func TestFindFetchesEvidenceFromMainAndKeepsItForOfflineUse(t *testing.T) {
	s := newReleaseServer(t)
	s.publish(t, "map-2026.10.01", fixtureMapDir(t), nil)
	for p, content := range stubUIEvidence() {
		s.evidence[p] = []byte(content)
	}
	repo := releaseRepo(t, s, t.TempDir())

	expectStubUIStates(t, repo.Automap(t, "find", "stub-ui", "--json"))
	s.Close()
	expectStubUIStates(t, repo.Automap(t, "find", "stub-ui", "--json"))
}

func TestVerifyWarnsAboutStaleOrBrokenPublishedEvidenceForTheLocksPins(t *testing.T) {
	repo := buildableRepo(t)
	repo.Automap(t, startTestbed...).ExpectExit(t, 0)
	for p, content := range stubUIEvidence() {
		repo.WriteFile(t, "map/"+strings.Replace(p, "linux-x64-clang", hostPlatform, 1), strings.ReplaceAll(content, "linux-x64-clang", hostPlatform))
	}

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	run.ExpectExit(t, 0)
	for _, code := range []string{"EVIDENCE_STALE(stub-ui--diagnostics," + hostPlatform + ")", "EVIDENCE_BROKEN(stub-ui--stub-window," + hostPlatform + ")"} {
		if !strings.Contains(string(run.Stdout), code) {
			t.Errorf("no %s warning:\n%s", code, run.Stdout)
		}
	}
}

// verifier is the evidence's verifier.
func verifier(t *testing.T, run *clitest.Run) (kind, id string, signed bool) {
	t.Helper()
	var e struct {
		Predicate struct {
			MapRelease string `json:"map_release"`
			Verifier   struct {
				Kind, ID string
				Signed   bool
			}
		}
	}
	json.Unmarshal([]byte(run.ReadFile(t, "evidence/"+hostPlatform+".json")), &e)
	return e.Predicate.Verifier.Kind, e.Predicate.Verifier.ID, e.Predicate.Verifier.Signed
}

// githubActions is the environment of a GitHub Actions job in the
// developer's repo, with an identity token when withToken.
func githubActions(repo *clitest.Repo, withToken bool) {
	repo.Setenv("GITHUB_ACTIONS", "true")
	repo.Setenv("GITHUB_SERVER_URL", "https://github.com")
	repo.Setenv("GITHUB_WORKFLOW_REF", "acme/engine/.github/workflows/automap.yml@refs/heads/main")
	if withToken {
		repo.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", "https://token.example.invalid")
		repo.Setenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN", "token")
	}
}

func TestVerifyInCIWithAnIdentityTokenIsSignedByTheWorkflow(t *testing.T) {
	repo := buildableRepo(t)
	commitMap(t, repo) // a clean checkout of the map: git:<commit>, not unreleased
	repo.Automap(t, startTestbed...).ExpectExit(t, 0)
	githubActions(repo, true)

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	run.ExpectExit(t, 0)
	kind, id, signed := verifier(t, run)
	if kind != "developer-ci" || id != "https://github.com/acme/engine/.github/workflows/automap.yml@refs/heads/main" || !signed {
		t.Errorf("verifier %s %s signed=%v, want developer-ci, the workflow, signed", kind, id, signed)
	}
}

func TestVerifyInCIWithoutAnIdentityTokenOrAgainstAnUnreleasedMapIsUnsigned(t *testing.T) {
	repo := buildableRepo(t)
	commitMap(t, repo)
	repo.Automap(t, startTestbed...).ExpectExit(t, 0)
	githubActions(repo, false)

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	if kind, _, signed := verifier(t, run); kind != "developer-ci" || signed {
		t.Errorf("without an identity token: verifier %s signed=%v, want developer-ci, unsigned", kind, signed)
	}

	unreleased := buildableRepo(t)
	unreleased.Automap(t, startTestbed...).ExpectExit(t, 0)
	githubActions(unreleased, true)
	run = unreleased.Automap(t, "verify", "--map", "map", "--json")
	if _, _, signed := verifier(t, run); signed {
		t.Error("evidence against an unreleased map is signed")
	}
}

// commitMap makes the repo's map directory a clean git checkout.
func commitMap(t *testing.T, repo *clitest.Repo) {
	t.Helper()
	for _, args := range [][]string{{"init", "-q"}, {"add", "."}, {"commit", "-q", "-m", "map"}} {
		cmd := exec.Command("git", append([]string{"-C", repo.Dir + "/map", "-c", "user.name=t", "-c", "user.email=t@example.invalid", "-c", "maintenance.auto=false", "-c", "gc.auto=0"}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}
