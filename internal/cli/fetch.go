package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/zach/automap/internal/composition"
	"github.com/zach/automap/internal/mapmodel"
	"github.com/zach/automap/internal/release"
	"github.com/zach/automap/internal/verify"
)

// fetchCmd fills the caches the composition needs to build and verify on
// the host platform without the network (ADR-0038): the map release it
// records, the published evidence for its bindings and pairings, the vcpkg
// tool when VCPKG_ROOT isn't set, the registries' ports, and every port's
// source downloads.
func fetchCmd(a *args, r *result) int {
	m, man, lock, status := loadForFetch(a, r)
	if m == nil {
		return status
	}
	platform := verify.HostPlatform()
	written, fails := regenerate(m, man, lock, ".")
	if fails != nil {
		return failWith(r, fails, "The composition's files can't be generated.", exitFailed)
	}
	r.FilesWritten = append(r.FilesWritten, written...)
	warnEvidence(r, compositionEvidence(m, man, lock, time.Now()))
	vcpkg, err := verify.VcpkgRoot(m)
	if err != nil {
		r.Summary = "vcpkg isn't available."
		r.fail("FETCH_FAILED", err.Error())
		return exitFailed
	}
	log, installRoot := logFile("fetch-"+platform+".log"), filepath.Join("build", "automap", platform, "vcpkg_installed")
	if _, err := verify.Vcpkg(vcpkg, ".", abs(installRoot), log, nil, "install", "--only-downloads"); err != nil {
		r.Summary = "vcpkg couldn't download what the composition needs."
		r.fail("FETCH_FAILED", err.Error()+"; the log has the details.")
		r.Failures[len(r.Failures)-1].DetailsFile = filepath.ToSlash(log)
		return exitFailed
	}
	r.OK = true
	r.Summary = "Fetched what the composition needs to build and verify on " + platform + " without the network."
	return exitOK
}

