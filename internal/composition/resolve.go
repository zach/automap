package composition

import (
	"encoding/json"
	"path"
	"slices"
	"sort"
	"strings"

	"github.com/zach/automap/internal/mapmodel"
)

// V1Platforms are the platforms start composes for unless told otherwise.
var V1Platforms = []string{"windows-x64-msvc", "linux-x64-clang", "macos-arm64-appleclang"}

// LibraryRequest is one --with: a library and, optionally, its variant.
type LibraryRequest struct {
	Library string
	Variant string
}

// SlotRequest is one --slot <slot>=<library>.
type SlotRequest struct {
	Slot    string
	Library string
}

// Request is what start is asked to compose.
type Request struct {
	With       []LibraryRequest
	Slots      []SlotRequest
	Platforms  []string
	MapRelease string
	Executable Executable
}

// Default glue locations (SPEC §2.3).
const (
	InterfacesDir = "glue/automap"
	GlueDir       = "glue"
)

// NewManifest assigns each requested library to the slots it fills, or to
// the slots --slot names for it, and lays out the composition's glue. It
// is a pure function of the map and the request.
func NewManifest(m *mapmodel.Map, req Request) (*Manifest, []Failure) {
	var fails []Failure
	man := &Manifest{
		Schema:       "https://github.com/zach/automap/schemas/manifest.schema.json",
		Format:       "automap-manifest/1",
		MapRelease:   req.MapRelease,
		BuildBackend: "vcpkg",
		Platforms:    req.Platforms,
		Executable:   req.Executable,
		Slots:        map[string]string{},
		Variants:     map[string]VariantChoice{},
		Choices:      map[string]map[string]any{},
	}
	if len(man.Platforms) == 0 {
		man.Platforms = slices.Clone(V1Platforms)
	}

	variants := map[string]string{}
	var order []string
	addLibrary := func(lib, variant string) bool {
		l := m.Libraries[lib]
		if l == nil {
			fails = append(fails, failure("UNKNOWN_ENTRY(library "+lib+")", "The map has no library %s; automap find lists libraries.", lib))
			return false
		}
		v := l.Newest()
		if variant == "" {
			variant = v.Variants[0].Name
		}
		if v.Variant(variant) == nil {
			var names []string
			for _, x := range v.Variants {
				names = append(names, x.Name)
			}
			fails = append(fails, failure("UNKNOWN_VARIANT("+lib+":"+variant+")", "%s %s has no variant %s; its variants are %s.", lib, v.Version, variant, strings.Join(names, ", ")))
			return false
		}
		if prev, ok := variants[lib]; ok && prev != variant {
			fails = append(fails, failure("VARIANT_CONFLICT("+lib+")", "%s is asked for as both %s and %s.", lib, prev, variant))
			return false
		}
		if _, ok := variants[lib]; !ok {
			order = append(order, lib)
		}
		variants[lib] = variant
		return true
	}

	explicit := map[string]bool{}
	for _, s := range req.Slots {
		explicit[s.Library] = true
	}
	for _, w := range req.With {
		addLibrary(w.Library, w.Variant)
	}
	for _, s := range req.Slots {
		if _, ok := variants[s.Library]; !ok && !addLibrary(s.Library, "") {
			continue
		}
		switch {
		case m.Slots[s.Slot] == nil:
			fails = append(fails, failure("UNKNOWN_ENTRY(slot "+s.Slot+")", "The map has no slot %s.", s.Slot))
		case !slices.Contains(m.Libraries[s.Library].Slots, s.Slot):
			fails = append(fails, failure("SLOT_MISMATCH("+s.Slot+"="+s.Library+")", "%s doesn't fill %s; it fills %s.", s.Library, s.Slot, strings.Join(m.Libraries[s.Library].Slots, ", ")))
		case man.Slots[s.Slot] != "":
			fails = append(fails, failure("SLOT_TAKEN("+s.Slot+")", "%s already fills %s.", man.Slots[s.Slot], s.Slot))
		default:
			man.Slots[s.Slot] = s.Library
		}
	}
	for _, lib := range order {
		if explicit[lib] {
			continue
		}
		for _, slot := range m.Libraries[lib].Slots {
			if other := man.Slots[slot]; other != "" && other != lib {
				fails = append(fails, failure("SLOT_TAKEN("+slot+")", "%s and %s both fill %s; name the one you want with --slot %s=<library>.", other, lib, slot, slot))
				continue
			}
			man.Slots[slot] = lib
		}
	}
	if len(fails) > 0 {
		return nil, fails
	}
	for lib, variant := range variants {
		man.Variants[lib] = VariantChoice{Default: variant}
	}
	for _, c := range Contracts(m, man) {
		contract := m.LatestContract(c)
		if len(contract.Choices) == 0 {
			continue
		}
		man.Choices[c] = map[string]any{}
		for id, ch := range contract.Choices {
			var v any
			json.Unmarshal(ch.Default, &v)
			man.Choices[c][id] = v
		}
	}
	slotChoices(m, man)
	man.Glue = defaultGlue(m, man)
	return man, nil
}

