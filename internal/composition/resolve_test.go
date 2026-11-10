package composition_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/zach/automap/internal/composition"
	"github.com/zach/automap/internal/mapmodel"
)

// fixtureMap loads the fixture map: the stub libraries in testdata/map with
// the public map's contracts and probe harness.
func fixtureMap(t *testing.T) *mapmodel.Map {
	t.Helper()
	dir := t.TempDir()
	for src, dst := range map[string]string{"../../testdata/map": "", "../../map/contracts": "contracts", "../../map/probe-harness": "probe-harness"} {
		if err := os.CopyFS(filepath.Join(dir, dst), os.DirFS(src)); err != nil {
			t.Fatal(err)
		}
	}
	m, problems := mapmodel.Load(dir, nil)
	if len(problems) > 0 {
		t.Fatalf("fixture map: %v", problems)
	}
	return m
}

func codes(fails []composition.Failure) []string {
	var out []string
	for _, f := range fails {
		out = append(out, f.Code)
	}
	return out
}

func TestEachLibraryFillsEverySlotItCan(t *testing.T) {
	man, fails := composition.NewManifest(fixtureMap(t), composition.Request{
		With: []composition.LibraryRequest{{Library: "stub-window"}, {Library: "stub-ui"}},
	})
	if fails != nil {
		t.Fatal(fails)
	}
	want := map[string]string{"windowing": "stub-window", "input": "stub-window", "tools-ui": "stub-ui"}
	if !reflect.DeepEqual(man.Slots, want) {
		t.Errorf("slots %v, want %v", man.Slots, want)
	}
	if man.Variants["stub-window"].Default != "default" {
		t.Errorf("stub-window variant %q, want its first variant, default", man.Variants["stub-window"].Default)
	}
}

func TestSlotRequestsAssignOnlyTheSlotsNamed(t *testing.T) {
	man, fails := composition.NewManifest(fixtureMap(t), composition.Request{
		With:  []composition.LibraryRequest{{Library: "stub-window", Variant: "wide-events"}},
		Slots: []composition.SlotRequest{{Slot: "windowing", Library: "stub-window"}},
	})
	if fails != nil {
		t.Fatal(fails)
	}
	if want := map[string]string{"windowing": "stub-window"}; !reflect.DeepEqual(man.Slots, want) {
		t.Errorf("slots %v, want %v", man.Slots, want)
	}
	if man.Variants["stub-window"].Default != "wide-events" {
		t.Errorf("variant %q, want wide-events", man.Variants["stub-window"].Default)
	}
}

func TestSlotRequestFailures(t *testing.T) {
	m := fixtureMap(t)
	for _, c := range []struct {
		name string
		req  composition.Request
		want string
	}{
		{"slot taken", composition.Request{Slots: []composition.SlotRequest{{Slot: "windowing", Library: "stub-window"}, {Slot: "windowing", Library: "stub-window"}}}, "SLOT_TAKEN(windowing)"},
		{"library doesn't fill the slot", composition.Request{Slots: []composition.SlotRequest{{Slot: "tools-ui", Library: "stub-window"}}}, "SLOT_MISMATCH(tools-ui=stub-window)"},
		{"unknown library", composition.Request{With: []composition.LibraryRequest{{Library: "jolt"}}}, "UNKNOWN_ENTRY(library jolt)"},
		{"unknown variant", composition.Request{With: []composition.LibraryRequest{{Library: "stub-ui", Variant: "docking"}}}, "UNKNOWN_VARIANT(stub-ui:docking)"},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, fails := composition.NewManifest(m, c.req)
			if got := codes(fails); len(got) != 1 || got[0] != c.want {
				t.Errorf("failures %v, want [%s]", got, c.want)
			}
		})
	}
}

