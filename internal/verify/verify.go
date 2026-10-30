// Package verify builds a composition with CMake and vcpkg, runs its
// executable headless, and writes evidence at the level it reached (SPEC
// §6.2).
package verify

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/zach/automap/internal/composition"
	"github.com/zach/automap/internal/mapmodel"
)

// SmokeFrames is how many headless frames the smoke run runs (SPEC §5.6).
const SmokeFrames = 120

// smokeTimeout bounds the smoke run, so a hang fails instead of stalling.
const smokeTimeout = 2 * time.Minute

// Failure is one failed check, with its failure code and details file.
type Failure struct {
	Code        string
	Rule        string
	Library     string
	Message     string
	DetailsFile string
}

// Outcome is what one verification run reached.
type Outcome struct {
	Platform     string
	Level        string
	Results      []Result
	Failures     []Failure
	EvidenceFile string
}

// Options say where a verification run happens and what it writes.
type Options struct {
	Root     string // the composition's directory
	Platform string
	Evidence bool // write evidence/<platform>.json
}

// HostPlatform is the v1 platform id of the machine the CLI runs on.
func HostPlatform() string {
	arch := map[string]string{"amd64": "x64", "arm64": "arm64"}[runtime.GOARCH]
	switch runtime.GOOS {
	case "windows":
		return "windows-" + arch + "-msvc"
	case "darwin":
		return "macos-" + arch + "-appleclang"
	default:
		return "linux-" + arch + "-clang"
	}
}

// Run verifies a composition: it builds and runs the smoke test, then
// builds with AUTOMAP_PROBE in a separate directory and runs every probe,
// judges lifecycle order and platform targets from the probe run, assigns
// every applicable probed rule a status, and computes the level.
func Run(m *mapmodel.Map, man *composition.Manifest, lock *composition.Lock, opts Options) *Outcome {
	platform := opts.Platform
	out := &Outcome{Platform: platform}
	const logs = "build/automap"
	paths := struct{ build, probeBuild, buildLog, probeBuildLog, smokeLog, probeLog, results string }{
		build:         path.Join(logs, platform),
		probeBuild:    path.Join(logs, platform+"-probe"),
		buildLog:      path.Join(logs, "build-"+platform+".log"),
		probeBuildLog: path.Join(logs, "probe-build-"+platform+".log"),
		smokeLog:      path.Join(logs, "smoke-"+platform+".log"),
		probeLog:      path.Join(logs, "probe-"+platform+".log"),
		results:       path.Join(logs, "probe-results-"+platform+".json"),
	}
	in := func(p string) string { return filepath.Join(opts.Root, filepath.FromSlash(p)) }
	os.MkdirAll(in(logs), 0o755)
	os.Remove(in(paths.results))

	env := headlessEnv(m, man)
	var built, smoked, probesRan bool
	if err := build(opts.Root, paths.build, paths.buildLog, platform, man.Executable.Target, false); err != nil {
		out.Failures = append(out.Failures, Failure{Code: "BUILD_FAILED", Message: err.Error(), DetailsFile: paths.buildLog})
	} else {
		built = true
	}
	if built {
		if err := runExecutable(opts.Root, paths.build, paths.smokeLog, man.Executable.Target, env, ""); err != nil {
			out.Failures = append(out.Failures, Failure{Code: "SMOKE_FAILED", Message: err.Error(), DetailsFile: paths.smokeLog})
		} else {
			smoked = true
		}
	}
	reached := Level(built, smoked, false, nil)
	obs := Observed{Level: reached}
	if smoked {
		if err := build(opts.Root, paths.probeBuild, paths.probeBuildLog, platform, man.Executable.Target, true); err != nil {
			out.Failures = append(out.Failures, Failure{Code: "BUILD_FAILED", Message: "The probe build failed: " + err.Error(), DetailsFile: paths.probeBuildLog})
		} else {
			err := runExecutable(opts.Root, paths.probeBuild, paths.probeLog, man.Executable.Target, env, paths.results)
			harness, readErr := readResults(in(paths.results))
			switch {
			case readErr != nil && err != nil:
				out.Failures = append(out.Failures, Failure{Code: "PROBE_FAILED", Message: "The probe run failed: " + err.Error(), DetailsFile: paths.probeLog})
			case readErr != nil:
				out.Failures = append(out.Failures, Failure{Code: "PROBE_FAILED", Message: "The probe run wrote no results: " + readErr.Error(), DetailsFile: paths.probeLog})
			default:
				probesRan = true
				obs.Ran = true
				obs.Reported = harness.Results
				obs.Computed = append(CheckLifecycle(DeclaredLifecycle(m, man), harness.InitSteps, harness.ShutdownSteps), CheckTargets(m, man))
				if err != nil {
					out.Failures = append(out.Failures, Failure{Code: "PROBE_FAILED", Message: "The probe run didn't exit cleanly: " + err.Error(), DetailsFile: paths.probeLog})
				}
			}
		}
	}
	out.Results = Results(ExpectedRules(m, man), platform, obs)
	out.Level = Level(built, smoked, probesRan, out.Results)
	texts := map[string]string{}
	for _, e := range ExpectedRules(m, man) {
		texts[e.Rule.ID] = e.Rule.Text
	}
	for _, r := range out.Results {
		if r.Status == "fail" {
			msg := r.Reason
			if msg == "" {
				msg = texts[r.Rule]
			}
			out.Failures = append(out.Failures, Failure{Code: r.FailureCode, Rule: r.Rule, Library: r.Library, Message: msg, DetailsFile: paths.probeLog})
		}
	}
	if !opts.Evidence {
		return out
	}
	evidence, err := Evidence(lock, EvidenceInput{
		LockFile:  in(composition.LockFile),
		Platform:  platform,
		Level:     out.Level,
		Compiler:  compilerVersion(platform),
		BuildType: "Debug",
		Env:       env,
		Time:      time.Now().UTC(),
		Results:   out.Results,
	})
	if err != nil {
		out.Failures = append(out.Failures, Failure{Code: "EVIDENCE_FAILED", Message: err.Error()})
		return out
	}
	out.EvidenceFile = path.Join("evidence", platform+".json")
	if err := os.MkdirAll(in("evidence"), 0o755); err == nil {
		err = os.WriteFile(in(out.EvidenceFile), evidence, 0o644)
	}
	if err != nil {
		out.Failures = append(out.Failures, Failure{Code: "EVIDENCE_FAILED", Message: err.Error()})
	}
	return out
}

