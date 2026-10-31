// Package clitest runs the automap CLI as a process against fixture repos
// and checks what it prints and writes against golden files and the
// published CLI result schema (ADR-0077).
//
// Fixture repos live in testdata/repos/<name>/. A case's golden files live
// in testdata/golden/<case>/: result.json is the expected stdout, and files/
// holds every file the run is expected to add or change in the repo. Run
// `go test ./... -update` to rewrite golden files from the CLI's output.
package clitest

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/zach/automap/schemas"
)

var update = flag.Bool("update", false, "rewrite golden files from the CLI's output")

// Result is a CLI result as the published schema defines it.
type Result struct {
	Command      string    `json:"command"`
	CLIVersion   string    `json:"cli_version"`
	OK           bool      `json:"ok"`
	Summary      string    `json:"summary"`
	Level        string    `json:"level"`
	Warnings     []Warning `json:"warnings"`
	Failures     []Failure `json:"failures"`
	FilesWritten []string  `json:"files_written"`
	EvidenceFile string    `json:"evidence_file"`
	Entries      []Entry   `json:"entries"`
}

// Entry is one map entry in a find or show result.
type Entry map[string]any

// Get walks a path of object keys and array indexes through the entry.
func (e Entry) Get(path ...any) any {
	var v any = map[string]any(e)
	for _, k := range path {
		switch k := k.(type) {
		case string:
			m, _ := v.(map[string]any)
			v = m[k]
		case int:
			a, _ := v.([]any)
			if k >= len(a) {
				return nil
			}
			v = a[k]
		}
	}
	return v
}

type Warning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Failure struct {
	Code        string `json:"code"`
	Rule        string `json:"rule"`
	Library     string `json:"library"`
	Message     string `json:"message"`
	DetailsFile string `json:"details_file"`
}

// Repo is a temporary copy of a fixture repo that a test prepares, then
// runs the CLI in.
type Repo struct {
	Dir string
	env []string
}

// Run is one finished run of the CLI.
type Run struct {
	Dir      string            // the repo the CLI ran in
	Before   map[string][]byte // the repo's files just before the run
	Stdout   []byte
	Stderr   []byte
	Exit     int
	Duration time.Duration
	Result   Result // set when the run was given --json
}

// NewRepo copies the fixture repo testdata/repos/<fixture> to a temporary
// directory.
func NewRepo(t *testing.T, fixture string) *Repo {
	t.Helper()
	dir := t.TempDir()
	copyTree(t, filepath.Join(Root(t), "testdata", "repos", fixture), dir)
	return &Repo{Dir: dir}
}

// Copy copies a file or directory from the module root (such as
// testdata/map) to dst in the repo.
func (r *Repo) Copy(t *testing.T, src, dst string) {
	t.Helper()
	copyTree(t, filepath.Join(Root(t), filepath.FromSlash(src)), filepath.Join(r.Dir, filepath.FromSlash(dst)))
}

// FixtureMap assembles the fixture map at dst in the repo: the stub
// libraries in testdata/map with the public map's contracts and probe
// harness, so the CLI's tests exercise the real shared interfaces.
func (r *Repo) FixtureMap(t *testing.T, dst string) {
	t.Helper()
	r.Copy(t, "testdata/map", dst)
	r.Copy(t, "map/contracts", dst+"/contracts")
	r.Copy(t, "map/probe-harness", dst+"/probe-harness")
}

// EditJSON decodes the JSON file at path in the repo, passes it to edit, and
// writes back what edit returns.
func (r *Repo) EditJSON(t *testing.T, path string, edit func(doc any) any) {
	t.Helper()
	full := filepath.Join(r.Dir, filepath.FromSlash(path))
	raw, err := os.ReadFile(full)
	if err != nil {
		t.Fatal(err)
	}
	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	out, err := json.MarshalIndent(edit(doc), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, full, append(out, '\n'))
}

// WriteFile writes data to path in the repo.
func (r *Repo) WriteFile(t *testing.T, path string, data string) {
	t.Helper()
	writeFile(t, filepath.Join(r.Dir, filepath.FromSlash(path)), []byte(data))
}

// Setenv sets an environment variable for the repo's runs.
func (r *Repo) Setenv(key, value string) {
	r.env = append(r.env, key+"="+value)
}

// Automap copies the fixture repo and runs the CLI in it with args.
func Automap(t *testing.T, fixture string, args ...string) *Run {
	t.Helper()
	return NewRepo(t, fixture).Automap(t, args...)
}

