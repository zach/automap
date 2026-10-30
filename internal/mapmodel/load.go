package mapmodel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/zach/automap/schemas"
)

// Problem is one reason a map file is invalid.
type Problem struct {
	File   string // slash-separated, as the map's directory was given
	Reason string
}

// PrivateMap is a private map a manifest names, in layering order.
type PrivateMap struct {
	ID        string
	Location  string
	Platforms []string
}

// Map is the public map with any private maps layered over it.
type Map struct {
	Maps      []*MapMeta            // the public map, then private maps in layering order
	Slots     map[string]*Slot      // by id
	Contracts map[string][]Contract // by contract id, oldest version first
	Libraries map[string]*Library   // by id, with private extensions merged
	Recipes   map[string]*Recipe    // by recipe id

	origins    map[string][]string // "<kind> <id>" -> ids of the maps that define it
	recipeDirs map[string]string   // recipe id -> the directory its reference glue paths are relative to
	problems   []Problem
}

// file is one JSON file in a map, decoded once for schema validation.
type file struct {
	meta *MapMeta
	rel  string // relative to the map's directory
	raw  []byte
	doc  any
}

func (f *file) name() string { return path.Join(filepath.ToSlash(f.meta.Dir), f.rel) }

// Load reads the public map at dir and the private maps over it, validating
// every file. It returns the problems it found; a map with problems is
// incomplete and should not be used.
func Load(dir string, privates []PrivateMap) (*Map, []Problem) {
	m := &Map{
		Slots:     map[string]*Slot{},
		Contracts: map[string][]Contract{},
		Libraries: map[string]*Library{},
		Recipes:   map[string]*Recipe{},
		origins:   map[string][]string{},

		recipeDirs: map[string]string{},
	}
	var files []*file
	public := m.readMap(dir, "public", nil)
	if public != nil {
		files = append(files, m.mapFiles(public)...)
	}
	for _, pm := range privates {
		meta := m.readMap(pm.Location, "private", &pm)
		if meta != nil {
			files = append(files, m.mapFiles(meta)...)
		}
	}
	byKind := func(prefix string) []*file {
		var out []*file
		for _, f := range files {
			if strings.HasPrefix(f.rel, prefix) {
				out = append(out, f)
			}
		}
		return out
	}
	m.loadSlots(byKind("slots/"))
	m.loadContracts(byKind("contracts/"))
	m.loadLibraries(byKind("libraries/"))
	m.loadRecipes(byKind("recipes/"))
	m.checkSeeds()
	m.loadEvidence(byKind("evidence/"))
	sort.SliceStable(m.problems, func(i, j int) bool { return m.problems[i].File < m.problems[j].File })
	return m, m.problems
}

func (m *Map) fail(file, format string, args ...any) {
	m.problems = append(m.problems, Problem{File: file, Reason: fmt.Sprintf(format, args...)})
}

// readMap reads and checks a map's map.json.
func (m *Map) readMap(dir, kind string, private *PrivateMap) *MapMeta {
	meta := &MapMeta{Dir: dir}
	f := &file{meta: meta, rel: "map.json"}
	if _, err := os.Stat(dir); err != nil {
		m.fail(filepath.ToSlash(dir), "not a map directory: %v", err)
		return nil
	}
	if !m.decode(f, "map", meta) {
		return nil
	}
	meta.Dir = dir
	if (meta.Kind == "public") != (kind == "public") {
		m.fail(f.name(), "only the public map has kind public")
	}
	m.Maps = append(m.Maps, meta)
	return meta
}

// mapFiles lists the JSON files in a map that the map model reads.
func (m *Map) mapFiles(meta *MapMeta) []*file {
	var out []*file
	patterns := []string{"slots/*.json", "contracts/*/*.json", "libraries/*.json", "recipes/*/*.json"}
	root := os.DirFS(meta.Dir)
	for _, p := range patterns {
		matches, _ := fs.Glob(root, p)
		for _, rel := range matches {
			out = append(out, &file{meta: meta, rel: rel})
		}
	}
	fs.WalkDir(root, "evidence", func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".json") && !strings.HasSuffix(p, ".sigstore.json") {
			out = append(out, &file{meta: meta, rel: p})
		}
		return nil
	})
	return out
}