// harnessResults is the probe harness's results file (SPEC §5.6).
type harnessResults struct {
	Results       []Result `json:"results"`
	InitSteps     []string `json:"init_steps"`
	ShutdownSteps []string `json:"shutdown_steps"`
}

func readResults(file string) (*harnessResults, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var h harnessResults
	if err := json.Unmarshal(raw, &h); err != nil {
		return nil, fmt.Errorf("%s isn't valid JSON: %v", file, err)
	}
	return &h, nil
}

// build configures and builds the executable through CMake with vcpkg's
// toolchain, logging to logFile. A probe build defines AUTOMAP_PROBE.
func build(root, buildDir, logFile, platform, target string, probe bool) error {
	vcpkg := os.Getenv("VCPKG_ROOT")
	if vcpkg == "" {
		return errors.New("VCPKG_ROOT isn't set: set it to the root of a vcpkg installation")
	}
	log, err := os.Create(filepath.Join(root, filepath.FromSlash(logFile)))
	if err != nil {
		return err
	}
	defer log.Close()
	configure := []string{"-S", root, "-B", filepath.Join(root, filepath.FromSlash(buildDir)), "-G", "Ninja",
		"-DCMAKE_BUILD_TYPE=Debug",
		"-DCMAKE_TOOLCHAIN_FILE=" + filepath.Join(vcpkg, "scripts", "buildsystems", "vcpkg.cmake")}
	if probe {
		configure = append(configure, "-DAUTOMAP_PROBE=ON")
	}
	cmdEnv := os.Environ()
	if strings.HasSuffix(platform, "-clang") {
		configure = append(configure, "-DCMAKE_C_COMPILER=clang", "-DCMAKE_CXX_COMPILER=clang++")
		cmdEnv = append(cmdEnv, "CC=clang", "CXX=clang++")
	}
	for _, step := range [][]string{configure, {"--build", filepath.Join(root, filepath.FromSlash(buildDir)), "--target", target}} {
		fmt.Fprintf(log, "$ cmake %s\n", strings.Join(step, " "))
		cmd := exec.Command("cmake", step...)
		cmd.Stdout, cmd.Stderr, cmd.Env = log, log, cmdEnv
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("cmake %s failed (%v); the build log has the details", step[0], err)
		}
	}
	return nil
}

