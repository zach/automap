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
	"slices"
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
	// Repo is the starting repo, a directory relative to task.json (another
	// task's, such as ../custom-adopt/repo), or "" for an empty repo.
	Repo string `json:"repo,omitempty"`
	// Overlay is a directory beside task.json copied over the starting
	// repo, such as a private map and a patch.
	Overlay string `json:"overlay,omitempty"`
	// Setup is automap invocations that bring the repo to the task's
	// starting point before the session, such as starting the testbed.
	Setup [][]string `json:"setup,omitempty"`
	// Pass is what grading requires: the level automap verify --json
	// reports, the library each named slot holds in the manifest, the
	// libraries the composition has, and the libraries the lock pins to a
	// private map's variant.
	Pass struct {
		Level           string            `json:"level"`
		Slots           map[string]string `json:"slots,omitempty"`
		Libraries       []string          `json:"libraries,omitempty"`
		PrivateVariants []string          `json:"private_variants,omitempty"`
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
	if t.Overlay != "" {
		overlay(filepath.Join(root, "evals", "tasks", t.ID, t.Overlay), repo)
	}
	// A studio's repo is a git repository with its history, which the lock
	// pins a private map in.
	for _, args := range [][]string{{"init", "-q"}, {"add", "-A"}, {"commit", "-q", "--allow-empty", "-m", "Starting point"}} {
		git := exec.Command("git", append([]string{"-C", repo}, args...)...)
		git.Env = append(os.Environ(), "GIT_AUTHOR_NAME=acme", "GIT_AUTHOR_EMAIL=acme@example.invalid",
			"GIT_COMMITTER_NAME=acme", "GIT_COMMITTER_EMAIL=acme@example.invalid")
		git.Run()
	}
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
	o := outcome{Level: result.Level, Summary: result.Summary}
	var manifest struct {
		Slots    map[string]string `json:"slots"`
		Variants map[string]any    `json:"variants"`
	}
	if raw, err := os.ReadFile(filepath.Join(repo, "automap.manifest.json")); err == nil {
		json.Unmarshal(raw, &manifest)
	}
	o.Slots = manifest.Slots
	for lib := range manifest.Variants {
		o.Libraries = append(o.Libraries, lib)
	}
	var lock struct {
		Pins map[string]struct {
			Registry *struct {
				Kind string `json:"kind"`
			} `json:"registry"`
		} `json:"pins"`
	}
	if raw, err := os.ReadFile(filepath.Join(repo, "automap.lock.json")); err == nil {
		json.Unmarshal(raw, &lock)
	}
	for lib, pin := range lock.Pins {
		if pin.Registry != nil && pin.Registry.Kind == "private" {
			o.PrivateVariants = append(o.PrivateVariants, lib)
		}
	}
	r.Status, r.Reason = grade(t, o)
	return r, used
}

// overlay copies the files under src over dst.
func overlay(src, dst string) {
	filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		data, err := os.ReadFile(p)
		if err == nil {
			os.MkdirAll(filepath.Join(dst, filepath.Dir(rel)), 0o755)
			err = os.WriteFile(filepath.Join(dst, rel), data, 0o644)
		}
		return err
	})
}

// outcome is what a session left: the level verify reached, its summary,
// the manifest's slots and libraries, and the libraries the lock pins to a
// private map's variant.
type outcome struct {
	Level, Summary  string
	Slots           map[string]string
	Libraries       []string
	PrivateVariants []string
}

// grade judges a session's outcome: verify reached the task's level, and
// the manifest puts each library the task names in its slot. It is a pure
// function.
func grade(t *Task, o outcome) (status, reason string) {
	if o.Level != t.Pass.Level {
		return "fail", fmt.Sprintf("verify reports %q, want %s: %s", o.Level, t.Pass.Level, o.Summary)
	}
	for slot, lib := range t.Pass.Slots {
		if o.Slots[slot] != lib {
			return "fail", fmt.Sprintf("the manifest's %s slot holds %q, want %s", slot, o.Slots[slot], lib)
		}
	}
	for _, lib := range t.Pass.Libraries {
		if !slices.Contains(o.Libraries, lib) {
			return "fail", fmt.Sprintf("the composition doesn't have %s", lib)
		}
	}
	for _, lib := range t.Pass.PrivateVariants {
		if !slices.Contains(o.PrivateVariants, lib) {
			return "fail", fmt.Sprintf("the lock doesn't pin %s to a private map's variant", lib)
		}
	}
	return "pass", ""
}