func TestResolvePinsTheNewestVersionFromTheMap(t *testing.T) {
	m := fixtureMap(t)
	man, fails := composition.NewManifest(m, composition.Request{
		With:       []composition.LibraryRequest{{Library: "stub-window"}, {Library: "stub-ui"}},
		MapRelease: "unreleased",
	})
	if fails != nil {
		t.Fatal(fails)
	}
	lock, fails := composition.Resolve(m, man)
	if fails != nil {
		t.Fatal(fails)
	}
	want := composition.Pin{
		Version:     "1.0.0",
		Commit:      "1111111111111111111111111111111111111111",
		Variant:     "default",
		Registry:    &composition.Registry{Kind: "automap", Commit: "3333333333333333333333333333333333333333"},
		ContentHash: "sha512:00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001",
		License:     "MIT-0",
	}
	if !reflect.DeepEqual(lock.Pins["stub-window"], want) {
		t.Errorf("stub-window pin\n%+v, want\n%+v", lock.Pins["stub-window"], want)
	}
	wantContracts := map[string]string{"allocation": "1.0.0", "build-flags": "1.0.0", "diagnostics": "1.0.0", "lifecycle": "1.0.0", "platform": "1.1.0"}
	if !reflect.DeepEqual(lock.Contracts, wantContracts) {
		t.Errorf("contracts %v, want %v", lock.Contracts, wantContracts)
	}
	if lock.MapRelease != "unreleased" {
		t.Errorf("map_release %q, want unreleased", lock.MapRelease)
	}
}

func TestPerPlatformVariantsGetTheirOwnPin(t *testing.T) {
	m := fixtureMap(t)
	man, _ := composition.NewManifest(m, composition.Request{With: []composition.LibraryRequest{{Library: "stub-window"}}})
	man.Variants["stub-window"] = composition.VariantChoice{Default: "default", PerPlatform: map[string]string{"linux-x64-clang": "wide-events"}}
	lock, fails := composition.Resolve(m, man)
	if fails != nil {
		t.Fatal(fails)
	}
	pin := lock.Pins["stub-window"]
	if len(pin.Platforms) != 1 || pin.Platforms["linux-x64-clang"].Variant != "wide-events" {
		t.Errorf("platform pins %+v, want only linux-x64-clang with wide-events", pin.Platforms)
	}
}

// A port two libraries share keeps its default features unless both turn
// them off, and takes the union of their features.
func TestASharedPortMergesWhatItsLibrariesAsk(t *testing.T) {
	for _, c := range []struct {
		name          string
		first, second map[string]string
		want          string
	}{
		{"both off", map[string]string{"vcpkg_default_features": "false", "vcpkg_features": "a"}, map[string]string{"vcpkg_default_features": "false", "vcpkg_features": "b,a"},
			`{"name":"stub-ui","default-features":false,"features":["a","b"]}`},
		{"one on", map[string]string{"vcpkg_default_features": "false", "vcpkg_features": "x11[linux]"}, map[string]string{},
			`{"name":"stub-ui","features":[{"name":"x11","platform":"linux"}]}`},
	} {
		t.Run(c.name, func(t *testing.T) {
			m := fixtureMap(t)
			twin := *m.Libraries["stub-ui"]
			twin.ID = "stub-ui-twin"
			twin.Versions = append([]mapmodel.Version{}, twin.Versions...)
			twin.Versions[0].Variants = []mapmodel.Variant{{Name: "default", BuildOptions: c.second}}
			m.Libraries["stub-ui-twin"] = &twin
			m.Libraries["stub-ui"].Versions[0].Variants = []mapmodel.Variant{{Name: "default", BuildOptions: c.first}}
			man := &composition.Manifest{Slots: map[string]string{"tools-ui": "stub-ui", "input": "stub-ui-twin"}}
			lock := &composition.Lock{Pins: map[string]composition.Pin{"stub-ui": {Variant: "default"}, "stub-ui-twin": {Variant: "default"}}}

			var doc struct {
				Dependencies []json.RawMessage `json:"dependencies"`
			}
			json.Unmarshal(composition.VcpkgJSON(m, man, lock), &doc)

			if len(doc.Dependencies) != 1 {
				t.Fatalf("dependencies %s, want the shared port once", doc.Dependencies)
			}
			var compact bytes.Buffer
			json.Compact(&compact, doc.Dependencies[0])
			if compact.String() != c.want {
				t.Errorf("dependency %s, want %s", compact.String(), c.want)
			}
		})
	}
}
