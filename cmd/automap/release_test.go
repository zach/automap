package main_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/zach/automap/internal/clitest"
)

// releaseServer serves map releases the way GitHub's releases API lists
// them, each an archive of the fixture map with its SHA-256, and published
// evidence the way the map's main branch serves it, from files tests add.
type releaseServer struct {
	*httptest.Server
	mu       sync.Mutex
	tags     []string
	files    map[string][]byte // download path -> content
	evidence map[string][]byte // path in the map -> content
	listing  bool              // whether to answer the releases API
}

func newReleaseServer(t *testing.T) *releaseServer {
	t.Helper()
	s := &releaseServer{files: map[string][]byte{}, evidence: map[string][]byte{}, listing: true}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		switch {
		case r.URL.Path == "/releases" && !s.listing:
			http.Error(w, "rate limit exceeded", http.StatusForbidden)
		case r.URL.Path == "/releases":
			var list []map[string]any
			for _, tag := range s.tags {
				list = append(list, map[string]any{"tag_name": tag, "assets": []map[string]string{
					{"name": tag + ".tar.gz", "browser_download_url": s.URL + "/download/" + tag + ".tar.gz"},
					{"name": tag + ".tar.gz.sha256", "browser_download_url": s.URL + "/download/" + tag + ".tar.gz.sha256"},
				}})
			}
			json.NewEncoder(w).Encode(list)
		case strings.HasPrefix(r.URL.Path, "/releases/download/"):
			// <tag>/<file>, as GitHub serves a release's files directly
			parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/releases/download/"), "/", 2)
			if data, ok := s.files["/download/"+parts[len(parts)-1]]; ok {
				w.Write(data)
				return
			}
			http.NotFound(w, r)
		case strings.HasPrefix(r.URL.Path, "/download/"):
			if data, ok := s.files[r.URL.Path]; ok {
				w.Write(data)
				return
			}
			http.NotFound(w, r)
		case strings.HasPrefix(r.URL.Path, "/main/"):
			if data, ok := s.evidence[strings.TrimPrefix(r.URL.Path, "/main/")]; ok {
				w.Write(data)
				return
			}
			http.NotFound(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(s.Close)
	return s
}

// publish adds a release of the fixture map as it is in repo's map/
// directory, with edit applied to the archive's SHA-256 file.
func (s *releaseServer) publish(t *testing.T, tag string, mapDir string, sum func(string) string) {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	filepath.WalkDir(mapDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(mapDir, p)
		data, _ := os.ReadFile(p)
		tw.WriteHeader(&tar.Header{Name: "./" + filepath.ToSlash(rel), Mode: 0o644, Size: int64(len(data)), Typeflag: tar.TypeReg})
		tw.Write(data)
		return nil
	})
	tw.Close()
	gz.Close()
	digest := sha256.Sum256(buf.Bytes())
	line := hex.EncodeToString(digest[:]) + "  " + tag + ".tar.gz\n"
	if sum != nil {
		line = sum(line)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tags = append(s.tags, tag)
	s.files["/download/"+tag+".tar.gz"] = buf.Bytes()
	s.files["/download/"+tag+".tar.gz.sha256"] = []byte(line)
}

// releaseRepo is an empty repo whose CLI reads map releases from the
// server, with a cache of its own, and no map directory.
func releaseRepo(t *testing.T, s *releaseServer, cache string) *clitest.Repo {
	t.Helper()
	repo := clitest.NewRepo(t, "no-manifest")
	repo.Setenv("AUTOMAP_RELEASES", s.URL)
	repo.Setenv("AUTOMAP_EVIDENCE_URL", s.URL+"/main")
	repo.Setenv("AUTOMAP_CACHE", cache)
	return repo
}

// fixtureMapDir assembles the fixture map in a directory of its own.
func fixtureMapDir(t *testing.T) string {
	t.Helper()
	repo := clitest.NewRepo(t, "no-manifest")
	repo.FixtureMap(t, "map")
	return filepath.Join(repo.Dir, "map")
}

func lockRelease(t *testing.T, repo *clitest.Repo) string {
	t.Helper()
	var lock struct {
		MapRelease string `json:"map_release"`
	}
	json.Unmarshal([]byte(clitest.Read(t, repo, "automap.lock.json")), &lock)
	return lock.MapRelease
}

var startWindowFromRelease = []string{"start", "--with", "stub-window", "--no-verify", "--json"}

func TestStartResolvesAgainstTheNewestRelease(t *testing.T) {
	s := newReleaseServer(t)
	mapDir := fixtureMapDir(t)
	s.publish(t, "map-2026.09.01", mapDir, nil)
	s.publish(t, "map-2026.10.01", mapDir, nil)
	repo := releaseRepo(t, s, t.TempDir())

	repo.Automap(t, startWindowFromRelease...).ExpectExit(t, 0)

	if got := lockRelease(t, repo); got != "map-2026.10.01" {
		t.Errorf("the lock records map release %q, want map-2026.10.01", got)
	}
}

func TestEveryCommandWorksOfflineOnceTheReleaseIsCached(t *testing.T) {
	s := newReleaseServer(t)
	s.publish(t, "map-2026.10.01", fixtureMapDir(t), nil)
	cache := t.TempDir()
	repo := releaseRepo(t, s, cache)
	repo.Automap(t, startWindowFromRelease...).ExpectExit(t, 0)
	s.Close() // the network is gone

	repo.Automap(t, "show", "library", "stub-window", "--json").ExpectExit(t, 0)
	repo.Automap(t, "extend", "stub-ui", "--no-verify", "--json").ExpectExit(t, 0)
	other := releaseRepo(t, s, cache)
	other.Automap(t, startWindowFromRelease...).ExpectExit(t, 0)
	if got := lockRelease(t, other); got != "map-2026.10.01" {
		t.Errorf("offline, start resolved against %q, want the newest cached release map-2026.10.01", got)
	}
}

func TestExtendKeepsTheLocksRelease(t *testing.T) {
	s := newReleaseServer(t)
	mapDir := fixtureMapDir(t)
	s.publish(t, "map-2026.09.01", mapDir, nil)
	repo := releaseRepo(t, s, t.TempDir())
	repo.Automap(t, startWindowFromRelease...).ExpectExit(t, 0)
	s.publish(t, "map-2026.10.01", mapDir, nil)

	repo.Automap(t, "extend", "stub-ui", "--no-verify", "--json").ExpectExit(t, 0)

	if got := lockRelease(t, repo); got != "map-2026.09.01" {
		t.Errorf("extend moved the lock to %q; moving to a newer release is upgrade's job", got)
	}
}

func TestAReleaseThatDoesntMatchItsSumIsRefused(t *testing.T) {
	s := newReleaseServer(t)
	s.publish(t, "map-2026.10.01", fixtureMapDir(t), func(line string) string { return strings.Repeat("0", 64) + line[64:] })
	repo := releaseRepo(t, s, t.TempDir())

	run := repo.Automap(t, startWindowFromRelease...)

	run.ExpectExit(t, 2)
	run.ExpectFailure(t, "MAP_RELEASE_INVALID(map-2026.10.01)", "SHA-256")
}

func TestWithNoReleaseAndNothingCachedThereIsNoMap(t *testing.T) {
	s := newReleaseServer(t)
	repo := releaseRepo(t, s, t.TempDir())
	s.Close()

	run := repo.Automap(t, startWindowFromRelease...)

	run.ExpectExit(t, 2)
	run.ExpectFailure(t, "NO_MAP", "")
}

func TestAMapDirectoryStillWinsOverReleases(t *testing.T) {
	s := newReleaseServer(t)
	s.publish(t, "map-2026.10.01", fixtureMapDir(t), nil)
	repo := releaseRepo(t, s, t.TempDir())
	repo.FixtureMap(t, "map")

	repo.Automap(t, "start", "--with", "stub-window", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)

	if got := lockRelease(t, repo); got != "unreleased" {
		t.Errorf("with --map the lock records %q, want unreleased", got)
	}
}

// A lock's release downloads without GitHub's releases API, which shared
// CI runners exhaust.
func TestTheLocksReleaseDownloadsWithoutTheReleasesAPI(t *testing.T) {
	s := newReleaseServer(t)
	s.publish(t, "map-2026.10.01", fixtureMapDir(t), nil)
	repo := releaseRepo(t, s, t.TempDir())
	repo.Automap(t, startWindowFromRelease...).ExpectExit(t, 0)
	s.mu.Lock()
	s.listing = false
	s.mu.Unlock()

	fresh := releaseRepo(t, s, t.TempDir()) // the same composition with nothing cached
	for _, f := range []string{"automap.manifest.json", "automap.lock.json"} {
		fresh.WriteFile(t, f, clitest.Read(t, repo, f))
	}

	fresh.Automap(t, "show", "library", "stub-window", "--json").ExpectExit(t, 0)
}
