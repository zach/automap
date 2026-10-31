package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zach/automap/internal/agentfiles"
	"github.com/zach/automap/internal/composition"
	"github.com/zach/automap/internal/mapmodel"
	"github.com/zach/automap/internal/verify"
	"github.com/zach/automap/schemas"
)

// verifyCmd builds the composition in the current directory, runs its
// smoke test, and writes evidence (SPEC §5.4).
func verifyCmd(a *args, r *result) int {
	if _, err := os.Stat(manifestFile); os.IsNotExist(err) {
		r.Summary = "Nothing to verify: no " + manifestFile + " in the current directory."
		r.fail("NO_MANIFEST", "No composition here. Run automap start to create one, or automap adopt to adopt an existing engine.")
		return exitUsage
	}
	m, man, lock, status := loadComposition(a, r)
	if m == nil {
		return status
	}
	platform := a.value("platform")
	if platform == "" {
		platform = verify.HostPlatform()
	}
	if platform != verify.HostPlatform() {
		r.Summary = "verify runs on the host platform, " + verify.HostPlatform() + "."
		r.fail("USAGE", "Run verify on a "+platform+" machine to verify that platform.")
		return exitUsage
	}
	if seed := a.value("seed"); seed != "" {
		return runSeed(m, man, lock, platform, seed, r)
	}
	return runVerify(m, man, lock, platform, r)
}

// runVerify verifies on platform and records the outcome in r.
func runVerify(m *mapmodel.Map, man *composition.Manifest, lock *composition.Lock, platform string, r *result) int {
	if !contains(man.Platforms, platform) {
		r.Summary = "The composition doesn't target " + platform + "."
		r.fail("PLATFORM_NOT_IN_COMPOSITION("+platform+")", "The manifest's platforms are "+fmt.Sprint(man.Platforms)+".")
		return exitUsage
	}
	out := verify.Run(m, man, lock, verify.Options{Root: ".", Platform: platform, Evidence: true})
	r.Level = out.Level
	r.EvidenceFile = out.EvidenceFile
	for _, f := range out.Failures {
		r.Failures = append(r.Failures, failure{Code: f.Code, Rule: f.Rule, Library: f.Library, Message: f.Message, DetailsFile: f.DetailsFile})
	}
	var unverifiable int
	for _, res := range out.Results {
		if res.Status == "unverifiable" && out.Level != "none" && out.Level != "builds" {
			unverifiable++
			code := "UNVERIFIABLE(" + res.Rule + ")"
			if res.Library != "" {
				code = "UNVERIFIABLE(" + res.Rule + "," + res.Library + ")"
			}
			r.Warnings = append(r.Warnings, warning{Code: code, Message: res.Reason})
		}
	}
	warnEvidence(r, compositionEvidence(m, man, lock, time.Now()))
	r.OK = len(r.Failures) == 0
	switch {
	case !r.OK:
		var rules int
		for _, f := range out.Failures {
			if f.Rule != "" {
				rules++
			}
		}
		detail := out.Failures[0].Code
		if rules > 0 {
			detail = fmt.Sprintf("%d probed %s failed", rules, plural(rules, "rule", "rules"))
		}
		r.Summary = fmt.Sprintf("Composition is %s on %s: %s.", out.Level, platform, detail)
		return exitFailed
	case unverifiable > 0:
		r.Summary = fmt.Sprintf("Composition is %s on %s: %d probed %s unverifiable.", out.Level, platform, unverifiable, plural(unverifiable, "rule", "rules"))
	default:
		r.Summary = fmt.Sprintf("Composition is %s on %s.", out.Level, platform)
	}
	return exitOK
}

