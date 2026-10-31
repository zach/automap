package composition

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/zach/automap/internal/mapmodel"
)

// ExtendRequest is what extend is asked to add: one library, optionally a
// variant, and optionally the slots it should fill.
type ExtendRequest struct {
	Library    string
	Variant    string
	Slots      []string
	MapRelease string
}

// Extend adds one library to a composition: it fills the slots asked for
// (or every slot the library fills), chooses the variant, adds the
// contract defaults for any contract the composition starts using, and
// lays out the new recipes' glue. The developer's existing glue entries,
// choices and startup stay as they are. It is a pure function of the map,
// the manifest and the request.
func Extend(m *mapmodel.Map, base *Manifest, req ExtendRequest) (*Manifest, []Failure) {
	lib := m.Libraries[req.Library]
	if lib == nil {
		return nil, []Failure{failure("UNKNOWN_ENTRY(library "+req.Library+")", "The map has no library %s; automap find lists libraries.", req.Library)}
	}
	if _, present := base.Variants[req.Library]; present {
		return nil, []Failure{failure("ALREADY_PRESENT("+req.Library+")", "%s is already in the composition; use automap upgrade to change its version or variant.", req.Library)}
	}
	v := lib.Newest()
	variant := req.Variant
	if variant == "" {
		variant = v.Variants[0].Name
	}
	if v.Variant(variant) == nil {
		var names []string
		for _, x := range v.Variants {
			names = append(names, x.Name)
		}
		return nil, []Failure{failure("UNKNOWN_VARIANT("+req.Library+":"+variant+")", "%s %s has no variant %s; its variants are %s.", req.Library, v.Version, variant, strings.Join(names, ", "))}
	}
	slots := req.Slots
	if len(slots) == 0 {
		slots = lib.Slots
	}
	var fails []Failure
	for _, slot := range slots {
		switch {
		case m.Slots[slot] == nil:
			fails = append(fails, failure("UNKNOWN_ENTRY(slot "+slot+")", "The map has no slot %s.", slot))
		case !slices.Contains(lib.Slots, slot):
			fails = append(fails, failure("SLOT_MISMATCH("+slot+"="+req.Library+")", "%s doesn't fill %s; it fills %s.", req.Library, slot, strings.Join(lib.Slots, ", ")))
		case base.Slots[slot] != "":
			fails = append(fails, failure("SLOT_TAKEN("+slot+")", "%s already fills %s; name the slots %s should fill with --slot.", base.Slots[slot], slot, req.Library))
		}
	}
	if fails != nil {
		return nil, fails
	}

	man := copyManifest(base)
	man.MapRelease = req.MapRelease
	for _, slot := range slots {
		man.Slots[slot] = req.Library
	}
	man.Variants[req.Library] = VariantChoice{Default: variant}
	addDefaults(m, man)
	return man, nil
}

// Upgrade moves a composition to another map release: the manifest records
// the release, any contract the composition starts using brings its choice
// defaults, and any recipe the composition gains has its glue laid out.
// The developer's slots, variants, choices and glue entries stay as they
// are. m is the new release's map, with the libraries that don't move held
// at their versions. It is a pure function.
func Upgrade(m *mapmodel.Map, base *Manifest, mapRelease string) *Manifest {
	man := copyManifest(base)
	man.MapRelease = mapRelease
	addDefaults(m, man)
	return man
}

// addDefaults adds the choice defaults of every contract the composition
// uses that it has no choice for yet, and lays out the glue of every
// recipe it has no glue entry for.
func addDefaults(m *mapmodel.Map, man *Manifest) {
	for _, cid := range Contracts(m, man) {
		c := m.LatestContract(cid)
		for id, ch := range c.Choices {
			if _, set := man.Choices[cid][id]; set {
				continue
			}
			if man.Choices[cid] == nil {
				man.Choices[cid] = map[string]any{}
			}
			var value any
			json.Unmarshal(ch.Default, &value)
			man.Choices[cid][id] = value
		}
	}
	slotChoices(m, man)
	fresh := defaultGlue(m, man)
	for key, entry := range fresh.Bindings {
		if _, kept := man.Glue.Bindings[key]; !kept {
			man.Glue.Bindings[key] = entry
		}
	}
	for key, entry := range fresh.Pairings {
		if _, kept := man.Glue.Pairings[key]; !kept {
			man.Glue.Pairings[key] = entry
		}
	}
	for key, entry := range fresh.Setup {
		if _, kept := man.Glue.Setup[key]; !kept {
			man.Glue.Setup[key] = entry
		}
	}
}

// copyManifest deep-copies a manifest through JSON.
func copyManifest(m *Manifest) *Manifest {
	raw, _ := json.Marshal(m)
	var out Manifest
	json.Unmarshal(raw, &out)
	for _, mp := range []*map[string]GlueEntry{&out.Glue.Bindings, &out.Glue.Pairings, &out.Glue.Setup} {
		if *mp == nil {
			*mp = map[string]GlueEntry{}
		}
	}
	if out.Choices == nil {
		out.Choices = map[string]map[string]any{}
	}
	return &out
}
