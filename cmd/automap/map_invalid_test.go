package main_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/zach/automap/internal/clitest"
)

// at walks a slash-separated path of keys and indexes to the parent of its
// last element, and returns the parent and that last element.
func at(doc any, path string) (any, string) {
	parts := strings.Split(path, "/")
	for _, p := range parts[:len(parts)-1] {
		switch d := doc.(type) {
		case map[string]any:
			doc = d[p]
		case []any:
			i, _ := strconv.Atoi(p)
			doc = d[i]
		}
	}
	return doc, parts[len(parts)-1]
}

// set returns an edit that sets the value at path.
func set(path string, value any) func(any) any {
	return func(doc any) any {
		parent, key := at(doc, path)
		switch p := parent.(type) {
		case map[string]any:
			p[key] = value
		case []any:
			i, _ := strconv.Atoi(key)
			p[i] = value
		}
		return doc
	}
}

// push returns an edit that appends value to the array at path.
func push(path string, value any) func(any) any {
	return func(doc any) any {
		parent, key := at(doc, path)
		m := parent.(map[string]any)
		a, _ := m[key].([]any)
		m[key] = append(a, value)
		return doc
	}
}

// remove returns an edit that deletes the key or array element at path.
func remove(path string) func(any) any {
	return func(doc any) any {
		parent, key := at(doc, path)
		switch p := parent.(type) {
		case map[string]any:
			delete(p, key)
		case []any:
			i, _ := strconv.Atoi(key)
			grand, gkey := at(doc, path[:strings.LastIndex(path, "/")])
			grand.(map[string]any)[gkey] = append(p[:i:i], p[i+1:]...)
		}
		return doc
	}
}

type mapCase struct {
	name    string
	edit    func(t *testing.T, repo *clitest.Repo)
	file    string // the file named in MAP_INVALID(<file>)
	message string
}

func editJSON(path string, edit func(any) any) func(*testing.T, *clitest.Repo) {
	return func(t *testing.T, repo *clitest.Repo) { repo.EditJSON(t, path, edit) }
}

func writeFile(path, data string) func(*testing.T, *clitest.Repo) {
	return func(t *testing.T, repo *clitest.Repo) { repo.WriteFile(t, path, data) }
}

func runMapCases(t *testing.T, newRepo func(*testing.T) *clitest.Repo, cases []mapCase) {
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := newRepo(t)
			c.edit(t, repo)
			run := repo.Automap(t, "find", "--map", "map", "--json")

			run.ExpectExit(t, 2)
			run.ExpectFailure(t, "MAP_INVALID("+c.file+")", c.message)
		})
	}
}