// slotChoices sets every contract choice that can name a filled slot to
// that slot: a choice whose values include "<slot>-slot", such as the jobs
// contract's scheduler, takes "jobs-slot" once a library fills the jobs
// slot (SPEC §2.5).
func slotChoices(m *mapmodel.Map, man *Manifest) {
	for _, cid := range Contracts(m, man) {
		for id, ch := range m.LatestContract(cid).Choices {
			var schema struct {
				Enum []any `json:"enum"`
			}
			json.Unmarshal(ch.Schema, &schema)
			for slot, lib := range man.Slots {
				if lib != "" && slices.Contains(schema.Enum, any(slot+"-slot")) {
					if man.Choices[cid] == nil {
						man.Choices[cid] = map[string]any{}
					}
					man.Choices[cid][id] = slot + "-slot"
				}
			}
		}
	}
}

// Contracts lists the contracts a composition uses: those its libraries
// declare facts for, and those its manifest makes choices at.
func Contracts(m *mapmodel.Map, man *Manifest) []string {
	seen := map[string]bool{}
	for _, lib := range man.Libraries() {
		l := m.Libraries[lib]
		if l == nil {
			continue
		}
		v := l.Newest()
		for c := range v.Facts {
			seen[c] = true
		}
		if variant := v.Variant(man.Variants[lib].Default); variant != nil {
			for c := range variant.Facts {
				seen[c] = true
			}
		}
	}
	for c := range man.Choices {
		seen[c] = true
	}
	var out []string
	for c := range seen {
		if m.LatestContract(c) != nil {
			out = append(out, c)
		}
	}
	sort.Strings(out)
	return out
}

// Recipes lists the recipes a composition uses, in a fixed order: each
// library's bindings, then its library recipe, then the pairings.
func Recipes(m *mapmodel.Map, man *Manifest) []string {
	libs := man.Libraries()
	var bindings, libraries, pairings []string
	for _, rid := range sortedKeys(m.Recipes) {
		r := m.Recipes[rid]
		switch r.Kind {
		case "binding":
			if slices.Contains(libs, r.Subject.Library) {
				bindings = append(bindings, rid)
			}
		case "library":
			if slices.Contains(libs, r.Subject.Library) {
				libraries = append(libraries, rid)
			}
		case "pairing":
			if slices.Contains(libs, r.Subject.Libraries[0]) && slices.Contains(libs, r.Subject.Libraries[1]) {
				pairings = append(pairings, rid)
			}
		}
	}
	return slices.Concat(bindings, libraries, pairings)
}

// GlueDirFor is the directory a recipe's glue is written to: glue/<library>
// for a binding or library recipe, glue/<a>--<b> for a pairing.
func GlueDirFor(r *mapmodel.Recipe) string {
	if r.Kind == "pairing" {
		return path.Join(GlueDir, r.ID())
	}
	return path.Join(GlueDir, r.Subject.Library)
}

// defaultGlue lays out each recipe's glue at its default location.
func defaultGlue(m *mapmodel.Map, man *Manifest) Glue {
	g := Glue{
		Interfaces:    InterfacesDir,
		ChoicesHeader: path.Join(InterfacesDir, "choices.h"),
		Startup:       Startup{Files: []string{path.Join(InterfacesDir, "context.h"), path.Join(InterfacesDir, "startup.h"), path.Join(InterfacesDir, "startup.cpp")}},
		Bindings:      map[string]GlueEntry{},
		Pairings:      map[string]GlueEntry{},
		Setup:         map[string]GlueEntry{},
	}
	for _, rid := range Recipes(m, man) {
		r := m.Recipes[rid]
		var files []string
		for _, gl := range r.ReferenceGlue {
			if gl.Role != "build-fragment" {
				files = append(files, path.Join(GlueDirFor(r), path.Base(gl.Path)))
			}
		}
		entry := GlueEntry{Files: files}
		if files == nil {
			var reasons []string
			for _, na := range r.NotApplicable {
				reasons = append(reasons, na.Reason)
			}
			entry.NotApplicable = strings.Join(reasons, " ")
			if entry.NotApplicable == "" {
				entry.NotApplicable = "The " + rid + " recipe ships no glue."
			}
		}
		switch r.Kind {
		case "binding":
			g.Bindings[r.Subject.Library+"/"+r.Subject.Contract] = entry
		case "pairing":
			g.Pairings[strings.Join(r.Subject.Libraries, "/")] = entry
		case "library":
			g.Setup[r.Subject.Library] = entry
		}
	}
	return g
}

