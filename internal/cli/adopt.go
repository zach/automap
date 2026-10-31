package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zach/automap/internal/composition"
	"github.com/zach/automap/internal/mapmodel"
)

// adopt brings an existing CMake and vcpkg project under Automap (SPEC
// §2.6): it matches the libraries the project links to map libraries,
// reports any no map describes, and once every one is described, writes a
// manifest that points at the glue the engine already has, keeps the
// engine's own startup, writes only what is missing, and verifies.
func adopt(a *args, r *result) int {
	for _, f := range []string{composition.ManifestFile, composition.LockFile} {
		if _, err := os.Stat(f); err == nil {
			r.Summary = "A composition already exists here."
			r.fail("COMPOSITION_EXISTS", f+" already exists. Use automap extend to add a library, or automap upgrade to move to a newer map release.")
			return exitRefused
		}
	}
	if _, err := os.Stat("CMakeLists.txt"); err != nil {
		r.Summary = "No CMake project here."
		r.fail("NO_CMAKE_PROJECT", "adopt reads an existing CMake and vcpkg project from CMakeLists.txt in the current directory. For a new engine, run automap start.")
		return exitUsage
	}
	repoFiles, cmakeFiles := readProject()
	project := composition.ScanCMake(cmakeFiles)
	exe, ok := adoptedExecutable(project, a.value("target"), r)
	if !ok {
		return exitUsage
	}
	privates, ok := adoptPrivateMaps(a, r)
	if !ok {
		return exitUsage
	}
	m, status := loadMapWith(a, privates, r)
	if m == nil {
		return status
	}

	// Match each library target the project links to the map library
	// whose CMake targets name it.
	byTarget := map[string]string{}
	for id, l := range m.Libraries {
		for _, v := range l.Versions {
			for _, t := range v.CMake.Targets {
				byTarget[t] = id
			}
		}
	}
	var libs []string
	seen := map[string]bool{}
	for _, t := range project.LibraryTargets() {
		lib, ok := byTarget[t]
		if !ok {
			r.fail("UNMAPPED_TARGET("+t+")", fmt.Sprintf("No map describes %s. Describe the library that provides it in a private map (automap-private/), with source.kind path for code in this repo, then run automap adopt again.", t))
			continue
		}
		if !seen[lib] {
			seen[lib] = true
			libs = append(libs, lib)
		}
	}
	if len(r.Failures) > 0 {
		r.Summary = fmt.Sprintf("%d %s no map describes.", len(r.Failures), plural(len(r.Failures), "target", "targets"))
		return exitFailed
	}
	sort.Strings(libs)

	req := composition.Request{Platforms: a.values["platform"], MapRelease: mapRelease(m.PublicDir()), Executable: composition.Executable{Target: exe.Target, Sources: exe.Sources}}
	for _, lib := range libs {
		req.With = append(req.With, composition.LibraryRequest{Library: lib})
	}
	man, fails := composition.NewManifest(m, req)
	if fails != nil {
		return failWith(r, fails, "The composition can't be made from the project's libraries.", exitUsage)
	}
	for _, pm := range privates {
		man.PrivateMaps = append(man.PrivateMaps, composition.PrivateMap{ID: pm.ID, Location: pm.Location, Platforms: pm.Platforms})
	}
	// The glue goes where the engine keeps its own, and the engine keeps
	// its own startup code (ADR-0080).
	man.Glue = composition.RootGlue(man.Glue, glueRoot(man.Glue, repoFiles))
	ownStartup := ""
	for _, src := range exe.Sources {
		if data, err := os.ReadFile(filepath.FromSlash(src)); err == nil && strings.Contains(string(data), "main(") {
			ownStartup = src
			break
		}
	}
	if ownStartup != "" {
		man.Glue.Startup = composition.Startup{DeveloperOwned: []string{ownStartup}}
	}
	lock, fails := composition.Resolve(m, man)
	if fails != nil {
		return failWith(r, fails, "The composition can't be resolved against the map.", exitUsage)
	}
	plan, fails := composition.PlanStartup(m, man)
	if fails != nil {
		return failWith(r, fails, "The composition's startup can't be planned.", exitFailed)
	}
	files, fails := composition.Files(m, man, lock, plan, composition.FileOptions{RegistryURL: registryURL(m)})
	if fails != nil {
		return failWith(r, fails, "The composition's files can't be written.", exitFailed)
	}
	if !includesAutomapCMake(cmakeFiles) {
		r.warn("AUTOMAP_CMAKE_NOT_INCLUDED", "No CMake file includes cmake/automap/automap.cmake. Include it once from CMakeLists.txt, before "+exe.Target+" links automap_glue, so the glue builds.")
	}
	// The CLI owns what it generates; everything else adopt writes only
	// where nothing is there yet.
	generated := map[string]bool{}
	for _, f := range generatedFiles(man) {
		generated[f] = true
	}
	keep := func(p string) bool { return !generated[p] }
	var pinned []string
	for _, lib := range man.Libraries() {
		pin := lock.Pins[lib]
		pinned = append(pinned, fmt.Sprintf("%s %s %s", lib, pin.Version, pin.Variant))
	}
	return writeAndVerify(a, m, man, files, keep, fmt.Sprintf("Adopted %s (%s)", exe.Target, strings.Join(pinned, ", ")), r)
}

