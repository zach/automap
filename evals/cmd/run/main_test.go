package main

import (
	"os"
	"testing"

	"github.com/zach/automap/internal/release"
)

func TestEveryTaskLoads(t *testing.T) {
	ids, err := taskIDs("../../..", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) == 0 {
		t.Fatal("no eval tasks")
	}
	for _, id := range ids {
		task, err := loadTask("../../..", id)
		if err != nil {
			t.Fatal(err)
		}
		if task.ID != id || task.Request == "" || task.Pass.Level == "" && !task.Pass.Ejected {
			t.Errorf("task %s: %+v needs its id, a request and the level it must reach, or to be an eject task", id, task)
		}
		if task.Repo != "" {
			if _, err := os.Stat("../../tasks/" + id + "/" + task.Repo); err != nil {
				t.Errorf("task %s: starting repo: %v", id, err)
			}
		}
	}
}

func TestASpendingCapStopIsSkippedNotFailed(t *testing.T) {
	for _, c := range []struct {
		output string
		failed bool
		want   bool
	}{
		{`{"is_error":true,"result":"Your credit balance is too low to access the Anthropic API."}`, true, true},
		{`API Error: 400 You have reached your specified API usage limits.`, true, true},
		{`{"is_error":false,"result":"Started; probed."}`, false, false},
		{`error: build failed`, true, false},
	} {
		if got := capped(c.output, c.failed); got != c.want {
			t.Errorf("capped(%q, %v) = %v, want %v", c.output, c.failed, got, c.want)
		}
	}
}

func TestGrade(t *testing.T) {
	var task Task
	task.Pass.Level = "probed"
	task.Pass.Slots = map[string]string{"physics-3d": "jolt"}
	for _, c := range []struct {
		name, level string
		slots       map[string]string
		pass        bool
	}{
		{"probed with jolt in physics-3d", "probed", map[string]string{"physics-3d": "jolt", "windowing": "sdl3"}, true},
		{"probed without the library", "probed", map[string]string{"windowing": "sdl3"}, false},
		{"jolt below probed", "smoke", map[string]string{"physics-3d": "jolt"}, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			status, reason := grade(&task, outcome{Level: c.level, Summary: "summary", Slots: c.slots})
			if (status == "pass") != c.pass {
				t.Errorf("grade %s (%s), want pass %v", status, reason, c.pass)
			}
			if status != "pass" && reason == "" {
				t.Error("a failed grade gives no reason")
			}
		})
	}
}

func TestGradeChecksLibrariesAndPrivateVariants(t *testing.T) {
	var task Task
	task.Pass.Level = "probed"
	task.Pass.Libraries = []string{"acme-jobs"}
	task.Pass.PrivateVariants = []string{"sdl3"}
	for _, c := range []struct {
		name string
		o    outcome
		pass bool
	}{
		{"both", outcome{Level: "probed", Libraries: []string{"sdl3", "acme-jobs"}, PrivateVariants: []string{"sdl3"}}, true},
		{"no in-house library", outcome{Level: "probed", Libraries: []string{"sdl3"}, PrivateVariants: []string{"sdl3"}}, false},
		{"SDL3 from the public map", outcome{Level: "probed", Libraries: []string{"sdl3", "acme-jobs"}}, false},
	} {
		if status, _ := grade(&task, c.o); (status == "pass") != c.pass {
			t.Errorf("%s: grade %s, want pass %v", c.name, status, c.pass)
		}
	}
}

func TestGradeAnEjectTask(t *testing.T) {
	var task Task
	task.Pass.Ejected = true
	for _, c := range []struct {
		name string
		o    outcome
		pass bool
	}{
		{"ejected and builds", outcome{Ejected: true}, true},
		{"the lock is still there", outcome{Ejected: false}, false},
		{"ejected but the smoke run fails", outcome{Ejected: true, Plain: "the smoke run: exit status 1"}, false},
	} {
		if status, _ := grade(&task, c.o); (status == "pass") != c.pass {
			t.Errorf("%s: grade %s, want pass %v", c.name, status, c.pass)
		}
	}
}

func TestEjectMeasure(t *testing.T) {
	yes, no := true, false
	if m := ejectMeasure([]TaskResult{{ID: "a"}, {ID: "b", Status: "fail"}}); m != nil {
		t.Errorf("eject measure %+v with nothing ejected, want none", m)
	}
	m := ejectMeasure([]TaskResult{{ID: "a", Ejected: &yes}, {ID: "b", Ejected: &no}, {ID: "c", Ejected: &yes}, {ID: "d"}})
	if m == nil || m.Ejected != 3 || m.Built != 2 {
		t.Fatalf("eject measure %+v, want 2 of 3", m)
	}
}

