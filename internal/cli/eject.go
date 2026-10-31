package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zach/automap/internal/composition"
	"github.com/zach/automap/internal/verify"
)

// EjectedPorts is where eject puts the Automap registry's ports.
const EjectedPorts = "ports"

// eject leaves Automap in one step (ADR-0040): it brings the composition's
// generated files up to date, copies every port the lock pins from the
// Automap registry into ports/, points vcpkg-configuration.json at them
// instead of the registry, and removes the manifest and lock. What is left
// is a plain CMake and vcpkg project; glue, shared interfaces and build
// files stay as they are.
func eject(a *args, r *result) int {
	m, man, lock, status := loadForFetch(a, r)
	if m == nil {
		return status
	}
	written, fails := regenerate(m, man, lock, ".")
	if fails != nil {
		return failWith(r, fails, "The composition's files can't be generated.", exitFailed)
	}
	r.FilesWritten = append(r.FilesWritten, written...)

	// The ports the composition takes from the Automap registry, on any
	// platform, at the commits the lock pins.
	commits := map[string]string{}
	for _, lib := range man.Libraries() {
		pins := []composition.Pin{lock.Pins[lib]}
		for _, p := range lock.Pins[lib].Platforms {
			pins = append(pins, p)
		}
		for _, p := range pins {
			if p.Registry != nil && p.Registry.Kind == "automap" {
				commits[m.Libraries[lib].Newest().BuildBackend.Vcpkg.Port] = p.Registry.Commit
			}
		}
	}
	var ports []string
	for port := range commits {
		ports = append(ports, port)
	}
	sort.Strings(ports)
	url := registryURL(m)
	for _, port := range ports {
		dst := filepath.Join(EjectedPorts, port)
		os.RemoveAll(dst)
		if err := registryPort(url, commits[port], port, dst); err != nil {
			r.Summary = "The Automap registry's ports can't be copied into the repo."
			r.fail("EJECT_FAILED", fmt.Sprintf("copying port %s from %s at %s: %v", port, url, commits[port], err))
			return exitFailed
		}
		r.FilesWritten = append(r.FilesWritten, listFiles(dst)...)
	}

	// vcpkg reads those ports as overlay ports, ahead of any others, and
	// no longer the Automap registry.
	var config map[string]any
	raw, err := os.ReadFile("vcpkg-configuration.json")
	if err == nil {
		err = json.Unmarshal(raw, &config)
	}
	if err != nil {
		r.Summary = "vcpkg-configuration.json can't be read."
		r.fail("EJECT_FAILED", err.Error())
		return exitFailed
	}
	registries, _ := config["registries"].([]any)
	kept := []any{}
	for _, reg := range registries {
		if entry, ok := reg.(map[string]any); !ok || entry["repository"] != url {
			kept = append(kept, reg)
		}
	}
	config["registries"] = kept
	var overlays []any
	for _, port := range ports {
		overlays = append(overlays, path.Join(EjectedPorts, port))
	}
	if existing, ok := config["overlay-ports"].([]any); ok {
		overlays = append(overlays, existing...)
	}
	if len(overlays) > 0 {
		config["overlay-ports"] = overlays
	}
	out, _ := json.MarshalIndent(config, "", "  ")
	if err := os.WriteFile("vcpkg-configuration.json", append(out, '\n'), 0o644); err != nil {
		r.Summary = "vcpkg-configuration.json can't be written."
		r.fail("EJECT_FAILED", err.Error())
		return exitFailed
	}
	if !contains(r.FilesWritten, "vcpkg-configuration.json") {
		r.FilesWritten = append(r.FilesWritten, "vcpkg-configuration.json")
	}
	for _, f := range []string{composition.ManifestFile, composition.LockFile} {
		if err := os.Remove(f); err != nil {
			r.Summary = "The composition's manifest and lock can't be removed."
			r.fail("EJECT_FAILED", err.Error())
			return exitFailed
		}
	}
	r.OK = true
	r.Summary = fmt.Sprintf("Ejected: %d Automap registry %s copied into %s/, and the manifest and lock removed. The repo is a plain CMake and vcpkg project.",
		len(ports), plural(len(ports), "port", "ports"), EjectedPorts)
	return exitOK
}

// registryPort copies a port from a git registry at a commit into dst,
// with its files as committed.
func registryPort(url, commit, port, dst string) error {
	tmp, err := os.MkdirTemp("", "automap-eject-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	git := func(args ...string) *exec.Cmd {
		return exec.Command("git", append([]string{"-C", tmp, "-c", "core.autocrlf=false", "-c", "core.eol=lf"}, args...)...)
	}
	for _, args := range [][]string{{"init", "-q"}, {"fetch", "-q", "--depth", "1", url, commit}} {
		if out, err := git(args...).CombinedOutput(); err != nil {
			return fmt.Errorf("git %s: %v: %s", args[0], err, strings.TrimSpace(string(out)))
		}
	}
	return verify.ArchiveTree(git("archive", "--format=tar", "FETCH_HEAD:ports/"+port), dst)
}
