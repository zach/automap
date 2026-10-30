package composition_test

import (
	"reflect"
	"testing"

	"github.com/zach/automap/internal/composition"
	"github.com/zach/automap/internal/mapmodel"
)

func fixtureMap(t *testing.T) *mapmodel.Map {
	t.Helper()
	m, problems := mapmodel.Load("../../testdata/map", nil)
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
	wantContracts := map[string]string{"allocation": "1.0.0", "build-flags": "1.0.0", "diagnostics": "1.0.0", "lifecycle": "1.0.0", "platform": "1.0.0"}
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
