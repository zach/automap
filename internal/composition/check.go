package composition

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/zach/automap/internal/mapmodel"
	"github.com/zach/automap/schemas"
)

// Check checks a composition the CLI reads against the map: the manifest's
// slots, variants, choices and glue, and the lock's pins (the composition
// rules of the spec's example checker). Problems with the manifest are
// MANIFEST_INVALID; a lock that disagrees with the manifest or the map is
// LOCK_MISMATCH(<library>).
func Check(m *mapmodel.Map, man *Manifest, lock *Lock) []Failure {
	var fails []Failure
	manifest := func(format string, args ...any) {
		fails = append(fails, failure("MANIFEST_INVALID", format, args...))
	}
	used := man.Libraries()
	for _, slot := range sortedKeys(man.Slots) {
		lib := man.Slots[slot]
		l := m.Libraries[lib]
		switch {
		case l == nil:
			manifest("slot %s: the map has no library %s", slot, lib)
		case !slices.Contains(l.Slots, slot):
			manifest("%s doesn't fill slot %s", lib, slot)
		}
	}
	if !slices.Equal(sortedKeys(man.Variants), used) {
		manifest("variants must name exactly the libraries in slots: %v", used)
	}
	for _, cid := range sortedKeys(man.Choices) {
		c := m.LatestContract(cid)
		if c == nil {
			manifest("choices: the map has no %s contract", cid)
			continue
		}
		props := map[string]any{}
		for id, ch := range c.Choices {
			var schema any
			json.Unmarshal(ch.Schema, &schema)
			props[id] = schema
		}
		schema := map[string]any{"type": "object", "additionalProperties": false, "properties": props}
		v, _ := schemas.ValidateAgainst(schema, toAny(man.Choices[cid]))
		for _, msg := range v {
			manifest("choices/%s: %s", cid, msg)
		}
	}
	for _, cid := range sortedKeys(man.Choices) {
		for _, id := range sortedKeys(man.Choices[cid]) {
			value, _ := man.Choices[cid][id].(string)
			if slot, ok := strings.CutSuffix(value, "-slot"); ok && man.Slots[slot] == "" {
				manifest("choices/%s/%s %s needs a library in the %s slot", cid, id, value, slot)
			}
		}
	}
	for _, entries := range []map[string]GlueEntry{man.Glue.Bindings, man.Glue.Pairings} {
		for _, key := range sortedKeys(entries) {
			for _, part := range strings.Split(key, "/") {
				if !slices.Contains(used, part) && m.LatestContract(part) == nil {
					manifest("glue entry %s names %s, which isn't in this composition", key, part)
				}
			}
		}
	}

	mismatch := func(lib, format string, args ...any) {
		fails = append(fails, failure("LOCK_MISMATCH("+lib+")", format, args...))
	}
	if !slices.Equal(sortedKeys(lock.Pins), used) {
		fails = append(fails, failure("LOCK_MISMATCH", "the lock pins %v but the manifest's slots hold %v; re-resolve with automap extend or upgrade", sortedKeys(lock.Pins), used))
	}
	if lock.MapRelease != man.MapRelease {
		fails = append(fails, failure("LOCK_MISMATCH", "the lock's map_release %s differs from the manifest's %s", lock.MapRelease, man.MapRelease))
	}
	for _, lib := range used {
		pin, ok := lock.Pins[lib]
		l := m.Libraries[lib]
		if !ok || l == nil {
			continue
		}
		choice := man.Variants[lib]
		if pin.Variant != choice.Default {
			mismatch(lib, "%s is pinned as variant %s, but the manifest asks for %s", lib, pin.Variant, choice.Default)
		}
		for _, p := range man.Platforms {
			ppin := pin
			if pp, ok := pin.Platforms[p]; ok {
				ppin = pp
			}
			if ppin.Variant != choice.On(p) {
				mismatch(lib, "%s on %s is pinned as variant %s, but the manifest asks for %s", lib, p, ppin.Variant, choice.On(p))
			}
			if msg := pinAgainstMap(m, l, ppin); msg != "" {
				mismatch(lib, "%s", msg)
			}
		}
	}
	for _, c := range sortedKeys(lock.Contracts) {
		found := false
		for _, v := range m.Contracts[c] {
			found = found || v.Version == lock.Contracts[c]
		}
		if !found {
			fails = append(fails, failure("LOCK_MISMATCH", "the lock names contract %s %s, which the map doesn't have", c, lock.Contracts[c]))
		}
	}
	return fails
}

// pinAgainstMap says how a pin differs from what the map records for its
// library version and variant, or "" when it matches.
func pinAgainstMap(m *mapmodel.Map, l *mapmodel.Library, pin Pin) string {
	var v *mapmodel.Version
	for i := range l.Versions {
		if l.Versions[i].Version == pin.Version {
			v = &l.Versions[i]
		}
	}
	if v == nil || v.Variant(pin.Variant) == nil {
		return fmt.Sprintf("%s %s (%s) is not in the map", l.ID, pin.Version, pin.Variant)
	}
	if pin.License != l.License.SPDX {
		return fmt.Sprintf("%s's license %s differs from the map's %s", l.ID, pin.License, l.License.SPDX)
	}
	if l.Source != nil && l.Source.Kind == "path" {
		if pin.Path != l.Source.Path {
			return fmt.Sprintf("%s is pinned at path %q, but the map says %q", l.ID, pin.Path, l.Source.Path)
		}
		return ""
	}
	src := v.Pin
	if variant := v.Variant(pin.Variant); variant.Pin != nil {
		src = variant.Pin
	}
	if src != nil && (pin.Commit != src.Commit || pin.ContentHash != src.ContentHash) {
		return fmt.Sprintf("%s %s's pinned commit or content hash differs from the map", l.ID, pin.Version)
	}
	if pin.Registry != nil && pin.Registry.Kind == "vcpkg-curated" && pin.Registry.Commit != vcpkgBaseline(m) {
		return fmt.Sprintf("%s's curated registry commit differs from the map's vcpkg baseline", l.ID)
	}
	return ""
}

// toAny round-trips v through JSON to the generic form schema validation
// expects.
func toAny(v any) any {
	raw, _ := json.Marshal(v)
	var out any
	json.Unmarshal(raw, &out)
	return out
}
