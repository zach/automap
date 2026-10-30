// Command verifymap is Automap's CI for the public map (ADR-0033): it plans
// which declared edges need verifying, verifies one composition per job and
// platform, runs its seeded violations, and writes each edge's evidence for
// the map.
//
//	verifymap plan -map map -platforms linux-x64-clang,... -out matrix.json
//	verifymap run -map map -job job.json -automap <cli> -out <dir>
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/zach/automap/internal/composition"
	"github.com/zach/automap/internal/mapmodel"
	"github.com/zach/automap/internal/verify"
)

// runners are the hosted runners each v1 platform verifies on (SPEC §6.4).
var runners = map[string]string{
	"linux-x64-clang":        "ubuntu-latest",
	"windows-x64-msvc":       "windows-latest",
	"macos-arm64-appleclang": "macos-latest",
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: verifymap plan|run [flags]")
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "plan":
		err = plan(os.Args[2:])
	case "run":
		err = run(os.Args[2:])
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "verifymap:", err)
		os.Exit(1)
	}
}

func loadMap(dir string) (*mapmodel.Map, error) {
	m, problems := mapmodel.Load(dir, nil)
	if len(problems) > 0 {
		return nil, fmt.Errorf("the map has problems: %v", problems)
	}
	return m, nil
}

// plan writes the GitHub Actions matrix of jobs that need running.
func plan(args []string) error {
	fs := flag.NewFlagSet("plan", flag.ExitOnError)
	mapDir := fs.String("map", "map", "the public map")
	refsFile := fs.String("compositions", "ci/compositions.json", "the reference compositions")
	platforms := fs.String("platforms", "linux-x64-clang,windows-x64-msvc,macos-arm64-appleclang", "comma-separated platforms")
	all := fs.Bool("all", false, "verify every edge, whatever its evidence")
	out := fs.String("out", "matrix.json", "where to write the matrix")
	fs.Parse(args)

	m, err := loadMap(*mapDir)
	if err != nil {
		return err
	}
	refs, err := readReferences(*refsFile)
	if err != nil {
		return err
	}
	evidence := readPublished(*mapDir)
	published := func(e Edge, platform string) *Published {
		if *all {
			return nil
		}
		return evidence[publishedKey(e, platform)]
	}
	jobs := Plan(m, Edges(m, refs), strings.Split(*platforms, ","), published, gitChanged, time.Now())
	type entry struct {
		Name string `json:"name"`
		OS   string `json:"os"`
		Job  string `json:"job"`
	}
	matrix := struct {
		Include []entry `json:"include"`
	}{Include: []entry{}}
	for _, j := range jobs {
		raw, _ := json.Marshal(j)
		matrix.Include = append(matrix.Include, entry{Name: j.Name, OS: runners[j.Platform], Job: string(raw)})
	}
	raw, _ := json.Marshal(matrix)
	fmt.Printf("%d jobs, %d edge runs\n", len(jobs), edgeRuns(jobs))
	return os.WriteFile(*out, raw, 0o644)
}

func edgeRuns(jobs []Job) int {
	n := 0
	for _, j := range jobs {
		n += len(j.Edges)
	}
	return n
}

// publishedKey identifies an edge's evidence on a platform: a binding's
// or pairing's by its directory, a composition's by its id alone, since
// its directory names a lock hash known only after composing.
func publishedKey(e Edge, platform string) string {
	if e.Kind == "composition" {
		return "evidence/composition/" + e.ID + " " + platform
	}
	return e.Dir("") + " " + platform
}

