package cli

import (
	"encoding/json"
	"fmt"
	"os"

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
	return runVerify(m, man, lock, platform, r)
}

// runVerify verifies on platform and records the outcome in r.
func runVerify(m *mapmodel.Map, man *composition.Manifest, lock *composition.Lock, platform string, r *result) int {
	if !contains(man.Platforms, platform) {
		r.Summary = "The composition doesn't target " + platform + "."
		r.fail("PLATFORM_NOT_IN_COMPOSITION("+platform+")", "The manifest's platforms are "+fmt.Sprint(man.Platforms)+".")
		return exitUsage
	}
	out := verify.Run(m, man, lock, platform)
	r.Level = out.Level
	r.EvidenceFile = out.EvidenceFile
	for _, f := range out.Failures {
		r.Failures = append(r.Failures, failure{Code: f.Code, Message: f.Message, DetailsFile: f.DetailsFile})
	}
	r.OK = len(r.Failures) == 0
	r.Summary = fmt.Sprintf("Composition is %s on %s.", out.Level, platform)
	if !r.OK {
		r.Summary = fmt.Sprintf("Composition is %s on %s: %s.", out.Level, platform, out.Failures[0].Code)
		return exitFailed
	}
	return exitOK
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
		r.fail("LOCK_EDITED", err.Error())
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
