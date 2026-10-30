package composition

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zach/automap/internal/mapmodel"
)

// FileOptions are the choices about which files a composition gets.
type FileOptions struct {
	Testbed     bool   // write the map's testbed template: CMakeLists.txt and main.cpp
	RegistryURL string // the Automap registry's URL for vcpkg-configuration.json
}

// TestbedTarget is the CMake target of the testbed template's executable.
const TestbedTarget = "testbed"

// Files returns every file a composition is made of, by its path in the
// developer's repo: the manifest and lock, the vcpkg files, the build
// fragments and the include file, the shared interfaces and probe harness,
// each recipe's glue, the choices header and the startup, and for a
// testbed, its CMakeLists.txt and main.cpp.
func Files(m *mapmodel.Map, man *Manifest, lock *Lock, plan *StartupPlan, opts FileOptions) (map[string][]byte, []Failure) {
	out := map[string][]byte{}
	var fails []Failure
	add := func(p string, data []byte, from string) {
		if _, dup := out[p]; dup {
			fails = append(fails, failure("GLUE_COLLISION("+p+")", "%s and another file would both be written to %s.", from, p))
			return
		}
		out[p] = data
	}
	copyFile := func(src, dst, owner string) {
		data, err := os.ReadFile(src)
		if err != nil {
			fails = append(fails, failure("MAP_INVALID("+filepath.ToSlash(src)+")", "%s names %s, which can't be read: %v", owner, filepath.ToSlash(src), err))
			return
		}
		add(dst, data, filepath.ToSlash(src))
	}

	add(ManifestFile, indentJSON(man), "the manifest")
	lockBytes, err := EncodeLock(*lock)
	if err != nil {
		fails = append(fails, failure("INTERNAL", "encoding the lock: %v", err))
	}
	add(LockFile, lockBytes, "the lock")
	add("vcpkg.json", VcpkgJSON(m, man, lock), "the vcpkg manifest")
	add("vcpkg-configuration.json", VcpkgConfiguration(m, man, lock, opts.RegistryURL), "the vcpkg configuration")

	if opts.Testbed {
		dir := filepath.Join(m.PublicDir(), "testbed")
		for _, rel := range walk(dir) {
			copyFile(filepath.Join(dir, filepath.FromSlash(rel)), rel, "the map's testbed")
		}
	}

	includeRoot := path.Dir(man.Glue.Interfaces)
	for _, cid := range sortedKeys(lock.Contracts) {
		c := m.LatestContract(cid)
		if c == nil || c.SharedInterface == nil {
			continue
		}
		for _, f := range c.SharedInterface.Files {
			copyFile(m.ContractFile(c, f), path.Join(includeRoot, f), "contract "+cid+" "+c.Version)
		}
	}
	harness := filepath.Join(m.PublicDir(), "probe-harness")
	for _, rel := range walk(harness) {
		copyFile(filepath.Join(harness, filepath.FromSlash(rel)), path.Join(includeRoot, rel), "the map's probe harness")
	}
	add(man.Glue.ChoicesHeader, []byte(ChoicesHeader(m, man, lock)), "the choices header")
	add(path.Join(man.Glue.Interfaces, "context.h"), []byte(ContextHeader(plan)), "the context")
	add(path.Join(man.Glue.Interfaces, "startup.h"), []byte(StartupHeader()), "the startup")
	add(path.Join(man.Glue.Interfaces, "startup.cpp"), []byte(StartupSource(plan)), "the startup")

	for _, rid := range Recipes(m, man) {
		r := m.Recipes[rid]
		for _, g := range r.ReferenceGlue {
			dst := path.Join(GlueDirFor(r), path.Base(g.Path))
			if g.Role == "build-fragment" {
				dst = path.Join("cmake/automap", r.Subject.Library+".cmake")
			}
			copyFile(m.RecipeFile(rid, g.Path), dst, "recipe "+rid)
		}
	}
	for _, lib := range man.Libraries() {
		if _, ok := out[path.Join("cmake/automap", lib+".cmake")]; !ok {
			fails = append(fails, failure("NO_BUILD_FRAGMENT("+lib+")", "The map has no library recipe with a build fragment for %s.", lib))
		}
	}

	var sources []string
	for p := range out {
		if strings.HasSuffix(p, ".cpp") && strings.HasPrefix(p, includeRoot+"/") {
			sources = append(sources, p)
		}
	}
	sort.Strings(sources)
	cxx := "20"
	if v, ok := man.Choices["build-flags"]["cxx_standard"]; ok {
		cxx = fmt.Sprint(v)
	}
	add("cmake/automap/automap.cmake", []byte(AutomapCMake(man, sources, cxx)), "the include file")
	if len(fails) > 0 {
		return nil, fails
	}
	return out, nil
}

// walk lists the regular files under dir, slash-separated and relative to
// it, or nothing when dir doesn't exist.
func walk(dir string) []string {
	var out []string
	fs.WalkDir(os.DirFS(dir), ".", func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			out = append(out, p)
		}
		return nil
	})
	return out
}