// readPublished reads when and at which commit each piece of the map's
// evidence was verified, keeping a composition's newest.
func readPublished(mapDir string) map[string]*Published {
	out := map[string]*Published{}
	root := os.DirFS(mapDir)
	fs.WalkDir(root, "evidence", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".json") || strings.HasSuffix(p, ".sigstore.json") {
			return nil
		}
		raw, err := fs.ReadFile(root, p)
		if err != nil {
			return nil
		}
		var s Statement
		if json.Unmarshal(raw, &s) != nil {
			return nil
		}
		at, _ := time.Parse(time.RFC3339, s.Predicate.ProducedAt)
		pub := &Published{ProducedAt: at}
		pub.Commit, _ = strings.CutPrefix(s.Predicate.MapRelease, "git:")
		if pub.Commit == s.Predicate.MapRelease {
			pub.Commit = ""
		}
		parts := strings.Split(p, "/") // evidence/<kind>/<id>/<pins>/<platform>.json
		if len(parts) != 5 {
			return nil
		}
		platform := strings.TrimSuffix(parts[4], ".json")
		key := strings.Join(parts[:4], "/") + " " + platform
		if parts[1] == "composition" {
			key = strings.Join(parts[:3], "/") + " " + platform
		}
		if prev := out[key]; prev == nil || at.After(prev.ProducedAt) {
			out[key] = pub
		}
		return nil
	})
	return out
}

// gitChanged reports whether any input changed between a commit and HEAD.
// A commit git doesn't know counts as a change.
func gitChanged(commit string, inputs []string) bool {
	out, err := exec.Command("git", "diff", "--name-only", commit, "HEAD", "--").Output()
	if err != nil {
		return true
	}
	return touches(strings.Fields(string(out)))(commit, inputs)
}

