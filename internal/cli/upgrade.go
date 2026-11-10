package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/zach/automap/internal/agentfiles"
	"github.com/zach/automap/internal/composition"
	"github.com/zach/automap/internal/mapmodel"
	"github.com/zach/automap/internal/release"
	"github.com/zach/automap/internal/verify"
	"github.com/zach/automap/schemas"
)

// upgrade moves the composition in the current directory to a newer map
// release (SPEC §5.4): it re-resolves the libraries named, or all of them,
// against the new release, updates every file the CLI copied from the map
// that the developer hasn't changed since, leaves changed ones alone with
// the new reference beside them, and ends with verify.
func upgrade(a *args, r *result) int {
	base, lock, privates, status := openComposition(r, "upgrade")
	if base == nil {
		return status
	}
	if !release.Tag.MatchString(lock.MapRelease) {
		r.Summary = "Nothing to upgrade from: the lock names no map release."
		r.fail("NO_MAP", "The lock records map release "+lock.MapRelease+"; upgrade moves a composition from one published map release to another. automap resolve re-resolves against a map directory.")
		return exitUsage
	}
	for _, lib := range a.positional {
		if _, ok := lock.Pins[lib]; !ok {
			r.Summary = lib + " isn't in the composition."
			r.fail("UNKNOWN_ENTRY(library "+lib+")", "The composition has no library "+lib+"; automap extend adds one.")
			return exitUsage
		}
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
	newMap, ok := loadMapAt(newDir, privates, r)
	if !ok {
		return exitUsage
	}
	// Libraries not named stay at their versions.
	hold := map[string]string{}
	if len(a.positional) > 0 {
		for lib, pin := range lock.Pins {
			if !slices.Contains(a.positional, lib) {
				hold[lib] = pin.Version
			}
		}
	}
	newMap = newMap.Hold(hold)
	old, ok := lockedFiles(base, lock, oldDir, privates, r, "upgrade")
	if !ok {
		return exitUsage
	}
	tag := mapRelease(newMap.PublicDir())
	return recompose(a, r, lock, old, newMap, composition.Upgrade(newMap, base, tag), "Upgraded to "+tag, "upgrade")
}

// resolve re-resolves the composition against the map release its lock
// records, after the manifest or a private map changed (SPEC §5.4): the
// release and every library's version stay where they are; the lock, the
// files the CLI generates and the glue the developer hasn't changed are
// brought up to date; and it ends with verify.
func resolve(a *args, r *result) int {
	base, lock, privates, status := openComposition(r, "resolve")
	if base == nil {
		return status
	}
	if len(a.positional) > 0 {
		r.Summary = "Usage error: resolve takes no arguments."
		r.fail("USAGE", "automap resolve re-resolves the whole composition; automap upgrade <library> moves one library to a newer version.")
		return exitUsage
	}
	m, old, ok := lockedMap(a, r, base, lock, privates, "resolve")
	if !ok {
		return exitUsage
	}
	tag := mapRelease(m.PublicDir())
	return recompose(a, r, lock, old, m, composition.Upgrade(m, base, tag), "Resolved against "+tag, "resolve")
}

// remove takes one library out of the composition (SPEC §5.4), the other
// half of extend: it empties the library's slots, drops its variant and
// glue entries, resets what depended on it, re-resolves, deletes the files
// the CLI copied for it that the developer hasn't changed, and ends with
// verify.
func remove(a *args, r *result) int {
	if len(a.positional) != 1 {
		r.Summary = "Usage error: remove takes one library."
		r.fail("USAGE", "automap remove <library>")
		return exitUsage
	}
	base, lock, privates, status := openComposition(r, "remove")
	if base == nil {
		return status
	}
	lib := a.positional[0]
	m, old, ok := lockedMap(a, r, base, lock, privates, "remove")
	if !ok {
		return exitUsage
	}
	man, slots, fails := composition.Remove(m, base, lib)
	if fails != nil {
		return failWith(r, fails, lib+" can't be removed.", exitUsage)
	}
	man.MapRelease = mapRelease(m.PublicDir())
	return recompose(a, r, lock, old, m, man, fmt.Sprintf("Removed %s from %s", lib, strings.Join(slots, ", ")), "remove")
}

// openComposition reads the manifest and lock in the current directory and
// the private maps the manifest names, for a command that changes the
// composition.
func openComposition(r *result, command string) (*composition.Manifest, *composition.Lock, []mapmodel.PrivateMap, int) {
	if _, err := os.Stat(manifestFile); os.IsNotExist(err) {
		r.Summary = "Nothing to " + command + ": no " + manifestFile + " in the current directory."
		r.fail("NO_MANIFEST", "No composition here. Run automap start to create one, or automap adopt to adopt an existing engine.")
		return nil, nil, nil, exitUsage
	}
	base, lock, status := readComposition(r)
	if base == nil {
		return nil, nil, nil, status
	}
	privates, ok := manifestPrivateMaps(r)
	if !ok {
		return nil, nil, nil, exitUsage
	}
	return base, lock, privates, exitOK
}

// lockedMap loads the map the lock was resolved against, with every
// library held at the version the lock pins, and the composition's files
// as the lock pinned them: what resolve and remove start from.
func lockedMap(a *args, r *result, base *composition.Manifest, lock *composition.Lock, privates []mapmodel.PrivateMap, command string) (*mapmodel.Map, map[string][]byte, bool) {
	dir := a.value("map")
	if dir == "" {
		dir = os.Getenv("AUTOMAP_MAP")
	}
	if dir == "" {
		var ok bool
		if dir, ok = releaseDir(r); !ok {
			return nil, nil, false
		}
	}
	m, ok := loadMapAt(dir, privates, r)
	if !ok {
		return nil, nil, false
	}
	hold := map[string]string{}
	for lib, pin := range lock.Pins {
		hold[lib] = pin.Version
	}
	old, ok := lockedFiles(base, lock, dir, privates, r, command)
	if !ok {
		return nil, nil, false
	}
	return m.Hold(hold), old, true
}

// fileOptions are the options for writing a composition's files.
func fileOptions(m *mapmodel.Map, man *composition.Manifest) composition.FileOptions {
	return composition.FileOptions{Testbed: man.Executable.Target == composition.TestbedTarget, RegistryURL: registryURL(m), Vendored: composition.ReadVendoredPorts(".")}
}

// lockedFiles is the composition's files as the lock pinned them: the
// public map in oldDir with each private map at the commit the lock
// records, so a private recipe's changed reference shows as a change. A
// private map that can't be read at its commit, or doesn't describe the
// composition there, is read as it is now.
func lockedFiles(base *composition.Manifest, lock *composition.Lock, oldDir string, privates []mapmodel.PrivateMap, r *result, command string) (map[string][]byte, bool) {
	locked := map[string]string{}
	for lib, pin := range lock.Pins {
		locked[lib] = pin.Version
	}
	filesWith := func(privates []mapmodel.PrivateMap, r *result) (map[string][]byte, bool) {
		oldMap, ok := loadMapAt(oldDir, privates, r)
		if !ok {
			return nil, false
		}
		oldMap = oldMap.Hold(locked)
		oldPlan, fails := composition.PlanStartup(oldMap, base)
		if fails == nil {
			var files map[string][]byte
			if files, fails = composition.Files(oldMap, base, lock, oldPlan, fileOptions(oldMap, base)); fails == nil {
				return files, true
			}
		}
		failWith(r, fails, "", exitUsage)
		return nil, false
	}
	pinned, cleanup := lockedPrivateMaps(privates, lock)
	defer cleanup()
	files, ok := filesWith(pinned, newResult(r.Command))
	if !ok {
		if files, ok = filesWith(privates, r); !ok {
			r.Summary = "The composition can't be regenerated as the lock pinned it (" + lockedMaps(lock) + "), so " + command + " can't tell which files have changed since."
			return nil, false
		}
	}
	return files, true
}

// recompose writes a changed composition: it resolves man against m,
// rewrites the files the CLI generates, updates each file copied from the
// map that still matches old (the files as the lock pinned them), leaves a
// changed one alone with the new reference beside it, deletes the files
// the composition no longer has where they still match old, and ends with
// verify. done begins the summary; command names the caller in warnings.
func recompose(a *args, r *result, lock *composition.Lock, old map[string][]byte, m *mapmodel.Map, man *composition.Manifest, done, command string) int {
	tag := man.MapRelease
	newLock, fails := composition.Resolve(m, man)
	if fails != nil {
		return failWith(r, fails, "The composition can't be resolved against "+tag+".", exitUsage)
	}
	plan, fails := composition.PlanStartup(m, man)
	if fails != nil {
		return failWith(r, fails, "The composition's startup can't be generated against "+tag+".", exitFailed)
	}
	files, fails := composition.Files(m, man, newLock, plan, fileOptions(m, man))
	if fails != nil {
		return failWith(r, fails, "The composition's files can't be written against "+tag+".", exitFailed)
	}

	write, updated, diverged := upgradeFiles(man, old, files)
	for _, p := range diverged {
		r.warn("GLUE_DIVERGED("+p+")", fmt.Sprintf("%s differs from the reference it was copied from, so %s left it as it is; the new reference is beside it at %s.reference. Carry over what applies, using the recipe's adaptation notes (automap show recipe <id>).", p, command, p))
	}
	// Files the composition had and no longer has, such as a removed
	// library's glue: the CLI's to delete where the developer hasn't
	// changed them.
	var gone []string
	for p := range old {
		if _, kept := files[p]; !kept {
			gone = append(gone, p)
		}
	}
	sort.Strings(gone)
	var deleted int
	for _, p := range gone {
		current, err := os.ReadFile(filepath.FromSlash(p))
		switch {
		case err != nil:
		case bytes.Equal(current, old[p]):
			if os.Remove(filepath.FromSlash(p)) == nil {
				deleted++
				// And its directories, where it was the last file in them.
				for dir := path.Dir(p); dir != "." && os.Remove(filepath.FromSlash(dir)) == nil; dir = path.Dir(dir) {
				}
			}
		default:
			r.warn("GLUE_LEFT("+p+")", fmt.Sprintf("%s is no longer part of the composition, but it differs from the reference it was copied from, so %s left it. Delete it if nothing of yours needs it.", p, command))
		}
	}

	var moved []string
	for _, lib := range man.Libraries() {
		from, to := lock.Pins[lib], newLock.Pins[lib]
		if from.Version != "" && (from.Version != to.Version || from.Variant != to.Variant) {
			change := fmt.Sprintf("%s %s → %s", lib, from.Version, to.Version)
			if from.Version == to.Version {
				change = fmt.Sprintf("%s %s → %s", lib, from.Variant, to.Variant)
			}
			moved = append(moved, change)
		}
	}
	if len(moved) > 0 {
		done += " (" + strings.Join(moved, ", ") + ")"
	}
	done += fmt.Sprintf("; %d %s updated, %d diverged", updated, plural(updated, "file", "files"), len(diverged))
	if deleted > 0 {
		done += fmt.Sprintf(", %d deleted", deleted)
	}
	return writeAndVerify(a, m, man, write, nil, done, r)
}

// lockedPrivateMaps is each private map as the lock pinned it: its files at
// the commit the lock records, in a scratch directory. A map the lock has
// no commit for, or whose commit git can't read, stays where it is. The
// caller runs cleanup when done.
func lockedPrivateMaps(privates []mapmodel.PrivateMap, lock *composition.Lock) ([]mapmodel.PrivateMap, func()) {
	scratch, err := os.MkdirTemp("", "automap-private-")
	if err != nil {
		return privates, func() {}
	}
	out := slices.Clone(privates)
	for i, pm := range out {
		commit := lock.PrivateMaps[pm.ID].Commit
		if commit == "" {
			continue
		}
		// The map's directory within its repository, and the repository's
		// root, where git archive reads a tree's path from.
		where, err := exec.Command("git", "-C", pm.Location, "rev-parse", "--show-prefix", "--show-toplevel").Output()
		lines := strings.Split(strings.TrimRight(string(where), "\n"), "\n")
		if err != nil || len(lines) != 2 {
			continue
		}
		tree := commit
		if dir := strings.TrimSuffix(lines[0], "/"); dir != "" {
			tree += ":" + dir
		}
		dst := filepath.Join(scratch, fmt.Sprint(i))
		archive := exec.Command("git", "-C", lines[1], "-c", "core.autocrlf=false", "archive", "--format=tar", tree)
		if verify.ArchiveTree(archive, dst) == nil {
			out[i].Location = dst
		}
	}
	return out, func() { os.RemoveAll(scratch) }
}

// lockedMaps names the maps a lock was resolved against.
func lockedMaps(lock *composition.Lock) string {
	out := "map release " + lock.MapRelease
	var ids []string
	for id := range lock.PrivateMaps {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		out += ", private map " + id + " at " + lock.PrivateMaps[id].Commit[:min(7, len(lock.PrivateMaps[id].Commit))]
	}
	return out
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
			r.Summary = "Usage error: --map-release takes map-YYYY.MM.DD or map-YYYY.MM.DD.N."
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