// runSeed applies one seeded violation in a scratch copy of the
// composition, verifies there, and reports whether the rule's probe caught
// it. It writes no evidence (SPEC §6.1).
func runSeed(m *mapmodel.Map, man *composition.Manifest, lock *composition.Lock, platform, spec string, r *result) int {
	rule, library, _ := strings.Cut(spec, ":")
	var matches []verify.Seed
	for _, s := range verify.Seeds(m, man) {
		// A composition rule's seed is named by the library or pairing
		// whose recipe carries it.
		if s.Rule == rule && (library == "" || s.Library == library || s.Library == "" && s.Subject == library) {
			matches = append(matches, s)
		}
	}
	switch {
	case len(matches) == 0:
		r.Summary = "No seeded violation for " + spec + " applies to this composition."
		r.fail("UNKNOWN_SEED("+spec+")", "No recipe or contract in the composition seeds "+spec+".")
		return exitUsage
	case len(matches) > 1:
		var names []string
		for _, s := range matches {
			name := s.Library
			if name == "" {
				name = s.Subject
			}
			names = append(names, s.Rule+":"+name)
		}
		r.Summary = "More than one seeded violation matches " + spec + "."
		r.fail("USAGE", "Name the library: "+strings.Join(names, ", "))
		return exitUsage
	}
	seed := matches[0]
	scratch, err := os.MkdirTemp("", "automap-seed-")
	if err != nil {
		r.Summary = "Can't make a scratch copy."
		r.fail("SEED_FAILED", err.Error())
		return exitFailed
	}
	defer os.RemoveAll(scratch)
	if err := copyComposition(".", scratch); err == nil {
		err = verify.ApplySeed(scratch, seed)
	}
	if err != nil {
		r.Summary = "The seeded violation can't be applied."
		r.fail("SEED_FAILED", err.Error())
		return exitFailed
	}
	seeded := *man
	if raw, err := os.ReadFile(filepath.Join(scratch, composition.ManifestFile)); err == nil {
		json.Unmarshal(raw, &seeded)
	}
	out := verify.Run(m, &seeded, lock, verify.Options{Root: scratch, Platform: platform})
	r.Level = out.Level
	for _, res := range out.Results {
		if res.Rule == seed.Rule && res.Library == seed.Library && res.Status == "fail" {
			r.OK = true
			r.Summary = fmt.Sprintf("The probe caught the seeded violation of %s: %s.", spec, res.FailureCode)
			return exitOK
		}
	}
	// The scratch copy goes away: keep the log that says why, the failing
	// run's or else the probe run's (ThreadSanitizer's for its rule).
	got, log := "it reported no failure", "build/automap/probe-"+platform+".log"
	if seed.Rule == verify.TSanRule {
		log = "build/automap/tsan-" + platform + ".log"
	}
	for _, f := range out.Failures {
		if f.Rule == "" {
			got, log = "the run failed with "+f.Code+" instead", f.DetailsFile
			break
		}
	}
	details := ""
	if data, err := os.ReadFile(filepath.Join(scratch, filepath.FromSlash(log))); err == nil && log != "" {
		details = "build/automap/seed-" + strings.NewReplacer(":", "-", "/", "-").Replace(spec) + ".log"
		os.MkdirAll("build/automap", 0o755)
		os.WriteFile(filepath.FromSlash(details), data, 0o644)
	}
	r.Summary = fmt.Sprintf("The probe didn't catch the seeded violation of %s.", spec)
	r.fail("SEED_NOT_CAUGHT("+spec+")", seed.Description+" The probe for "+seed.Rule+" didn't fail: "+got+".")
	r.Failures[len(r.Failures)-1].DetailsFile = details
	return exitFailed
}

// copyComposition copies a composition's files to dst, leaving out build
// output, evidence and version control.
func copyComposition(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		if d.IsDir() {
			switch filepath.ToSlash(rel) {
			case "build", "evidence", ".git":
				return filepath.SkipDir
			}
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dst, rel), data, 0o644)
	})
}

// loadComposition reads the manifest and lock in the current directory and
// the maps they name, and checks them against each other.
func loadComposition(a *args, r *result) (*mapmodel.Map, *composition.Manifest, *composition.Lock, int) {
	raw, err := os.ReadFile(manifestFile)
	var doc any
	if err == nil {
		err = json.Unmarshal(raw, &doc)
	}
	if err != nil {
		r.Summary = "The manifest can't be read."
		r.fail("MANIFEST_INVALID", err.Error())
		return nil, nil, nil, exitUsage
	}
	if v := schemas.Validate("manifest", doc); v != nil {
		r.Summary = "The manifest doesn't match its schema."
		for _, msg := range v {
			r.fail("MANIFEST_INVALID", "manifest: "+msg)
		}
		return nil, nil, nil, exitUsage
	}
	var man composition.Manifest
	json.Unmarshal(raw, &man)
	rawLock, err := os.ReadFile(composition.LockFile)
	if err != nil {
		r.Summary = "The composition has no lock."
		r.fail("NO_LOCK", composition.LockFile+" is missing. Only the CLI writes the lock: restore it from version control.")
		return nil, nil, nil, exitUsage
	}
	lock, err := composition.CheckLock(rawLock)
	if err != nil {
		r.Summary = "Refusing to act on a hand-edited lock."
		r.fail("LOCK_EDITED", composition.LockFile+" was edited by hand: "+err.Error()+". "+agentfiles.Never(1))
		return nil, nil, nil, exitRefused
	}
	m, status := loadMap(a, r)
	if m == nil {
		return nil, nil, nil, status
	}
	if fails := composition.Check(m, &man, lock); fails != nil {
		failWith(r, fails, "The composition doesn't match the map.", exitUsage)
		return nil, nil, nil, exitUsage
	}
	return m, &man, lock, exitOK
}

func contains(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}
