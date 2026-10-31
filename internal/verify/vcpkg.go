package verify

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/zach/automap/internal/mapmodel"
	"github.com/zach/automap/internal/release"
)

// DefaultVcpkgTool is where vcpkg's tool releases download from.
const DefaultVcpkgTool = "https://github.com/microsoft/vcpkg-tool/releases/download"

// CuratedRegistry is vcpkg's curated registry, which a vcpkg without a
// ports tree of its own reads as a git registry.
const CuratedRegistry = "https://github.com/microsoft/vcpkg"

// VcpkgTool is the vcpkg tool release the public map's map.json names.
func VcpkgTool(m *mapmodel.Map) string {
	var bb struct {
		Vcpkg struct {
			Tool string `json:"tool"`
		} `json:"vcpkg"`
	}
	json.Unmarshal(m.Maps[0].BuildBackend, &bb)
	return bb.Vcpkg.Tool
}

// VcpkgRoot is the vcpkg the CLI builds with: VCPKG_ROOT when it is set,
// and otherwise the vcpkg tool release the public map names, bootstrapped
// into the cache the first time (SPEC §5.5).
func VcpkgRoot(m *mapmodel.Map) (string, error) {
	if root := os.Getenv("VCPKG_ROOT"); root != "" {
		return root, nil
	}
	tool := VcpkgTool(m)
	if tool == "" {
		return "", errors.New("VCPKG_ROOT isn't set and the map names no vcpkg tool release to bootstrap")
	}
	dir := filepath.Join(release.CacheDir(), "vcpkg", tool)
	if _, err := os.Stat(filepath.Join(dir, vcpkgExe())); err == nil {
		return dir, nil
	}
	if err := bootstrap(tool, dir); err != nil {
		return "", fmt.Errorf("VCPKG_ROOT isn't set and the vcpkg tool release %s can't be bootstrapped: %v", tool, err)
	}
	return dir, nil
}

func vcpkgExe() string {
	if runtime.GOOS == "windows" {
		return "vcpkg.exe"
	}
	return "vcpkg"
}

// toolAsset is the vcpkg tool release's executable for this machine.
func toolAsset() (string, error) {
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "linux/amd64":
		return "vcpkg-glibc", nil
	case "linux/arm64":
		return "vcpkg-glibc-arm64", nil
	case "darwin/amd64", "darwin/arm64":
		return "vcpkg-macos", nil
	case "windows/amd64":
		return "vcpkg.exe", nil
	case "windows/arm64":
		return "vcpkg-arm64.exe", nil
	}
	return "", fmt.Errorf("vcpkg publishes no tool for %s/%s", runtime.GOOS, runtime.GOARCH)
}

// bootstrap downloads a vcpkg tool release, its standalone bundle of
// scripts and triplets and its executable, into dir.
func bootstrap(tool, dir string) error {
	asset, err := toolAsset()
	if err != nil {
		return err
	}
	base := DefaultVcpkgTool
	if u := os.Getenv("AUTOMAP_VCPKG_TOOL_URL"); u != "" {
		base = strings.TrimSuffix(u, "/")
	}
	if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp(filepath.Dir(dir), "bootstrap-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	client := &http.Client{Timeout: 10 * time.Minute}
	get := func(name string) (io.ReadCloser, error) {
		resp, err := client.Get(base + "/" + tool + "/" + name)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("%s/%s/%s: %s", base, tool, name, resp.Status)
		}
		return resp.Body, nil
	}
	bundle, err := get("vcpkg-standalone-bundle.tar.gz")
	if err != nil {
		return err
	}
	err = untar(bundle, tmp)
	bundle.Close()
	if err != nil {
		return err
	}
	exe, err := get(asset)
	if err != nil {
		return err
	}
	defer exe.Close()
	f, err := os.OpenFile(filepath.Join(tmp, vcpkgExe()), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, exe); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, dir); err != nil {
		if _, statErr := os.Stat(filepath.Join(dir, vcpkgExe())); statErr != nil {
			return err
		}
	}
	return nil
}

// untar extracts a gzipped tar into dir.
func untar(r io.Reader, dir string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	return extract(tar.NewReader(gz), dir)
}

