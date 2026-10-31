package main

import (
	"os"
	"testing"
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
		if task.ID != id || task.Request == "" || task.Pass.Level == "" {
			t.Errorf("task %s: %+v needs its id, a request and the level it must reach", id, task)
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
