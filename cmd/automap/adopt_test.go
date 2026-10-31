package main_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/zach/automap/internal/clitest"
)

// acmeAdoptRepo is Acme's hand-built engine with the fixture map beside it.
func acmeAdoptRepo(t *testing.T) *clitest.Repo {
	t.Helper()
	repo := clitest.NewRepo(t, "acme-adopt")
	repo.FixtureMap(t, "map")
	return repo
}

// describeAcme adds the private map that describes Acme's in-house
// libraries, committed, as the lock pins it at its last commit.
func describeAcme(t *testing.T, repo *clitest.Repo) {
	t.Helper()
	repo.Copy(t, "testdata/private-maps/acme-adopt", "automap-private")
	commitRepo(t, repo)
}

func TestAdoptReportsTheTargetsNoMapDescribes(t *testing.T) {
	repo := acmeAdoptRepo(t)

	run := repo.Automap(t, "adopt", "--map", "map", "--no-verify", "--json")

	run.ExpectExit(t, 1)
	run.ExpectFailure(t, "UNMAPPED_TARGET(Acme::Jobs)", "private map")
	run.ExpectFailure(t, "UNMAPPED_TARGET(Acme::Renderer)", "private map")
	if len(run.Result.Failures) != 2 {
		t.Errorf("failures %v, want only the two in-house targets: the stub libraries are in the map, and Threads::Threads is CMake's own", run.Result.Failures)
	}
	run.ExpectUnchanged(t)
}

func TestAdoptWritesTheManifestOnceEveryTargetIsDescribed(t *testing.T) {
	repo := acmeAdoptRepo(t)
	describeAcme(t, repo)

	run := repo.Automap(t, "adopt", "--map", "map", "--no-verify", "--json")

	run.ExpectGoldenResult(t, "adopt-acme")
	var man struct {
		PrivateMaps []struct{ ID, Location string } `json:"private_maps"`
		Executable  struct {
			Target  string
			Sources []string
		}
		Slots map[string]string
		Glue  struct {
			Interfaces    string
			ChoicesHeader string `json:"choices_header"`
			Startup       struct {
				DeveloperOwned []string `json:"developer_owned"`
			}
			Bindings, Setup map[string]any
		}
	}
	if err := json.Unmarshal([]byte(clitest.Read(t, repo, "automap.manifest.json")), &man); err != nil {
		t.Fatal(err)
	}
	if len(man.PrivateMaps) != 1 || man.PrivateMaps[0].ID != "acme" || man.PrivateMaps[0].Location != "automap-private" {
		t.Errorf("private maps %v, want acme at automap-private", man.PrivateMaps)
	}
	if man.Executable.Target != "acme_game" {
		t.Errorf("executable %q, want acme_game", man.Executable.Target)
	}
	for slot, lib := range map[string]string{"windowing": "stub-window", "physics-3d": "stub-physics", "task-graph": "acme-jobs", "renderer": "acme-renderer"} {
		if man.Slots[slot] != lib {
			t.Errorf("slot %s holds %q, want %s", slot, man.Slots[slot], lib)
		}
	}
	if man.Glue.Interfaces != "engine/glue/automap" || man.Glue.ChoicesHeader != "engine/glue/automap/choices.h" {
		t.Errorf("interfaces %q and choices header %q, want them beside the engine's glue in engine/glue/automap", man.Glue.Interfaces, man.Glue.ChoicesHeader)
	}
	if strings.Join(man.Glue.Startup.DeveloperOwned, ",") != "engine/core/engine_main.cpp" {
		t.Errorf("developer-owned startup %v, want the engine's own engine/core/engine_main.cpp", man.Glue.Startup.DeveloperOwned)
	}
	// Every glue entry points at a file the engine already has.
	for _, entries := range []map[string]any{man.Glue.Bindings, man.Glue.Setup} {
		for key, e := range entries {
			files, ok := e.([]any)
			if !ok {
				continue
			}
			for _, f := range files {
				if _, existed := run.Before[f.(string)]; !existed {
					t.Errorf("glue %s points at %s, which the engine didn't have", key, f)
				}
			}
		}
	}
	for _, f := range []string{"engine/glue/automap/startup.cpp", "engine/glue/automap/startup.h"} {
		if _, err := os.Stat(filepath.Join(repo.Dir, filepath.FromSlash(f))); err == nil {
			t.Errorf("adopt generated %s; the engine keeps its own startup", f)
		}
	}
}

