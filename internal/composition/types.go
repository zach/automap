// Package composition holds a composition's manifest and lock, resolves a
// manifest against the map, plans its startup, and writes the files a
// composition is made of (SPEC §2.3, §4.7, §4.8).
package composition

import (
	"encoding/json"
	"fmt"
	"sort"
)

// ManifestFile and LockFile are where a composition's manifest and lock
// live, at the root of the developer's repo.
const (
	ManifestFile = "automap.manifest.json"
	LockFile     = "automap.lock.json"
)

// Failure is one reason the CLI can't do what was asked, with a failure
// code (ADR-0041).
type Failure struct {
	Code    string
	Message string
}

func failure(code, format string, args ...any) Failure {
	return Failure{Code: code, Message: fmt.Sprintf(format, args...)}
}

// Manifest is automap.manifest.json.
type Manifest struct {
	Schema       string                    `json:"$schema"`
	Format       string                    `json:"format"`
	MapRelease   string                    `json:"map_release"`
	BuildBackend string                    `json:"build_backend"`
	Platforms    []string                  `json:"platforms"`
	Executable   Executable                `json:"executable"`
	PrivateMaps  []PrivateMap              `json:"private_maps,omitempty"`
	Slots        map[string]string         `json:"slots"`
	Variants     map[string]VariantChoice  `json:"variants"`
	Choices      map[string]map[string]any `json:"choices"`
	Glue         Glue                      `json:"glue"`
}

type Executable struct {
	Target  string   `json:"target"`
	Sources []string `json:"sources,omitempty"`
}

type PrivateMap struct {
	ID        string   `json:"id"`
	Location  string   `json:"location"`
	Platforms []string `json:"platforms,omitempty"`
}

// VariantChoice is a library's variant: one for every platform, or a
// default with the platforms that differ (ADR-0075).
type VariantChoice struct {
	Default     string
	PerPlatform map[string]string
}

// On returns the variant chosen for a platform.
func (v VariantChoice) On(platform string) string {
	if x, ok := v.PerPlatform[platform]; ok {
		return x
	}
	return v.Default
}

func (v VariantChoice) MarshalJSON() ([]byte, error) {
	if len(v.PerPlatform) == 0 {
		return json.Marshal(v.Default)
	}
	m := map[string]string{"default": v.Default}
	for k, x := range v.PerPlatform {
		m[k] = x
	}
	return json.Marshal(m)
}

func (v *VariantChoice) UnmarshalJSON(raw []byte) error {
	if err := json.Unmarshal(raw, &v.Default); err == nil {
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal(raw, &m); err != nil {
		return err
	}
	v.Default = m["default"]
	delete(m, "default")
	v.PerPlatform = m
	return nil
}

// Glue says where a composition's glue lives.
type Glue struct {
	Interfaces    string               `json:"interfaces"`
	ChoicesHeader string               `json:"choices_header"`
	Startup       Startup              `json:"startup"`
	Bindings      map[string]GlueEntry `json:"bindings"`
	Pairings      map[string]GlueEntry `json:"pairings"`
	Setup         map[string]GlueEntry `json:"setup"`
}

// Startup is the startup files the CLI writes, or the developer's own
// startup code (ADR-0080).
type Startup struct {
	Files          []string
	DeveloperOwned []string
}

func (s Startup) MarshalJSON() ([]byte, error) {
	if s.DeveloperOwned != nil {
		return json.Marshal(map[string][]string{"developer_owned": s.DeveloperOwned})
	}
	return json.Marshal(s.Files)
}

func (s *Startup) UnmarshalJSON(raw []byte) error {
	if err := json.Unmarshal(raw, &s.Files); err == nil {
		return nil
	}
	var m struct {
		DeveloperOwned []string `json:"developer_owned"`
	}
	err := json.Unmarshal(raw, &m)
	s.DeveloperOwned = m.DeveloperOwned
	return err
}

// GlueEntry is a binding's, pairing's or library's glue: its files, or why
// it has none (ADR-0074).
type GlueEntry struct {
	Files         []string
	NotApplicable string
}

func (g GlueEntry) MarshalJSON() ([]byte, error) {
	if g.Files == nil {
		return json.Marshal(map[string]string{"not_applicable": g.NotApplicable})
	}
	return json.Marshal(g.Files)
}

func (g *GlueEntry) UnmarshalJSON(raw []byte) error {
	if err := json.Unmarshal(raw, &g.Files); err == nil {
		return nil
	}
	var m struct {
		NotApplicable string `json:"not_applicable"`
	}
	err := json.Unmarshal(raw, &m)
	g.NotApplicable = m.NotApplicable
	return err
}

// Libraries lists the composition's libraries, sorted.
func (m *Manifest) Libraries() []string {
	seen := map[string]bool{}
	var out []string
	for _, lib := range m.Slots {
		if !seen[lib] {
			seen[lib] = true
			out = append(out, lib)
		}
	}
	sort.Strings(out)
	return out
}

// Lock is automap.lock.json.
type Lock struct {
	Format      string                   `json:"format"`
	MapRelease  string                   `json:"map_release"`
	PrivateMaps map[string]PrivateMapPin `json:"private_maps,omitempty"`
	Pins        map[string]Pin           `json:"pins"`
	Contracts   map[string]string        `json:"contracts"`
	LockHash    string                   `json:"lock_hash,omitempty"`
}

type PrivateMapPin struct {
	Commit string `json:"commit"`
}

// Pin is one library's exact version in the lock (ADR-0036).
type Pin struct {
	Version     string         `json:"version"`
	Commit      string         `json:"commit,omitempty"`
	Variant     string         `json:"variant"`
	Registry    *Registry      `json:"registry,omitempty"`
	ContentHash string         `json:"content_hash"`
	License     string         `json:"license"`
	Platforms   map[string]Pin `json:"platforms,omitempty"`
	Path        string         `json:"path,omitempty"`
}

type Registry struct {
	Kind   string `json:"kind"`
	ID     string `json:"id,omitempty"`
	Commit string `json:"commit"`
}