// Automap runs the CLI in the repo with args and returns the run. When args
// include --json, stdout must be exactly one CLI result that validates
// against the published schema.
func (r *Repo) Automap(t *testing.T, args ...string) *Run {
	t.Helper()
	bin := binary(t)
	run := &Run{Dir: r.Dir, Before: readTree(t, r.Dir)}

	cmd := exec.Command(bin, args...)
	cmd.Dir = r.Dir
	// The CLI never reaches the published map or its evidence from a test:
	// releases and evidence come from nowhere, and fetches go to a cache of
	// the repo's own, unless the test sets them.
	offline := []string{"AUTOMAP_RELEASES=http://127.0.0.1:1", "AUTOMAP_EVIDENCE_URL=http://127.0.0.1:1", "AUTOMAP_CACHE=" + filepath.Join(r.Dir, ".automap-test-cache")}
	cmd.Env = append(append(cleanEnv(), offline...), r.env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	began := time.Now()
	err := cmd.Run()
	run.Duration = time.Since(began)
	run.Stdout, run.Stderr = stdout.Bytes(), stderr.Bytes()
	var exit *exec.ExitError
	switch {
	case errors.As(err, &exit):
		run.Exit = exit.ExitCode()
	case err != nil:
		t.Fatalf("running automap %s: %v", strings.Join(args, " "), err)
	}

	for _, a := range args {
		if a == "--json" {
			validateResult(t, run)
			if err := json.Unmarshal(run.Stdout, &run.Result); err != nil {
				t.Fatalf("CLI result: %v", err)
			}
			if run.Result.OK != (len(run.Result.Failures) == 0) {
				t.Errorf("ok is %v with %d failures; ok must be true exactly when there are none", run.Result.OK, len(run.Result.Failures))
			}
		}
	}
	return run
}

// cleanEnv is the test process's environment without AUTOMAP_ variables, so
// a developer's own settings never reach the CLI under test, and without
// the CI's identity (GITHUB_ and ACTIONS_ variables), so tests that run in
// CI see what they see locally unless they set it.
func cleanEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "AUTOMAP_") && !strings.HasPrefix(kv, "GITHUB_") && !strings.HasPrefix(kv, "ACTIONS_") {
			env = append(env, kv)
		}
	}
	return env
}

// ReadFile returns a file in the repo after the run, or fails the test.
func (r *Run) ReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(r.Dir, filepath.FromSlash(path)))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// ExpectUnchanged fails the test if the run added, changed or removed any
// file in the repo.
func (r *Run) ExpectUnchanged(t *testing.T) {
	t.Helper()
	after := readTree(t, r.Dir)
	for _, p := range union(r.Before, after) {
		if !bytes.Equal(r.Before[p], after[p]) {
			t.Errorf("%s changed, but the run should write nothing", p)
		}
	}
}

// Read returns a file in the repo, or fails the test.
func Read(t *testing.T, r *Repo, path string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(r.Dir, filepath.FromSlash(path)))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// ExpectSchema fails the test unless raw is JSON valid against the named
// published schema.
func ExpectSchema(t *testing.T, name, raw string) {
	t.Helper()
	var doc any
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	if v := schemas.Validate(name, doc); v != nil {
		t.Errorf("not valid against the %s schema: %s", name, strings.Join(v, "; "))
	}
}

// LogDetails logs the last lines of every failure's details file, so a
// failing run in CI shows why.
func (r *Run) LogDetails(t *testing.T) {
	t.Helper()
	for _, f := range r.Result.Failures {
		if f.DetailsFile == "" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(r.Dir, filepath.FromSlash(f.DetailsFile)))
		if err != nil {
			continue
		}
		lines := strings.Split(string(raw), "\n")
		var errs []string
		for _, l := range lines {
			if ll := strings.ToLower(l); strings.Contains(ll, "error") || strings.Contains(ll, "threadsanitizer") {
				errs = append(errs, l)
			}
		}
		if len(lines) > 40 {
			lines = lines[len(lines)-40:]
		}
		t.Logf("%s (%s), lines mentioning errors:\n%s\nlast lines:\n%s", f.DetailsFile, f.Code, strings.Join(errs, "\n"), strings.Join(lines, "\n"))
	}
}

// ExpectFailure fails the test unless the run's result has a failure with
// code whose message contains message.
func (r *Run) ExpectFailure(t *testing.T, code, message string) {
	t.Helper()
	for _, f := range r.Result.Failures {
		if f.Code == code && strings.Contains(f.Message, message) {
			return
		}
	}
	t.Errorf("no failure %s with message containing %q; failures:\n%s", code, message, r.Stdout)
}

// ExpectExit fails the test unless the CLI exited with status want.
func (r *Run) ExpectExit(t *testing.T, want int) {
	t.Helper()
	if r.Exit != want {
		t.Errorf("exit status %d, want %d\nstdout:\n%s\nstderr:\n%s", r.Exit, want, r.Stdout, r.Stderr)
	}
}

