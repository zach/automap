// Package release finds, fetches and caches the public map's dated
// releases (ADR-0053) and the published evidence for exact pins
// (ADR-0082). Fetching is the only thing that needs the network; once a
// release or a piece of evidence is cached, it is read from the cache when
// the network is unavailable (ADR-0038).
package release

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"
)

// DefaultSource is where the public map's releases are listed: the
// repository's GitHub releases API.
const DefaultSource = "https://api.github.com/repos/zach/automap"

// DefaultEvidence is where published evidence is read: the map on the
// repository's main branch, where Automap's CI adds it.
const DefaultEvidence = "https://raw.githubusercontent.com/zach/automap/main/map"

// Tag matches a map release's tag, map-YYYY.MM.DD.
var Tag = regexp.MustCompile(`^map-\d{4}\.\d{2}\.\d{2}$`)

var client = &http.Client{Timeout: 30 * time.Second}

// Source is the releases API, from AUTOMAP_RELEASES or the default.
func Source() string {
	if s := os.Getenv("AUTOMAP_RELEASES"); s != "" {
		return strings.TrimSuffix(s, "/")
	}
	return DefaultSource
}

// EvidenceSource is where published evidence is read, from
// AUTOMAP_EVIDENCE_URL or the default.
func EvidenceSource() string {
	if s := os.Getenv("AUTOMAP_EVIDENCE_URL"); s != "" {
		return strings.TrimSuffix(s, "/")
	}
	return DefaultEvidence
}

// CacheDir is where fetched releases and evidence are kept: AUTOMAP_CACHE,
// or automap in the user's cache directory.
func CacheDir() string {
	if c := os.Getenv("AUTOMAP_CACHE"); c != "" {
		return c
	}
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "automap")
}

// Dir is where a cached release's map is.
func Dir(tag string) string { return filepath.Join(CacheDir(), "maps", tag) }

// Cached lists the releases in the cache, newest first.
func Cached() []string {
	entries, _ := os.ReadDir(filepath.Join(CacheDir(), "maps"))
	var tags []string
	for _, e := range entries {
		if e.IsDir() && Tag.MatchString(e.Name()) {
			tags = append(tags, e.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(tags)))
	return tags
}

type asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

type listed struct {
	Tag    string  `json:"tag_name"`
	Assets []asset `json:"assets"`
}

// releases lists the published map releases, newest first.
func releases() ([]listed, error) {
	resp, err := client.Get(Source() + "/releases?per_page=100")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s/releases: %s", Source(), resp.Status)
	}
	var all []listed
	if err := json.NewDecoder(resp.Body).Decode(&all); err != nil {
		return nil, err
	}
	var out []listed
	for _, r := range all {
		if Tag.MatchString(r.Tag) {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Tag > out[j].Tag })
	return out, nil
}

// InvalidError is a release whose archive doesn't match its SHA-256 or
// can't be unpacked.
type InvalidError struct {
	Tag    string
	Reason string
}

func (e *InvalidError) Error() string { return e.Tag + ": " + e.Reason }

// Newest returns the newest map release and its directory, fetching it
// into the cache when it isn't there. Without the network it falls back
// to the newest cached release.
func Newest() (tag, dir string, err error) {
	list, listErr := releases()
	if listErr == nil {
		if len(list) == 0 {
			return "", "", errors.New("no map release is published yet")
		}
		dir, err := fetch(list[0])
		return list[0].Tag, dir, err
	}
	if cached := Cached(); len(cached) > 0 {
		return cached[0], Dir(cached[0]), nil
	}
	return "", "", fmt.Errorf("can't list map releases and none is cached: %v", listErr)
}

// Open returns a release's directory, fetching it when it isn't cached.
func Open(tag string) (string, error) {
	if _, err := os.Stat(Dir(tag)); err == nil {
		return Dir(tag), nil
	}
	list, err := releases()
	if err != nil {
		return "", fmt.Errorf("%s isn't cached and can't be fetched: %v", tag, err)
	}
	for _, r := range list {
		if r.Tag == tag {
			return fetch(r)
		}
	}
	return "", fmt.Errorf("no map release %s is published", tag)
}

