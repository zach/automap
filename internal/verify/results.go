package verify

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/zach/automap/internal/composition"
	"github.com/zach/automap/internal/mapmodel"
)

// Result is one rule's result in evidence (ADR-0072).
type Result struct {
	Rule        string `json:"rule"`
	Library     string `json:"library,omitempty"`
	Status      string `json:"status"`
	FailureCode string `json:"failure_code,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

// Expected is a probed rule that applies to the composition: once for a
// composition or pairing rule, once per library for a library rule.
type Expected struct {
	Rule          mapmodel.Rule
	Library       string
	Platforms     []string
	NotApplicable string // why the library's binding recipe marks the rule not-applicable, if it does
}

// ExpectedRules lists every probed rule that applies to the composition.
func ExpectedRules(m *mapmodel.Map, man *composition.Manifest) []Expected {
	var out []Expected
	for _, cid := range composition.Contracts(m, man) {
		for _, rule := range m.LatestContract(cid).Rules {
			if rule.Kind != "probed" {
				continue
			}
			if rule.AppliesTo != "library" {
				e := Expected{Rule: rule, Platforms: rule.Platforms}
				if rule.ID == "build-flags.abi-defines-match" && !declaresABIDefines(m, man) {
					e.NotApplicable = "No library in the composition declares an ABI-affecting define."
				}
				out = append(out, e)
				continue
			}
			for _, lib := range man.Libraries() {
				if _, ok := m.Libraries[lib].Newest().Facts[cid]; ok {
					e := Expected{Rule: rule, Library: lib, Platforms: rule.Platforms}
					if binding := m.Recipes[lib+"--"+cid]; binding != nil {
						for _, na := range binding.NotApplicable {
							if na.Rule == rule.ID {
								e.NotApplicable = na.Reason
							}
						}
					}
					out = append(out, e)
				}
			}
		}
	}
	for _, rid := range composition.Recipes(m, man) {
		for _, rule := range m.Recipes[rid].Rules {
			if rule.Kind == "probed" {
				out = append(out, Expected{Rule: rule, Platforms: rule.Platforms})
			}
		}
	}
	return out
}

// declaresABIDefines reports whether any library in the composition, in its
// chosen variant, declares an ABI-affecting define.
func declaresABIDefines(m *mapmodel.Map, man *composition.Manifest) bool {
	for _, lib := range man.Libraries() {
		v := m.Libraries[lib].Newest()
		facts := v.Facts["build-flags"]
		if variant := v.Variant(man.Variants[lib].Default); variant != nil && variant.Facts["build-flags"] != nil {
			facts = variant.Facts["build-flags"]
		}
		if defines, ok := facts["abi_defines"].([]any); ok && len(defines) > 0 {
			return true
		}
	}
	return false
}

// Observed is what a probe run reported, and the results verify computed
// itself from it.
type Observed struct {
	Ran      bool     // whether the probes ran at all
	Level    string   // the level reached before probing, for reasons
	Reported []Result // results the probe harness recorded
	Computed []Result // results verify computed: lifecycle order, platform targets
}

// Results assigns every expected rule its status: not-applicable on a
// platform its probe doesn't run on, the computed or reported result when
// there is one, and unverifiable with a reason otherwise. It is a pure
// function.
func Results(expected []Expected, platform string, obs Observed) []Result {
	var out []Result
	find := func(results []Result, e Expected) *Result {
		for i := range results {
			if results[i].Rule == e.Rule.ID && results[i].Library == e.Library {
				return &results[i]
			}
		}
		return nil
	}
	for _, e := range expected {
		r := Result{Rule: e.Rule.ID, Library: e.Library}
		computed := find(obs.Computed, e)
		switch got := firstNonNil(computed, find(obs.Reported, e)); {
		case e.NotApplicable != "":
			r.Status, r.Reason = "not-applicable", e.NotApplicable
		case len(e.Platforms) > 0 && !slices.Contains(e.Platforms, platform):
			r.Status, r.Reason = "not-applicable", "The rule's probe runs only on "+strings.Join(e.Platforms, ", ")+"."
		case computed != nil:
			// Some results don't need the probes, such as a mismatch the
			// linker found.
			r = *computed
		case !obs.Ran:
			r.Status, r.Reason = "unverifiable", "No probe ran: the composition reached "+obs.Level+"."
		case got != nil:
			r = *got
		default:
			r.Status, r.Reason = "unverifiable", "No probe reported a result for this rule."
		}
		out = append(out, r)
	}
	return out
}

func firstNonNil(rs ...*Result) *Result {
	for _, r := range rs {
		if r != nil {
			return r
		}
	}
	return nil
}

// Level is the highest level a run reached (SPEC §6.2): none if it didn't
// build, builds if its smoke run failed, smoke unless the probes ran and
// every applicable result passed, and then probed. It is a pure function.
func Level(built, smoked, probesRan bool, results []Result) string {
	switch {
	case !built:
		return "none"
	case !smoked:
		return "builds"
	case !probesRan:
		return "smoke"
	}
	for _, r := range results {
		if r.Status == "fail" || r.Status == "unverifiable" {
			return "smoke"
		}
	}
	return "probed"
}

// StepInfo is what the map declares about one lifecycle step.
type StepInfo struct {
	Phase int
	After []string
}

// Lifecycle is the composition's declared init and shutdown steps, with
// the pairings' ordering constraints.
type Lifecycle struct {
	Init, Shutdown map[string]StepInfo
	Ordering       []mapmodel.Ordering
}

// DeclaredLifecycle collects the steps the composition's libraries declare
// in their lifecycle facts and the steps its recipes' entry points perform,
// with the after lists of both, and the pairings' ordering constraints.
func DeclaredLifecycle(m *mapmodel.Map, man *composition.Manifest) Lifecycle {
	lc := Lifecycle{Init: map[string]StepInfo{}, Shutdown: map[string]StepInfo{}}
	add := func(steps map[string]StepInfo, step, phase string, after []string) {
		info := steps[step]
		info.Phase = slices.Index(composition.Phases, phase)
		for _, a := range after {
			if !slices.Contains(info.After, a) {
				info.After = append(info.After, a)
			}
		}
		steps[step] = info
	}
	for _, lib := range man.Libraries() {
		v := m.Libraries[lib].Newest()
		for kind, steps := range map[string]map[string]StepInfo{"init_steps": lc.Init, "shutdown_steps": lc.Shutdown} {
			raw, _ := v.Facts["lifecycle"][kind].([]any)
			for _, x := range raw {
				s, _ := x.(map[string]any)
				name, _ := s["step"].(string)
				phase, _ := s["phase"].(string)
				var after []string
				if a, ok := s["after"].([]any); ok {
					for _, y := range a {
						after = append(after, fmt.Sprint(y))
					}
				}
				add(steps, name, phase, after)
			}
		}
	}
	for _, c := range composition.InterfaceCalls(m, man) {
		steps := lc.Init
		if c.When == "shutdown" {
			steps = lc.Shutdown
		}
		for _, s := range c.Steps {
			add(steps, s, c.Phase, c.After)
		}
	}
	for _, rid := range composition.Recipes(m, man) {
		r := m.Recipes[rid]
		for _, ep := range r.EntryPoints {
			steps := lc.Init
			if ep.When == "shutdown" {
				steps = lc.Shutdown
			}
			for _, s := range ep.Steps {
				add(steps, s, ep.Phase, ep.After)
			}
		}
		lc.Ordering = append(lc.Ordering, r.Ordering...)
	}
	return lc
}

// CheckLifecycle judges the recorded init and shutdown steps against the
// declared lifecycle: init steps in non-decreasing phase order and shutdown
// steps in non-increasing, each after every step it must follow, and every
// ordering constraint holding. It is a pure function.
func CheckLifecycle(lc Lifecycle, init, shutdown []string) []Result {
	return []Result{
		checkOrder("lifecycle.init-order", "INIT_ORDER", lc.Init, lc.Ordering, "init", init, 1),
		checkOrder("lifecycle.shutdown-reverse", "SHUTDOWN_ORDER", lc.Shutdown, lc.Ordering, "shutdown", shutdown, -1),
	}
}

// checkOrder checks one recorded sequence. direction is 1 when phases must
// not decrease and -1 when they must not increase.
func checkOrder(rule, code string, steps map[string]StepInfo, ordering []mapmodel.Ordering, when string, recorded []string, direction int) Result {
	if len(recorded) == 0 {
		return Result{Rule: rule, Status: "unverifiable", Reason: "No " + when + " steps were recorded."}
	}
	pos := map[string]int{}
	for i, s := range recorded {
		if _, seen := pos[s]; !seen {
			pos[s] = i
		}
	}
	fail := func(first, second, why string) Result {
		return Result{Rule: rule, Status: "fail", FailureCode: fmt.Sprintf("%s(%s before %s)", code, first, second), Reason: why}
	}
	for i, s := range recorded {
		info, ok := steps[s]
		if !ok {
			continue
		}
		for _, earlier := range recorded[:i] {
			if e, ok := steps[earlier]; ok && (info.Phase-e.Phase)*direction < 0 {
				return fail(earlier, s, fmt.Sprintf("%s ran before %s, but its phase, %s, comes after %s's.", earlier, s, composition.Phases[e.Phase], s))
			}
		}
		for _, a := range info.After {
			if p, ok := pos[a]; ok && p > i {
				return fail(s, a, fmt.Sprintf("%s ran before %s, which it must follow.", s, a))
			}
		}
	}
	for _, o := range ordering {
		if o.When != when {
			continue
		}
		b, okB := pos[o.Before]
		a, okA := pos[o.After]
		if okA && okB && a < b {
			return fail(o.After, o.Before, fmt.Sprintf("%s ran before %s, but a pairing requires %s first.", o.After, o.Before, o.Before))
		}
	}
	return Result{Rule: rule, Status: "pass"}
}

// CheckTargets is the platform.targets-supported result: every library
// lists every platform the composition names among its targets.
func CheckTargets(m *mapmodel.Map, man *composition.Manifest) Result {
	r := Result{Rule: "platform.targets-supported", Status: "pass"}
	for _, lib := range man.Libraries() {
		v := m.Libraries[lib].Newest()
		var targets []string
		facts := v.Facts["platform"]
		if variant := v.Variant(man.Variants[lib].Default); variant != nil && variant.Facts["platform"]["targets"] != nil {
			facts = variant.Facts["platform"]
		}
		if raw, ok := facts["targets"].([]any); ok {
			for _, t := range raw {
				targets = append(targets, fmt.Sprint(t))
			}
		}
		for _, p := range man.Platforms {
			if !slices.Contains(targets, p) {
				return Result{Rule: r.Rule, Status: "fail", FailureCode: fmt.Sprintf("PLATFORM_UNSUPPORTED(%s,%s)", lib, p),
					Reason: fmt.Sprintf("%s doesn't list %s among its targets.", lib, p)}
			}
		}
	}
	return r
}

// Precisions is each library's position_precision math-types fact, in its
// chosen variant, for the libraries that declare one.
func Precisions(m *mapmodel.Map, man *composition.Manifest) map[string]string {
	out := map[string]string{}
	for _, lib := range man.Libraries() {
		v := m.Libraries[lib].Newest()
		facts := v.Facts["math-types"]
		if variant := v.Variant(man.Variants[lib].Default); variant != nil && variant.Facts["math-types"]["position_precision"] != nil {
			facts = variant.Facts["math-types"]
		}
		if p, ok := facts["position_precision"].(string); ok {
			out[lib] = p
		}
	}
	return out
}

// CheckPrecision is the math-types.precision-agrees result: every library
// that exposes world positions uses the composition's position_precision.
// It is a pure function.
func CheckPrecision(choice string, precisions map[string]string) Result {
	libs := make([]string, 0, len(precisions))
	for lib := range precisions {
		libs = append(libs, lib)
	}
	slices.Sort(libs)
	for _, lib := range libs {
		if precisions[lib] != choice {
			return Result{Rule: "math-types.precision-agrees", Status: "fail", FailureCode: fmt.Sprintf("PRECISION_MISMATCH(%s)", lib),
				Reason: fmt.Sprintf("%s uses %s world positions, but the composition chooses %s.", lib, precisions[lib], choice)}
		}
	}
	return Result{Rule: "math-types.precision-agrees", Status: "pass"}
}

// linkMismatch is MSVC's report of a #pragma detect_mismatch conflict.
var linkMismatch = regexp.MustCompile(`LNK2038: mismatch detected for '([^']+)': (.*)`)

// LinkMismatch is the build-flags.abi-defines-match result a failed build's
// log shows, when MSVC's linker found two translation units built with
// different values of an ABI define declared with #pragma detect_mismatch,
// or nil. It is a pure function.
func LinkMismatch(buildLog string) *Result {
	m := linkMismatch.FindStringSubmatch(buildLog)
	if m == nil {
		return nil
	}
	return &Result{Rule: "build-flags.abi-defines-match", Status: "fail", FailureCode: "ABI_DEFINE_MISMATCH(" + m[1] + ")",
		Reason: "The linker found translation units built with different " + m[1] + ": " + strings.TrimSpace(m[2]) + "."}
}
