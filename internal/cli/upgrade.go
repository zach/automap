package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/zach/automap/internal/agentfiles"
	"github.com/zach/automap/internal/composition"
	"github.com/zach/automap/internal/mapmodel"
	"github.com/zach/automap/internal/release"
	"github.com/zach/automap/schemas"
)

// upgrade moves the composition in the current directory to a newer map
// release (SPEC §5.4): it re-resolves the libraries named, or all of them,
// against the new release, updates every file the CLI copied from the map
// that the developer hasn't changed since, leaves changed ones alone with
// the new reference beside them, and ends with verify.
func upgrade(a *args, r *result) int {
	if _, err := os.Stat(manifestFile); os.IsNotExist(err) {
		r.Summary = "Nothing to upgrade: no " + manifestFile + " in the current directory."
		r.fail("NO_MANIFEST", "No composition here. Run automap start to create one, or automap adopt to adopt an existing engine.")
		return exitUsage
	}
	base, lock, status := readComposition(r)
	if base == nil {
		return status
	}
	if !release.Tag.MatchString(lock.MapRelease) {
		r.Summary = "Nothing to upgrade from: the lock names no map release."
		r.fail("NO_MAP", "The lock records map release "+lock.MapRelease+"; upgrade moves a composition from one published map release to another.")
		return exitUsage
	}
	for _, lib := range a.positional {
		if _, ok := lock.Pins[lib]; !ok {
			r.Summary = lib + " isn't in the composition."
			r.fail("UNKNOWN_ENTRY(library "+lib+")", "The composition has no library "+lib+"; automap extend adds one.")
			return exitUsage
		}
	}
	privates, ok := manifestPrivateMaps(r)
	if !ok {
		return exitUsage
	}

	// The map the composition was resolved against, and the one it moves to.
	oldDir, err := release.Open(lock.MapRelease)
	if err != nil {
		r.Summary = "The lock's map release can't be read."
		r.fail("NO_MAP", err.Error())
		return exitUsage
	}
	newDir, ok := upgradeTarget(a, r)
	if !ok {
		return exitUsage
	}
	oldMap, ok := loadMapAt(oldDir, privates, r)
	if !ok {
		return exitUsage
	}
	newMap, ok := loadMapAt(newDir, privates, r)
	if !ok {
		return exitUsage
	}
	locked := map[string]string{}
	for lib, pin := range lock.Pins {
		locked[lib] = pin.Version
	}
	oldMap = oldMap.Hold(locked)
	// Libraries not named stay at their versions.
	hold := map[string]string{}
	if len(a.positional) > 0 {
		for lib, v := range locked {
			if !slices.Contains(a.positional, lib) {
				hold[lib] = v
			}
		}
	}
	newMap = newMap.Hold(hold)

	opts := func(m *mapmodel.Map) composition.FileOptions {
		return composition.FileOptions{Testbed: base.Executable.Target == composition.TestbedTarget, RegistryURL: registryURL(m), Vendored: composition.ReadVendoredPorts(".")}
	}
	oldPlan, fails := composition.PlanStartup(oldMap, base)
	if fails != nil {
		return failWith(r, fails, "The composition doesn't match its own map release.", exitUsage)
	}
	oldFiles, fails := composition.Files(oldMap, base, lock, oldPlan, opts(oldMap))
	if fails != nil {
		return failWith(r, fails, "The composition doesn't match its own map release.", exitUsage)
	}
	tag := mapRelease(newMap.PublicDir())
	man := composition.Upgrade(newMap, base, tag)
	newLock, fails := composition.Resolve(newMap, man)
	if fails != nil {
		return failWith(r, fails, "The composition can't be resolved against "+tag+".", exitUsage)
	}
	plan, fails := composition.PlanStartup(newMap, man)
	if fails != nil {
		return failWith(r, fails, "The composition's startup can't be generated against "+tag+".", exitFailed)
	}
	files, fails := composition.Files(newMap, man, newLock, plan, opts(newMap))
	if fails != nil {
		return failWith(r, fails, "The composition's files can't be written against "+tag+".", exitFailed)
	}

	write, updated, diverged := upgradeFiles(man, oldFiles, files)
	for _, p := range diverged {
		r.warn("GLUE_DIVERGED("+p+")", fmt.Sprintf("%s differs from %s's reference, so upgrade left it as it is; %s's reference is beside it at %s.reference. Carry over what applies, using the recipe's adaptation notes (automap show recipe <id>).", p, lock.MapRelease, tag, p))
	}

	var moved []string
	for _, lib := range man.Libraries() {
		from, to := lock.Pins[lib], newLock.Pins[lib]
		if from.Version != to.Version || from.Variant != to.Variant {
			moved = append(moved, fmt.Sprintf("%s %s → %s", lib, from.Version, to.Version))
		}
	}
	done := "Upgraded to " + tag
	if len(moved) > 0 {
		done += " (" + strings.Join(moved, ", ") + ")"
	}
	done += fmt.Sprintf("; %d %s updated, %d diverged", updated, plural(updated, "file", "files"), len(diverged))
	return writeAndVerify(a, newMap, man, write, nil, done, r)
}

