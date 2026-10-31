package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zach/automap/internal/agentfiles"
	"github.com/zach/automap/internal/composition"
	"github.com/zach/automap/internal/mapmodel"
	"github.com/zach/automap/internal/release"
	"github.com/zach/automap/internal/verify"
)

// start creates a composition in the current directory (SPEC §5.4).
func start(a *args, r *result) int {
	for _, f := range []string{composition.ManifestFile, composition.LockFile} {
		if _, err := os.Stat(f); err == nil {
			r.Summary = "A composition already exists here."
			r.fail("COMPOSITION_EXISTS", f+" already exists. Use automap extend to add a library, or automap upgrade to move to a newer map release.")
			return exitRefused
		}
	}
	if _, err := os.Stat("CMakeLists.txt"); err == nil {
		r.Summary = "This repo already has a CMake project."
		r.fail("CMAKE_PROJECT_EXISTS", "start writes a testbed into a repo with no CMake project. For an existing engine, run automap adopt.")
		return exitRefused
	}
	req := composition.Request{Platforms: a.values["platform"]}
	for _, w := range a.values["with"] {
		lib, variant, _ := strings.Cut(w, ":")
		req.With = append(req.With, composition.LibraryRequest{Library: lib, Variant: variant})
	}
	for _, s := range a.values["slot"] {
		slot, lib, ok := strings.Cut(s, "=")
		if !ok {
			r.Summary = "Usage error: --slot takes <slot>=<library>."
			r.fail("USAGE", "--slot "+s+" isn't <slot>=<library>")
			return exitUsage
		}
		req.Slots = append(req.Slots, composition.SlotRequest{Slot: slot, Library: lib})
	}
	if len(req.With) == 0 && len(req.Slots) == 0 {
		r.Summary = "Usage error: start needs at least one library."
		r.fail("USAGE", "automap start --with <library>[:<variant>]... [--slot <slot>=<library>]... [--platform <id>]...")
		return exitUsage
	}

	m, status := loadMap(a, r)
	if m == nil {
		return status
	}
	req.MapRelease = mapRelease(m.PublicDir())
	req.Executable = composition.Executable{Target: composition.TestbedTarget, Sources: testbedSources(m)}

	man, fails := composition.NewManifest(m, req)
	if fails != nil {
		return failWith(r, fails, "The composition can't be made as asked.", exitUsage)
	}
	lock, fails := composition.Resolve(m, man)
	if fails != nil {
		return failWith(r, fails, "The composition can't be resolved against the map.", exitUsage)
	}
	plan, fails := composition.PlanStartup(m, man)
	if fails != nil {
		return failWith(r, fails, "The composition's startup can't be generated.", exitFailed)
	}
	files, fails := composition.Files(m, man, lock, plan, composition.FileOptions{Testbed: true, RegistryURL: registryURL(m)})
	if fails != nil {
		return failWith(r, fails, "The composition's files can't be written.", exitFailed)
	}
	var pinned []string
	for _, lib := range man.Libraries() {
		pin := lock.Pins[lib]
		pinned = append(pinned, fmt.Sprintf("%s %s %s", lib, pin.Version, pin.Variant))
	}
	return writeAndVerify(a, m, man, files, nil, fmt.Sprintf("Started a testbed (%s)", strings.Join(pinned, ", ")), r)
}

// writeAndVerify adds the agent files to a composition's files, writes
// them, and ends with verify on the host platform unless --no-verify says
// otherwise. With keep, a file keep accepts is written only where it
// doesn't exist yet: the developer owns it once written. done begins the
// summary.
func writeAndVerify(a *args, m *mapmodel.Map, man *composition.Manifest, files map[string][]byte, keep func(string) bool, done string, r *result) int {
	agents, err := agentfiles.Files(".", Version, agentfiles.SectionInput{Slots: man.Slots, Executable: man.Executable.Target})
	if err != nil {
		r.Summary = "Reading the repo's agent files failed."
		r.fail("WRITE_FAILED", err.Error())
		return exitFailed
	}
	for p, data := range agents {
		files[p] = data
	}
	changed := map[string][]byte{}
	for p, data := range files {
		existing, err := os.ReadFile(filepath.FromSlash(p))
		switch {
		case err != nil:
			changed[p] = data
		case keep != nil && keep(p):
		case !bytes.Equal(existing, data):
			changed[p] = data
		}
	}
	written, err := writeFiles(changed)
	r.FilesWritten = append(r.FilesWritten, written...)
	if err != nil {
		r.Summary = "Writing the composition failed."
		r.fail("WRITE_FAILED", err.Error())
		return exitFailed
	}

	host := verify.HostPlatform()
	switch {
	case a.bools["no-verify"]:
		r.OK = true
		r.Summary = done + "; it is not verified: run automap verify."
		return exitOK
	case !contains(man.Platforms, host):
		r.OK = true
		r.Summary = done + "; it is not verified: it doesn't target this machine's platform, " + host + "."
		return exitOK
	}
	// Verify the lock as written, with its lock_hash.
	checked, err := composition.CheckLock(files[composition.LockFile])
	if err != nil {
		r.Summary = "The lock just written doesn't check."
		r.fail("INTERNAL", err.Error())
		return exitFailed
	}
	status := runVerify(m, man, checked, host, r)
	r.Summary = done + "; " + strings.TrimSuffix(strings.Replace(r.Summary, "Composition is", "the composition is", 1), ".") + "."
	return status
}

// registryURL is the Automap registry's URL: AUTOMAP_REGISTRY, or the
// public map's.
func registryURL(m *mapmodel.Map) string {
	if u := os.Getenv("AUTOMAP_REGISTRY"); u != "" {
		return u
	}
	return composition.AutomapRegistryURL(m)
}

func sortStrings(s []string) { sort.Strings(s) }

func failWith(r *result, fails []composition.Failure, summary string, status int) int {
	r.Summary = summary
	for _, f := range fails {
		r.fail(f.Code, f.Message)
	}
	return status
}

// testbedSources lists the map's testbed template files other than its
// CMakeLists.txt: the executable's sources.
func testbedSources(m *mapmodel.Map) []string {
	var out []string
	dir := filepath.Join(m.PublicDir(), "testbed")
	filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			rel, _ := filepath.Rel(dir, p)
			if rel = filepath.ToSlash(rel); rel != "CMakeLists.txt" {
				out = append(out, rel)
			}
		}
		return nil
	})
	sort.Strings(out)
	return out
}

// writeFiles writes every file and returns their paths, sorted.
func writeFiles(files map[string][]byte) ([]string, error) {
	var paths []string
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		full := filepath.FromSlash(p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(full, files[p], 0o644); err != nil {
			return nil, err
		}
	}
	return paths, nil
}

// mapRelease records where a map read from a directory came from: its
// commit when it is a clean git checkout, and unreleased otherwise
// (ADR-0079).
func mapRelease(dir string) string {
	// A map read from the release cache, or vendored, is that release.
	if abs, err := filepath.Abs(dir); err == nil {
		if tag := filepath.Base(abs); release.Tag.MatchString(tag) && (abs == release.Dir(tag) || abs == release.VendoredDir(tag)) {
			return tag
		}
	}
	git := func(args ...string) (string, error) {
		out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
		return strings.TrimSpace(string(out)), err
	}
	status, err := git("status", "--porcelain", "--untracked-files=all", "--", ".")
	if err != nil || status != "" {
		return "unreleased"
	}
	commit, err := git("rev-parse", "HEAD")
	if err != nil || len(commit) != 40 {
		return "unreleased"
	}
	return "git:" + commit
}
