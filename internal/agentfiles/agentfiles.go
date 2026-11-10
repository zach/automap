// Package agentfiles writes what the developer's agents read: the Automap
// section of AGENTS.md, the CLAUDE.md import line, and the router skill
// (ADR-0043, ADR-0044).
package agentfiles

import (
	_ "embed"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

// Nevers are the four things an agent must never do (ADR-0046). The
// router skill and the CLI's refusals share this one list (ADR-0074).
var Nevers = []string{
	"Never edit the lock by hand: only the CLI writes automap.lock.json.",
	"Never create or claim evidence except by running the CLI.",
	"Never weaken, skip or delete a probe to get a pass: fix the glue, or report the failure.",
	"Never open an issue or pull request on an upstream project whose contribution policy bans AI-generated contributions (each library's contribution_policy in the map).",
}

// Never returns never n, counting from 1.
func Never(n int) string { return Nevers[n-1] }

const (
	beginMarker = "<!-- automap:begin -->"
	endMarker   = "<!-- automap:end -->"
)

// SectionInput is what the AGENTS.md section describes.
type SectionInput struct {
	Slots      map[string]string
	Executable string
}

// Section is the Automap section of AGENTS.md, markers included: the
// build, test and run commands, where the manifest and lock live, the slots
// in use, and the instruction to go through the router skill and CLI.
func Section(in SectionInput) string {
	var slots []string
	for slot, lib := range in.Slots {
		slots = append(slots, slot+" = "+lib)
	}
	sort.Strings(slots)
	return beginMarker + `
## Automap

This engine is an Automap composition. Before changing any glue, read the router skill (` + "`.claude/skills/automap/SKILL.md`" + `) and go through the ` + "`automap`" + ` CLI.

- Manifest: ` + "`automap.manifest.json`" + `. Lock: ` + "`automap.lock.json`" + `, which only the CLI writes.
- Slots: ` + strings.Join(slots, ", ") + `.
- Test: ` + "`automap verify --json`" + `, which builds a Debug build in ` + "`build/automap/<platform>/`" + `, runs it headless and runs the probes.
- Build: ` + "`cmake -S . -B build -G Ninja -DCMAKE_TOOLCHAIN_FILE=$VCPKG_ROOT/scripts/buildsystems/vcpkg.cmake && cmake --build build`" + `. Without ` + "`VCPKG_ROOT`" + `, use the vcpkg the CLI keeps: the first line of ` + "`build/automap/build-<platform>.log`" + ` is the configure command ` + "`automap verify`" + ` ran, toolchain path included.
- Run: ` + "`./build/" + in.Executable + "`" + `
` + endMarker + "\n"
}

// UpdateAgentsMD returns existing with its Automap section replaced, or
// with the section appended when it has none. Everything outside the
// markers keeps its bytes.
func UpdateAgentsMD(existing, section string) string {
	begin := strings.Index(existing, beginMarker)
	end := strings.Index(existing, endMarker)
	if begin >= 0 && end > begin {
		after := existing[end+len(endMarker):]
		after = strings.TrimPrefix(after, "\n")
		return existing[:begin] + section + after
	}
	if existing == "" {
		return section
	}
	sep := "\n"
	if !strings.HasSuffix(existing, "\n") {
		sep = "\n\n"
	}
	return existing + sep + section
}

// RemoveSection returns existing without its Automap section, and whether
// it had one. Everything outside the markers keeps its bytes, but for the
// blank line the section was appended after.
func RemoveSection(existing string) (string, bool) {
	begin := strings.Index(existing, beginMarker)
	end := strings.Index(existing, endMarker)
	if begin < 0 || end < begin {
		return existing, false
	}
	before := existing[:begin]
	after := strings.TrimPrefix(existing[end+len(endMarker):], "\n")
	if after == "" && strings.HasSuffix(before, "\n\n") {
		before = strings.TrimSuffix(before, "\n")
	}
	return before + after, true
}

// RemoveClaudeImport returns CLAUDE.md without the import line.
func RemoveClaudeImport(existing string) string {
	var kept []string
	for _, line := range strings.SplitAfter(existing, "\n") {
		if strings.TrimSpace(line) != ClaudeImport {
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "")
}

// ClaudeImport is the line that makes Claude Code read AGENTS.md.
const ClaudeImport = "@AGENTS.md"

// AddClaudeImport returns CLAUDE.md with the import line added, and
// whether it changed. A file that already has the line is unchanged.
func AddClaudeImport(existing string) (string, bool) {
	for _, line := range strings.Split(existing, "\n") {
		if strings.TrimSpace(line) == ClaudeImport {
			return existing, false
		}
	}
	if existing != "" && !strings.HasSuffix(existing, "\n") {
		existing += "\n"
	}
	return existing + ClaudeImport + "\n", true
}

//go:embed SKILL.md.tmpl
var skillTemplate string

//go:embed game.md.tmpl
var gameTemplate string

//go:embed private-maps.md.tmpl
var privateMapsTemplate string

// skillFiles are the router skill's files, by name in its directory:
// SKILL.md routes, and the guides beside it hold what is too long for it.
var skillFiles = map[string]*string{"SKILL.md": &skillTemplate, "game.md": &gameTemplate, "private-maps.md": &privateMapsTemplate}

// RouterSkill is the router skill's SKILL.md at a CLI version.
func RouterSkill(version string) string { return skillFile("SKILL.md", version) }

// SkillGuide is one of the guides beside SKILL.md, game.md or
// private-maps.md, at a CLI version.
func SkillGuide(name, version string) string { return skillFile(name, version) }

func skillFile(name, version string) string {
	t := template.Must(template.New(name).Funcs(template.FuncMap{"inc": func(i int) int { return i + 1 }}).Parse(*skillFiles[name]))
	var b strings.Builder
	if err := t.Execute(&b, struct {
		Version string
		Nevers  []string
	}{version, Nevers}); err != nil {
		panic(err)
	}
	return b.String()
}

// SkillFiles is every file of the router skill at a CLI version, by its
// name in the skill's directory.
func SkillFiles(version string) map[string]string {
	out := map[string]string{}
	for name := range skillFiles {
		out[name] = skillFile(name, version)
	}
	return out
}

// skillDirFiles is every file of the router skill in the directory that
// holds its SKILL.md at p.
func skillDirFiles(p, version string) map[string][]byte {
	out := map[string][]byte{}
	for name := range skillFiles {
		out[path.Join(path.Dir(p), name)] = []byte(skillFile(name, version))
	}
	return out
}

// IsAgentFile says whether a path in the repo is one of the files this
// package maintains: commands merge into them where they exist, never keep
// them as found.
func IsAgentFile(p string) bool {
	return p == "AGENTS.md" || p == "CLAUDE.md" || strings.Contains(p, "/skills/automap/")
}

// otherSkillDirs are other agents' skill directories; the router skill goes
// into each one a repo already has.
var otherSkillDirs = []string{".agents/skills", ".github/skills", ".codex/skills", ".cursor/skills"}

// SkillPaths lists where the router skill goes in root: Claude Code's skill
// directory, and every other agent's skill directory present.
func SkillPaths(root string) []string {
	out := []string{".claude/skills/automap/SKILL.md"}
	for _, d := range otherSkillDirs {
		if info, err := os.Stat(filepath.Join(root, filepath.FromSlash(d))); err == nil && info.IsDir() {
			out = append(out, path.Join(d, "automap", "SKILL.md"))
		}
	}
	return out
}

// Files returns the agent files a command that changes the composition
// writes in root: AGENTS.md with its section, CLAUDE.md with the import if
// it exists and lacks it, and the router skill in each skill directory.
func Files(root, version string, in SectionInput) (map[string][]byte, error) {
	out := map[string][]byte{}
	existing, err := readOptional(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		return nil, err
	}
	out["AGENTS.md"] = []byte(UpdateAgentsMD(existing, Section(in)))
	if claude, err := os.ReadFile(filepath.Join(root, "CLAUDE.md")); err == nil {
		if updated, changed := AddClaudeImport(string(claude)); changed {
			out["CLAUDE.md"] = []byte(updated)
		}
	}
	for _, p := range SkillPaths(root) {
		for file, data := range skillDirFiles(p, version) {
			out[file] = data
		}
	}
	return out, nil
}

// StaleSkills returns the router skills in root written by an older CLI
// than version, rewritten at version (ADR-0044).
func StaleSkills(root, version string) map[string][]byte {
	out := map[string][]byte{}
	for _, p := range SkillPaths(root) {
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
		if err != nil {
			continue
		}
		if m := skillVersion.FindStringSubmatch(string(raw)); m != nil && compareSemver(m[1], version) < 0 {
			for file, data := range skillDirFiles(p, version) {
				out[file] = data
			}
		}
	}
	return out
}

var skillVersion = regexp.MustCompile(`automap-cli-version: "([^"]+)"`)

func readOptional(file string) (string, error) {
	raw, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		return "", nil
	}
	return string(raw), err
}

// compareSemver orders semantic versions, including pre-releases.
func compareSemver(a, b string) int {
	coreA, preA, _ := strings.Cut(a, "-")
	coreB, preB, _ := strings.Cut(b, "-")
	pa, pb := strings.Split(coreA, "."), strings.Split(coreB, ".")
	for i := 0; i < 3; i++ {
		x, y := 0, 0
		if i < len(pa) {
			x, _ = strconv.Atoi(pa[i])
		}
		if i < len(pb) {
			y, _ = strconv.Atoi(pb[i])
		}
		if x != y {
			return x - y
		}
	}
	switch {
	case preA == preB:
		return 0
	case preA == "":
		return 1
	case preB == "":
		return -1
	}
	return strings.Compare(preA, preB)
}