// runExecutable runs the executable headless for SmokeFrames frames, with
// --probe <results> when results isn't empty, and checks it exits cleanly.
func runExecutable(root, buildDir, logFile, target string, env []string, results string) error {
	log, err := os.Create(filepath.Join(root, filepath.FromSlash(logFile)))
	if err != nil {
		return err
	}
	defer log.Close()
	exe, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(buildDir), target))
	if err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	ctx, cancel := context.WithTimeout(context.Background(), smokeTimeout)
	defer cancel()
	args := []string{"--frames", fmt.Sprint(SmokeFrames)}
	if results != "" {
		abs, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(results)))
		if err != nil {
			return err
		}
		args = append(args, "--probe", abs)
	}
	fmt.Fprintf(log, "$ %s %s %s\n", strings.Join(env, " "), exe, strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, exe, args...)
	var stderr bytes.Buffer
	// Run from the build directory, so files a library writes as it runs,
	// such as imgui.ini, never land in the developer's repo.
	cmd.Dir = filepath.Join(root, filepath.FromSlash(buildDir))
	cmd.Stdout, cmd.Stderr = log, &stderr
	cmd.Env = append(os.Environ(), env...)
	err = cmd.Run()
	log.Write(stderr.Bytes())
	if ctx.Err() != nil {
		return fmt.Errorf("%s ran %d frames for longer than %s", target, SmokeFrames, smokeTimeout)
	}
	if err != nil {
		if code := initFailed.FindString(stderr.String()); code != "" {
			return fmt.Errorf("%s didn't start: %s (%v)", target, code, err)
		}
		return fmt.Errorf("%s didn't run %d headless frames and shut down cleanly: %v", target, SmokeFrames, err)
	}
	return nil
}

var initFailed = regexp.MustCompile(`INIT_FAILED\([^)]*\)`)

// headlessEnv is the environment each library's platform facts say it
// needs to run headless, such as STUB_WINDOW_HEADLESS=1.
func headlessEnv(m *mapmodel.Map, man *composition.Manifest) []string {
	assignment := regexp.MustCompile(`^[A-Z_][A-Z0-9_]*=\S*$`)
	var env []string
	for _, lib := range man.Libraries() {
		l := m.Libraries[lib]
		if l == nil {
			continue
		}
		if h, ok := l.Newest().Facts["platform"]["headless"].(string); ok && assignment.MatchString(h) && !slices.Contains(env, h) {
			env = append(env, h)
		}
	}
	return env
}

// compilerVersion names the C++ compiler the platform builds with.
func compilerVersion(platform string) string {
	name := "c++"
	switch {
	case strings.HasSuffix(platform, "-clang"):
		name = "clang++"
	case strings.HasSuffix(platform, "-gcc"):
		name = "g++"
	case strings.HasSuffix(platform, "-msvc"):
		return "msvc"
	}
	out, err := exec.Command(name, "--version").Output()
	if err != nil {
		return name
	}
	first, _, _ := strings.Cut(string(out), "\n")
	return strings.TrimSpace(first)
}

// EvidenceInput is what a verification run observed.
type EvidenceInput struct {
	LockFile  string
	Platform  string
	Level     string
	Compiler  string
	BuildType string
	Env       []string
	Time      time.Time
	Results   []Result
}

// Evidence is the composition's evidence for one run: an in-toto v1
// Statement about the lock file, unsigned, from the developer's machine
// (SPEC §4.9, §6.3).
func Evidence(lock *composition.Lock, in EvidenceInput) ([]byte, error) {
	raw, err := os.ReadFile(in.LockFile)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(raw)
	environment := map[string]string{}
	for _, kv := range in.Env {
		k, v, _ := strings.Cut(kv, "=")
		environment[k] = v
	}
	conditions := map[string]any{
		"build_type":  in.BuildType,
		"compiler":    in.Compiler,
		"environment": environment,
		"sanitizers":  map[string]string{},
	}
	if img := os.Getenv("ImageOS"); img != "" {
		conditions["runner_image"] = img + " " + os.Getenv("ImageVersion")
	}
	statement := map[string]any{
		"_type":         "https://in-toto.io/Statement/v1",
		"subject":       []any{map[string]any{"name": composition.LockFile, "digest": map[string]string{"sha256": hex.EncodeToString(sum[:])}}},
		"predicateType": "https://github.com/zach/automap/evidence/v1",
		"predicate": map[string]any{
			"subject_kind":   "composition",
			"subject_detail": map[string]string{"lock_hash": lock.LockHash},
			"platform":       in.Platform,
			"level":          in.Level,
			"produced_at":    in.Time.Format(time.RFC3339),
			"map_release":    lock.MapRelease,
			"conditions":     conditions,
			"results":        in.Results,
			"verifier":       map[string]any{"kind": "developer-local", "id": "local", "signed": false},
		},
	}
	out, err := json.MarshalIndent(statement, "", "  ")
	return append(out, '\n'), err
}