// Resolve pins every library in the manifest at its newest version in the
// map, records the contracts' versions, and returns the lock. It is a pure
// function of the map and the manifest.
func Resolve(m *mapmodel.Map, man *Manifest) (*Lock, []Failure) {
	var fails []Failure
	lock := &Lock{Format: "automap-lock/1", MapRelease: man.MapRelease, Pins: map[string]Pin{}, Contracts: map[string]string{}}
	var buildBackend struct {
		Vcpkg struct {
			Baseline        string `json:"baseline"`
			AutomapRegistry *struct {
				Commit string `json:"commit"`
			} `json:"automap_registry"`
		} `json:"vcpkg"`
	}
	json.Unmarshal(m.Maps[0].BuildBackend, &buildBackend)

	pinFor := func(lib *mapmodel.Library, v *mapmodel.Version, variantName string) (Pin, bool) {
		variant := v.Variant(variantName)
		if variant == nil {
			fails = append(fails, failure("UNKNOWN_VARIANT("+lib.ID+":"+variantName+")", "%s %s has no variant %s.", lib.ID, v.Version, variantName))
			return Pin{}, false
		}
		pin := Pin{Version: v.Version, Variant: variantName, License: lib.License.SPDX}
		if lib.Source != nil && lib.Source.Kind == "path" {
			pin.Path = lib.Source.Path
			return pin, true
		}
		src := v.Pin
		if variant.Pin != nil {
			src = variant.Pin
		}
		pin.Commit, pin.ContentHash = src.Commit, src.ContentHash
		kind := v.BuildBackend.Vcpkg.Registry
		if variant.Registry != "" {
			kind = variant.Registry
		}
		switch kind {
		case "curated":
			pin.Registry = &Registry{Kind: "vcpkg-curated", Commit: buildBackend.Vcpkg.Baseline}
		case "automap":
			if buildBackend.Vcpkg.AutomapRegistry == nil {
				fails = append(fails, failure("MAP_INVALID("+m.Maps[0].Dir+"/map.json)", "%s comes from the Automap registry, but map.json names no automap_registry.", lib.ID))
				return Pin{}, false
			}
			pin.Registry = &Registry{Kind: "automap", Commit: buildBackend.Vcpkg.AutomapRegistry.Commit}
		default:
			fails = append(fails, failure("UNSUPPORTED("+lib.ID+")", "%s comes from a private registry, which start doesn't support yet.", lib.ID))
			return Pin{}, false
		}
		return pin, true
	}

	for _, id := range man.Libraries() {
		lib := m.Libraries[id]
		if lib == nil {
			fails = append(fails, failure("UNKNOWN_ENTRY(library "+id+")", "The map has no library %s.", id))
			continue
		}
		v := lib.Newest()
		choice := man.Variants[id]
		pin, ok := pinFor(lib, v, choice.Default)
		if !ok {
			continue
		}
		for _, p := range man.Platforms {
			if name := choice.On(p); name != choice.Default {
				pp, ok := pinFor(lib, v, name)
				if !ok {
					continue
				}
				if pin.Platforms == nil {
					pin.Platforms = map[string]Pin{}
				}
				pin.Platforms[p] = pp
			}
		}
		lock.Pins[id] = pin
	}
	for _, c := range Contracts(m, man) {
		lock.Contracts[c] = m.LatestContract(c).Version
	}
	if len(fails) > 0 {
		return nil, fails
	}
	return lock, nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// GlueEntryFor is the manifest's glue entry for a recipe.
func GlueEntryFor(man *Manifest, r *mapmodel.Recipe) GlueEntry {
	switch r.Kind {
	case "binding":
		return man.Glue.Bindings[r.Subject.Library+"/"+r.Subject.Contract]
	case "pairing":
		return man.Glue.Pairings[strings.Join(r.Subject.Libraries, "/")]
	default:
		return man.Glue.Setup[r.Subject.Library]
	}
}
