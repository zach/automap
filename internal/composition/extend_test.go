package composition_test

import (
	"reflect"
	"slices"
	"testing"

	"github.com/zach/automap/internal/composition"
)

func windowOnly(t *testing.T) *composition.Manifest {
	t.Helper()
	man, fails := composition.NewManifest(fixtureMap(t), composition.Request{
		With: []composition.LibraryRequest{{Library: "stub-window"}}, MapRelease: "unreleased",
	})
	if fails != nil {
		t.Fatal(fails)
	}
	return man
}

func TestExtendAddsTheLibraryAndThePairingsItCompletes(t *testing.T) {
	m := fixtureMap(t)
	base := windowOnly(t)
	base.Glue.Setup["stub-window"] = composition.GlueEntry{Files: []string{"engine/window_setup.h", "engine/window_setup.cpp"}}

	man, fails := composition.Extend(m, base, composition.ExtendRequest{Library: "stub-ui", MapRelease: "unreleased"})

	if fails != nil {
		t.Fatal(fails)
	}
	if man.Slots["tools-ui"] != "stub-ui" || man.Slots["windowing"] != "stub-window" {
		t.Errorf("slots %v, want stub-ui added to tools-ui beside stub-window", man.Slots)
	}
	if man.Variants["stub-ui"].Default != "default" {
		t.Errorf("stub-ui variant %q, want its first variant", man.Variants["stub-ui"].Default)
	}
	if _, ok := man.Glue.Pairings["stub-ui/stub-window"]; !ok {
		t.Error("no glue entry for the stub-ui--stub-window pairing")
	}
	if got := man.Glue.Setup["stub-window"].Files; !slices.Equal(got, []string{"engine/window_setup.h", "engine/window_setup.cpp"}) {
		t.Errorf("stub-window's setup glue moved to %v; the developer's entry must stay", got)
	}
	if !reflect.DeepEqual(base.Slots, map[string]string{"windowing": "stub-window", "input": "stub-window"}) {
		t.Errorf("Extend changed the base manifest's slots: %v", base.Slots)
	}
}

func TestExtendFailures(t *testing.T) {
	m := fixtureMap(t)
	for _, c := range []struct {
		name string
		req  composition.ExtendRequest
		want string
	}{
		{"already present", composition.ExtendRequest{Library: "stub-window"}, "ALREADY_PRESENT(stub-window)"},
		{"unknown library", composition.ExtendRequest{Library: "jolt"}, "UNKNOWN_ENTRY(library jolt)"},
		{"unknown variant", composition.ExtendRequest{Library: "stub-ui", Variant: "docking"}, "UNKNOWN_VARIANT(stub-ui:docking)"},
		{"slot it doesn't fill", composition.ExtendRequest{Library: "stub-ui", Slots: []string{"windowing"}}, "SLOT_MISMATCH(windowing=stub-ui)"},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, fails := composition.Extend(m, windowOnly(t), c.req)
			if got := codes(fails); len(got) != 1 || got[0] != c.want {
				t.Errorf("failures %v, want [%s]", got, c.want)
			}
		})
	}
}

func TestExtendRejectsATakenSlot(t *testing.T) {
	m := fixtureMap(t)
	m.Libraries["stub-ui"].Slots = append(m.Libraries["stub-ui"].Slots, "windowing")

	_, fails := composition.Extend(m, windowOnly(t), composition.ExtendRequest{Library: "stub-ui"})

	if got := codes(fails); len(got) != 1 || got[0] != "SLOT_TAKEN(windowing)" {
		t.Errorf("failures %v, want [SLOT_TAKEN(windowing)]", got)
	}
}
