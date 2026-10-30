package composition

import (
	"fmt"
	"slices"

	"github.com/zach/automap/internal/mapmodel"
)

// Phases is the lifecycle contract's fixed phase order (ADR-0026).
var Phases = []string{"memory", "diagnostics", "platform", "jobs", "systems", "game"}

// Call is one entry point the startup calls.
type Call struct {
	Recipe string
	mapmodel.EntryPoint
}

// interfaceSetup is the setup each shared interface with running state
// needs: entry points the startup calls like a recipe's, owned by automap
// (SPEC §4.1).
var interfaceSetup = map[string][]mapmodel.EntryPoint{
	"jobs": {
		{Phase: "jobs", When: "init", Call: "am_jobs_setup_init", Steps: []string{"automap:jobs.start"}},
		{Phase: "jobs", When: "shutdown", Call: "am_jobs_setup_shutdown", Steps: []string{"automap:jobs.stop"}},
	},
}

// InterfaceCalls lists the entry points of the composition's shared
// interfaces' setup.
func InterfaceCalls(m *mapmodel.Map, man *Manifest) []Call {
	var out []Call
	for _, cid := range Contracts(m, man) {
		c := m.LatestContract(cid)
		if c.SharedInterface == nil {
			continue
		}
		for _, ep := range interfaceSetup[cid] {
			out = append(out, Call{Recipe: "automap", EntryPoint: ep})
		}
	}
	return out
}

// StartupPlan is the order the startup calls entry points in, and the
// context fields they share (ADR-0080).
type StartupPlan struct {
	Init     []Call
	Shutdown []Call
	Fields   []mapmodel.ContextField // in the order init provides them
}

// PlanStartup orders the composition's entry points: init in phase order,
// shutdown in reverse; within a phase, each entry point after every entry
// point performing a step it must follow, by its own after list, its
// library's lifecycle facts, or a pairing's ordering constraint. It checks
// that every context field an entry point reads is provided earlier. It is
// a pure function of the map and the manifest.
func PlanStartup(m *mapmodel.Map, man *Manifest) (*StartupPlan, []Failure) {
	var fails []Failure
	calls := InterfaceCalls(m, man)
	for _, rid := range Recipes(m, man) {
		for _, ep := range m.Recipes[rid].EntryPoints {
			calls = append(calls, Call{Recipe: rid, EntryPoint: ep})
		}
	}
	var init, shutdown []Call
	for _, c := range calls {
		if c.When == "init" {
			init = append(init, c)
		} else {
			shutdown = append(shutdown, c)
		}
	}

	plan := &StartupPlan{}
	var f []Failure
	plan.Init, f = order(m, man, init, "init", nil)
	fails = append(fails, f...)
	initPos := map[string]int{}
	for i, c := range plan.Init {
		if _, ok := initPos[c.Recipe]; !ok {
			initPos[c.Recipe] = i
		}
	}
	plan.Shutdown, f = order(m, man, shutdown, "shutdown", initPos)
	fails = append(fails, f...)

	provided := map[string]bool{}
	for _, c := range plan.Init {
		for _, field := range c.Reads {
			if !provided[field] {
				fails = append(fails, failure("CONTEXT_UNPROVIDED("+field+")", "%s reads %s, which %s.", c.Call, field, providedLater(plan.Init, field)))
			}
		}
		for _, field := range c.Provides {
			if provided[field.Name] {
				fails = append(fails, failure("CONTEXT_PROVIDED_TWICE("+field.Name+")", "%s provides %s, which an earlier entry point already provides.", c.Call, field.Name))
				continue
			}
			provided[field.Name] = true
			plan.Fields = append(plan.Fields, field)
		}
	}
	for _, c := range plan.Shutdown {
		for _, field := range c.Reads {
			if !provided[field] {
				fails = append(fails, failure("CONTEXT_UNPROVIDED("+field+")", "%s reads %s, which no init entry point in the composition provides.", c.Call, field))
			}
		}
	}
	if len(fails) > 0 {
		return nil, fails
	}
	return plan, nil
}