// extract extracts a tar into dir, keeping executable bits and refusing
// any entry that would land outside it.
func extract(tr *tar.Reader, dir string) error {
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		name := path.Clean(strings.TrimPrefix(h.Name, "./"))
		if name == "." || h.Typeflag == tar.TypeXGlobalHeader {
			continue
		}
		if strings.HasPrefix(name, "../") || path.IsAbs(name) {
			return fmt.Errorf("entry %s is outside the archive", h.Name)
		}
		target := filepath.Join(dir, filepath.FromSlash(name))
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			os.MkdirAll(filepath.Dir(target), 0o755)
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(h.Mode)&0o755|0o644)
			if err != nil {
				return err
			}
			_, err = io.Copy(f, tr)
			f.Close()
			if err != nil {
				return err
			}
		}
	}
}

// WriteVcpkgLock records, in vcpkg's lock file under installRoot, each git
// registry the composition's vcpkg files name at its baseline, merged into
// what the file already records. vcpkg then reads a registry at that commit
// from its registries cache, and fetches only what the cache lacks, rather
// than asking the network where the registry's reference points on every
// run (ADR-0038).
func WriteVcpkgLock(root, installRoot string) error {
	var config struct {
		Registries []struct {
			Kind       string `json:"kind"`
			Repository string `json:"repository"`
			Reference  string `json:"reference"`
			Baseline   string `json:"baseline"`
		} `json:"registries"`
	}
	var manifest struct {
		BuiltinBaseline string `json:"builtin-baseline"`
	}
	if raw, err := os.ReadFile(filepath.Join(root, "vcpkg-configuration.json")); err == nil {
		json.Unmarshal(raw, &config)
	}
	if raw, err := os.ReadFile(filepath.Join(root, "vcpkg.json")); err == nil {
		json.Unmarshal(raw, &manifest)
	}
	pins := map[string]map[string]string{}
	pin := func(repo, ref, commit string) {
		if repo == "" || commit == "" {
			return
		}
		if ref == "" {
			ref = "HEAD"
		}
		if pins[repo] == nil {
			pins[repo] = map[string]string{}
		}
		pins[repo][ref] = commit
	}
	for _, r := range config.Registries {
		if r.Kind == "git" {
			pin(r.Repository, r.Reference, r.Baseline)
		}
	}
	// A vcpkg with no ports tree of its own, such as a bootstrapped tool
	// release, reads the curated registry from git at the builtin baseline.
	pin(CuratedRegistry, "HEAD", manifest.BuiltinBaseline)
	if len(pins) == 0 {
		return nil
	}
	file := filepath.Join(installRoot, "vcpkg", "vcpkg-lock.json")
	lock := map[string]map[string]string{}
	if raw, err := os.ReadFile(file); err == nil {
		json.Unmarshal(raw, &lock)
	}
	for repo, refs := range pins {
		if lock[repo] == nil {
			lock[repo] = map[string]string{}
		}
		for ref, commit := range refs {
			lock[repo][ref] = commit
		}
	}
	raw, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	return os.WriteFile(file, append(raw, '\n'), 0o644)
}

// Vcpkg runs vcpkg from root in the composition's directory in manifest
// mode, with installRoot as its install root, appending its output to
// logFile, and returns that output.
func Vcpkg(vcpkgRoot, root, installRoot, logFile string, env []string, args ...string) (string, error) {
	if err := WriteVcpkgLock(root, installRoot); err != nil {
		return "", err
	}
	args = append(args, "--x-install-root="+installRoot)
	cmd := exec.Command(filepath.Join(vcpkgRoot, vcpkgExe()), args...)
	cmd.Dir = root
	cmd.Env = append(append(os.Environ(), "VCPKG_ROOT="+vcpkgRoot), env...)
	out, err := cmd.CombinedOutput()
	if f, ferr := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644); ferr == nil {
		fmt.Fprintf(f, "$ vcpkg %s\n%s", strings.Join(args, " "), out)
		f.Close()
	}
	if err != nil {
		return string(out), fmt.Errorf("vcpkg %s failed (%v)", args[0], err)
	}
	return string(out), nil
}

// PlannedPort is one port in vcpkg's install plan, with where vcpkg reads
// it from: a git registry's tree, an overlay port's directory, or neither
// for the ports tree of a vcpkg with one.
type PlannedPort struct {
	Name, Version string
	Repository    string // a git registry's URL
	Tree          string // the port's git tree in that registry
	Dir           string // an overlay port's directory
}

var planLine = regexp.MustCompile(`^\s+(?:\*\s+)?([a-z0-9-]+)(?:\[[^\]]*\])?:[^@\s]+@(\S+)(?:\s+--\s+(.+))?$`)

