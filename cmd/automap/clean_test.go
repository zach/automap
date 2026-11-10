package main_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zach/automap/internal/clitest"
)

// start ignores the directory it and the developer build in: a new
// .gitignore, or one more line in the repo's own.
func TestStartIgnoresTheBuildDirectory(t *testing.T) {
	for _, c := range []struct{ name, before, after string }{
		{"no .gitignore", "", "/build/\n"},
		{"the repo's own", "*.o\n.vscode/", "*.o\n.vscode/\n/build/\n"},
		{"already ignored", "build/\n*.o\n", "build/\n*.o\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			repo := fixtureRepo(t)
			if c.before != "" {
				repo.WriteFile(t, ".gitignore", c.before)
			}

			run := repo.Automap(t, startTestbed...)

			run.ExpectExit(t, 0)
			if got := run.ReadFile(t, ".gitignore"); got != c.after {
				t.Errorf(".gitignore is %q, want %q", got, c.after)
			}
		})
	}
}

// adopt ignores only what verify writes: where an existing engine builds
// is its own business.
func TestAdoptIgnoresWhatVerifyWrites(t *testing.T) {
	repo := acmeAdoptRepo(t)
	describeAcme(t, repo)

	run := repo.Automap(t, "adopt", "--map", "map", "--private-map", "automap-private", "--no-verify", "--json")

	run.ExpectExit(t, 0)
	if got := run.ReadFile(t, ".gitignore"); !strings.Contains("\n"+got, "\n/build/automap/\n") {
		t.Errorf(".gitignore is %q, want it to ignore /build/automap/", got)
	}
}

// After eject nothing in the repo points an agent at the CLI: the Automap
// section of AGENTS.md and the router skill go with the manifest and lock.
func TestEjectRemovesTheAgentFiles(t *testing.T) {
	repo := buildableRepo(t)
	repo.WriteFile(t, "AGENTS.md", "# Spacefarer\n\nBuild with care.\n")
	os.MkdirAll(filepath.Join(repo.Dir, ".agents", "skills"), 0o755)
	repo.Automap(t, startTestbed...).ExpectExit(t, 0)

	run := repo.Automap(t, "eject", "--map", "map", "--json")

	run.ExpectExit(t, 0)
	if got := run.ReadFile(t, "AGENTS.md"); got != "# Spacefarer\n\nBuild with care.\n" {
		t.Errorf("AGENTS.md after eject is %q, want the developer's own content alone", got)
	}
	for _, dir := range []string{".claude/skills/automap", ".agents/skills/automap", ".claude"} {
		if _, err := os.Stat(filepath.Join(repo.Dir, filepath.FromSlash(dir))); err == nil {
			t.Errorf("eject left %s", dir)
		}
	}
	if _, err := os.Stat(filepath.Join(repo.Dir, ".agents", "skills")); err != nil {
		t.Errorf("eject removed .agents/skills, the developer's own directory: %v", err)
	}
	if !strings.Contains(run.Result.Summary, "evidence/") {
		t.Errorf("summary %q, want it to say evidence/ stays", run.Result.Summary)
	}
}

// An AGENTS.md that held only the Automap section goes with it.
func TestEjectRemovesAnAgentsMDItWroteAlone(t *testing.T) {
	repo := buildableRepo(t)
	repo.Automap(t, startTestbed...).ExpectExit(t, 0)

	repo.Automap(t, "eject", "--map", "map", "--json").ExpectExit(t, 0)

	if _, err := os.Stat(filepath.Join(repo.Dir, "AGENTS.md")); err == nil {
		t.Errorf("eject left an AGENTS.md with nothing in it:\n%s", clitest.Read(t, repo, "AGENTS.md"))
	}
}

// The executable is run wherever the project's CMake defines it: verify
// puts it at the build root.
func TestVerifyRunsAnExecutableDefinedInASubdirectory(t *testing.T) {
	repo := buildableRepo(t)
	repo.Automap(t, "start", "--with", "stub-window", "--with", "stub-ui", "--map", "map", "--no-verify", "--json").ExpectExit(t, 0)
	repo.WriteFile(t, "game/app/main.cpp", clitest.Read(t, repo, "main.cpp"))
	os.Remove(filepath.Join(repo.Dir, "main.cpp"))
	repo.WriteFile(t, "game/CMakeLists.txt", "add_executable(game app/main.cpp)\ntarget_link_libraries(game PRIVATE automap_glue)\n")
	repo.WriteFile(t, "CMakeLists.txt", "cmake_minimum_required(VERSION 3.25)\nproject(game LANGUAGES C CXX)\ninclude(cmake/automap/automap.cmake)\nadd_subdirectory(game)\n")
	repo.EditJSON(t, "automap.manifest.json", set("executable", map[string]any{"target": "game", "sources": []any{"game/app/main.cpp"}}))

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	run.ExpectExit(t, 0)
	if run.Result.Level != "probed" {
		t.Errorf("level %q, want probed\n%s", run.Result.Level, run.Stdout)
	}
}