// vendorCmd copies what the composition needs on the host platform into
// vendor/automap/ in the repo (ADR-0038): every port in vcpkg's install
// plan, the sources vcpkg downloads for them, the map release the lock
// records and its published evidence. The vendored ports join the vcpkg
// configuration as overlay ports, so a fresh clone builds and verifies
// with no network and no caches.
func vendorCmd(a *args, r *result) int {
	m, man, lock, status := loadForFetch(a, r)
	if m == nil {
		return status
	}
	platform := verify.HostPlatform()
	warnEvidence(r, compositionEvidence(m, man, lock, time.Now()))
	vcpkg, err := verify.VcpkgRoot(m)
	if err != nil {
		r.Summary = "vcpkg isn't available."
		r.fail("VENDOR_FAILED", err.Error())
		return exitFailed
	}
	vendorFailed := func(err error, details string) int {
		r.Summary = "The composition can't be vendored."
		r.fail("VENDOR_FAILED", err.Error())
		r.Failures[len(r.Failures)-1].DetailsFile = filepath.ToSlash(details)
		return exitFailed
	}
	// vcpkg plans against the registries, not the ports vendored before,
	// and keeps their ports and its downloads where vendor can copy them.
	if _, fails := regenerateWith(m, man, lock, ".", nil); fails != nil {
		return failWith(r, fails, "The composition's files can't be generated.", exitFailed)
	}
	scratch := filepath.Join("build", "automap", "vendor")
	os.RemoveAll(scratch)
	registries, downloads := filepath.Join(release.CacheDir(), "vcpkg", "registries"), abs(filepath.Join(scratch, "downloads"))
	for _, d := range []string{registries, downloads} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return vendorFailed(err, "")
		}
	}
	env := []string{"X_VCPKG_REGISTRIES_CACHE=" + abs(registries), "VCPKG_DOWNLOADS=" + downloads}
	log, installRoot := logFile("vendor-"+platform+".log"), abs(filepath.Join(scratch, "installed"))
	plan, err := verify.Vcpkg(vcpkg, ".", installRoot, log, env, "install", "--dry-run")
	if err != nil {
		return vendorFailed(err, log)
	}
	ports := verify.ParsePlan(plan)
	var baseline struct {
		BuiltinBaseline string `json:"builtin-baseline"`
	}
	raw, _ := os.ReadFile("vcpkg.json")
	json.Unmarshal(raw, &baseline)
	var written []string
	for _, p := range ports {
		dst := filepath.Join(filepath.FromSlash(composition.VendoredPorts), p.Name)
		os.RemoveAll(dst)
		switch {
		case p.Tree != "":
			err = copyDir(filepath.Join(registries, "git-trees", p.Tree), dst)
		case p.Dir != "":
			err = copyDir(p.Dir, dst)
		default:
			err = verify.CuratedPort(vcpkg, baseline.BuiltinBaseline, p, dst)
		}
		if err != nil {
			return vendorFailed(fmt.Errorf("vendoring port %s: %v", p.Name, err), log)
		}
		written = append(written, listFiles(dst)...)
	}
	if _, err := verify.Vcpkg(vcpkg, ".", installRoot, log, env, "install", "--only-downloads"); err != nil {
		return vendorFailed(err, log)
	}
	entries, _ := os.ReadDir(downloads)
	var sources int
	for _, e := range entries {
		if e.IsDir() || strings.HasSuffix(e.Name(), ".part") {
			continue
		}
		dst := filepath.Join(filepath.FromSlash(release.Vendored), "downloads", e.Name())
		if err := copyPath(filepath.Join(downloads, e.Name()), dst); err != nil {
			return vendorFailed(err, "")
		}
		written = append(written, filepath.ToSlash(dst))
		sources++
	}
	// The map release the lock records, and its published evidence, as
	// fetching left them in the cache.
	if tag := mapRelease(m.PublicDir()); release.Tag.MatchString(tag) {
		if dir := release.VendoredDir(tag); abs(m.PublicDir()) != dir {
			os.RemoveAll(dir)
			if err := copyDir(m.PublicDir(), dir); err != nil {
				return vendorFailed(err, "")
			}
			written = append(written, listFiles(filepath.Join(filepath.FromSlash(release.Vendored), "maps", tag))...)
		}
		for _, e := range compositionEdges(m, man, lock) {
			for _, platform := range man.Platforms {
				rel := evidencePath(e.kind, e.id, e.pins, platform)
				cached := release.CachedEvidence(rel)
				if cached == "" {
					continue
				}
				dst := filepath.Join(filepath.FromSlash(release.Vendored), filepath.FromSlash(rel))
				if err := copyPath(cached, dst); err != nil {
					return vendorFailed(err, "")
				}
				written = append(written, filepath.ToSlash(dst))
			}
		}
	}
	regenerated, fails := regenerate(m, man, lock, ".")
	if fails != nil {
		return failWith(r, fails, "The composition's files can't be generated.", exitFailed)
	}
	sort.Strings(written)
	r.FilesWritten = append(r.FilesWritten, append(written, regenerated...)...)
	r.OK = true
	r.Summary = fmt.Sprintf("Vendored %d %s and %d %s into %s for %s.", len(ports), plural(len(ports), "port", "ports"),
		sources, plural(sources, "source download", "source downloads"), release.Vendored, platform)
	return exitOK
}

// loadForFetch loads the composition in the current directory, for fetch
// and vendor.
func loadForFetch(a *args, r *result) (*mapmodel.Map, *composition.Manifest, *composition.Lock, int) {
	if _, err := os.Stat(manifestFile); os.IsNotExist(err) {
		r.Summary = "Nothing to " + r.Command + ": no " + manifestFile + " in the current directory."
		r.fail("NO_MANIFEST", "No composition here. Run automap start to create one, or automap adopt to adopt an existing engine.")
		return nil, nil, nil, exitUsage
	}
	return loadComposition(a, r)
}

// logFile is a log in build/automap, emptied.
func logFile(name string) string {
	p := filepath.Join("build", "automap", name)
	os.MkdirAll(filepath.Dir(p), 0o755)
	os.Remove(p)
	return p
}

func abs(p string) string {
	a, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return a
}

// copyDir copies a directory tree.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		return copyPath(p, filepath.Join(dst, rel))
	})
}

// copyPath copies a file, creating its directory.
func copyPath(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

// listFiles lists the files under dir, slash-separated.
func listFiles(dir string) []string {
	var out []string
	filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			out = append(out, path.Clean(filepath.ToSlash(p)))
		}
		return nil
	})
	return out
}
