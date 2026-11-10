package verify

import (
	"fmt"
	"os/exec"
	"slices"
	"sort"
	"strings"

	"github.com/zach/automap/internal/composition"
	"github.com/zach/automap/internal/mapmodel"
)

// SystemPackages lists the system packages the composition's libraries
// state for a platform and package manager (their platform facts'
// system_packages): those the build needs, and those only a headless run
// needs, each sorted and without repeats.
func SystemPackages(m *mapmodel.Map, man *composition.Manifest, platform, manager string) (build, headless []string) {
	collect := func(key string) []string {
		var out []string
		for _, lib := range man.Libraries() {
			l := m.Libraries[lib]
			if l == nil {
				continue
			}
			byPlatform, _ := l.Newest().Facts["platform"]["system_packages"].(map[string]any)
			byManager, _ := byPlatform[platform].(map[string]any)
			names, _ := byManager[key].([]any)
			for _, n := range names {
				if name := fmt.Sprint(n); !slices.Contains(out, name) {
					out = append(out, name)
				}
			}
		}
		sort.Strings(out)
		return out
	}
	return collect(manager), collect(manager + "_headless")
}

// installedApt lists the packages dpkg has installed on this machine, and
// the names they provide, or nil when the machine has no dpkg-query: then
// nothing can be said about its packages.
func installedApt() map[string]bool {
	out, err := exec.Command("dpkg-query", "-W", "-f", `${db:Status-Abbrev}\t${Package}\t${Provides}\n`).Output()
	if err != nil {
		return nil
	}
	installed := map[string]bool{}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) < 2 || !strings.HasPrefix(fields[0], "ii") {
			continue
		}
		// A package can be named with its architecture, such as zlib1g:amd64.
		name, _, _ := strings.Cut(fields[1], ":")
		installed[name] = true
		if len(fields) > 2 {
			// "libfoo-dev (= 1.2), virtual-foo"
			for _, p := range strings.Split(fields[2], ",") {
				if name := strings.Fields(p); len(name) > 0 {
					installed[name[0]] = true
				}
			}
		}
	}
	return installed
}

// missingPackages checks the composition's system packages against what
// this machine has installed. It returns a failure for each package the
// build needs and the machine lacks, and a warning for each a headless
// run needs. A machine without dpkg-query, or a platform that isn't
// Linux, is not checked.
func missingPackages(m *mapmodel.Map, man *composition.Manifest, platform string) (failures, warnings []Failure) {
	if !strings.HasPrefix(platform, "linux-") {
		return nil, nil
	}
	build, headless := SystemPackages(m, man, platform, "apt")
	if len(build)+len(headless) == 0 {
		return nil, nil
	}
	installed := installedApt()
	if installed == nil {
		return nil, nil
	}
	lacking := func(names []string) []string {
		var out []string
		for _, n := range names {
			if !installed[n] {
				out = append(out, n)
			}
		}
		return out
	}
	missing := lacking(build)
	for _, name := range missing {
		failures = append(failures, Failure{Code: "SYSTEM_PACKAGE_MISSING(" + name + ")",
			Message: "The composition's libraries need the system package " + name + " to build, and it isn't installed. Install what is missing: sudo apt-get install -y " + strings.Join(missing, " ")})
	}
	absent := lacking(headless)
	for _, name := range absent {
		warnings = append(warnings, Failure{Code: "SYSTEM_PACKAGE_MISSING(" + name + ")",
			Message: "A headless run needs the system package " + name + " on a machine without a display or a GPU driver, and it isn't installed. If the smoke run fails to start, install it: sudo apt-get install -y " + strings.Join(absent, " ")})
	}
	return failures, warnings
}