// run verifies one job's composition, runs its seeds, and writes each
// edge's evidence under the output directory, at its path in the map. It
// fails when a probe missed its seed, after writing the evidence.
func run(args []string) error {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	mapDir := fs.String("map", "map", "the public map")
	jobJSON := fs.String("job", "", "the job, as the plan's matrix gives it")
	automap := fs.String("automap", "automap", "the automap CLI")
	out := fs.String("out", "evidence-out", "where to write the edges' evidence, laid out as in the map")
	fs.Parse(args)

	var job Job
	if err := json.Unmarshal([]byte(*jobJSON), &job); err != nil {
		return fmt.Errorf("-job: %v", err)
	}
	mapAbs, _ := filepath.Abs(*mapDir)
	m, err := loadMap(mapAbs)
	if err != nil {
		return err
	}
	head, err := cleanHead(mapAbs)
	if err != nil {
		return err
	}
	work, err := os.MkdirTemp("", "verifymap-")
	if err != nil {
		return err
	}
	// The composition's map: the public map itself when every library is
	// at its newest version, or a copy keeping only the picked versions.
	compMap := mapAbs
	if !allNewest(m, job.Picks) {
		compMap = filepath.Join(work, "map")
		if err := copyMapWithVersions(mapAbs, compMap, job.Picks); err != nil {
			return err
		}
	}
	repo := filepath.Join(work, "repo")
	os.MkdirAll(repo, 0o755)
	cli := func(args ...string) (map[string]any, error) {
		cmd := exec.Command(*automap, append(args, "--map", compMap, "--platform", job.Platform, "--json")...)
		cmd.Dir = repo
		var stdout bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, os.Stderr
		err := cmd.Run()
		var result map[string]any
		json.Unmarshal(stdout.Bytes(), &result)
		fmt.Printf("automap %s: %v\n", strings.Join(args, " "), result["summary"])
		return result, err
	}

	start := []string{"start"}
	for _, p := range job.Picks {
		start = append(start, "--with", p.Library+":"+p.Variant)
	}
	_, startErr := cli(start...) // a failing verification still writes evidence
	raw, err := os.ReadFile(filepath.Join(repo, "evidence", job.Platform+".json"))
	if err != nil {
		return fmt.Errorf("the composition wrote no evidence (automap start: %v): %v", startErr, err)
	}
	var comp Statement
	if err := json.Unmarshal(raw, &comp); err != nil {
		return err
	}
	comp.Predicate.MapRelease = "git:" + head
	lockFile, _ := os.ReadFile(filepath.Join(repo, composition.LockFile))
	var lock composition.Lock
	json.Unmarshal(lockFile, &lock)
	var man composition.Manifest
	if raw, err := os.ReadFile(filepath.Join(repo, composition.ManifestFile)); err == nil {
		json.Unmarshal(raw, &man)
	}

	// Every seed whose rule applies here must be caught (ADR-0071). A
	// library rule's seeds run wherever the library is; a composition
	// rule's mistake depends on the libraries around it, so its seeds run
	// in the reference compositions (SPEC §6.1). Each seed runs in its own
	// scratch copy, so a few run at once.
	reference := slices.ContainsFunc(job.Edges, func(e Edge) bool { return e.Kind == "composition" })
	var (
		uncaught []Uncaught
		mu       sync.Mutex
		wg       sync.WaitGroup
		slots    = make(chan struct{}, 4)
	)
	for _, s := range verify.Seeds(m, &man) {
		if !applies(comp, s) || s.Library == "" && !reference {
			continue
		}
		spec := s.Rule
		if name := cmp(s.Library, s.Subject); name != "" {
			spec += ":" + name
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			slots <- struct{}{}
			defer func() { <-slots }()
			result, err := cli("verify", "--seed", spec)
			if err != nil || result["ok"] != true {
				fmt.Printf("::error::the probe for %s didn't catch its seeded violation on %s\n", spec, job.Platform)
				mu.Lock()
				uncaught = append(uncaught, Uncaught{Rule: s.Rule, Library: s.Library})
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	verifier := os.Getenv("GITHUB_SERVER_URL") + "/" + os.Getenv("GITHUB_WORKFLOW_REF")
	if os.Getenv("GITHUB_WORKFLOW_REF") == "" {
		verifier = "verifymap"
	}
	for _, e := range job.Edges {
		s := EdgeEvidence(m, e, comp, lock, lockFile, uncaught, verifier)
		if s == nil {
			fmt.Printf("%s %s: the composition reports no result for its rules\n", e.Kind, e.ID)
			continue
		}
		file := filepath.Join(*out, filepath.FromSlash(e.Dir(lock.LockHash)), job.Platform+".json")
		os.MkdirAll(filepath.Dir(file), 0o755)
		raw, _ := json.MarshalIndent(s, "", "  ")
		if err := os.WriteFile(file, append(raw, '\n'), 0o644); err != nil {
			return err
		}
		fmt.Printf("%s %s: %s\n", e.Kind, e.ID, s.Predicate.Level)
	}
	if len(uncaught) > 0 {
		return fmt.Errorf("%d probes didn't catch their seeded violations", len(uncaught))
	}
	return nil
}

// applies reports whether a seed's rule applies in the composition's run:
// a rule not-applicable there, such as ThreadSanitizer's on Windows, has
// nothing to catch.
func applies(comp Statement, s verify.Seed) bool {
	for _, r := range comp.Predicate.Results {
		if r.Rule == s.Rule && (s.Library == "" || r.Library == s.Library) {
			return r.Status != "not-applicable"
		}
	}
	return false
}

func cmp(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// cleanHead is the commit the map is checked out at, which the evidence
// names; a map with uncommitted changes can't be published.
func cleanHead(dir string) (string, error) {
	status, err := exec.Command("git", "-C", dir, "status", "--porcelain", "--", ".").Output()
	if err != nil {
		return "", fmt.Errorf("the map isn't a git checkout: %v", err)
	}
	if len(bytes.TrimSpace(status)) > 0 {
		return "", errors.New("the map has uncommitted changes")
	}
	head, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD").Output()
	return strings.TrimSpace(string(head)), err
}

func allNewest(m *mapmodel.Map, picks []Pick) bool {
	return !slices.ContainsFunc(picks, func(p Pick) bool {
		l := m.Libraries[p.Library]
		return l == nil || l.Newest().Version != p.Version
	})
}

// copyMapWithVersions copies the map, keeping only the picked version of
// each picked library, so automap start composes it.
func copyMapWithVersions(src, dst string, picks []Pick) error {
	if err := os.CopyFS(dst, os.DirFS(src)); err != nil {
		return err
	}
	for _, p := range picks {
		file := filepath.Join(dst, "libraries", p.Library+".json")
		raw, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		var doc map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil {
			return err
		}
		versions, _ := doc["versions"].([]any)
		var keep []any
		for _, v := range versions {
			if vm, _ := v.(map[string]any); vm["version"] == p.Version {
				keep = append(keep, v)
			}
		}
		doc["versions"] = keep
		raw, _ = json.MarshalIndent(doc, "", "  ")
		if err := os.WriteFile(file, raw, 0o644); err != nil {
			return err
		}
	}
	return nil
}
