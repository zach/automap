package main_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/zach/automap/internal/clitest"
	"github.com/zach/automap/internal/mapmodel"
)

// planLine is one port in vcpkg's install plan: its name and, in brackets,
// its features ("core" and those the plan turns on).
var planLine = regexp.MustCompile(`^\s*\*?\s*([a-z0-9-]+)(?:\[([^\]]*)\])?:[a-z0-9-]+@`)

// installPlan is the features vcpkg would build each port with for the
// composition in repo, from a dry run: nothing is downloaded or built.
func installPlan(t *testing.T, repo *clitest.Repo) map[string][]string {
	t.Helper()
	vcpkg := filepath.Join(os.Getenv("VCPKG_ROOT"), "vcpkg")
	if runtime.GOOS == "windows" {
		vcpkg += ".exe"
	}
	cmd := exec.Command(vcpkg, "install", "--dry-run", "--x-install-root="+filepath.Join(t.TempDir(), "installed"))
	cmd.Dir = repo.Dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("vcpkg install --dry-run: %v\n%s", err, tail(out))
	}
	plan := map[string][]string{}
	for _, line := range strings.Split(string(out), "\n") {
		if m := planLine.FindStringSubmatch(strings.TrimRight(line, "\r")); m != nil {
			features := []string{}
			for _, f := range strings.Split(m[2], ",") {
				if f = strings.TrimSpace(f); f != "" && f != "core" {
					features = append(features, f)
				}
			}
			sort.Strings(features)
			plan[m[1]] = features
		}
	}
	if len(plan) == 0 {
		t.Fatalf("vcpkg's dry run named no ports:\n%s", tail(out))
	}
	return plan
}

// A variant's description says what vcpkg builds, whatever else is in the
// composition: vcpkg turns on the union of the features every port asks
// of a port, so a library's port must get the same features in each
// reference composition as it gets alone. A feature another library's
// port switches on belongs in the variant, where its description states
// it.
func TestALibrarysPortBuildsTheSameFeaturesInEveryReferenceComposition(t *testing.T) {
	requirePublicMapBuilds(t)
	m, problems := mapmodel.Load(filepath.Join(clitest.Root(t), "map"), nil)
	if len(problems) > 0 {
		t.Fatal(problems)
	}
	var refs []struct {
		ID   string
		With []string
	}
	raw, err := os.ReadFile(filepath.Join(clitest.Root(t), "ci", "compositions.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &refs); err != nil {
		t.Fatal(err)
	}
	plan := func(with ...string) map[string][]string {
		repo := testbedRepo(t)
		args := []string{"start", "--map", "map", "--no-verify", "--json"}
		for _, w := range with {
			args = append(args, "--with", w)
		}
		repo.Automap(t, args...).ExpectExit(t, 0)
		return installPlan(t, repo)
	}
	alone := map[string][]string{}
	for _, ref := range refs {
		in := plan(ref.With...)
		for _, with := range ref.With {
			lib, _, _ := strings.Cut(with, ":")
			v := m.Libraries[lib].Newest()
			if v.BuildBackend == nil {
				continue
			}
			port := v.BuildBackend.Vcpkg.Port
			if _, ok := alone[with]; !ok {
				alone[with] = plan(with)[port]
			}
			if !slices.Equal(in[port], alone[with]) {
				t.Errorf("in %s, vcpkg builds %s's port %s with features %v; alone, with %v. Another port in the composition switches the difference on: state it in %s's variant.",
					ref.ID, with, port, in[port], alone[with], lib)
			}
		}
	}
}
