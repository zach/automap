package verify_test

import (
	"strings"
	"testing"

	"github.com/zach/automap/internal/mapmodel"
	"github.com/zach/automap/internal/verify"
)

func TestLevel(t *testing.T) {
	pass := verify.Result{Rule: "a.b", Status: "pass"}
	na := verify.Result{Rule: "a.c", Status: "not-applicable", Reason: "elsewhere"}
	fail := verify.Result{Rule: "a.d", Status: "fail", FailureCode: "X"}
	unverifiable := verify.Result{Rule: "a.e", Status: "unverifiable", Reason: "can't check"}
	for _, c := range []struct {
		name                     string
		built, smoked, probesRan bool
		results                  []verify.Result
		want                     string
	}{
		{"didn't build", false, false, false, nil, "none"},
		{"built, smoke failed", true, false, false, nil, "builds"},
		{"smoke, no probes", true, true, false, nil, "smoke"},
		{"every applicable rule passed", true, true, true, []verify.Result{pass, na}, "probed"},
		{"a rule failed", true, true, true, []verify.Result{pass, fail}, "smoke"},
		{"an applicable rule was unverifiable", true, true, true, []verify.Result{pass, unverifiable}, "smoke"},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := verify.Level(c.built, c.smoked, c.probesRan, c.results); got != c.want {
				t.Errorf("level %s, want %s", got, c.want)
			}
		})
	}
}

func lifecycle() verify.Lifecycle {
	return verify.Lifecycle{
		Init: map[string]verify.StepInfo{
			"win:set_allocator": {Phase: 0},
			"win:init":          {Phase: 2},
			"win:open":          {Phase: 2, After: []string{"win:init"}},
			"ui:create":         {Phase: 4},
			"pair:attach":       {Phase: 4, After: []string{"ui:create"}},
		},
		Shutdown: map[string]verify.StepInfo{
			"pair:detach": {Phase: 4},
			"ui:destroy":  {Phase: 4},
			"win:close":   {Phase: 2},
			"win:quit":    {Phase: 2, After: []string{"win:close"}},
		},
		Ordering: []mapmodel.Ordering{
			{Before: "pair:detach", After: "ui:destroy", When: "shutdown"},
		},
	}
}

func TestCheckLifecycle(t *testing.T) {
	goodInit := []string{"win:set_allocator", "win:init", "win:open", "ui:create", "pair:attach"}
	goodShutdown := []string{"pair:detach", "ui:destroy", "win:close", "win:quit"}
	for _, c := range []struct {
		name               string
		init, shutdown     []string
		wantInit, wantShut string // failure codes, or "pass"
	}{
		{"declared order", goodInit, goodShutdown, "pass", "pass"},
		{"init out of phase", []string{"win:init", "win:set_allocator", "win:open", "ui:create", "pair:attach"}, goodShutdown, "INIT_ORDER(win:init before win:set_allocator)", "pass"},
		{"init before its after", []string{"win:set_allocator", "win:open", "win:init", "ui:create", "pair:attach"}, goodShutdown, "INIT_ORDER(win:open before win:init)", "pass"},
		{"shutdown out of phase", goodInit, []string{"win:close", "pair:detach", "ui:destroy", "win:quit"}, "pass", "SHUTDOWN_ORDER(win:close before pair:detach)"},
		{"pairing ordering", goodInit, []string{"ui:destroy", "pair:detach", "win:close", "win:quit"}, "pass", "SHUTDOWN_ORDER(ui:destroy before pair:detach)"},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := verify.CheckLifecycle(lifecycle(), c.init, c.shutdown)
			for i, want := range []string{c.wantInit, c.wantShut} {
				code := got[i].Status
				if code == "fail" {
					code = got[i].FailureCode
				}
				if code != want {
					t.Errorf("%s: %s, want %s", got[i].Rule, code, want)
				}
			}
		})
	}
}

func TestCheckLifecycleWithNothingRecordedIsUnverifiable(t *testing.T) {
	for _, r := range verify.CheckLifecycle(lifecycle(), nil, nil) {
		if r.Status != "unverifiable" || r.Reason == "" {
			t.Errorf("%s: %s %q, want unverifiable with a reason", r.Rule, r.Status, r.Reason)
		}
	}
}

