package composition

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/zach/automap/internal/mapmodel"
)

const generatedNote = "Written by automap from automap.manifest.json. Edit the manifest, not this file."

var nonAlnum = regexp.MustCompile(`[^A-Za-z0-9]+`)

// macroName upper-cases s and replaces every non-alphanumeric run with _.
func macroName(parts ...string) string {
	var out []string
	for _, p := range parts {
		out = append(out, strings.Trim(nonAlnum.ReplaceAllString(strings.ToUpper(p), "_"), "_"))
	}
	return strings.Join(out, "_")
}

// ChoicesHeader is the choices header: every choice of the composition's
// contracts, the manifest's value or the contract's default (SPEC §4.10).
func ChoicesHeader(m *mapmodel.Map, man *Manifest, lock *Lock) string {
	var b strings.Builder
	b.WriteString("// " + generatedNote + "\n#pragma once\n")
	for _, cid := range sortedKeys(lock.Contracts) {
		c := m.LatestContract(cid)
		if c == nil {
			continue
		}
		for _, id := range sortedKeys(c.Choices) {
			var value any
			json.Unmarshal(c.Choices[id].Default, &value)
			if v, ok := man.Choices[cid][id]; ok {
				value = v
			}
			name := "AUTOMAP_" + macroName(cid, id)
			switch v := value.(type) {
			case bool:
				fmt.Fprintf(&b, "#define %s %d\n", name, map[bool]int{false: 0, true: 1}[v])
			case float64:
				fmt.Fprintf(&b, "#define %s %s\n", name, strconv.FormatFloat(v, 'f', -1, 64))
			case string:
				fmt.Fprintf(&b, "#define %s_%s 1\n", name, macroName(v))
			}
		}
	}
	return b.String()
}

// pointee is the identifier a pointer type points to, such as stub_window
// for "stub_window*", or "" for any other type.
var pointerType = regexp.MustCompile(`^\s*((?:[A-Za-z_][A-Za-z0-9_]*::)*[A-Za-z_][A-Za-z0-9_]*)\s*\*\s*$`)

func cIdent(s string) string { return nonAlnum.ReplaceAllString(s, "_") }

// ContextHeader declares the context every entry point takes: one struct
// per owner of the fields recipes provide (ADR-0080). Pointer types are
// forward-declared and qualified, so the header needs no library headers.
func ContextHeader(plan *StartupPlan) string {
	var b strings.Builder
	b.WriteString("// Written by automap from the composition's recipes. Edit the manifest, not this file.\n#pragma once\n")
	var forward []string
	owners := map[string][]mapmodel.ContextField{}
	var order []string
	for _, f := range plan.Fields {
		owner, _, _ := strings.Cut(f.Name, ".")
		if _, ok := owners[owner]; !ok {
			order = append(order, owner)
		}
		owners[owner] = append(owners[owner], f)
		if m := pointerType.FindStringSubmatch(f.Type); m != nil {
			forward = append(forward, m[1])
		}
	}
	sort.Strings(forward)
	if len(forward) > 0 {
		b.WriteString("\n")
		seen := map[string]bool{}
		for _, f := range forward {
			if seen[f] {
				continue
			}
			seen[f] = true
			// A C type is a struct; a namespaced C++ type is declared as a
			// class inside its namespaces.
			if i := strings.LastIndex(f, "::"); i >= 0 {
				var open, close string
				for _, ns := range strings.Split(f[:i], "::") {
					open += "namespace " + ns + " { "
					close += " }"
				}
				fmt.Fprintf(&b, "%sclass %s;%s\n", open, f[i+2:], close)
			} else {
				fmt.Fprintf(&b, "struct %s;\n", f)
			}
		}
	}
	b.WriteString("\nnamespace am {\n\n// The context every entry point takes: the fields recipes provide and read.\nstruct Context {\n")
	for _, owner := range order {
		b.WriteString("    struct {\n")
		for _, f := range owners[owner] {
			_, field, _ := strings.Cut(f.Name, ".")
			typ := f.Type
			init := "{}"
			if m := pointerType.FindStringSubmatch(f.Type); m != nil {
				typ, init = "::"+m[1]+"*", "nullptr"
			}
			fmt.Fprintf(&b, "        %s %s = %s;\n", typ, field, init)
		}
		fmt.Fprintf(&b, "    } %s;\n", cIdent(owner))
	}
	b.WriteString("};\n\n} // namespace am\n")
	return b.String()
}

