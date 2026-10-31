package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/zach/automap/internal/composition"
)

// extend adds one library to the composition in the current directory
// (SPEC §5.4).
func extend(a *args, r *result) int {
	if len(a.positional) != 1 {
		r.Summary = "Usage error: extend takes one library."
		r.fail("USAGE", "automap extend <library>[:<variant>] [--slot <slot>]...")
		return exitUsage
	}
	if _, err := os.Stat(manifestFile); os.IsNotExist(err) {
		r.Summary = "Nothing to extend: no " + manifestFile + " in the current directory."
		r.fail("NO_MANIFEST", "No composition here. Run automap start to create one, or automap adopt to adopt an existing engine.")
		return exitUsage
	}
	m, base, _, status := loadComposition(a, r)
	if m == nil {
		return status
	}
	lib, variant, _ := strings.Cut(a.positional[0], ":")
	man, fails := composition.Extend(m, base, composition.ExtendRequest{
		Library: lib, Variant: variant, Slots: a.values["slot"], MapRelease: mapRelease(m.PublicDir()),
	})
	if fails != nil {
		return failWith(r, fails, "The library can't be added as asked.", exitUsage)
	}
	if l := m.Libraries[lib]; l.License.Class != "permissive" {
		reason := l.License.Warning
		if reason == "" {
			reason = fmt.Sprintf("%s is %s (%s).", lib, l.License.Class, l.License.SPDX)
		}
		r.Warnings = append(r.Warnings, warning{Code: "LICENSE_WARNING(" + lib + ")", Message: reason})
	}
	lock, fails := composition.Resolve(m, man)
	if fails != nil {
		return failWith(r, fails, "The composition can't be resolved against the map.", exitUsage)
	}
	plan, fails := composition.PlanStartup(m, man)
	if fails != nil {
		return failWith(r, fails, "The composition's startup can't be generated.", exitFailed)
	}
	files, fails := composition.Files(m, man, lock, plan, composition.FileOptions{RegistryURL: registryURL(m), Vendored: composition.ReadVendoredPorts(".")})
	if fails != nil {
		return failWith(r, fails, "The composition's files can't be written.", exitFailed)
	}
	if man.Glue.Startup.DeveloperOwned != nil {
		for _, f := range []string{"context.h", "startup.h", "startup.cpp"} {
			delete(files, man.Glue.Interfaces+"/"+f)
		}
	}
	// The CLI rewrites what it generates; glue, build fragments and shared
	// interfaces already in the repo are the developer's.
	generated := map[string]bool{
		composition.ManifestFile: true, composition.LockFile: true,
		"vcpkg.json": true, "vcpkg-configuration.json": true,
		"cmake/automap/automap.cmake": true, man.Glue.ChoicesHeader: true,
	}
	for _, f := range man.Glue.Startup.Files {
		generated[f] = true
	}
	keep := func(p string) bool {
		return !generated[p] && p != "AGENTS.md" && p != "CLAUDE.md" && !strings.HasSuffix(p, "/automap/SKILL.md")
	}
	pin := lock.Pins[lib]
	var slots []string
	for slot, l := range man.Slots {
		if l == lib {
			slots = append(slots, slot)
		}
	}
	sortStrings(slots)
	done := fmt.Sprintf("Added %s %s (%s) to %s", lib, pin.Version, pin.Variant, strings.Join(slots, ", "))
	return writeAndVerify(a, m, man, files, keep, done, r)
}
