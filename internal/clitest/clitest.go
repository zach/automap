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
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
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

// Run is one finished run of the CLI.
type Run struct {
	Fixture string // the fixture repo's name
	Dir     string // the temporary copy the CLI ran in
	Stdout  []byte
	Stderr  []byte
	Exit    int
	Result  Result // set when the run was given --json
}

// Automap copies the fixture repo testdata/repos/<fixture> to a temporary
// directory, runs the CLI there with args, and returns the run. When args
// include --json, stdout must be exactly one CLI result that validates
// against the published schema.
func Automap(t *testing.T, fixture string, args ...string) *Run {
	t.Helper()
	bin := binary(t)
	dir := t.TempDir()
	copyTree(t, filepath.Join(Root(t), "testdata", "repos", fixture), dir)

	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	r := &Run{Fixture: fixture, Dir: dir, Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}
	var exit *exec.ExitError
	switch {
	case errors.As(err, &exit):
		r.Exit = exit.ExitCode()
	case err != nil:
		t.Fatalf("running automap %s: %v", strings.Join(args, " "), err)
	}

	for _, a := range args {
		if a == "--json" {
			validateResult(t, r)
			if err := json.Unmarshal(r.Stdout, &r.Result); err != nil {
				t.Fatalf("CLI result: %v", err)
			}
		}
	}
	return r
}

// ExpectExit fails the test unless the CLI exited with status want.
func (r *Run) ExpectExit(t *testing.T, want int) {
	t.Helper()
	if r.Exit != want {
		t.Errorf("exit status %d, want %d\nstdout:\n%s\nstderr:\n%s", r.Exit, want, r.Stdout, r.Stderr)
	}
}

// ExpectGolden compares stdout with testdata/golden/<name>/result.json, and
// the repo after the run with the fixture overlaid by
// testdata/golden/<name>/files/.
func (r *Run) ExpectGolden(t *testing.T, name string) {
	t.Helper()
	golden := filepath.Join(Root(t), "testdata", "golden", name)
	fixture := readTree(t, filepath.Join(Root(t), "testdata", "repos", r.Fixture))
	after := readTree(t, r.Dir)

	if *update {
		if err := os.RemoveAll(golden); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(golden, "result.json"), r.Stdout)
		for path, data := range after {
			if old, ok := fixture[path]; !ok || !bytes.Equal(old, data) {
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

	expected := fixture
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

var cliResultSchema = sync.OnceValues(func() (*jsonschema.Schema, error) {
	return schemas.Compile("cli-result")
})

func validateResult(t *testing.T, r *Run) {
	t.Helper()
	schema, err := cliResultSchema()
	if err != nil {
		t.Fatalf("compiling cli-result schema: %v", err)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(r.Stdout))
	if err != nil {
		t.Fatalf("stdout is not one JSON value: %v\nstdout:\n%s\nstderr:\n%s", err, r.Stdout, r.Stderr)
	}
	if err := schema.Validate(doc); err != nil {
		t.Fatalf("stdout does not conform to the cli-result schema: %v\n%s", err, r.Stdout)
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