// StartupHeader declares the startup and shutdown functions.
func StartupHeader() string {
	return `// Written by automap from the composition's recipes. Edit the manifest, not this file.
#pragma once
#include "automap/context.h"

namespace am {

// Calls every init entry point in phase order. When one fails, reports
// INIT_FAILED(<call>) on stderr and returns false.
bool startup(Context& ctx);

// Calls every shutdown entry point in reverse phase order.
void shutdown(Context& ctx);

} // namespace am
`
}

// StartupSource defines the startup and shutdown functions from the plan.
func StartupSource(plan *StartupPlan) string {
	var b strings.Builder
	b.WriteString("// Written by automap from the composition's recipes. Edit the manifest, not this file.\n")
	b.WriteString("#include \"automap/startup.h\"\n\n#include <cstdio>\n\n")
	for _, c := range plan.Init {
		fmt.Fprintf(&b, "bool %s(am::Context& ctx);\n", c.Call)
	}
	for _, c := range plan.Shutdown {
		fmt.Fprintf(&b, "void %s(am::Context& ctx);\n", c.Call)
	}
	b.WriteString("\nnamespace am {\n\nnamespace {\nbool init_failed(const char* call) {\n    std::fprintf(stderr, \"INIT_FAILED(%s)\\n\", call);\n    return false;\n}\n} // namespace\n\n")
	b.WriteString("bool startup(Context& ctx) {\n")
	phase := ""
	for _, c := range plan.Init {
		if c.Phase != phase {
			phase = c.Phase
			fmt.Fprintf(&b, "    // %s\n", phase)
		}
		fmt.Fprintf(&b, "    if (!%s(ctx)) return init_failed(\"%s\");\n", c.Call, c.Call)
	}
	b.WriteString("    return true;\n}\n\nvoid shutdown(Context& ctx) {\n")
	phase = ""
	for _, c := range plan.Shutdown {
		if c.Phase != phase {
			phase = c.Phase
			fmt.Fprintf(&b, "    // %s\n", phase)
		}
		fmt.Fprintf(&b, "    %s(ctx);\n", c.Call)
	}
	b.WriteString("}\n\n} // namespace am\n")
	return b.String()
}

// AutomapCMake is cmake/automap/automap.cmake: it defines the glue target
// from every glue source, adds each in-repo library's directory, and
// includes each library's build fragment. The developer's CMakeLists.txt
// includes it once. The glue target is an object library, so every probe's
// registration links into the executable.
func AutomapCMake(man *Manifest, lock *Lock, sources []string, cxxStandard string) string {
	var b strings.Builder
	b.WriteString("# " + generatedNote + "\n")
	b.WriteString("# The developer's CMakeLists.txt includes this file once.\n\n")
	b.WriteString("set(AUTOMAP_ROOT ${CMAKE_CURRENT_LIST_DIR}/../..)\n")
	fmt.Fprintf(&b, "set(AUTOMAP_CXX_STANDARD %s)\n\n", cxxStandard)
	b.WriteString("add_library(automap_glue OBJECT\n")
	for _, s := range sources {
		fmt.Fprintf(&b, "    ${AUTOMAP_ROOT}/%s\n", s)
	}
	b.WriteString(")\n")
	include := strings.TrimSuffix(man.Glue.Interfaces, "/automap")
	if include == man.Glue.Interfaces {
		include = man.Glue.Interfaces + "/.."
	}
	fmt.Fprintf(&b, "target_include_directories(automap_glue PUBLIC ${AUTOMAP_ROOT}/%s)\n", include)
	b.WriteString("target_compile_features(automap_glue PUBLIC cxx_std_${AUTOMAP_CXX_STANDARD})\n")
	b.WriteString("if(AUTOMAP_PROBE)\n    target_compile_definitions(automap_glue PUBLIC AUTOMAP_PROBE)\nendif()\n")
	// Code that uses a library only when the composition has it tests these.
	b.WriteString("target_compile_definitions(automap_glue PUBLIC")
	for _, lib := range man.Libraries() {
		fmt.Fprintf(&b, " AUTOMAP_LIBRARY_%s=1", macroName(lib))
	}
	b.WriteString(")\n")
	// Glue that keys data by a library's build, such as a rebuild key, reads
	// the variant here.
	b.WriteString("target_compile_definitions(automap_glue PUBLIC")
	for _, lib := range man.Libraries() {
		fmt.Fprintf(&b, ` AUTOMAP_VARIANT_%s="%s"`, macroName(lib), man.Variants[lib].Default)
	}
	b.WriteString(")\n")
	if fmt.Sprint(man.Choices["determinism"]["profile"]) == "on" {
		// The determinism profile: no floating-point contraction in glue.
		b.WriteString("target_compile_options(automap_glue PUBLIC $<IF:$<CXX_COMPILER_ID:MSVC>,/fp:precise,-ffp-contract=off>)\n")
	}
	b.WriteString("\n")
	// In-repo libraries join the build as the developer's own CMake targets,
	// before the build fragments that link them.
	for _, lib := range man.Libraries() {
		if p := lock.Pins[lib].Path; p != "" {
			fmt.Fprintf(&b, "add_subdirectory(${AUTOMAP_ROOT}/%s ${CMAKE_BINARY_DIR}/automap-in-repo/%s)\n", p, lib)
		}
	}
	for _, lib := range man.Libraries() {
		fmt.Fprintf(&b, "include(${CMAKE_CURRENT_LIST_DIR}/%s.cmake)\n", lib)
	}
	return b.String()
}