// readProject lists the repo's files and reads its CMake files, leaving out
// build output, version control, vendored files, what the CLI generates,
// and maps, which hold CMake files of their own.
func readProject() (all []string, cmake map[string]string) {
	cmake = map[string]string{}
	filepath.WalkDir(".", func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel := filepath.ToSlash(p)
		if d.IsDir() {
			if _, err := os.Stat(filepath.Join(p, "map.json")); err == nil {
				return filepath.SkipDir
			}
			switch rel {
			case ".git", "build", "vendor", "cmake/automap":
				return filepath.SkipDir
			}
			if strings.HasPrefix(d.Name(), ".") && rel != "." {
				return filepath.SkipDir
			}
			return nil
		}
		all = append(all, rel)
		if d.Name() == "CMakeLists.txt" || strings.HasSuffix(d.Name(), ".cmake") {
			if data, err := os.ReadFile(p); err == nil {
				cmake[rel] = string(data)
			}
		}
		return nil
	})
	return all, cmake
}

// adoptedExecutable is the project's executable: the one --target names,
// or its only one.
func adoptedExecutable(p composition.Project, target string, r *result) (composition.ProjectExecutable, bool) {
	var names []string
	for _, e := range p.Executables {
		if e.Target == target || target == "" && len(p.Executables) == 1 {
			return e, true
		}
		names = append(names, e.Target)
	}
	switch {
	case target != "":
		r.Summary = "The project has no executable " + target + "."
		r.fail("USAGE", "The project's executables are "+strings.Join(names, ", ")+".")
	case len(names) == 0:
		r.Summary = "The project has no executable."
		r.fail("USAGE", "adopt needs the engine's executable: CMakeLists.txt defines none with add_executable.")
	default:
		r.Summary = "The project has more than one executable."
		r.fail("USAGE", "Name the engine's executable with --target: "+strings.Join(names, ", ")+".")
	}
	return composition.ProjectExecutable{}, false
}

// adoptPrivateMaps are the private maps adopt layers over the public map,
// in order: those --private-map names, or automap-private when it holds a
// map.
func adoptPrivateMaps(a *args, r *result) ([]mapmodel.PrivateMap, bool) {
	dirs := a.values["private-map"]
	if len(dirs) == 0 {
		if _, err := os.Stat(filepath.Join("automap-private", "map.json")); err == nil {
			dirs = []string{"automap-private"}
		}
	}
	var out []mapmodel.PrivateMap
	for _, dir := range dirs {
		var meta struct {
			ID        string   `json:"id"`
			Platforms []string `json:"platforms"`
		}
		raw, err := os.ReadFile(filepath.Join(dir, "map.json"))
		if err == nil {
			err = json.Unmarshal(raw, &meta)
		}
		if err != nil {
			r.Summary = "A private map can't be read."
			r.fail("MAP_INVALID("+path.Join(filepath.ToSlash(dir), "map.json")+")", err.Error())
			return nil, false
		}
		out = append(out, mapmodel.PrivateMap{ID: meta.ID, Location: filepath.ToSlash(dir), Platforms: meta.Platforms})
	}
	return out, true
}

// glueRoot is where the engine keeps its glue: the directory under which
// the most reference glue files already sit at their usual place, such as
// engine/glue for engine/glue/sdl3/sdl3_setup.cpp, or the default glue
// directory when none do.
func glueRoot(g composition.Glue, files []string) string {
	counts := map[string]int{}
	for _, entries := range []map[string]composition.GlueEntry{g.Bindings, g.Pairings, g.Setup} {
		for _, e := range entries {
			for _, f := range e.Files {
				rest, ok := strings.CutPrefix(f, composition.GlueDir+"/")
				if !ok {
					continue
				}
				for _, have := range files {
					if root, ok := strings.CutSuffix(have, "/"+rest); ok {
						counts[root]++
					}
				}
			}
		}
	}
	best := composition.GlueDir
	for root, n := range counts {
		if n > counts[best] || n == counts[best] && root < best {
			best = root
		}
	}
	return best
}

func includesAutomapCMake(files map[string]string) bool {
	for _, data := range files {
		if strings.Contains(data, "cmake/automap/automap.cmake") {
			return true
		}
	}
	return false
}
