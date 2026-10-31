// Command run runs eval tasks: each a fresh Claude Code session in headless
// mode, in a clean repo, given a one-line request and no human input, then
// graded through the CLI's JSON output (ADR-0081). It writes a
// machine-readable report (ADR-0063).
//
//	go run ./evals/cmd/run -task testbed-start -report report.json
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/zach/automap/internal/agentfiles"
	"github.com/zach/automap/internal/cli"
)

// Task is one eval task, evals/tasks/<id>/task.json.
type Task struct {
	ID      string `json:"id"`
	Request string `json:"request"`
	// Repo is the starting repo, a directory beside task.json, or "" for an
	// empty repo.
	Repo string `json:"repo,omitempty"`
	// Setup is automap invocations that bring the repo to the task's
	// starting point before the session, such as starting the testbed.
	Setup [][]string `json:"setup,omitempty"`
	// Pass is what grading requires: the level automap verify --json
	// reports, and the library each named slot holds in the manifest.
	Pass struct {
		Level string            `json:"level"`
		Slots map[string]string `json:"slots,omitempty"`
	} `json:"pass"`
	TimeoutMinutes int `json:"timeout_minutes"`
}

// Report is the eval run's machine-readable report.
type Report struct {
	RunAt      string       `json:"run_at"`
	Model      string       `json:"model"`
	CLIVersion string       `json:"cli_version"`
	Cache      string       `json:"cache"` // warm or cold vcpkg binary cache
	Tasks      []TaskResult `json:"tasks"`
}

// TaskResult is one task's outcome.
type TaskResult struct {
	ID          string  `json:"id"`
	Status      string  `json:"status"` // pass, fail or skipped
	DurationSec float64 `json:"duration_seconds"`
	Level       string  `json:"level,omitempty"`
	Reason      string  `json:"reason,omitempty"`
	CostUSD     float64 `json:"cost_usd,omitempty"`
}

func main() {
	tasks := flag.String("task", "", "comma-separated task ids; all tasks when empty")
	reportPath := flag.String("report", "eval-report.json", "where to write the report")
	model := flag.String("model", "opus", "the Claude model to ask for; the report names the model that answered")
	automap := flag.String("automap", "automap", "the automap binary the agent uses")
	mapDir := flag.String("map", "", "a public map directory to read instead of the newest published map release")
	flag.Parse()

	root, _ := os.Getwd()
	ids, err := taskIDs(root, *tasks)
	if err != nil {
		fail(err)
	}
	cache := vcpkgCache()
	report := Report{RunAt: time.Now().UTC().Format(time.RFC3339), CLIVersion: cli.Version, Cache: "cold"}
	if entries, err := os.ReadDir(cache); err == nil && len(entries) > 0 {
		report.Cache = "warm"
	}
	absMap := ""
	if *mapDir != "" {
		absMap, _ = filepath.Abs(*mapDir)
	}
	absCLI, err := exec.LookPath(*automap)
	if err != nil {
		fail(err)
	}
	absCLI, _ = filepath.Abs(absCLI)
	for _, id := range ids {
		t, err := loadTask(root, id)
		if err != nil {
			fail(err)
		}
		r, used := runTask(root, t, *model, absCLI, absMap, cache)
		if used != "" {
			report.Model = used
		}
		report.Tasks = append(report.Tasks, r)
		fmt.Printf("%s: %s in %.0fs, model %s, level %q, cost $%.2f. %s\n", r.ID, r.Status, r.DurationSec, used, r.Level, r.CostUSD, r.Reason)
	}
	if report.Model == "" {
		report.Model = *model
	}
	raw, _ := json.MarshalIndent(report, "", "  ")
	os.MkdirAll(filepath.Dir(*reportPath), 0o755)
	if err := os.WriteFile(*reportPath, append(raw, '\n'), 0o644); err != nil {
		fail(err)
	}
	for _, r := range report.Tasks {
		if r.Status == "fail" {
			os.Exit(1)
		}
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "eval:", err)
	os.Exit(2)
}

func taskIDs(root, list string) ([]string, error) {
	if list != "" {
		return strings.Split(list, ","), nil
	}
	entries, err := os.ReadDir(filepath.Join(root, "evals", "tasks"))
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			ids = append(ids, e.Name())
		}
	}
	sort.Strings(ids)
	return ids, nil
}

func loadTask(root, id string) (*Task, error) {
	raw, err := os.ReadFile(filepath.Join(root, "evals", "tasks", id, "task.json"))
	if err != nil {
		return nil, err
	}
	var t Task
	if err := json.Unmarshal(raw, &t); err != nil {
		return nil, fmt.Errorf("task %s: %v", id, err)
	}
	if t.TimeoutMinutes == 0 {
		t.TimeoutMinutes = 60
	}
	return &t, nil
}

// spendingCap matches the API's errors when the key's spending cap or
// credit runs out; such a run is skipped, not failed (ADR-0081).
var spendingCap = regexp.MustCompile(`(?i)credit balance is too low|spend(ing)? (limit|cap)|usage limit|billing`)

// capped reports whether a failed session was stopped by the spending cap.
func capped(output string, failed bool) bool {
	return failed && spendingCap.MatchString(output)
}