func TestAdoptNeverOverwritesExistingGlue(t *testing.T) {
	repo := acmeAdoptRepo(t)
	describeAcme(t, repo)
	// Acme's glue differs from the reference glue.
	glue := "engine/glue/stub-window/stub_window_allocation.cpp"
	repo.WriteFile(t, glue, clitest.Read(t, repo, glue)+"// Acme's change.\n")

	run := repo.Automap(t, "adopt", "--map", "map", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	after := map[string]bool{}
	for _, f := range run.Result.FilesWritten {
		after[f] = true
	}
	for p, before := range run.Before {
		if strings.HasPrefix(p, "engine/") {
			data, err := os.ReadFile(filepath.Join(repo.Dir, filepath.FromSlash(p)))
			if err != nil || !bytes.Equal(data, before) {
				t.Errorf("adopt changed %s", p)
			}
			if after[p] {
				t.Errorf("adopt lists %s as written", p)
			}
		}
	}
}

func TestAdoptRefusesWhereACompositionExists(t *testing.T) {
	repo := acmeAdoptRepo(t)
	describeAcme(t, repo)
	repo.Automap(t, "adopt", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)

	run := repo.Automap(t, "adopt", "--map", "map", "--no-verify", "--json")

	run.ExpectExit(t, 3)
	run.ExpectFailure(t, "COMPOSITION_EXISTS", "")
}

func TestAdoptNeedsACMakeProject(t *testing.T) {
	run := clitest.Automap(t, "no-manifest", "adopt", "--json")

	run.ExpectExit(t, 2)
	run.ExpectFailure(t, "NO_CMAKE_PROJECT", "automap start")
}

// The engine keeps its own startup; probes judge its glue where it is,
// and the lifecycle probe judges the engine's order.
func TestAnAdoptedEngineReachesProbed(t *testing.T) {
	if hostPlatform != "windows-x64-msvc" {
		requireTSan(t) // the jobs contract's ThreadSanitizer rule
	}
	repo := buildableAcme(t)

	run := repo.Automap(t, "adopt", "--map", "map", "--json")

	expectProbed(t, run)
}

func TestTheLifecycleProbeJudgesTheAdoptedEnginesOrder(t *testing.T) {
	repo := buildableAcme(t)
	// Physics comes up before the scheduler it runs on.
	main := "engine/core/engine_main.cpp"
	src := clitest.Read(t, repo, main)
	swapped := strings.Replace(src, "am_stub_window_setup_init(ctx) && am_jobs_setup_init(ctx) &&\n           am_bind_stub_physics_jobs(ctx) && am_stub_physics_setup_init(ctx);",
		"am_stub_window_setup_init(ctx) && am_stub_physics_setup_init(ctx) &&\n           am_jobs_setup_init(ctx) && am_bind_stub_physics_jobs(ctx);", 1)
	if swapped == src {
		t.Fatal("the engine's init order isn't where the test expects it")
	}
	repo.WriteFile(t, main, swapped)

	run := repo.Automap(t, "adopt", "--map", "map", "--json")

	run.ExpectExit(t, 1)
	for _, f := range run.Result.Failures {
		if f.Rule == "lifecycle.init-order" && strings.HasPrefix(f.Code, "INIT_ORDER(stub-physics:create_world before ") {
			return
		}
	}
	t.Errorf("no lifecycle.init-order failure for physics before the scheduler:\n%s", run.Stdout)
}

// buildableAcme is Acme's engine, described, with the fixture map pointing
// at the fixture registry.
func buildableAcme(t *testing.T) *clitest.Repo {
	t.Helper()
	requireBuilds(t)
	url, commit := fixtureRegistry(t)
	repo := acmeAdoptRepo(t)
	describeAcme(t, repo)
	repo.EditJSON(t, "map/map.json", set("build_backend/vcpkg/automap_registry", map[string]any{"url": url, "commit": commit}))
	return repo
}

// The custom-adopt eval's engine can be adopted: with the private map the
// custom-variant eval starts from, it reaches probed on the public map.
// Ejected then, as the custom-eject eval asks, it builds and passes its
// smoke test without Automap.
func TestTheCustomAdoptEvalsEngineCanBeAdoptedAndEjected(t *testing.T) {
	requirePublicMapBuilds(t)
	if hostPlatform != "windows-x64-msvc" {
		requireTSan(t)
	}
	repo := testbedRepo(t)
	repo.Copy(t, "evals/tasks/custom-adopt/repo", ".")
	repo.Copy(t, "evals/tasks/custom-variant/files/automap-private", "automap-private")
	commitRepo(t, repo)

	run := repo.Automap(t, "adopt", "--map", "map", "--json")

	if run.Exit != 0 || run.Result.Level != "probed" {
		logVcpkgErrors(t)
	}
	expectProbed(t, run)

	repo.Automap(t, "eject", "--map", "map", "--json").ExpectExit(t, 0)
	os.RemoveAll(filepath.Join(repo.Dir, "map"))
	var headless []string
	if runtime.GOOS == "linux" {
		headless = append(headless, "SDL_VIDEO_DRIVER=offscreen")
	}
	buildWithoutAutomap(t, repo, "acme_game", headless...)
}