func TestMeasures(t *testing.T) {
	yes, no := true, false
	tasks := []TaskResult{
		{ID: "testbed-start", Status: "pass", DurationSec: 600, Ejected: &yes},
		{ID: "testbed-extend", Status: "pass", DurationSec: 200, Ejected: &yes},
		{ID: "custom-adopt", Status: "fail", DurationSec: 900, Ejected: &no},
		{ID: "custom-eject", Status: "skipped"},
	}
	fresh := &mapMeasure{}
	fresh.Freshness.Libraries, fresh.Freshness.Fresh = 12, 12
	fresh.Evidence.Pieces, fresh.Evidence.Current = 40, 36
	m := measures(tasks, "warm", ejectMeasure(tasks), fresh)
	if m.RequestToTestbed == nil || !m.RequestToTestbed.Met || m.RequestToTestbed.Seconds != 600 {
		t.Errorf("request to testbed %+v, want met at 600s warm", m.RequestToTestbed)
	}
	if s := m.FirstAttemptSuccess; s == nil || s.Of != 2 || s.Out != 3 || s.Met {
		t.Errorf("first-attempt success %+v, want 2 of 3 (skipped excluded), not met", s)
	}
	if f := m.Freshness; f == nil || !f.Libraries.Met || f.Evidence.Met || f.Met {
		t.Errorf("freshness %+v, want libraries met, evidence at 90%% below its 95%% target, not met", f)
	}
	if e := m.Eject; e == nil || e.Of != 2 || e.Out != 3 || e.Met {
		t.Errorf("eject %+v, want 2 of 3, not met", e)
	}
	if cold := measures(tasks, "cold", nil, nil); cold.RequestToTestbed.Met || cold.Freshness != nil || cold.Eject != nil {
		t.Errorf("with a cold cache and no map measures: %+v", cold)
	}
}

// An upgrade task passes when the lock has moved to the newest map
// release: an unchanged lock at probed is a session that did nothing.
func TestGradeChecksTheMapRelease(t *testing.T) {
	var task Task
	task.Pass.Level = "probed"
	task.Pass.MapRelease = "newest"
	for _, c := range []struct {
		name string
		o    outcome
		pass bool
	}{
		{"moved to the newest release", outcome{Level: "probed", MapRelease: "map-2026.11.07", NewestRelease: "map-2026.11.07"}, true},
		{"still one release behind", outcome{Level: "probed", MapRelease: "map-2026.11.06", NewestRelease: "map-2026.11.07"}, false},
		{"moved but below probed", outcome{Level: "smoke", MapRelease: "map-2026.11.07", NewestRelease: "map-2026.11.07"}, false},
	} {
		if status, reason := grade(&task, c.o); (status == "pass") != c.pass {
			t.Errorf("%s: grade %s (%s), want pass %v", c.name, status, reason, c.pass)
		}
	}
}

// The starter-upgrade task starts its composition against the release
// before the newest, so the session has a real release to move across.
func TestTheUpgradeTaskStartsOneReleaseBehind(t *testing.T) {
	task, err := loadTask("../../..", "starter-upgrade")
	if err != nil {
		t.Fatal(err)
	}
	if task.SetupMapRelease != "previous" || task.Pass.MapRelease != "newest" || task.Pass.Level != "probed" || len(task.Setup) == 0 {
		t.Errorf("starter-upgrade: %+v should set up against the previous release and pass at probed on the newest", task)
	}
}

// The runner gives each task's CLI its own home, so the CLI's release cache
// must be named outright: an upgrade task's setup reads the previous
// release from the runner's cache, and the lock records the release only
// when the CLI's cache is the same one.
func TestTheCLISharesTheRunnersReleaseCache(t *testing.T) {
	t.Setenv("AUTOMAP_CACHE", "/caches/automap")
	env := cliEnv("/work/home", "/bin/automap", "/caches/vcpkg", "")
	want := "AUTOMAP_CACHE=" + release.CacheDir()
	for _, e := range env {
		if e == want {
			return
		}
	}
	t.Errorf("the CLI's env %q lacks %q", env, want)
}