// upgradeTarget is the directory of the map to upgrade to: a map directory
// when one is given, the release --map-release names, or the newest.
func upgradeTarget(a *args, r *result) (string, bool) {
	dir := a.value("map")
	if dir == "" {
		dir = os.Getenv("AUTOMAP_MAP")
	}
	if dir != "" {
		return dir, true
	}
	var err error
	if tag := a.value("map-release"); tag != "" {
		if !release.Tag.MatchString(tag) {
			r.Summary = "Usage error: --map-release takes map-YYYY.MM.DD."
			r.fail("USAGE", tag+" isn't a map release name")
			return "", false
		}
		dir, err = release.Open(tag)
	} else {
		_, dir, err = release.Newest()
	}
	if err != nil {
		r.Summary = "The map release to upgrade to can't be read."
		r.fail("NO_MAP", err.Error())
		return "", false
	}
	return dir, true
}

// upgradeFiles decides which of the composition's files an upgrade writes.
// The files the CLI generates from the manifest are rewritten. A file the
// CLI copied from the map is updated to the new release's version when it
// is missing or still the old release's; when the developer has changed it
// and the reference changed too, it is left alone and the new reference
// is written beside it as <file>.reference. It returns the files to write,
// how many copied files were updated, and the diverged ones. It reads the
// repo's files; the decision is otherwise a pure function.
func upgradeFiles(man *composition.Manifest, before, after map[string][]byte) (write map[string][]byte, updated int, diverged []string) {
	generated := map[string]bool{}
	for _, p := range generatedFiles(man) {
		generated[p] = true
	}
	write = map[string][]byte{}
	paths := make([]string, 0, len(after))
	for p := range after {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		next := after[p]
		if generated[p] {
			write[p] = next
			continue
		}
		current, err := os.ReadFile(filepath.FromSlash(p))
		prev, known := before[p]
		switch {
		case err != nil:
			write[p] = next
			updated++
		case bytes.Equal(current, next):
		case known && bytes.Equal(current, prev):
			write[p] = next
			updated++
		case known && bytes.Equal(prev, next):
			// The developer changed it; the reference didn't change.
		default:
			write[p+".reference"] = next
			diverged = append(diverged, p)
		}
	}
	return write, updated, diverged
}

// readComposition reads the manifest and lock in the current directory and
// checks them, without reading the map.
func readComposition(r *result) (*composition.Manifest, *composition.Lock, int) {
	raw, err := os.ReadFile(manifestFile)
	var doc any
	if err == nil {
		err = json.Unmarshal(raw, &doc)
	}
	if err != nil {
		r.Summary = "The manifest can't be read."
		r.fail("MANIFEST_INVALID", err.Error())
		return nil, nil, exitUsage
	}
	if v := schemas.Validate("manifest", doc); v != nil {
		r.Summary = "The manifest doesn't match its schema."
		for _, msg := range v {
			r.fail("MANIFEST_INVALID", "manifest: "+msg)
		}
		return nil, nil, exitUsage
	}
	var man composition.Manifest
	json.Unmarshal(raw, &man)
	rawLock, err := os.ReadFile(composition.LockFile)
	if err != nil {
		r.Summary = "The composition has no lock."
		r.fail("NO_LOCK", composition.LockFile+" is missing. Only the CLI writes the lock: restore it from version control.")
		return nil, nil, exitUsage
	}
	lock, err := composition.CheckLock(rawLock)
	if err != nil {
		r.Summary = "Refusing to act on a hand-edited lock."
		r.fail("LOCK_EDITED", composition.LockFile+" was edited by hand: "+err.Error()+". "+agentfiles.Never(1))
		return nil, nil, exitRefused
	}
	return &man, lock, exitOK
}

// generatedFiles lists the files the CLI generates from the manifest and
// the lock, which it rewrites rather than merges.
func generatedFiles(man *composition.Manifest) []string {
	out := []string{composition.ManifestFile, composition.LockFile, "vcpkg.json", "vcpkg-configuration.json",
		"cmake/automap/automap.cmake", man.Glue.ChoicesHeader}
	if man.Glue.Startup.DeveloperOwned == nil {
		for _, f := range []string{"context.h", "startup.h", "startup.cpp"} {
			out = append(out, path.Join(man.Glue.Interfaces, f))
		}
	}
	return out
}