// ParsePlan reads the ports in the install plan vcpkg prints for
// `install --dry-run`, each once.
func ParsePlan(out string) []PlannedPort {
	var ports []PlannedPort
	seen := map[string]bool{}
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		m := planLine.FindStringSubmatch(strings.TrimRight(sc.Text(), "\r"))
		if m == nil || seen[m[1]] {
			continue
		}
		seen[m[1]] = true
		p := PlannedPort{Name: m[1], Version: m[2]}
		switch src := strings.TrimSpace(m[3]); {
		case strings.HasPrefix(src, "git+"):
			at := strings.LastIndex(src, "@")
			p.Repository, p.Tree = strings.TrimPrefix(src[:at], "git+"), src[at+1:]
		case src != "":
			p.Dir = src
		}
		ports = append(ports, p)
	}
	return ports
}

// Vendored is where `automap vendor` copies what a composition needs into
// its repo: ports, source downloads, the map release and its evidence.
const Vendored = release.Vendored

// DownloadsDir is where vcpkg keeps its downloads: VCPKG_DOWNLOADS, or
// downloads in its root.
func DownloadsDir(vcpkgRoot string) string {
	if d := os.Getenv("VCPKG_DOWNLOADS"); d != "" {
		return d
	}
	return filepath.Join(vcpkgRoot, "downloads")
}

// SeedDownloads copies a vendored composition's downloads into vcpkg's
// downloads, where vcpkg finds them without the network.
func SeedDownloads(vcpkgRoot, root string) error {
	src := filepath.Join(root, filepath.FromSlash(Vendored), "downloads")
	entries, err := os.ReadDir(src)
	if err != nil {
		return nil
	}
	dst := DownloadsDir(vcpkgRoot)
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	for _, e := range entries {
		target := filepath.Join(dst, e.Name())
		if e.IsDir() {
			continue
		}
		if _, err := os.Stat(target); err == nil {
			continue
		}
		if err := copyFile(filepath.Join(src, e.Name()), target); err != nil {
			return err
		}
	}
	return nil
}

// copyFile copies a file, through a temporary file beside the target so
// a reader never sees it half-written.
func copyFile(from, to string) error {
	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer in.Close()
	tmp, err := os.CreateTemp(filepath.Dir(to), ".copy-")
	if err != nil {
		return err
	}
	_, err = io.Copy(tmp, in)
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(tmp.Name(), to)
	}
	if err != nil {
		os.Remove(tmp.Name())
	}
	return err
}

// CuratedPort copies the port the curated registry's versions database, at
// baseline in the git clone that is vcpkg's root, records for p's version
// into dst. A vcpkg with a ports tree of its own reads the curated
// registry from it, and its install plan names no source for those ports.
func CuratedPort(vcpkgRoot, baseline string, p PlannedPort, dst string) error {
	// The port's files as committed: a patch with its line endings
	// converted no longer applies.
	git := func(args ...string) *exec.Cmd {
		return exec.Command("git", append([]string{"-C", vcpkgRoot, "-c", "core.autocrlf=false", "-c", "core.eol=lf"}, args...)...)
	}
	raw, err := git("show", baseline+":versions/"+p.Name[:1]+"-/"+p.Name+".json").Output()
	if err != nil {
		return fmt.Errorf("the curated registry at %s has no versions of %s: %v", baseline, p.Name, err)
	}
	var db struct {
		Versions []struct {
			Version     string `json:"version"`
			Semver      string `json:"version-semver"`
			Date        string `json:"version-date"`
			String      string `json:"version-string"`
			PortVersion int    `json:"port-version"`
			GitTree     string `json:"git-tree"`
		} `json:"versions"`
	}
	if err := json.Unmarshal(raw, &db); err != nil {
		return err
	}
	for _, v := range db.Versions {
		version := v.Version + v.Semver + v.Date + v.String
		if v.PortVersion != 0 {
			version += fmt.Sprintf("#%d", v.PortVersion)
		}
		if version != p.Version {
			continue
		}
		cmd := git("archive", "--format=tar", v.GitTree)
		out, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		if err := cmd.Start(); err != nil {
			return err
		}
		if err := os.MkdirAll(dst, 0o755); err != nil {
			return err
		}
		extractErr := extract(tar.NewReader(out), dst)
		io.Copy(io.Discard, out)
		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("git archive %s: %v", v.GitTree, err)
		}
		return extractErr
	}
	return fmt.Errorf("the curated registry at %s has no %s %s", baseline, p.Name, p.Version)
}
