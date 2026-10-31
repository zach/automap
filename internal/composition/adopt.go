package composition

import (
	"path"
	"regexp"
	"sort"
	"strings"
)

// Project is what adopt reads from an existing CMake project: its
// executables, the libraries it defines under an alias, the packages it
// finds in config mode, and the namespaced targets it links.
type Project struct {
	Executables []ProjectExecutable
	Aliases     map[string]bool
	Packages    map[string]bool
	Linked      map[string]bool
}

// ProjectExecutable is one add_executable: its target and its sources,
// relative to the project's root.
type ProjectExecutable struct {
	Target  string
	Sources []string
}

var (
	cmakeComment = regexp.MustCompile(`#[^\n]*`)
	cmakeCommand = regexp.MustCompile(`(?i)\b(add_executable|add_library|find_package|target_link_libraries)\s*\(([^)]*)\)`)
)

// ScanCMake reads a project's CMake files, by path relative to the
// project's root. It is a pure function.
func ScanCMake(files map[string]string) Project {
	p := Project{Aliases: map[string]bool{}, Packages: map[string]bool{}, Linked: map[string]bool{}}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		dir := path.Dir(name)
		for _, c := range cmakeCommand.FindAllStringSubmatch(cmakeComment.ReplaceAllString(files[name], ""), -1) {
			args := strings.Fields(c[2])
			if len(args) == 0 {
				continue
			}
			switch strings.ToLower(c[1]) {
			case "add_executable":
				e := ProjectExecutable{Target: args[0]}
				for _, a := range args[1:] {
					switch a {
					case "WIN32", "MACOSX_BUNDLE", "EXCLUDE_FROM_ALL":
					default:
						if !strings.Contains(a, "$") {
							e.Sources = append(e.Sources, path.Clean(path.Join(dir, a)))
						}
					}
				}
				p.Executables = append(p.Executables, e)
			case "add_library":
				if len(args) == 3 && args[1] == "ALIAS" {
					p.Aliases[args[0]] = true
				}
			case "find_package":
				for _, a := range args[1:] {
					if a == "CONFIG" || a == "NO_MODULE" {
						p.Packages[args[0]] = true
					}
				}
			case "target_link_libraries":
				for _, a := range args[1:] {
					if strings.Contains(a, "::") {
						p.Linked[a] = true
					}
				}
			}
		}
	}
	return p
}

// LibraryTargets are the linked targets that are libraries: the project's
// own, defined under an alias, and those of packages found in config mode,
// whose namespace is the package's name. Targets CMake itself provides,
// such as Threads::Threads, are neither.
func (p Project) LibraryTargets() []string {
	var out []string
	for t := range p.Linked {
		namespace, _, _ := strings.Cut(t, "::")
		if p.Aliases[t] || p.Packages[namespace] {
			out = append(out, t)
		}
	}
	sort.Strings(out)
	return out
}

// RootGlue moves glue laid out under GlueDir to root, where an adopted
// engine keeps its glue. It is a pure function.
func RootGlue(g Glue, root string) Glue {
	if root == GlueDir {
		return g
	}
	move := func(p string) string {
		if rest, ok := strings.CutPrefix(p, GlueDir+"/"); ok {
			return path.Join(root, rest)
		}
		return p
	}
	out := g
	out.Interfaces = move(g.Interfaces)
	out.ChoicesHeader = move(g.ChoicesHeader)
	out.Startup.Files = nil
	for _, f := range g.Startup.Files {
		out.Startup.Files = append(out.Startup.Files, move(f))
	}
	for _, m := range []*map[string]GlueEntry{&out.Bindings, &out.Pairings, &out.Setup} {
		moved := map[string]GlueEntry{}
		for k, e := range *m {
			var files []string
			for _, f := range e.Files {
				files = append(files, move(f))
			}
			e.Files = files
			moved[k] = e
		}
		*m = moved
	}
	return out
}
