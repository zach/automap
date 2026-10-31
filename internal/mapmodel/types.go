// Package mapmodel loads the public map and private maps, validates every
// file against the published schemas and the cross-file rules, layers
// private maps over the public map, and answers queries (SPEC §4, §7.1).
package mapmodel

import "encoding/json"

// MapMeta is a map's map.json.
type MapMeta struct {
	Format       string          `json:"format"`
	ID           string          `json:"id"`
	Kind         string          `json:"kind"`
	Platforms    []string        `json:"platforms,omitempty"`
	BuildBackend json.RawMessage `json:"build_backend,omitempty"`

	Dir    string `json:"-"` // where the map was read from, as given
	Commit string `json:"-"` // a private map's last commit in its git repository, or ""
}

type Slot struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Examples    []string `json:"examples,omitempty"`
}

type Contract struct {
	Contract        string            `json:"contract"`
	Version         string            `json:"version"`
	Description     string            `json:"description"`
	Phases          []string          `json:"phases,omitempty"`
	Facts           map[string]Fact   `json:"facts"`
	Choices         map[string]Choice `json:"choices"`
	Rules           []Rule            `json:"rules"`
	SharedInterface *SharedInterface  `json:"shared_interface"`
}

type Fact struct {
	Description string          `json:"description"`
	Required    *bool           `json:"required,omitempty"`
	Schema      json.RawMessage `json:"schema"`
}

type Choice struct {
	Description string          `json:"description"`
	Schema      json.RawMessage `json:"schema"`
	Default     json.RawMessage `json:"default"`
}

type Rule struct {
	ID              string   `json:"id"`
	Kind            string   `json:"kind"`
	Text            string   `json:"text"`
	AppliesTo       string   `json:"applies_to"`
	FailureCode     string   `json:"failure_code,omitempty"`
	Procedure       string   `json:"procedure,omitempty"`
	Platforms       []string `json:"platforms,omitempty"`
	Slot            string   `json:"slot,omitempty"`
	SeededViolation *Seed    `json:"seeded_violation,omitempty"`
}

type Seed struct {
	Description string `json:"description"`
	Patch       string `json:"patch"`
}

type SharedInterface struct {
	Files []string `json:"files"`
}

type Library struct {
	ID                 string             `json:"id"`
	Name               string             `json:"name"`
	License            License            `json:"license"`
	ContributionPolicy ContributionPolicy `json:"contribution_policy"`
	Slots              []string           `json:"slots"`
	Source             *Source            `json:"source,omitempty"`
	ConsoleVariants    []ConsoleVariant   `json:"console_variants,omitempty"`
	Versions           []Version          `json:"versions"`
}

type License struct {
	SPDX    string `json:"spdx"`
	Class   string `json:"class"`
	Warning string `json:"warning,omitempty"`
}

type ContributionPolicy struct {
	AIGeneratedContributions string `json:"ai_generated_contributions"`
	Source                   string `json:"source"`
}

type Source struct {
	Kind string `json:"kind"`
	URL  string `json:"url,omitempty"`
	Path string `json:"path,omitempty"`
}

type ConsoleVariant struct {
	PlatformFamily string `json:"platform_family"`
	DistributedBy  string `json:"distributed_by"`
}

// Facts are a library's facts keyed by contract, then by fact id.
type Facts map[string]map[string]any

type Version struct {
	Version      string        `json:"version"`
	Released     string        `json:"released"`
	SecurityFix  bool          `json:"security_fix,omitempty"`
	Pin          *Pin          `json:"pin,omitempty"`
	BuildBackend *BuildBackend `json:"build_backend,omitempty"`
	CMake        CMake         `json:"cmake"`
	Bundled      []Bundled     `json:"bundled,omitempty"`
	Variants     []Variant     `json:"variants"`
	Facts        Facts         `json:"facts"`
}

type Pin struct {
	Tag         string `json:"tag,omitempty"`
	Commit      string `json:"commit"`
	ContentHash string `json:"content_hash"`
}

type BuildBackend struct {
	Vcpkg VcpkgPort `json:"vcpkg"`
}

type VcpkgPort struct {
	Port        string `json:"port"`
	Registry    string `json:"registry"`
	PortVersion *int   `json:"port_version,omitempty"`
}

type CMake struct {
	Package string   `json:"package"`
	Targets []string `json:"targets"`
}

type Bundled struct {
	Library string `json:"library"`
	Version string `json:"version"`
}

type Variant struct {
	Name         string            `json:"name"`
	BuildOptions map[string]string `json:"build_options"`
	Description  string            `json:"description,omitempty"`
	Facts        Facts             `json:"facts,omitempty"`
	Registry     string            `json:"registry,omitempty"`
	Pin          *Pin              `json:"pin,omitempty"`
}

type Recipe struct {
	Kind            string               `json:"kind"`
	Subject         Subject              `json:"subject"`
	AppliesTo       map[string]AppliesTo `json:"applies_to"`
	ReferenceGlue   []Glue               `json:"reference_glue"`
	EntryPoints     []EntryPoint         `json:"entry_points,omitempty"`
	Rules           []Rule               `json:"rules,omitempty"`
	Ordering        []Ordering           `json:"ordering,omitempty"`
	Probes          []Probe              `json:"probes"`
	NotApplicable   []NotApplicable      `json:"not_applicable,omitempty"`
	AdaptationNotes []string             `json:"adaptation_notes"`
}

type Subject struct {
	Library         string   `json:"library,omitempty"`
	Contract        string   `json:"contract,omitempty"`
	ContractVersion string   `json:"contract_version,omitempty"`
	Libraries       []string `json:"libraries,omitempty"`
}

type AppliesTo struct {
	Versions []string `json:"versions"`
	Variants []string `json:"variants"`
}

type Glue struct {
	Path string `json:"path"`
	Role string `json:"role"`
}

type EntryPoint struct {
	Phase    string         `json:"phase"`
	When     string         `json:"when"`
	Call     string         `json:"call"`
	After    []string       `json:"after,omitempty"`
	Provides []ContextField `json:"provides,omitempty"`
	Reads    []string       `json:"reads,omitempty"`
	Steps    []string       `json:"steps,omitempty"`
}

type ContextField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Ordering struct {
	Before string `json:"before"`
	After  string `json:"after"`
	When   string `json:"when"`
}

type Probe struct {
	Rule            string `json:"rule"`
	Check           string `json:"check"`
	SeededViolation Seed   `json:"seeded_violation"`
}

type NotApplicable struct {
	Rule   string `json:"rule"`
	Reason string `json:"reason"`
}

// Evidence is the part of an evidence statement the map model checks.
type Evidence struct {
	Predicate struct {
		SubjectKind   string `json:"subject_kind"`
		SubjectDetail struct {
			Pin      EvidencePin `json:"pin"`
			Contract string      `json:"contract"`
			LockHash string      `json:"lock_hash"`
		} `json:"subject_detail"`
		Platform   string           `json:"platform"`
		Level      string           `json:"level"`
		MapRelease string           `json:"map_release"`
		Results    []EvidenceResult `json:"results"`
		Verifier   struct {
			Kind   string `json:"kind"`
			Signed bool   `json:"signed"`
		} `json:"verifier"`
	} `json:"predicate"`
}

type EvidencePin struct {
	Library string `json:"library"`
	Version string `json:"version"`
	Variant string `json:"variant"`
}

type EvidenceResult struct {
	Rule    string `json:"rule"`
	Library string `json:"library,omitempty"`
	Status  string `json:"status"`
}