func providedLater(init []Call, field string) string {
	for _, c := range init {
		for _, p := range c.Provides {
			if p.Name == field {
				return "only " + c.Call + " provides, and it runs later"
			}
		}
	}
	return "no entry point in the composition provides"
}

// order sorts one kind of entry point (init or shutdown). Phases run in
// order for init and in reverse for shutdown. Within a phase, entry points
// keep their recipes' order for init, and the reverse of their init
// position for shutdown, except where a dependency says otherwise.
func order(m *mapmodel.Map, man *Manifest, calls []Call, when string, initPos map[string]int) ([]Call, []Failure) {
	var fails []Failure
	phase := func(c Call) int {
		i := slices.Index(Phases, c.Phase)
		if when == "shutdown" {
			return len(Phases) - i
		}
		return i
	}
	performer := map[string]int{}
	for i, c := range calls {
		for _, s := range c.Steps {
			performer[s] = i
		}
	}

	// deps[i] holds the calls that must run before calls[i].
	deps := make([]map[int]bool, len(calls))
	for i := range deps {
		deps[i] = map[int]bool{}
	}
	addDep := func(before, after string) {
		b, okB := performer[before]
		a, okA := performer[after]
		if !okB || !okA || a == b {
			return
		}
		switch {
		case phase(calls[b]) > phase(calls[a]):
			fails = append(fails, failure("ORDER_UNSATISFIABLE("+before+" before "+after+")",
				"%s must run before %s at %s, but %s's phase comes later.", before, after, when, calls[b].Call))
		case phase(calls[b]) == phase(calls[a]):
			deps[a][b] = true
		}
	}
	for _, c := range calls {
		for _, s := range c.Steps {
			for _, a := range c.After {
				addDep(a, s)
			}
		}
	}
	kind := "init_steps"
	if when == "shutdown" {
		kind = "shutdown_steps"
	}
	for _, lib := range man.Libraries() {
		if l := m.Libraries[lib]; l != nil {
			for _, s := range lifecycleSteps(l.Newest(), kind) {
				for _, a := range s.After {
					addDep(a, s.Step)
				}
			}
		}
	}
	for _, rid := range Recipes(m, man) {
		for _, o := range m.Recipes[rid].Ordering {
			if o.When == when {
				addDep(o.Before, o.After)
			}
		}
	}

	rank := func(i int) int {
		if initPos == nil {
			return i
		}
		if p, ok := initPos[calls[i].Recipe]; ok {
			return -p
		}
		return len(calls) + i
	}
	idx := make([]int, len(calls))
	for i := range idx {
		idx[i] = i
	}
	slices.SortStableFunc(idx, func(a, b int) int {
		if pa, pb := phase(calls[a]), phase(calls[b]); pa != pb {
			return pa - pb
		}
		return rank(a) - rank(b)
	})

	var out []Call
	done := map[int]bool{}
	for len(out) < len(calls) {
		progressed := false
		for _, i := range idx {
			if done[i] {
				continue
			}
			ready := true
			for d := range deps[i] {
				ready = ready && done[d]
			}
			if ready {
				done[i] = true
				out = append(out, calls[i])
				progressed = true
				break
			}
		}
		if !progressed {
			var stuck []string
			for _, i := range idx {
				if !done[i] {
					stuck = append(stuck, calls[i].Call)
				}
			}
			fails = append(fails, failure("ORDER_UNSATISFIABLE(cycle)", "the %s entry points %v depend on each other in a cycle.", when, stuck))
			return nil, fails
		}
	}
	return out, fails
}

func lifecycleSteps(v *mapmodel.Version, kind string) []mapmodel.Step {
	var steps []mapmodel.Step
	raw, ok := v.Facts["lifecycle"][kind].([]any)
	if !ok {
		return nil
	}
	for _, x := range raw {
		obj, _ := x.(map[string]any)
		s := mapmodel.Step{}
		s.Step, _ = obj["step"].(string)
		s.Phase, _ = obj["phase"].(string)
		after, _ := obj["after"].([]any)
		for _, a := range after {
			s.After = append(s.After, fmt.Sprint(a))
		}
		steps = append(steps, s)
	}
	return steps
}