// vcpkgOverride pins a port's version and port version in vcpkg.json.
type vcpkgOverride struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	PortVersion int    `json:"port-version,omitempty"`
}

// vcpkgOverrides pins each library's port at its version and, for the
// curated port, its port version: a variant from another registry has that
// registry's own.
func vcpkgOverrides(m *mapmodel.Map, man *Manifest, lock *Lock) []vcpkgOverride {
	var out []vcpkgOverride
	for _, lib := range man.Libraries() {
		v := m.Libraries[lib].Newest()
		if v.BuildBackend == nil {
			continue
		}
		variant := v.Variant(lock.Pins[lib].Variant)
		o := vcpkgOverride{Name: v.BuildBackend.Vcpkg.Port, Version: v.Version}
		if pv := v.BuildBackend.Vcpkg.PortVersion; pv != nil && (variant == nil || variant.Registry == "" || variant.Registry == v.BuildBackend.Vcpkg.Registry) {
			o.PortVersion = *pv
		}
		out = append(out, o)
	}
	return out
}

// VcpkgJSON is the vcpkg manifest: each library's port with the features
// its variant enables, an override pinning its version and port version,
// and the curated registry's baseline.
func VcpkgJSON(m *mapmodel.Map, man *Manifest, lock *Lock) []byte {
	type dependency struct {
		Name            string `json:"name"`
		DefaultFeatures *bool  `json:"default-features,omitempty"`
		Features        []any  `json:"features,omitempty"`
	}
	var doc struct {
		Dependencies    []dependency    `json:"dependencies"`
		Overrides       []vcpkgOverride `json:"overrides"`
		BuiltinBaseline string          `json:"builtin-baseline,omitempty"`
	}
	if usesCurated(lock) {
		doc.BuiltinBaseline = vcpkgBaseline(m)
	}
	for _, lib := range man.Libraries() {
		v := m.Libraries[lib].Newest()
		if v.BuildBackend == nil {
			continue
		}
		dep := dependency{Name: v.BuildBackend.Vcpkg.Port}
		if variant := v.Variant(lock.Pins[lib].Variant); variant != nil {
			// A feature may name its platforms as a vcpkg platform
			// expression in brackets, such as "x11[linux]".
			for _, f := range strings.Split(variant.BuildOptions["vcpkg_features"], ",") {
				f = strings.TrimSpace(f)
				name, platform, limited := strings.Cut(strings.TrimSuffix(f, "]"), "[")
				switch {
				case f == "":
				case limited:
					dep.Features = append(dep.Features, map[string]string{"name": name, "platform": platform})
				default:
					dep.Features = append(dep.Features, f)
				}
			}
			if variant.BuildOptions["vcpkg_default_features"] == "false" {
				off := false
				dep.DefaultFeatures = &off
			}
		}
		doc.Dependencies = append(doc.Dependencies, dep)
	}
	doc.Overrides = vcpkgOverrides(m, man, lock)
	return indentJSON(doc)
}

// VendoredPorts is where `automap vendor` copies a composition's ports.
const VendoredPorts = "vendor/automap/ports"

// ReadVendoredPorts lists the ports vendored into the repo at root, each
// with its version and #port-version when it has one.
func ReadVendoredPorts(root string) map[string]string {
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(VendoredPorts)))
	if err != nil {
		return nil
	}
	out := map[string]string{}
	for _, e := range entries {
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(VendoredPorts), e.Name(), "vcpkg.json"))
		if err != nil {
			continue
		}
		var doc struct {
			Version     string `json:"version"`
			Semver      string `json:"version-semver"`
			Date        string `json:"version-date"`
			String      string `json:"version-string"`
			PortVersion int    `json:"port-version"`
		}
		json.Unmarshal(raw, &doc)
		v := doc.Version + doc.Semver + doc.Date + doc.String
		if doc.PortVersion != 0 {
			v += fmt.Sprintf("#%d", doc.PortVersion)
		}
		out[e.Name()] = v
	}
	return out
}