func TestResults(t *testing.T) {
	expected := []verify.Expected{
		{Rule: mapmodel.Rule{ID: "alloc.routed"}, Library: "win"},
		{Rule: mapmodel.Rule{ID: "alloc.routed"}, Library: "ui"},
		{Rule: mapmodel.Rule{ID: "jobs.tsan"}, Platforms: []string{"macos-arm64-appleclang"}},
		{Rule: mapmodel.Rule{ID: "life.order"}},
		{Rule: mapmodel.Rule{ID: "diag.routed"}, Library: "ui", NotApplicable: "ui has no runtime hooks."},
	}
	got := verify.Results(expected, "linux-x64-clang", verify.Observed{
		Ran:      true,
		Reported: []verify.Result{{Rule: "alloc.routed", Library: "win", Status: "fail", FailureCode: "NOT_ROUTED(win)"}},
		Computed: []verify.Result{{Rule: "life.order", Status: "pass"}},
	})
	want := []string{"fail", "unverifiable", "not-applicable", "pass", "not-applicable"}
	for i, r := range got {
		if r.Status != want[i] {
			t.Errorf("%s %s: %s, want %s", r.Rule, r.Library, r.Status, want[i])
		}
		if (r.Status == "unverifiable" || r.Status == "not-applicable") && r.Reason == "" {
			t.Errorf("%s %s is %s without a reason", r.Rule, r.Library, r.Status)
		}
	}
}

func TestResultsWhenProbesDidntRun(t *testing.T) {
	got := verify.Results([]verify.Expected{{Rule: mapmodel.Rule{ID: "a.b"}}}, "linux-x64-clang", verify.Observed{Level: "builds"})
	if got[0].Status != "unverifiable" || got[0].Reason != "No probe ran: the composition reached builds." {
		t.Errorf("%+v, want unverifiable because no probe ran", got[0])
	}
}

func TestCheckPrecision(t *testing.T) {
	for _, c := range []struct {
		name       string
		choice     string
		precisions map[string]string
		want       verify.Result
	}{
		{"every library agrees", "float", map[string]string{"jolt": "float", "other": "float"},
			verify.Result{Rule: "math-types.precision-agrees", Status: "pass"}},
		{"a library differs", "double", map[string]string{"jolt": "float"},
			verify.Result{Rule: "math-types.precision-agrees", Status: "fail", FailureCode: "PRECISION_MISMATCH(jolt)",
				Reason: "jolt uses float world positions, but the composition chooses double."}},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := verify.CheckPrecision(c.choice, c.precisions); got != c.want {
				t.Errorf("got %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestLinkMismatch(t *testing.T) {
	log := `[12/14] Linking CXX executable testbed.exe
FAILED: testbed.exe
stub_ui.lib(stub_ui.c.obj) : error LNK2038: mismatch detected for 'STUB_UI_WIDE_IDS': value '1' doesn't match value '0' in stub_ui_setup.cpp.obj
LINK : fatal error LNK1319: 1 mismatches detected`
	r := verify.LinkMismatch(log)
	if r == nil || r.Rule != "build-flags.abi-defines-match" || r.Status != "fail" || r.FailureCode != "ABI_DEFINE_MISMATCH(STUB_UI_WIDE_IDS)" {
		t.Fatalf("got %+v, want the ABI rule failed with ABI_DEFINE_MISMATCH(STUB_UI_WIDE_IDS)", r)
	}
	if !strings.Contains(r.Reason, "value '1' doesn't match value '0'") {
		t.Errorf("reason %q doesn't quote the linker", r.Reason)
	}
	if verify.LinkMismatch("error: undefined reference to `foo'") != nil {
		t.Error("an unrelated link error is reported as an ABI mismatch")
	}
}

func TestAComputedResultCountsWhenNoProbeRan(t *testing.T) {
	expected := []verify.Expected{{Rule: mapmodel.Rule{ID: "build-flags.abi-defines-match"}}, {Rule: mapmodel.Rule{ID: "lifecycle.init-order"}}}
	link := verify.Result{Rule: "build-flags.abi-defines-match", Status: "fail", FailureCode: "ABI_DEFINE_MISMATCH(X)"}
	got := verify.Results(expected, "windows-x64-msvc", verify.Observed{Level: "none", Computed: []verify.Result{link}})
	if got[0] != link || got[1].Status != "unverifiable" {
		t.Errorf("results %+v, want the link-time failure and the other rule unverifiable", got)
	}
}