func TestInvalidFixtureMapFiles(t *testing.T) {
	const (
		stubUI     = "map/libraries/stub-ui.json"
		stubWindow = "map/libraries/stub-window.json"
		alloc      = "map/contracts/allocation/1.0.0.json"
		uiAlloc    = "map/recipes/binding/stub-ui--allocation.json"
		uiDiag     = "map/recipes/binding/stub-ui--diagnostics.json"
		uiLib      = "map/recipes/library/stub-ui.json"
		pairing    = "map/recipes/pairing/stub-ui--stub-window.json"
	)
	runMapCases(t, fixtureRepo, []mapCase{
		{"not JSON", writeFile("map/slots/input.json", "{"), "map/slots/input.json", "not valid JSON"},
		{"schema", editJSON(stubUI, remove("name")), stubUI, "name"},
		{"public map kind", editJSON("map/map.json", set("kind", "private")), "map/map.json", "only the public map has kind public"},
		{"slot file name", editJSON("map/slots/input.json", set("id", "inputs")), "map/slots/input.json", "file name must be the slot id"},
		{"contract path", editJSON(alloc, set("version", "1.0.1")), alloc, "path must be contracts/<contract>/<version>.json"},
		{"choice default", editJSON(alloc, set("choices/allocator/default", "heap")), alloc, "choices/allocator/default"},
		{"lifecycle phases", editJSON("map/contracts/lifecycle/1.0.0.json", remove("phases")), "map/contracts/lifecycle/1.0.0.json", "the lifecycle contract lists its phases"},
		{"rule prefix", editJSON(alloc, set("rules/2/id", "memory.same-allocator")), alloc, "must be prefixed with the contract id"},
		{"library file name", editJSON(stubUI, set("id", "stub-uis")), stubUI, "file name must be the library id"},
		{"unknown slot", editJSON(stubUI, push("slots", "audio")), stubUI, "slot audio"},
		{"facts schema", editJSON(stubUI, set("versions/0/facts/allocation/hook_style", "magic")), stubUI, "1.0.0/facts/allocation"},
		{"step prefix", editJSON(stubUI, set("versions/0/facts/lifecycle/init_steps/0/step", "ui:set_allocator")), stubUI, "must be prefixed with 'stub-ui:'"},
		{"unknown after", editJSON(stubWindow, set("versions/0/facts/lifecycle/init_steps/3/after", []any{"stub-window:boot"})), stubWindow, "follows unknown step stub-window:boot"},
		{"variant facts", editJSON(stubWindow, set("versions/0/variants/1/facts/build-flags/abi_defines", "X")), stubWindow, "1.0.0/wide-events/facts/build-flags"},
		{"console variants on a path library", editJSON(stubUI, func(doc any) any {
			d := doc.(map[string]any)
			d["source"] = map[string]any{"kind": "path", "path": "engine/ui"}
			d["console_variants"] = []any{map[string]any{"platform_family": "Nebula", "distributed_by": "Nebula"}}
			return d
		}), stubUI, "console_variants belong to the public map only"},
		{"recipe directory", func(t *testing.T, repo *clitest.Repo) {
			repo.Copy(t, "testdata/map/recipes/library/stub-ui.json", "map/recipes/binding/stub-ui.json")
		}, "map/recipes/binding/stub-ui.json", "a library recipe lives in recipes/library/"},
		{"recipe file name", editJSON(uiDiag, set("subject/contract", "allocation")), uiDiag, "file name must be stub-ui--allocation"},
		{"contract version", editJSON(uiAlloc, set("subject/contract_version", "1.1.0")), uiAlloc, "contract_version 1.1.0"},
		{"pairing order", editJSON(pairing, set("subject/libraries", []any{"stub-window", "stub-ui"})), pairing, "pairing libraries are listed in sorted order"},
		{"pairing rule prefix", editJSON(pairing, set("rules/0/id", "pair.input-forwarded")), pairing, "must be prefixed with 'stub-ui--stub-window.'"},
		{"applies_to", editJSON(uiAlloc, set("applies_to/stub-ui/versions", []any{"9.9.9"})), uiAlloc, "applies_to names stub-ui 9.9.9 (default)"},
		{"probe for unknown rule", editJSON(uiAlloc, set("probes/0/rule", "allocation.hooks-early")), uiAlloc, "probe for unknown rule allocation.hooks-early"},
		{"probe for guidance rule", editJSON(uiAlloc, set("probes/0/rule", "allocation.same-allocator-across-modules")), uiAlloc, "probe for guidance rule"},
		{"not_applicable", editJSON(uiAlloc, set("not_applicable", []any{map[string]any{"rule": "allocation.nothing", "reason": "x"}})), uiAlloc, "not_applicable names unknown rule allocation.nothing"},
		{"build fragment", editJSON(uiLib, remove("reference_glue/0")), uiLib, "a library recipe ships a build fragment"},
		{"probe glue", editJSON(uiDiag, remove("reference_glue/1")), uiDiag, "binding and pairing recipes ship their probes"},
		{"ordering owner", editJSON(pairing, set("ordering/0/before", "stub-audio:play")), pairing, "ordering step stub-audio:play"},
		{"undeclared step", editJSON(uiAlloc, set("entry_points/0/steps", []any{"stub-ui:boot"})), uiAlloc, "which stub-ui's lifecycle facts don't declare"},
		{"step owner", editJSON(uiAlloc, set("entry_points/0/steps", []any{"stub-audio:boot"})), uiAlloc, "owned by neither a library nor this recipe"},
		{"unseeded composition rule", editJSON(pairing, remove("probes/1")), "map/contracts/lifecycle", "lifecycle.init-order has no seeded violation"},
		{"JSON Patch seed", editJSON(alloc, set("rules/0/seeded_violation", map[string]any{"description": "x", "patch": "1.0.0/seeded/x.json-patch"})), "map/contracts/allocation", "only manifest seeds are JSON Patches"},
		{"duplicate recipe", func(t *testing.T, repo *clitest.Repo) {
			repo.Copy(t, "testdata/map/recipes/binding/stub-ui--allocation.json", "map/recipes/binding/stub-ui--allocation-copy.json")
		}, "map/recipes/binding/stub-ui--allocation-copy.json", "file name must be stub-ui--allocation"},
	})
}

