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
	Message     string
	DetailsFile string
}

// Outcome is what one verification run reached.
type Outcome struct {
	Platform     string
	Level        string
	Failures     []Failure
	EvidenceFile string
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

// Run verifies the composition in the current directory on the host
// platform and writes its evidence.
func Run(m *mapmodel.Map, man *composition.Manifest, lock *composition.Lock, platform string) *Outcome {
	out := &Outcome{Platform: platform, Level: "none"}
	buildDir := path.Join("build/automap", platform)
	buildLog := path.Join("build/automap", "build-"+platform+".log")
	smokeLog := path.Join("build/automap", "smoke-"+platform+".log")
	if err := os.MkdirAll(filepath.FromSlash(buildDir), 0o755); err != nil {
		out.Failures = append(out.Failures, Failure{Code: "BUILD_FAILED", Message: err.Error()})
		return out
	}

	env := []string{}
	if err := build(buildDir, buildLog, platform, man.Executable.Target); err != nil {
		out.Failures = append(out.Failures, Failure{Code: "BUILD_FAILED", Message: err.Error(), DetailsFile: buildLog})
	} else {
		out.Level = "builds"
		env = headlessEnv(m, man)
		if err := smoke(buildDir, smokeLog, man.Executable.Target, env); err != nil {
			out.Failures = append(out.Failures, Failure{Code: "SMOKE_FAILED", Message: err.Error(), DetailsFile: smokeLog})
		} else {
			out.Level = "smoke"
		}
	}
	compiler := compilerVersion(platform)

	evidence, err := Evidence(m, man, lock, EvidenceInput{
		Platform:  platform,
		Level:     out.Level,
		Compiler:  compiler,
		BuildType: "Debug",
		Env:       env,
		Time:      time.Now().UTC(),
	})
	if err != nil {
		out.Failures = append(out.Failures, Failure{Code: "EVIDENCE_FAILED", Message: err.Error()})
		return out
	}
	out.EvidenceFile = path.Join("evidence", platform+".json")
	if err := os.MkdirAll("evidence", 0o755); err == nil {
		err = os.WriteFile(filepath.FromSlash(out.EvidenceFile), evidence, 0o644)
	}
	if err != nil {
		out.Failures = append(out.Failures, Failure{Code: "EVIDENCE_FAILED", Message: err.Error()})
	}
	return out
}

// build configures and builds the executable through CMake with vcpkg's
// toolchain, logging to logFile.
func build(buildDir, logFile, platform, target string) error {
	root := os.Getenv("VCPKG_ROOT")
	if root == "" {
		return errors.New("VCPKG_ROOT isn't set: set it to the root of a vcpkg installation")
	}
	log, err := os.Create(filepath.FromSlash(logFile))
	if err != nil {
		return err
	}
	defer log.Close()
	configure := []string{"-S", ".", "-B", buildDir, "-G", "Ninja",
		"-DCMAKE_BUILD_TYPE=Debug",
		"-DCMAKE_TOOLCHAIN_FILE=" + filepath.Join(root, "scripts", "buildsystems", "vcpkg.cmake")}
	cmdEnv := os.Environ()
	if strings.HasSuffix(platform, "-clang") {
		configure = append(configure, "-DCMAKE_C_COMPILER=clang", "-DCMAKE_CXX_COMPILER=clang++")
		cmdEnv = append(cmdEnv, "CC=clang", "CXX=clang++")
	}
	for _, step := range [][]string{configure, {"--build", buildDir, "--target", target}} {
		fmt.Fprintf(log, "$ cmake %s\n", strings.Join(step, " "))
		cmd := exec.Command("cmake", step...)
		cmd.Stdout, cmd.Stderr, cmd.Env = log, log, cmdEnv
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("cmake %s failed (%v); the build log has the details", step[0], err)
		}
	}
	return nil
}

// smoke runs the executable headless for SmokeFrames frames and checks it
// exits cleanly.
func smoke(buildDir, logFile, target string, env []string) error {
	log, err := os.Create(filepath.FromSlash(logFile))
	if err != nil {
		return err
	}
	defer log.Close()
	exe := filepath.Join(filepath.FromSlash(buildDir), target)
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	ctx, cancel := context.WithTimeout(context.Background(), smokeTimeout)
	defer cancel()
	args := []string{"--frames", fmt.Sprint(SmokeFrames)}
	fmt.Fprintf(log, "$ %s %s %s\n", strings.Join(env, " "), exe, strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, exe, args...)
	var stderr bytes.Buffer
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
	Platform  string
	Level     string
	Compiler  string
	BuildType string
	Env       []string
	Time      time.Time
}

// Evidence is the composition's evidence for one run: an in-toto v1
// Statement about the lock file, unsigned, from the developer's machine
// (SPEC §4.9, §6.3).
func Evidence(m *mapmodel.Map, man *composition.Manifest, lock *composition.Lock, in EvidenceInput) ([]byte, error) {
	raw, err := os.ReadFile(composition.LockFile)
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
			"results":        unrunResults(m, man, in.Platform, in.Level),
			"verifier":       map[string]any{"kind": "developer-local", "id": "local", "signed": false},
		},
	}
	out, err := json.MarshalIndent(statement, "", "  ")
	return append(out, '\n'), err
}

// unrunResults records every probed rule that applies to the composition
// as unverifiable, since no probe ran, and each platform-limited rule
// elsewhere as not-applicable (ADR-0072, ADR-0076).
func unrunResults(m *mapmodel.Map, man *composition.Manifest, platform, level string) []map[string]string {
	var out []map[string]string
	add := func(rule mapmodel.Rule, library string) {
		r := map[string]string{"rule": rule.ID}
		if library != "" {
			r["library"] = library
		}
		if len(rule.Platforms) > 0 && !slices.Contains(rule.Platforms, platform) {
			r["status"], r["reason"] = "not-applicable", "The rule's probe runs only on "+strings.Join(rule.Platforms, ", ")+"."
		} else {
			r["status"], r["reason"] = "unverifiable", "No probe ran: the composition reached "+level+"."
		}
		out = append(out, r)
	}
	for _, cid := range composition.Contracts(m, man) {
		for _, rule := range m.LatestContract(cid).Rules {
			if rule.Kind != "probed" {
				continue
			}
			if rule.AppliesTo != "library" {
				add(rule, "")
				continue
			}
			for _, lib := range man.Libraries() {
				if _, ok := m.Libraries[lib].Newest().Facts[cid]; ok {
					add(rule, lib)
				}
			}
		}
	}
	for _, rid := range composition.Recipes(m, man) {
		for _, rule := range m.Recipes[rid].Rules {
			if rule.Kind == "probed" {
				add(rule, "")
			}
		}
	}
	return out
}