// ExpectGolden compares stdout with testdata/golden/<name>/result.json, and
// the repo after the run with the repo before it overlaid by
// testdata/golden/<name>/files/.
func (r *Run) ExpectGolden(t *testing.T, name string) {
	t.Helper()
	golden := filepath.Join(Root(t), "testdata", "golden", name)
	// Build output and evidence change with every run: the result names
	// them, and the golden files leave them out.
	varies := func(tree map[string][]byte) map[string][]byte {
		out := map[string][]byte{}
		for p, data := range tree {
			if !strings.HasPrefix(p, "build/") && !strings.HasPrefix(p, "evidence/") {
				out[p] = data
			}
		}
		return out
	}
	before := varies(r.Before)
	after := varies(readTree(t, r.Dir))

	if *update {
		if err := os.RemoveAll(golden); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(golden, "result.json"), r.Stdout)
		for path, data := range after {
			if old, ok := before[path]; !ok || !bytes.Equal(old, data) {
				writeFile(t, filepath.Join(golden, "files", path), data)
			}
		}
		return
	}

	want, err := os.ReadFile(filepath.Join(golden, "result.json"))
	if err != nil {
		t.Fatalf("golden result: %v (run with -update to create it)", err)
	}
	if !bytes.Equal(r.Stdout, want) {
		t.Errorf("stdout differs from %s\ngot:\n%s\nwant:\n%s", rel(t, golden, "result.json"), r.Stdout, want)
	}

	expected := map[string][]byte{}
	for path, data := range before {
		expected[path] = data
	}
	if written, err := os.Stat(filepath.Join(golden, "files")); err == nil && written.IsDir() {
		for path, data := range readTree(t, filepath.Join(golden, "files")) {
			expected[path] = data
		}
	}
	for _, path := range union(expected, after) {
		want, inWant := expected[path]
		got, inGot := after[path]
		switch {
		case !inGot:
			t.Errorf("%s: expected in the repo after the run, missing", path)
		case !inWant:
			t.Errorf("%s: written by the run, not in %s", path, rel(t, golden, "files"))
		case !bytes.Equal(want, got):
			t.Errorf("%s differs from golden\ngot:\n%s\nwant:\n%s", path, got, want)
		}
	}
}

var (
	buildOnce sync.Once
	binPath   string
	buildErr  error
)

// binary builds cmd/automap once per test process.
func binary(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		dir, err := os.MkdirTemp("", "automap-bin-")
		if err != nil {
			buildErr = err
			return
		}
		binPath = filepath.Join(dir, "automap")
		if runtime.GOOS == "windows" {
			binPath += ".exe"
		}
		cmd := exec.Command("go", "build", "-o", binPath, "./cmd/automap")
		cmd.Dir = Root(t)
		if out, err := cmd.CombinedOutput(); err != nil {
			buildErr = fmt.Errorf("go build: %v\n%s", err, out)
		}
	})
	if buildErr != nil {
		t.Fatal(buildErr)
	}
	return binPath
}

func validateResult(t *testing.T, r *Run) {
	t.Helper()
	var doc any
	if err := json.Unmarshal(r.Stdout, &doc); err != nil {
		t.Fatalf("stdout is not one JSON value: %v\nstdout:\n%s\nstderr:\n%s", err, r.Stdout, r.Stderr)
	}
	if v := schemas.Validate("cli-result", doc); v != nil {
		t.Fatalf("stdout does not conform to the cli-result schema: %s\n%s", strings.Join(v, "; "), r.Stdout)
	}
}

// Root returns the module root, the directory holding go.mod.
func Root(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found above the test's directory")
		}
		dir = parent
	}
}

// readTree returns every regular file under root by slash-separated
// relative path.
func readTree(t *testing.T, root string) map[string][]byte {
	t.Helper()
	tree := map[string][]byte{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		tree[filepath.ToSlash(relPath)] = data
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return tree
}

func copyTree(t *testing.T, from, to string) {
	t.Helper()
	if _, err := os.Stat(from); err != nil {
		t.Fatalf("fixture repo: %v", err)
	}
	for path, data := range readTree(t, from) {
		writeFile(t, filepath.Join(to, filepath.FromSlash(path)), data)
	}
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func union(a, b map[string][]byte) []string {
	seen := map[string]bool{}
	for k := range a {
		seen[k] = true
	}
	for k := range b {
		seen[k] = true
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func rel(t *testing.T, elem ...string) string {
	r, err := filepath.Rel(Root(t), filepath.Join(elem...))
	if err != nil {
		return filepath.Join(elem...)
	}
	return filepath.ToSlash(r)
}