// vcpkgCache is the vcpkg binary cache the sessions share: the runner's
// own, even though each session gets a fresh HOME.
func vcpkgCache() string {
	if c := os.Getenv("VCPKG_DEFAULT_BINARY_CACHE"); c != "" {
		return c
	}
	home, _ := os.UserHomeDir()
	c := filepath.Join(home, ".cache", "vcpkg", "archives")
	os.MkdirAll(c, 0o755)
	return c
}

// runTask runs one task in a fresh repo and grades it. It returns the
// model Claude Code reports using.
func runTask(root string, t *Task, model, automap, mapDir, cache string) (TaskResult, string) {
	r := TaskResult{ID: t.ID}
	work, err := os.MkdirTemp("", "automap-eval-"+t.ID+"-")
	if err != nil {
		r.Status, r.Reason = "fail", err.Error()
		return r, ""
	}
	repo := filepath.Join(work, "repo")
	home := filepath.Join(work, "home")
	os.MkdirAll(repo, 0o755)
	if t.Repo != "" {
		os.CopyFS(repo, os.DirFS(filepath.Join(root, "evals", "tasks", t.ID, t.Repo)))
	}
	exec.Command("git", "-C", repo, "init", "-q").Run()
	// The agent finds Automap the way a developer's agent would: the CLI on
	// its PATH and the router skill installed at user scope.
	skill := filepath.Join(home, ".claude", "skills", "automap", "SKILL.md")
	os.MkdirAll(filepath.Dir(skill), 0o755)
	os.WriteFile(skill, []byte(agentfiles.RouterSkill(cli.Version)), 0o644)

	env := append(os.Environ(),
		"HOME="+home,
		"PATH="+filepath.Dir(automap)+string(os.PathListSeparator)+os.Getenv("PATH"),
		"VCPKG_DEFAULT_BINARY_CACHE="+cache,
	)
	// Without a map directory, the CLI reads the newest published map
	// release, as a developer's does.
	if mapDir != "" {
		env = append(env, "AUTOMAP_MAP="+mapDir)
	}
	for _, args := range t.Setup {
		setup := exec.Command(automap, args...)
		setup.Dir, setup.Env = repo, env
		if out, err := setup.CombinedOutput(); err != nil {
			r.Status, r.Reason = "fail", fmt.Sprintf("setup automap %s: %v\n%s", strings.Join(args, " "), err, out)
			return r, ""
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(t.TimeoutMinutes)*time.Minute)
	defer cancel()
	began := time.Now()
	cmd := exec.CommandContext(ctx, "claude", "-p", t.Request, "--output-format", "json", "--model", model, "--dangerously-skip-permissions")
	cmd.Dir, cmd.Env = repo, env
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	runErr := cmd.Run()
	r.DurationSec = time.Since(began).Seconds()
	if dir := os.Getenv("AUTOMAP_EVAL_LOGS"); dir != "" {
		os.MkdirAll(dir, 0o755)
		os.WriteFile(filepath.Join(dir, t.ID+"-session.json"), stdout.Bytes(), 0o644)
		os.WriteFile(filepath.Join(dir, t.ID+"-session.stderr"), stderr.Bytes(), 0o644)
	}

	var session struct {
		IsError    bool                       `json:"is_error"`
		Result     string                     `json:"result"`
		CostUSD    float64                    `json:"total_cost_usd"`
		ModelUsage map[string]json.RawMessage `json:"modelUsage"`
	}
	json.Unmarshal(stdout.Bytes(), &session)
	r.CostUSD = session.CostUSD
	used := ""
	for m := range session.ModelUsage {
		if used == "" || strings.Contains(m, "opus") {
			used = m
		}
	}
	if capped(stdout.String()+stderr.String(), runErr != nil || session.IsError) {
		r.Status, r.Reason = "skipped", "the API key's spending cap stopped the run"
		return r, used
	}
	if ctx.Err() != nil {
		r.Status, r.Reason = "fail", fmt.Sprintf("the session ran longer than %d minutes", t.TimeoutMinutes)
		return r, used
	}

	// Grade through the CLI, as the developer would: verify must reach the
	// task's level.
	verify := exec.Command(automap, "verify", "--json")
	verify.Dir, verify.Env = repo, env
	out, _ := verify.Output()
	var result struct {
		Level   string `json:"level"`
		Summary string `json:"summary"`
	}
	json.Unmarshal(out, &result)
	r.Level = result.Level
	var manifest struct {
		Slots map[string]string `json:"slots"`
	}
	if raw, err := os.ReadFile(filepath.Join(repo, "automap.manifest.json")); err == nil {
		json.Unmarshal(raw, &manifest)
	}
	r.Status, r.Reason = grade(t, result.Level, result.Summary, manifest.Slots)
	return r, used
}

// grade judges a session's outcome: verify reached the task's level, and
// the manifest puts each library the task names in its slot. It is a pure
// function.
func grade(t *Task, level, summary string, slots map[string]string) (status, reason string) {
	if level != t.Pass.Level {
		return "fail", fmt.Sprintf("verify reports %q, want %s: %s", level, t.Pass.Level, summary)
	}
	for slot, lib := range t.Pass.Slots {
		if slots[slot] != lib {
			return "fail", fmt.Sprintf("the manifest's %s slot holds %q, want %s", slot, slots[slot], lib)
		}
	}
	return "pass", ""
}
