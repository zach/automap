package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.invalid", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.invalid")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// identity gives Publish's commits an author, as the workflow does.
func identity(t *testing.T) {
	for _, k := range []string{"GIT_AUTHOR_NAME", "GIT_COMMITTER_NAME"} {
		t.Setenv(k, "t")
	}
	for _, k := range []string{"GIT_AUTHOR_EMAIL", "GIT_COMMITTER_EMAIL"} {
		t.Setenv(k, "t@example.invalid")
	}
}

func writePort(t *testing.T, ports, name, manifest, portfile string) {
	t.Helper()
	os.MkdirAll(filepath.Join(ports, name), 0o755)
	os.WriteFile(filepath.Join(ports, name, "vcpkg.json"), []byte(manifest), 0o644)
	os.WriteFile(filepath.Join(ports, name, "portfile.cmake"), []byte(portfile), 0o644)
}

func versions(t *testing.T, branch, port string) []map[string]any {
	t.Helper()
	var db struct{ Versions []map[string]any }
	raw, _ := os.ReadFile(filepath.Join(branch, "versions", port[:1]+"-", port+".json"))
	json.Unmarshal(raw, &db)
	return db.Versions
}

// Publishing copies the ports onto the registry branch and records each
// port's git tree in vcpkg's versions database, newest first.
func TestPublishRecordsEachPortVersion(t *testing.T) {
	identity(t)
	ports, branch := t.TempDir(), t.TempDir()
	git(t, branch, "init", "-q", "-b", "registry")
	writePort(t, ports, "joltphysics", `{"name": "joltphysics", "version": "5.6.0"}`, "# one\n")

	commit, err := Publish(ports, branch)
	if err != nil {
		t.Fatal(err)
	}
	if commit != git(t, branch, "rev-parse", "HEAD") {
		t.Errorf("Publish returned %s, want the branch's head", commit)
	}
	first := versions(t, branch, "joltphysics")
	if len(first) != 1 || first[0]["version"] != "5.6.0" || first[0]["git-tree"] != git(t, branch, "rev-parse", commit+":ports/joltphysics") {
		t.Fatalf("versions %v, want 5.6.0 at the port's tree", first)
	}
	var baseline struct{ Default map[string]map[string]any }
	raw, _ := os.ReadFile(filepath.Join(branch, "versions", "baseline.json"))
	json.Unmarshal(raw, &baseline)
	if b := baseline.Default["joltphysics"]; b["baseline"] != "5.6.0" || b["port-version"] != 0.0 {
		t.Errorf("baseline %v, want joltphysics 5.6.0#0", baseline.Default)
	}

	writePort(t, ports, "joltphysics", `{"name": "joltphysics", "version": "5.6.0", "port-version": 1}`, "# two\n")
	if _, err := Publish(ports, branch); err != nil {
		t.Fatal(err)
	}
	if second := versions(t, branch, "joltphysics"); len(second) != 2 || second[0]["port-version"] != 1.0 || second[1]["git-tree"] != first[0]["git-tree"] {
		t.Errorf("versions %v, want 5.6.0#1 first and 5.6.0#0 kept", second)
	}
}

// A changed port must carry a new version or port-version: vcpkg finds a
// version's files by the tree it recorded.
func TestPublishRefusesAChangedPortWithItsOldVersion(t *testing.T) {
	identity(t)
	ports, branch := t.TempDir(), t.TempDir()
	git(t, branch, "init", "-q", "-b", "registry")
	writePort(t, ports, "joltphysics", `{"name": "joltphysics", "version": "5.6.0"}`, "# one\n")
	if _, err := Publish(ports, branch); err != nil {
		t.Fatal(err)
	}
	writePort(t, ports, "joltphysics", `{"name": "joltphysics", "version": "5.6.0"}`, "# changed\n")

	if _, err := Publish(ports, branch); err == nil || !strings.Contains(err.Error(), "joltphysics 5.6.0#0") {
		t.Errorf("err %v, want a refusal naming joltphysics 5.6.0#0", err)
	}
}

// Publishing unchanged ports changes nothing.
func TestPublishingTwiceIsANoOp(t *testing.T) {
	identity(t)
	ports, branch := t.TempDir(), t.TempDir()
	git(t, branch, "init", "-q", "-b", "registry")
	writePort(t, ports, "joltphysics", `{"name": "joltphysics", "version": "5.6.0"}`, "# one\n")
	first, _ := Publish(ports, branch)

	second, err := Publish(ports, branch)

	if err != nil || second != first {
		t.Errorf("second publish: %s, %v; want the same commit %s", second, err, first)
	}
}
