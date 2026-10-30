package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/zach/automap/internal/composition"
	"github.com/zach/automap/internal/mapmodel"
)

// Seed is a seeded violation that applies to a composition (ADR-0071).
type Seed struct {
	Rule        string
	Library     string // the library a library rule's seed is about, or ""
	Owner       string // the recipe or contract that carries it
	Subject     string // the library or pairing whose recipe carries it, or ""
	Description string
	Patch       string            // the patch file on disk
	Targets     map[string]string // a file the patch names -> the composition's file
}

// Seeds lists every seeded violation that applies to the composition: each
// probe's in its recipes, and each composition rule's in its contracts.
func Seeds(m *mapmodel.Map, man *composition.Manifest) []Seed {
	var out []Seed
	for _, rid := range composition.Recipes(m, man) {
		r := m.Recipes[rid]
		entry := composition.GlueEntryFor(man, r)
		for _, p := range r.Probes {
			seed := Seed{Rule: p.Rule, Owner: rid, Description: p.SeededViolation.Description,
				Patch: m.RecipeFile(rid, p.SeededViolation.Patch), Targets: map[string]string{}}
			seed.Subject = r.Subject.Library
			if r.Kind == "pairing" {
				seed.Subject = rid
			} else if isLibraryRule(m, p.Rule) {
				seed.Library = r.Subject.Library
			}
			recipeDir := strings.Split(p.SeededViolation.Patch, "/")[0]
			for _, g := range r.ReferenceGlue {
				rel := strings.TrimPrefix(g.Path, recipeDir+"/")
				if g.Role == "build-fragment" {
					seed.Targets[rel] = path.Join("cmake/automap", r.Subject.Library+".cmake")
					continue
				}
				for _, f := range entry.Files {
					if path.Base(f) == path.Base(g.Path) {
						seed.Targets[rel] = f
					}
				}
			}
			out = append(out, seed)
		}
	}
	includeRoot := path.Dir(man.Glue.Interfaces)
	for _, cid := range composition.Contracts(m, man) {
		c := m.LatestContract(cid)
		for _, rule := range c.Rules {
			if rule.SeededViolation == nil {
				continue
			}
			seed := Seed{Rule: rule.ID, Owner: "contract " + cid, Description: rule.SeededViolation.Description,
				Patch: m.ContractFile(c, strings.TrimPrefix(rule.SeededViolation.Patch, c.Version+"/")), Targets: map[string]string{}}
			if c.SharedInterface != nil {
				for _, f := range c.SharedInterface.Files {
					seed.Targets[f] = path.Join(includeRoot, f)
				}
			}
			out = append(out, seed)
		}
	}
	return out
}

func isLibraryRule(m *mapmodel.Map, rule string) bool {
	cid, _, _ := strings.Cut(rule, ".")
	if c := m.LatestContract(cid); c != nil {
		for _, r := range c.Rules {
			if r.ID == rule {
				return r.AppliesTo == "library"
			}
		}
	}
	return false
}

// ApplySeed applies a seed to the composition at root: a unified diff to
// the files it maps to, or a JSON Patch to the manifest.
func ApplySeed(root string, s Seed) error {
	raw, err := os.ReadFile(s.Patch)
	if err != nil {
		return err
	}
	if strings.HasSuffix(s.Patch, ".json-patch") {
		return applyJSONPatch(filepath.Join(root, composition.ManifestFile), raw)
	}
	return applyDiff(root, s.Targets, string(raw))
}

// applyDiff applies a unified diff whose file names are mapped through
// targets to files under root.
func applyDiff(root string, targets map[string]string, diff string) error {
	lines := strings.SplitAfter(diff, "\n")
	for i := 0; i < len(lines); i++ {
		if !strings.HasPrefix(lines[i], "+++ ") {
			continue
		}
		name := strings.TrimSpace(strings.TrimPrefix(lines[i], "+++ "))
		name = strings.TrimPrefix(name, "b/")
		target, ok := targets[name]
		if !ok {
			return fmt.Errorf("the patch changes %s, which this composition doesn't have", name)
		}
		file := filepath.Join(root, filepath.FromSlash(target))
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		text := string(content)
		at := 0
		for i+1 < len(lines) && strings.HasPrefix(lines[i+1], "@@") {
			i++
			var old, new strings.Builder
			for i+1 < len(lines) && len(lines[i+1]) > 0 && strings.ContainsAny(lines[i+1][:1], " +-") && !strings.HasPrefix(lines[i+1], "--- ") {
				i++
				l := lines[i]
				switch l[0] {
				case ' ':
					old.WriteString(l[1:])
					new.WriteString(l[1:])
				case '-':
					old.WriteString(l[1:])
				case '+':
					new.WriteString(l[1:])
				}
			}
			idx := strings.Index(text[at:], old.String())
			if idx < 0 {
				return fmt.Errorf("the patch doesn't apply to %s: its context isn't there", target)
			}
			text = text[:at+idx] + new.String() + text[at+idx+old.Len():]
			at += idx + new.Len()
		}
		if err := os.WriteFile(file, []byte(text), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// applyJSONPatch applies an RFC 6902 patch's add, replace and remove
// operations to a JSON file.
func applyJSONPatch(file string, patch []byte) error {
	raw, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return err
	}
	var ops []struct {
		Op    string `json:"op"`
		Path  string `json:"path"`
		Value any    `json:"value"`
	}
	if err := json.Unmarshal(patch, &ops); err != nil {
		return err
	}
	for _, op := range ops {
		if doc, err = applyOp(doc, strings.Split(strings.TrimPrefix(op.Path, "/"), "/"), op.Op, op.Value); err != nil {
			return fmt.Errorf("%s %s: %v", op.Op, op.Path, err)
		}
	}
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, append(out, '\n'), 0o644)
}

func applyOp(doc any, path []string, op string, value any) (any, error) {
	key := strings.NewReplacer("~1", "/", "~0", "~").Replace(path[0])
	last := len(path) == 1
	switch d := doc.(type) {
	case map[string]any:
		if !last {
			child, err := applyOp(d[key], path[1:], op, value)
			d[key] = child
			return d, err
		}
		if op == "remove" {
			delete(d, key)
		} else {
			d[key] = value
		}
		return d, nil
	case []any:
		i := len(d)
		if key != "-" {
			n, err := strconv.Atoi(key)
			if err != nil || n < 0 || n > len(d) {
				return d, fmt.Errorf("bad index %s", key)
			}
			i = n
		}
		if !last {
			if i >= len(d) {
				return d, fmt.Errorf("bad index %s", key)
			}
			child, err := applyOp(d[i], path[1:], op, value)
			d[i] = child
			return d, err
		}
		switch op {
		case "add":
			return append(d[:i], append([]any{value}, d[i:]...)...), nil
		case "replace":
			d[i] = value
			return d, nil
		case "remove":
			return append(d[:i], d[i+1:]...), nil
		}
	}
	return doc, fmt.Errorf("can't %s here", op)
}
