package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Publish copies every port in ports onto the registry branch checked out
// at branch, commits them, and records each port's version and git tree in
// vcpkg's versions database (versions/<x>-/<port>.json, newest first, and
// versions/baseline.json). It returns the branch's head commit. A port
// whose files changed must carry a new version or port-version.
func Publish(ports, branch string) (string, error) {
	entries, err := os.ReadDir(ports)
	if err != nil {
		return "", err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		names = append(names, e.Name())
		dst := filepath.Join(branch, "ports", e.Name())
		os.RemoveAll(dst)
		if err := copyTree(filepath.Join(ports, e.Name()), dst); err != nil {
			return "", err
		}
	}
	sort.Strings(names)
	if err := commit(branch, "ports", "Ports from main"); err != nil {
		return "", err
	}

	baseline := map[string]map[string]any{}
	if raw, err := os.ReadFile(filepath.Join(branch, "versions", "baseline.json")); err == nil {
		var doc struct{ Default map[string]map[string]any }
		json.Unmarshal(raw, &doc)
		if doc.Default != nil {
			baseline = doc.Default
		}
	}
	for _, port := range names {
		version, portVersion, err := portManifest(filepath.Join(branch, "ports", port, "vcpkg.json"))
		if err != nil {
			return "", fmt.Errorf("%s: %v", port, err)
		}
		tree, err := run(branch, "rev-parse", "HEAD:ports/"+port)
		if err != nil {
			return "", err
		}
		file := filepath.Join(branch, "versions", port[:1]+"-", port+".json")
		var db struct {
			Versions []map[string]any `json:"versions"`
		}
		if raw, err := os.ReadFile(file); err == nil {
			json.Unmarshal(raw, &db)
		}
		known := false
		for _, v := range db.Versions {
			if v["version"] == version && v["port-version"] == float64(portVersion) {
				if v["git-tree"] != tree {
					return "", fmt.Errorf("ports/%s changed, but %s %s#%d is already published at another tree: raise its version or port-version", port, port, version, portVersion)
				}
				known = true
			}
		}
		if !known {
			db.Versions = append([]map[string]any{{"version": version, "port-version": portVersion, "git-tree": tree}}, db.Versions...)
		}
		if err := writeJSON(file, db); err != nil {
			return "", err
		}
		baseline[port] = map[string]any{"baseline": version, "port-version": portVersion}
	}
	if err := writeJSON(filepath.Join(branch, "versions", "baseline.json"), map[string]any{"default": baseline}); err != nil {
		return "", err
	}
	if err := commit(branch, "versions", "Versions database"); err != nil {
		return "", err
	}
	return run(branch, "rev-parse", "HEAD")
}

// portManifest reads a port's version, whichever version field it uses,
// and its port-version.
func portManifest(file string) (string, int, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return "", 0, err
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return "", 0, err
	}
	for _, field := range []string{"version", "version-semver", "version-date", "version-string"} {
		if v, ok := m[field].(string); ok {
			pv, _ := m["port-version"].(float64)
			return v, int(pv), nil
		}
	}
	return "", 0, fmt.Errorf("vcpkg.json has no version")
}

// commit stages a directory and commits it when anything changed.
func commit(dir, path, message string) error {
	if _, err := run(dir, "add", "-A", path); err != nil {
		return err
	}
	if _, err := run(dir, "diff", "--cached", "--quiet"); err == nil {
		return nil
	}
	_, err := run(dir, "commit", "-q", "-m", message)
	return err
}

func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

func writeJSON(file string, v any) error {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(file), 0o755)
	return os.WriteFile(file, append(raw, '\n'), 0o644)
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}
