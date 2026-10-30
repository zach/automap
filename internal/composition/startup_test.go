package composition_test

import (
	"slices"
	"testing"

	"github.com/zach/automap/internal/composition"
	"github.com/zach/automap/internal/mapmodel"
)

func stubComposition(t *testing.T, m *mapmodel.Map) *composition.Manifest {
	t.Helper()
	man, fails := composition.NewManifest(m, composition.Request{
		With: []composition.LibraryRequest{{Library: "stub-window"}, {Library: "stub-ui"}},
	})
	if fails != nil {
		t.Fatal(fails)
	}
	return man
}

func callNames(calls []composition.Call) []string {
	var out []string
	for _, c := range calls {
		out = append(out, c.Call)
	}
	return out
}

func TestStartupCallsEntryPointsInPhaseOrder(t *testing.T) {
	m := fixtureMap(t)
	plan, fails := composition.PlanStartup(m, stubComposition(t, m))
	if fails != nil {
		t.Fatal(fails)
	}
	wantInit := []string{
		"am_bind_stub_ui_allocation", "am_bind_stub_window_allocation", // memory
		"am_bind_stub_ui_diagnostics", "am_bind_stub_window_diagnostics", // diagnostics
		"am_stub_window_setup_init",                                 // platform
		"am_stub_ui_setup_init", "am_pair_stub_ui_stub_window_init", // systems: the pairing follows stub-ui:create_context
	}
	if got := callNames(plan.Init); !slices.Equal(got, wantInit) {
		t.Errorf("init\n%v, want\n%v", got, wantInit)
	}
	wantShutdown := []string{
		"am_pair_stub_ui_stub_window_shutdown", "am_stub_ui_setup_shutdown", // systems, reversed
		"am_stub_window_setup_shutdown", // platform
	}
	if got := callNames(plan.Shutdown); !slices.Equal(got, wantShutdown) {
		t.Errorf("shutdown\n%v, want\n%v", got, wantShutdown)
	}
	if len(plan.Fields) != 1 || plan.Fields[0].Name != "stub-window.window" || plan.Fields[0].Type != "stub_window*" {
		t.Errorf("context fields %+v, want stub-window.window of type stub_window*", plan.Fields)
	}
}

func TestAfterListsOrderEntryPointsWithinAPhase(t *testing.T) {
	m := fixtureMap(t)
	// Recipe order puts stub-ui's setup before the pairing; an after list
	// naming the pairing's step must put it behind the pairing instead.
	m.Recipes["stub-ui--stub-window"].EntryPoints[0].After = nil
	m.Recipes["stub-ui"].EntryPoints[0].After = []string{"stub-ui--stub-window:route_input"}
	plan, fails := composition.PlanStartup(m, stubComposition(t, m))
	if fails != nil {
		t.Fatal(fails)
	}
	got := callNames(plan.Init)
	if slices.Index(got, "am_stub_ui_setup_init") < slices.Index(got, "am_pair_stub_ui_stub_window_init") {
		t.Errorf("init %v runs stub-ui's setup before the pairing it must follow", got)
	}
}

func TestAContextFieldNobodyProvidesFails(t *testing.T) {
	m := fixtureMap(t)
	m.Recipes["stub-window"].EntryPoints[0].Provides = nil
	_, fails := composition.PlanStartup(m, stubComposition(t, m))
	if got := codes(fails); !slices.Equal(got, []string{"CONTEXT_UNPROVIDED(stub-window.window)"}) {
		t.Errorf("failures %v, want CONTEXT_UNPROVIDED(stub-window.window)", got)
	}
}

func TestAContextFieldProvidedLaterFails(t *testing.T) {
	m := fixtureMap(t)
	m.Recipes["stub-ui--stub-window"].EntryPoints[0].Phase = "diagnostics"
	m.Recipes["stub-ui--stub-window"].EntryPoints[0].After = nil
	_, fails := composition.PlanStartup(m, stubComposition(t, m))
	if got := codes(fails); !slices.Contains(got, "CONTEXT_UNPROVIDED(stub-window.window)") {
		t.Errorf("failures %v, want CONTEXT_UNPROVIDED(stub-window.window)", got)
	}
}

func TestAnOrderingConstraintAgainstThePhasesFails(t *testing.T) {
	m := fixtureMap(t)
	pairing := m.Recipes["stub-ui--stub-window"]
	pairing.Ordering = append(pairing.Ordering, mapmodel.Ordering{Before: "stub-ui--stub-window:attach", After: "stub-window:open_window", When: "init"})
	_, fails := composition.PlanStartup(m, stubComposition(t, m))
	if got := codes(fails); !slices.Contains(got, "ORDER_UNSATISFIABLE(stub-ui--stub-window:attach before stub-window:open_window)") {
		t.Errorf("failures %v, want ORDER_UNSATISFIABLE", got)
	}
}