// fetch downloads a release's archive and its SHA-256, checks one against
// the other, and unpacks the archive into the cache.
func fetch(r listed) (string, error) {
	dir := Dir(r.Tag)
	if _, err := os.Stat(dir); err == nil {
		return dir, nil
	}
	var archiveURL, sumURL string
	for _, a := range r.Assets {
		switch a.Name {
		case r.Tag + ".tar.gz":
			archiveURL = a.URL
		case r.Tag + ".tar.gz.sha256":
			sumURL = a.URL
		}
	}
	if archiveURL == "" || sumURL == "" {
		return "", fmt.Errorf("%s has no %s.tar.gz with its .sha256", r.Tag, r.Tag)
	}
	archive, err := get(archiveURL)
	if err != nil {
		return "", err
	}
	sumFile, err := get(sumURL)
	if err != nil {
		return "", err
	}
	want, _, _ := strings.Cut(strings.TrimSpace(string(sumFile)), " ")
	got := sha256.Sum256(archive)
	if hex.EncodeToString(got[:]) != want {
		return "", &InvalidError{r.Tag, r.Tag + ".tar.gz doesn't match its SHA-256"}
	}
	if err := os.MkdirAll(CacheDir(), 0o755); err != nil {
		return "", err
	}
	tmp, err := os.MkdirTemp(CacheDir(), "unpack-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmp)
	if err := unpack(archive, tmp); err != nil {
		return "", &InvalidError{r.Tag, fmt.Sprintf("%s.tar.gz: %v", r.Tag, err)}
	}
	os.MkdirAll(filepath.Dir(dir), 0o755)
	if err := os.Rename(tmp, dir); err != nil && !exists(dir) {
		return "", err
	}
	return dir, nil
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func get(url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// unpack extracts a gzipped tar of the map into dir, refusing any entry
// that would land outside it.
func unpack(archive []byte, dir string) error {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return err
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		name := path.Clean(strings.TrimPrefix(h.Name, "./"))
		if name == "." {
			continue
		}
		if strings.HasPrefix(name, "../") || path.IsAbs(name) {
			return fmt.Errorf("entry %s is outside the map", h.Name)
		}
		target := filepath.Join(dir, filepath.FromSlash(name))
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			os.MkdirAll(filepath.Dir(target), 0o755)
			data, err := io.ReadAll(tr)
			if err != nil {
				return err
			}
			if err := os.WriteFile(target, data, 0o644); err != nil {
				return err
			}
		}
	}
}

// Evidence reads one piece of published evidence, by its path in the map
// (evidence/<kind>/<id>/<pins>/<platform>.json), from the evidence source,
// keeping a copy in the cache. Without the network it reads the cached
// copy. It returns nil when none is published.
func Evidence(rel string) ([]byte, error) {
	cached := filepath.Join(CacheDir(), "evidence", filepath.FromSlash(rel))
	resp, err := client.Get(EvidenceSource() + "/" + rel)
	if err == nil {
		defer resp.Body.Close()
		switch resp.StatusCode {
		case http.StatusOK:
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			os.MkdirAll(filepath.Dir(cached), 0o755)
			os.WriteFile(cached, data, 0o644)
			return data, nil
		case http.StatusNotFound:
			os.Remove(cached)
			return nil, nil
		}
	}
	if data, err := os.ReadFile(cached); err == nil {
		return data, nil
	}
	return nil, nil
}

// MaxAge is how long evidence stays current (ADR-0031).
const MaxAge = 30 * 24 * time.Hour

// State is a piece of evidence's state (ADR-0031): broken when its run
// failed (a rule failed, or it didn't reach smoke), stale when it is older
// than MaxAge, and current otherwise. It returns the evidence's level too.
// It is a pure function.
func State(raw []byte, now time.Time) (level, state string, err error) {
	var e struct {
		Predicate struct {
			Level      string `json:"level"`
			ProducedAt string `json:"produced_at"`
			Results    []struct {
				Status string `json:"status"`
			} `json:"results"`
		} `json:"predicate"`
	}
	if err := json.Unmarshal(raw, &e); err != nil {
		return "", "", err
	}
	p := e.Predicate
	at, err := time.Parse(time.RFC3339, p.ProducedAt)
	if err != nil {
		return p.Level, "", fmt.Errorf("produced_at: %v", err)
	}
	switch {
	case p.Level == "none" || p.Level == "builds":
		return p.Level, "broken", nil
	case slices.ContainsFunc(p.Results, func(r struct {
		Status string `json:"status"`
	}) bool {
		return r.Status == "fail"
	}):
		return p.Level, "broken", nil
	case now.Sub(at) > MaxAge:
		return p.Level, "stale", nil
	}
	return p.Level, "current", nil
}