// VcpkgConfiguration is vcpkg-configuration.json: the Automap registry at
// its commit for the ports it supplies, the ports vendored into the repo
// as overlay ports, and the overlay ports of each private map a variant
// comes from. A vendored port the composition pins at another version is
// left out, so the registries supply that version until it is vendored
// again.
func VcpkgConfiguration(m *mapmodel.Map, man *Manifest, lock *Lock, registryURL string, vendored map[string]string) []byte {
	type registry struct {
		Kind       string   `json:"kind"`
		Repository string   `json:"repository"`
		Reference  string   `json:"reference,omitempty"`
		Baseline   string   `json:"baseline"`
		Packages   []string `json:"packages"`
	}
	var doc struct {
		DefaultRegistry *struct{}  `json:"default-registry"`
		Registries      []registry `json:"registries"`
		OverlayPorts    []string   `json:"overlay-ports,omitempty"`
	}
	doc.Registries = []registry{}
	pinned := map[string]string{}
	for _, o := range vcpkgOverrides(m, man, lock) {
		pinned[o.Name] = o.Version
		if o.PortVersion != 0 {
			pinned[o.Name] += fmt.Sprintf("#%d", o.PortVersion)
		}
	}
	for _, port := range sortedKeys(vendored) {
		if want, ok := pinned[port]; !ok || want == vendored[port] {
			doc.OverlayPorts = append(doc.OverlayPorts, path.Join(VendoredPorts, port))
		}
	}
	var automapPorts []string
	var commit string
	for _, lib := range man.Libraries() {
		pin := lock.Pins[lib]
		if pin.Registry != nil && pin.Registry.Kind == "automap" {
			automapPorts = append(automapPorts, m.Libraries[lib].Newest().BuildBackend.Vcpkg.Port)
			commit = pin.Registry.Commit
		}
	}
	// A private variant's port is an overlay port in the private map that
	// adds it (ADR-0047).
	for _, meta := range m.Maps[1:] {
		uses := false
		for _, lib := range man.Libraries() {
			pin := lock.Pins[lib]
			pins := []Pin{pin}
			for _, pp := range pin.Platforms {
				pins = append(pins, pp)
			}
			for _, p := range pins {
				uses = uses || p.Registry != nil && p.Registry.Kind == "private" && p.Registry.ID == meta.ID
			}
		}
		if uses {
			doc.OverlayPorts = append(doc.OverlayPorts, path.Join(filepath.ToSlash(meta.Dir), "ports"))
		}
	}
	if len(automapPorts) > 0 {
		doc.Registries = append(doc.Registries, registry{Kind: "git", Repository: registryURL, Reference: "registry", Baseline: commit, Packages: automapPorts})
	}
	if usesCurated(lock) {
		// vcpkg's curated registry stays the default registry, at the
		// baseline vcpkg.json names.
		raw := indentJSON(struct {
			Registries   []registry `json:"registries"`
			OverlayPorts []string   `json:"overlay-ports,omitempty"`
		}{doc.Registries, doc.OverlayPorts})
		return raw
	}
	return indentJSON(doc)
}

// usesCurated reports whether any library comes from vcpkg's curated
// registry. When none does, the composition names no curated baseline and
// disables vcpkg's default registry.
func usesCurated(lock *Lock) bool {
	for _, pin := range lock.Pins {
		if pin.Registry != nil && pin.Registry.Kind == "vcpkg-curated" {
			return true
		}
		for _, pp := range pin.Platforms {
			if pp.Registry != nil && pp.Registry.Kind == "vcpkg-curated" {
				return true
			}
		}
	}
	return false
}

// AutomapRegistryURL is the Automap registry's URL from the public map's
// map.json.
func AutomapRegistryURL(m *mapmodel.Map) string {
	var bb struct {
		Vcpkg struct {
			AutomapRegistry struct {
				URL string `json:"url"`
			} `json:"automap_registry"`
		} `json:"vcpkg"`
	}
	json.Unmarshal(m.Maps[0].BuildBackend, &bb)
	return bb.Vcpkg.AutomapRegistry.URL
}

func vcpkgBaseline(m *mapmodel.Map) string {
	var bb struct {
		Vcpkg struct {
			Baseline string `json:"baseline"`
		} `json:"vcpkg"`
	}
	json.Unmarshal(m.Maps[0].BuildBackend, &bb)
	return bb.Vcpkg.Baseline
}

func indentJSON(v any) []byte {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return append(raw, '\n')
}
