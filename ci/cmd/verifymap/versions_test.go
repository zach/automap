package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/zach/automap/internal/clitest"
	"github.com/zach/automap/internal/mapmodel"
)

// A library with two versions is verified at its older one through a copy
// of the map holding only that version. The copy must stay a valid map:
// recipes name only the versions kept, and a recipe that applies to none
// of a picked library's kept versions is left out.
func TestCopyingTheMapAtAnOlderVersionKeepsItValid(t *testing.T) {
	repo := clitest.NewRepo(t, "no-manifest")
	repo.FixtureMap(t, "map")
	src := filepath.Join(repo.Dir, "map")

	// stub-ui gets a newer version 1.1.0, which every stub-ui recipe
	// applies to; one recipe applies to 1.1.0 alone.
	edit(t, filepath.Join(src, "libraries", "stub-ui.json"), func(doc map[string]any) {
		versions := doc["versions"].([]any)
		raw, _ := json.Marshal(versions[0])
		var newer map[string]any
		json.Unmarshal(raw, &newer)
		newer["version"] = "1.1.0"
		doc["versions"] = append([]any{newer}, versions...)
	})
	recipes, _ := filepath.Glob(filepath.Join(src, "recipes", "*", "*.json"))
	var newerOnly string
	for _, file := range recipes {
		edit(t, file, func(doc map[string]any) {
			applies := doc["applies_to"].(map[string]any)
			a, ok := applies["stub-ui"].(map[string]any)
			if !ok {
				return
			}
			if newerOnly == "" && filepath.Base(filepath.Dir(file)) == "binding" {
				newerOnly = file
				a["versions"] = []any{"1.1.0"}
				return
			}
			a["versions"] = append(a["versions"].([]any), "1.1.0")
		})
	}
	if newerOnly == "" {
		t.Fatal("no binding recipe applies to stub-ui")
	}
	// Binding evidence pins stub-ui at each version.
	evidenceAt := map[string]string{}
	for _, version := range []string{"1.1.0", "1.0.0"} {
		evidenceAt[version] = writeBindingEvidence(t, src, "stub-ui", version)
	}
	if _, problems := mapmodel.Load(src, nil); len(problems) > 0 {
		t.Fatalf("the source map: %v", problems)
	}

	dst := filepath.Join(t.TempDir(), "map")
	if err := copyMapWithVersions(src, dst, []Pick{{Library: "stub-ui", Version: "1.0.0", Variant: "default"}}); err != nil {
		t.Fatal(err)
	}
	m, problems := mapmodel.Load(dst, nil)
	if len(problems) > 0 {
		t.Fatalf("the copy: %v", problems)
	}
	if got := m.Libraries["stub-ui"].Newest().Version; got != "1.0.0" {
		t.Errorf("stub-ui's version in the copy = %s, want 1.0.0", got)
	}
	rel, _ := filepath.Rel(src, newerOnly)
	if _, err := os.Stat(filepath.Join(dst, rel)); !os.IsNotExist(err) {
		t.Errorf("%s applies to stub-ui 1.1.0 alone and should be left out of the copy", rel)
	}
	if _, err := os.Stat(filepath.Join(dst, evidenceAt["1.1.0"])); !os.IsNotExist(err) {
		t.Errorf("%s pins stub-ui 1.1.0 and should be left out of the copy", evidenceAt["1.1.0"])
	}
	if _, err := os.Stat(filepath.Join(dst, evidenceAt["1.0.0"])); err != nil {
		t.Errorf("%s pins stub-ui 1.0.0 and should stay in the copy: %v", evidenceAt["1.0.0"], err)
	}
	for _, file := range recipes {
		rel, _ := filepath.Rel(src, file)
		raw, err := os.ReadFile(filepath.Join(dst, rel))
		if err != nil {
			continue
		}
		var doc struct {
			AppliesTo map[string]struct{ Versions []string } `json:"applies_to"`
		}
		json.Unmarshal(raw, &doc)
		if a, ok := doc.AppliesTo["stub-ui"]; ok && (len(a.Versions) != 1 || a.Versions[0] != "1.0.0") {
			t.Errorf("%s applies to stub-ui %v in the copy, want [1.0.0]", rel, a.Versions)
		}
	}
}

// writeBindingEvidence writes allocation binding evidence pinning a
// library's version into the map, as Verify map would, and returns its
// path in the map.
func writeBindingEvidence(t *testing.T, mapDir, library, version string) string {
	t.Helper()
	rel := filepath.Join("evidence", "binding", library+"--allocation", library+"-"+version+"-default", "linux-x64-clang.json")
	doc := map[string]any{
		"_type":         "https://in-toto.io/Statement/v1",
		"subject":       []any{map[string]any{"name": "pkg:generic/" + library + "@" + version, "digest": map[string]any{"sha512": "00"}}},
		"predicateType": "https://github.com/zach/automap/evidence/v1",
		"predicate": map[string]any{
			"subject_kind": "binding",
			"subject_detail": map[string]any{
				"contract":         "allocation",
				"contract_version": "1.0.0",
				"pin":              map[string]any{"library": library, "version": version, "variant": "default", "content_hash": "sha512:00"},
			},
			"platform":    "linux-x64-clang",
			"level":       "probed",
			"produced_at": "2026-11-10T00:00:00Z",
			"map_release": "unreleased",
			"conditions":  map[string]any{"build_type": "Debug", "compiler": "clang", "environment": map[string]any{}, "runner_image": "test", "sanitizers": map[string]any{}},
			"results":     []any{map[string]any{"rule": "allocation.hooks-before-first-allocation", "library": library, "status": "pass"}},
			"verifier":    map[string]any{"kind": "automap-ci", "id": "test", "signed": false},
		},
	}
	raw, _ := json.MarshalIndent(doc, "", "  ")
	full := filepath.Join(mapDir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	return rel
}

func edit(t *testing.T, file string, f func(doc map[string]any)) {
	t.Helper()
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	f(doc)
	raw, _ = json.MarshalIndent(doc, "", "  ")
	if err := os.WriteFile(file, raw, 0o644); err != nil {
		t.Fatal(err)
	}
}