// decode reads f, validates it against the named schema and decodes it
// into v. It reports problems and returns false when f can't be used.
func (m *Map) decode(f *file, schema string, v any) bool {
	raw, err := os.ReadFile(filepath.Join(f.meta.Dir, filepath.FromSlash(f.rel)))
	if err != nil {
		m.fail(f.name(), "%v", err)
		return false
	}
	f.raw = raw
	if err := json.Unmarshal(raw, &f.doc); err != nil {
		m.fail(f.name(), "not valid JSON: %v", err)
		return false
	}
	if v := schemas.Validate(schema, f.doc); v != nil {
		for _, msg := range v {
			m.fail(f.name(), "%s: %s", schema, msg)
		}
		return false
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	if err := dec.Decode(v); err != nil {
		m.fail(f.name(), "%v", err)
		return false
	}
	return true
}

func (m *Map) origin(kind, id, mapID string) {
	key := kind + " " + id
	if !slices.Contains(m.origins[key], mapID) {
		m.origins[key] = append(m.origins[key], mapID)
	}
}

// Origins names the maps that define or extend an entry, in layering order.
func (m *Map) Origins(kind, id string) []string { return m.origins[kind+" "+id] }

func (m *Map) loadSlots(files []*file) {
	for _, f := range files {
		var s Slot
		if !m.decode(f, "slot", &s) {
			continue
		}
		if f.rel != "slots/"+s.ID+".json" {
			m.fail(f.name(), "file name must be the slot id")
		}
		if prev := m.Origins("slot", s.ID); len(prev) > 0 {
			m.fail(f.name(), "slot %s is already defined in map %s", s.ID, prev[0])
			continue
		}
		m.Slots[s.ID] = &s
		m.origin("slot", s.ID, f.meta.ID)
	}
}

func (m *Map) loadContracts(files []*file) {
	for _, f := range files {
		var c Contract
		if !m.decode(f, "contract", &c) {
			continue
		}
		if f.rel != "contracts/"+c.Contract+"/"+c.Version+".json" {
			m.fail(f.name(), "path must be contracts/<contract>/<version>.json")
		}
		for _, id := range sortedKeys(c.Choices) {
			ch := c.Choices[id]
			m.checkAgainst(f.name(), "choices/"+id+"/default", jsonValue(ch.Default), jsonValue(ch.Schema))
		}
		if c.Contract == "lifecycle" && len(c.Phases) == 0 {
			m.fail(f.name(), "the lifecycle contract lists its phases")
		}
		for _, r := range c.Rules {
			if !strings.HasPrefix(r.ID, c.Contract+".") {
				m.fail(f.name(), "rule %s must be prefixed with the contract id", r.ID)
			}
		}
		if m.contractVersion(c.Contract, c.Version) != nil {
			m.fail(f.name(), "contract %s %s is already defined", c.Contract, c.Version)
			continue
		}
		m.Contracts[c.Contract] = append(m.Contracts[c.Contract], c)
		slices.SortStableFunc(m.Contracts[c.Contract], func(a, b Contract) int { return compareVersions(a.Version, b.Version) })
		m.origin("contract", c.Contract, f.meta.ID)
	}
}

// checkAgainst reports every way value violates schema.
func (m *Map) checkAgainst(file, where string, value, schema any) {
	v, err := schemas.ValidateAgainst(schema, value)
	if err != nil {
		m.fail(file, "%s: the contract's schema doesn't compile: %v", where, err)
		return
	}
	for _, msg := range v {
		m.fail(file, "%s: %s", where, msg)
	}
}

// factsSchema is a JSON Schema for one contract's facts.
func factsSchema(c *Contract) any {
	props := map[string]any{}
	required := []any{}
	for _, id := range sortedKeys(c.Facts) {
		f := c.Facts[id]
		props[id] = jsonValue(f.Schema)
		if f.Required == nil || *f.Required {
			required = append(required, id)
		}
	}
	return map[string]any{"type": "object", "additionalProperties": false, "properties": props, "required": required}
}

func (m *Map) loadLibraries(files []*file) {
	for _, f := range files {
		var lib Library
		if !m.decode(f, "library", &lib) {
			continue
		}
		if f.rel != "libraries/"+lib.ID+".json" {
			m.fail(f.name(), "file name must be the library id")
		}
		if base, ok := m.Libraries[lib.ID]; ok {
			m.extendLibrary(f, base, &lib)
		} else {
			m.Libraries[lib.ID] = &lib
		}
		m.origin("library", lib.ID, f.meta.ID)
		m.checkLibrary(f, &lib)
	}
}

// extendLibrary merges a private map's file for a library a map beneath it
// already defines. It may add versions, variants and, in an NDA platform's
// map, that platform as a target; nothing else (ADR-0075).
func (m *Map) extendLibrary(f *file, base, ext *Library) {
	name := f.name()
	for _, c := range []struct {
		field     string
		base, ext any
	}{
		{"name", base.Name, ext.Name},
		{"license", base.License, ext.License},
		{"contribution_policy", base.ContributionPolicy, ext.ContributionPolicy},
		{"slots", base.Slots, ext.Slots},
		{"source", base.Source, ext.Source},
	} {
		if !reflect.DeepEqual(c.base, c.ext) {
			m.fail(name, "extends %s but its %s differs from the map it extends", ext.ID, c.field)
		}
	}
	for _, v := range ext.Versions {
		i := slices.IndexFunc(base.Versions, func(b Version) bool { return b.Version == v.Version })
		if i < 0 {
			base.Versions = append(base.Versions, v)
			continue
		}
		bv := &base.Versions[i]
		for _, c := range []struct {
			field   string
			differs bool
		}{
			{"released", v.Released != bv.Released},
			{"security_fix", v.SecurityFix != bv.SecurityFix},
			{"pin", !reflect.DeepEqual(v.Pin, bv.Pin)},
			{"build_backend", !reflect.DeepEqual(v.BuildBackend, bv.BuildBackend)},
			{"cmake", !reflect.DeepEqual(v.CMake, bv.CMake)},
			{"bundled", !reflect.DeepEqual(v.Bundled, bv.Bundled)},
		} {
			if c.differs {
				m.fail(name, "extends %s %s but its %s differs", ext.ID, v.Version, c.field)
			}
		}
		newTargets, oldTargets := targets(v.Facts), targets(bv.Facts)
		ndaPlatforms := f.meta.Platforms
		switch {
		case !reflect.DeepEqual(withoutTargets(v.Facts), withoutTargets(bv.Facts)) || !subset(oldTargets, newTargets):
			m.fail(name, "extends %s %s: facts may only gain platform targets", ext.ID, v.Version)
		case !subset(difference(newTargets, oldTargets), ndaPlatforms):
			m.fail(name, "extends %s %s: only an NDA platform's map may add targets, and only its own platforms", ext.ID, v.Version)
		default:
			added := difference(newTargets, oldTargets)
			if len(added) > 0 {
				all := append(slices.Clone(oldTargets), added...)
				bv.Facts["platform"]["targets"] = toAny(all)
			}
		}
		for _, variant := range v.Variants {
			j := slices.IndexFunc(bv.Variants, func(b Variant) bool { return b.Name == variant.Name })
			switch {
			case j < 0:
				bv.Variants = append(bv.Variants, variant)
			case !reflect.DeepEqual(variant, bv.Variants[j]):
				m.fail(name, "variant %s differs from the map it extends", variant.Name)
			}
		}
	}
}

// checkLibrary checks one library file's own content.
func (m *Map) checkLibrary(f *file, lib *Library) {
	name := f.name()
	for _, s := range lib.Slots {
		if m.Slots[s] == nil {
			m.fail(name, "slot %s is not in the map", s)
		}
	}
	for _, v := range lib.Versions {
		for _, seam := range sortedKeys(v.Facts) {
			if c := m.latestContract(seam); c != nil {
				m.checkAgainst(name, v.Version+"/facts/"+seam, toAny(v.Facts[seam]), factsSchema(c))
			}
		}
		lc := v.Facts["lifecycle"]
		for _, kind := range []string{"init_steps", "shutdown_steps"} {
			steps := lifecycleSteps(lc, kind)
			names := map[string]bool{}
			for _, s := range steps {
				names[s.Step] = true
			}
			for _, s := range steps {
				if !strings.HasPrefix(s.Step, lib.ID+":") {
					m.fail(name, "step %s must be prefixed with '%s:'", s.Step, lib.ID)
				}
				for _, a := range s.After {
					if !names[a] && !strings.HasPrefix(a, "automap:") {
						m.fail(name, "step %s follows unknown step %s", s.Step, a)
					}
				}
			}
		}
		for _, variant := range v.Variants {
			for _, seam := range sortedKeys(variant.Facts) {
				if c := m.latestContract(seam); c != nil {
					merged := map[string]any{}
					for k, x := range v.Facts[seam] {
						merged[k] = x
					}
					for k, x := range variant.Facts[seam] {
						merged[k] = x
					}
					m.checkAgainst(name, v.Version+"/"+variant.Name+"/facts/"+seam, toAny(merged), factsSchema(c))
				}
			}
		}
	}
	if lib.Source != nil && lib.Source.Kind == "path" && len(lib.ConsoleVariants) > 0 {
		m.fail(name, "console_variants belong to the public map only")
	}
}

// Step is one lifecycle step a library declares.
type Step struct {
	Phase string   `json:"phase"`
	Step  string   `json:"step"`
	After []string `json:"after,omitempty"`
}

func lifecycleSteps(lc map[string]any, kind string) []Step {
	var steps []Step
	if raw, err := json.Marshal(lc[kind]); err == nil {
		json.Unmarshal(raw, &steps)
	}
	return steps
}

func (m *Map) loadRecipes(files []*file) {
	contractRules := m.contractRules()
	for _, f := range files {
		var r Recipe
		if !m.decode(f, "recipe", &r) {
			continue
		}
		name := f.name()
		if path.Dir(f.rel) != "recipes/"+r.Kind {
			m.fail(name, "a %s recipe lives in recipes/%s/", r.Kind, r.Kind)
		}
		rid := r.ID()
		switch r.Kind {
		case "binding":
			if m.contractVersion(r.Subject.Contract, r.Subject.ContractVersion) == nil {
				m.fail(name, "contract_version %s names no %s contract in the map", r.Subject.ContractVersion, r.Subject.Contract)
			}
		case "pairing":
			if !slices.IsSorted(r.Subject.Libraries) {
				m.fail(name, "pairing libraries are listed in sorted order")
			}
		}
		if path.Base(f.rel) != rid+".json" {
			m.fail(name, "file name must be %s", rid)
		}
		own := map[string]Rule{}
		for _, rule := range r.Rules {
			own[rule.ID] = rule
			if !strings.HasPrefix(rule.ID, rid+".") {
				m.fail(name, "pairing rule %s must be prefixed with '%s.'", rule.ID, rid)
			}
		}
		for _, lib := range sortedKeys(r.AppliesTo) {
			a := r.AppliesTo[lib]
			for _, ver := range a.Versions {
				for _, variant := range a.Variants {
					if m.libraryVersion(lib, ver, variant) == nil {
						m.fail(name, "applies_to names %s %s (%s), which no library in the map has", lib, ver, variant)
					}
				}
			}
		}
		for _, p := range r.Probes {
			rule, ok := own[p.Rule]
			if !ok {
				rule, ok = contractRules[p.Rule]
			}
			switch {
			case !ok:
				m.fail(name, "probe for unknown rule %s", p.Rule)
			case rule.Kind != "probed":
				m.fail(name, "probe for guidance rule %s", p.Rule)
			}
		}
		for _, na := range r.NotApplicable {
			if _, ok := contractRules[na.Rule]; !ok {
				m.fail(name, "not_applicable names unknown rule %s", na.Rule)
			}
		}
		hasRole := func(role string) bool {
			return slices.ContainsFunc(r.ReferenceGlue, func(g Glue) bool { return g.Role == role })
		}
		if r.Kind == "library" && !hasRole("build-fragment") {
			m.fail(name, "a library recipe ships a build fragment")
		}
		if r.Kind != "library" && !hasRole("probe") {
			var libRules []string
			if c := m.latestContract(r.Subject.Contract); c != nil {
				for _, rule := range c.Rules {
					if rule.Kind == "probed" && rule.AppliesTo == "library" {
						libRules = append(libRules, rule.ID)
					}
				}
			}
			var na []string
			for _, n := range r.NotApplicable {
				na = append(na, n.Rule)
			}
			if r.Kind == "pairing" || len(libRules) == 0 || !subset(libRules, na) {
				m.fail(name, "binding and pairing recipes ship their probes, unless every probed library rule is not-applicable")
			}
		}
		owners := append(slices.Clone(r.Subject.Libraries), rid)
		for _, o := range r.Ordering {
			for _, step := range []string{o.Before, o.After} {
				if !slices.Contains(owners, stepOwner(step)) {
					m.fail(name, "ordering step %s isn't owned by %v", step, owners)
				}
			}
		}
		for _, ep := range r.EntryPoints {
			for _, step := range ep.Steps {
				owner := stepOwner(step)
				if lib, ok := m.Libraries[owner]; ok {
					kind := "init_steps"
					if ep.When == "shutdown" {
						kind = "shutdown_steps"
					}
					declared := false
					for _, v := range lib.Versions {
						for _, s := range lifecycleSteps(v.Facts["lifecycle"], kind) {
							declared = declared || s.Step == step
						}
					}
					if !declared {
						m.fail(name, "%s performs %s, which %s's lifecycle facts don't declare", ep.Call, step, owner)
					}
				} else if owner != rid && owner != "automap" {
					m.fail(name, "step %s is owned by neither a library nor this recipe", step)
				}
			}
		}
		if prev := m.Origins("recipe", rid); len(prev) > 0 {
			m.fail(name, "recipe %s is already defined in map %s", rid, prev[0])
			continue
		}
		m.Recipes[rid] = &r
		m.recipeDirs[rid] = filepath.Join(f.meta.Dir, filepath.FromSlash(path.Dir(f.rel)))
		m.origin("recipe", rid, f.meta.ID)
	}
}

// checkSeeds checks that every probed composition rule is seeded by its
// contract or by some recipe (ADR-0071), and that only manifest seeds are
// JSON Patches.
func (m *Map) checkSeeds() {
	seeded := map[string]bool{}
	for _, r := range m.Recipes {
		for _, p := range r.Probes {
			seeded[p.Rule] = true
		}
	}
	for _, id := range sortedKeys(m.Contracts) {
		c := m.Contracts[id][len(m.Contracts[id])-1]
		where := path.Join(filepath.ToSlash(m.Maps[0].Dir), "contracts", id)
		for _, r := range c.Rules {
			if r.Kind == "probed" && r.AppliesTo == "composition" && r.SeededViolation == nil && !seeded[r.ID] {
				m.fail(where, "probed composition rule %s has no seeded violation in its contract or any recipe", r.ID)
			}
			if r.SeededViolation != nil && strings.HasSuffix(r.SeededViolation.Patch, ".json-patch") && r.AppliesTo != "composition" {
				m.fail(where, "%s: only manifest seeds are JSON Patches", r.ID)
			}
		}
	}
}

func (m *Map) loadEvidence(files []*file) {
	contractRules := m.contractRules()
	for _, f := range files {
		var e Evidence
		if !m.decode(f, "evidence", &e) {
			continue
		}
		name := f.name()
		p := &e.Predicate
		for _, r := range p.Results {
			if plats := contractRules[r.Rule].Platforms; len(plats) > 0 && !slices.Contains(plats, p.Platform) && r.Status != "not-applicable" {
				m.fail(name, "%s runs only on %v; on %s it is not-applicable", r.Rule, plats, p.Platform)
			}
		}
		want := "nda-holder"
		if f.meta.Kind == "public" {
			want = "automap-ci"
		}
		if p.Verifier.Kind != want {
			m.fail(name, "the public map's evidence comes from Automap's CI; a private map's from the NDA holder")
		}
		if p.SubjectKind == "binding" {
			pin := p.SubjectDetail.Pin
			if m.libraryVersion(pin.Library, pin.Version, pin.Variant) == nil {
				m.fail(name, "no library has %s %s (%s)", pin.Library, pin.Version, pin.Variant)
			}
			for _, r := range p.Results {
				if !strings.HasPrefix(r.Rule, p.SubjectDetail.Contract+".") {
					m.fail(name, "binding evidence for %s reports rule %s", p.SubjectDetail.Contract, r.Rule)
				}
			}
		}
		if p.MapRelease == "unreleased" && p.Verifier.Signed {
			m.fail(name, "evidence against an unreleased map is never signed (ADR-0079)")
		}
		if KnownGood(p.Level) && slices.ContainsFunc(p.Results, func(r EvidenceResult) bool {
			return r.Status == "fail" || r.Status == "unverifiable"
		}) {
			m.fail(name, "known-good level with failing or unverifiable results")
		}
	}
}

// RecipeFile is the path on disk of a file a recipe names, such as its
// reference glue or a seeded violation's patch.
func (m *Map) RecipeFile(rid, rel string) string {
	return filepath.Join(m.recipeDirs[rid], filepath.FromSlash(rel))
}

// ContractFile is the path on disk of a file a contract version names,
// such as a shared interface source.
func (m *Map) ContractFile(c *Contract, rel string) string {
	for _, meta := range m.Maps {
		p := filepath.Join(meta.Dir, "contracts", c.Contract, c.Version, filepath.FromSlash(rel))
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return filepath.Join(m.Maps[0].Dir, "contracts", c.Contract, c.Version, filepath.FromSlash(rel))
}

// PublicDir is the public map's directory, which holds the testbed
// template and the probe harness.
func (m *Map) PublicDir() string { return m.Maps[0].Dir }

// LatestContract returns a contract's newest version in the map, or nil.
func (m *Map) LatestContract(id string) *Contract { return m.latestContract(id) }

// Newest returns the library's most recently released version.
func (l *Library) Newest() *Version {
	best := &l.Versions[0]
	for i := range l.Versions[1:] {
		if v := &l.Versions[i+1]; v.Released >= best.Released {
			best = v
		}
	}
	return best
}

// Variant returns the version's variant with the given name, or nil.
func (v *Version) Variant(name string) *Variant {
	for i := range v.Variants {
		if v.Variants[i].Name == name {
			return &v.Variants[i]
		}
	}
	return nil
}

// KnownGood reports whether an evidence level means every applicable rule
// passed.
func KnownGood(level string) bool {
	return level == "probed" || level == "behavioral" || level == "hardware"
}

// ID is the recipe's id: <library>--<contract> for a binding, the two
// libraries joined by -- for a pairing, and the library for a library
// recipe.
func (r *Recipe) ID() string {
	switch r.Kind {
	case "binding":
		return r.Subject.Library + "--" + r.Subject.Contract
	case "pairing":
		return strings.Join(r.Subject.Libraries, "--")
	default:
		return r.Subject.Library
	}
}

func (m *Map) contractVersion(id, version string) *Contract {
	for i, c := range m.Contracts[id] {
		if c.Version == version {
			return &m.Contracts[id][i]
		}
	}
	return nil
}

func (m *Map) latestContract(id string) *Contract {
	cs := m.Contracts[id]
	if len(cs) == 0 {
		return nil
	}
	return &cs[len(cs)-1]
}

// contractRules maps every contract rule id to its rule, from the newest
// contract version that has it.
func (m *Map) contractRules() map[string]Rule {
	out := map[string]Rule{}
	for _, cs := range m.Contracts {
		for _, c := range cs {
			for _, r := range c.Rules {
				out[r.ID] = r
			}
		}
	}
	return out
}

func (m *Map) libraryVersion(lib, version, variant string) *Version {
	l := m.Libraries[lib]
	if l == nil {
		return nil
	}
	for i, v := range l.Versions {
		if v.Version == version && slices.ContainsFunc(v.Variants, func(x Variant) bool { return x.Name == variant }) {
			return &l.Versions[i]
		}
	}
	return nil
}

func stepOwner(step string) string {
	owner, _, _ := strings.Cut(step, ":")
	return owner
}

func targets(f Facts) []string {
	raw, _ := f["platform"]["targets"].([]any)
	var out []string
	for _, t := range raw {
		if s, ok := t.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func withoutTargets(f Facts) Facts {
	out := Facts{}
	for seam, facts := range f {
		out[seam] = map[string]any{}
		for k, v := range facts {
			if !(seam == "platform" && k == "targets") {
				out[seam][k] = v
			}
		}
	}
	return out
}

func subset(a, b []string) bool {
	for _, x := range a {
		if !slices.Contains(b, x) {
			return false
		}
	}
	return true
}

func difference(a, b []string) []string {
	var out []string
	for _, x := range a {
		if !slices.Contains(b, x) {
			out = append(out, x)
		}
	}
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// jsonValue decodes raw JSON to a generic value.
func jsonValue(raw json.RawMessage) any {
	var v any
	json.Unmarshal(raw, &v)
	return v
}

// toAny round-trips v through JSON, giving the generic value schema
// validation expects.
func toAny(v any) any {
	raw, _ := json.Marshal(v)
	var out any
	json.Unmarshal(raw, &out)
	return out
}

// compareVersions orders dotted numeric versions; other strings compare
// lexically.
func compareVersions(a, b string) int {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) && i < len(bs); i++ {
		var x, y int
		_, errA := fmt.Sscanf(as[i], "%d", &x)
		_, errB := fmt.Sscanf(bs[i], "%d", &y)
		if errA != nil || errB != nil {
			return strings.Compare(as[i], bs[i])
		}
		if x != y {
			return x - y
		}
	}
	return len(as) - len(bs)
}