// examplesRepo is a repo with only the spec's public example map, which
// carries published evidence.
func examplesRepo(t *testing.T) *clitest.Repo {
	t.Helper()
	repo := clitest.NewRepo(t, "no-manifest")
	repo.Copy(t, "docs/spec/examples/map", "map")
	return repo
}

func TestInvalidMapEvidence(t *testing.T) {
	const ev = "map/evidence/binding/jolt--jobs/jolt-5.6.0-default/linux-x64-clang.json"
	runMapCases(t, examplesRepo, []mapCase{
		{"schema", editJSON(ev, remove("predicate/level")), ev, "level"},
		{"verifier", editJSON(ev, set("predicate/verifier/kind", "nda-holder")), ev, "the public map's evidence comes from Automap's CI"},
		{"unknown pin", editJSON(ev, set("predicate/subject_detail/pin/version", "5.7.0")), ev, "no library has jolt 5.7.0 (default)"},
		{"rule outside the contract", editJSON(ev, set("predicate/results/0/rule", "allocation.routed-through-facade")), ev, "binding evidence for jobs reports rule allocation.routed-through-facade"},
		{"unreleased and signed", editJSON(ev, set("predicate/map_release", "unreleased")), ev, "evidence against an unreleased map is never signed"},
		{"known-good with a failure", editJSON(ev, set("predicate/results/0", map[string]any{"rule": "jobs.routed-through-interface", "library": "jolt", "status": "fail", "failure_code": "JOBS_NOT_ROUTED(jolt)"})), ev, "known-good level with failing or unverifiable results"},
		{"platform-limited rule", editJSON(ev, set("predicate/platform", "windows-x64-msvc")), ev, "jobs.tsan-clean runs only on"},
	})
}

func TestInvalidPrivateMaps(t *testing.T) {
	const (
		acmeSDL   = "private-map/libraries/sdl3.json"
		nebulaSDL = "private-map-nebula/libraries/sdl3.json"
	)
	runMapCases(t, acmeRepo, []mapCase{
		{"private map kind", editJSON("private-map/map.json", func(doc any) any {
			d := doc.(map[string]any)
			d["kind"] = "public"
			d["build_backend"] = map[string]any{"vcpkg": map[string]any{"baseline": strings.Repeat("0", 40), "tool": "2026-09-26"}}
			return d
		}), "private-map/map.json", "only the public map has kind public"},
		{"extension changes a library field", editJSON(acmeSDL, set("license/spdx", "MIT")), acmeSDL, "extends sdl3 but its license differs"},
		{"extension changes a version field", editJSON(acmeSDL, set("versions/0/released", "2026-10-03")), acmeSDL, "extends sdl3 3.4.18 but its released differs"},
		{"extension changes facts", editJSON(acmeSDL, set("versions/0/facts/diagnostics/query", "none")), acmeSDL, "facts may only gain platform targets"},
		{"non-NDA map adds a target", editJSON(acmeSDL, push("versions/0/facts/platform/targets", "linux-arm64-gcc")), acmeSDL, "only an NDA platform's map may add targets"},
		{"NDA map adds another platform", editJSON(nebulaSDL, push("versions/0/facts/platform/targets", "linux-arm64-gcc")), nebulaSDL, "only an NDA platform's map may add targets"},
		{"extension redefines a variant", editJSON(nebulaSDL, set("versions/0/variants/0/name", "acme-patched")), nebulaSDL, "variant acme-patched differs"},
		{"duplicate slot", func(t *testing.T, repo *clitest.Repo) {
			repo.Copy(t, "docs/spec/examples/map/slots/gpu.json", "private-map/slots/gpu.json")
		}, "private-map/slots/gpu.json", "slot gpu is already defined"},
	})
}

func TestManifestPrivateMapMismatch(t *testing.T) {
	for _, c := range []struct {
		name, message string
		edit          func(any) any
	}{
		{"id", "private map acme-games is at private-map, whose map.json has id acme", set("private_maps/0/id", "acme-games")},
		{"platforms", "private map acme-nebula: platforms differ from its map.json", set("private_maps/1/platforms", []any{"nebula-x64-clang"})},
	} {
		t.Run(c.name, func(t *testing.T) {
			repo := acmeRepo(t)
			repo.EditJSON(t, "automap.manifest.json", c.edit)
			run := repo.Automap(t, "find", "--map", "map", "--json")

			run.ExpectExit(t, 2)
			run.ExpectFailure(t, "MANIFEST_INVALID", c.message)
		})
	}
}
