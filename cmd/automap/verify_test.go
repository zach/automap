package main_test

import (
	"testing"

	"github.com/zach/automap/internal/clitest"
)

func TestVerifyWithNoManifestReportsNoManifest(t *testing.T) {
	run := clitest.Automap(t, "no-manifest", "verify", "--json")

	run.ExpectExit(t, 2)
	if run.Result.OK {
		t.Error("ok is true, want false")
	}
	if len(run.Result.Failures) != 1 || run.Result.Failures[0].Code != "NO_MANIFEST" {
		t.Errorf("failures %+v, want one NO_MANIFEST", run.Result.Failures)
	}
	run.ExpectGolden(t, "verify-no-manifest")
}
