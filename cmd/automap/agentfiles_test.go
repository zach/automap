package main_test

import (
	"strings"
	"testing"

	"github.com/zach/automap/internal/agentfiles"
	"github.com/zach/automap/internal/clitest"
)

func agentFilesRepo(t *testing.T) *clitest.Repo {
	t.Helper()
	repo := clitest.NewRepo(t, "agent-files")
	repo.FixtureMap(t, "map")
	return repo
}

func TestStartWritesTheAgentFiles(t *testing.T) {
	run := agentFilesRepo(t).Automap(t, startTestbed...)

	run.ExpectExit(t, 0)
	run.ExpectGolden(t, "start-agent-files")
}

func TestTheAgentsMDSectionLeavesTheRestOfTheFileAlone(t *testing.T) {
	repo := agentFilesRepo(t)
	before := clitest.Read(t, repo, "AGENTS.md")
	run := repo.Automap(t, startTestbed...)

	after := run.ReadFile(t, "AGENTS.md")
	if !strings.HasPrefix(after, before) {
		t.Errorf("AGENTS.md no longer starts with the developer's content:\n%s", after)
	}
	section := after[strings.Index(after, "<!-- automap:begin -->"):]
	for _, want := range []string{
		"automap.manifest.json", "automap.lock.json", // where the manifest and lock live
		"input = stub-window, tools-ui = stub-ui, windowing = stub-window", // the slots
		"cmake --build build", "automap verify --json", "./build/testbed", // build, test, run
		"router skill", // go through the skill and CLI
	} {
		if !strings.Contains(section, want) {
			t.Errorf("the Automap section lacks %q", want)
		}
	}
}

func TestTheAgentsMDSectionIsReplacedInPlace(t *testing.T) {
	repo := agentFilesRepo(t)
	original := "# Spacefarer\n\n<!-- automap:begin -->\nstale\n<!-- automap:end -->\n\n## After\n\nKeep me.\n"
	repo.WriteFile(t, "AGENTS.md", original)
	run := repo.Automap(t, startTestbed...)

	after := run.ReadFile(t, "AGENTS.md")
	if !strings.HasPrefix(after, "# Spacefarer\n\n<!-- automap:begin -->\n") || !strings.HasSuffix(after, "<!-- automap:end -->\n\n## After\n\nKeep me.\n") {
		t.Errorf("content around the markers changed:\n%s", after)
	}
	if strings.Contains(after, "stale") || strings.Count(after, "<!-- automap:begin -->") != 1 {
		t.Errorf("the old section wasn't replaced:\n%s", after)
	}
}

func TestTheClaudeImportIsAddedOnlyOnce(t *testing.T) {
	repo := agentFilesRepo(t)
	repo.WriteFile(t, "CLAUDE.md", "@AGENTS.md\nClaude: keep commits small.\n")
	run := repo.Automap(t, startTestbed...)

	if got := run.ReadFile(t, "CLAUDE.md"); got != "@AGENTS.md\nClaude: keep commits small.\n" {
		t.Errorf("CLAUDE.md changed:\n%s", got)
	}
}

func TestNoClaudeMDIsCreated(t *testing.T) {
	run := fixtureRepo(t).Automap(t, startTestbed...)

	for _, f := range run.Result.FilesWritten {
		if f == "CLAUDE.md" {
			t.Error("start created a CLAUDE.md where there was none")
		}
	}
}

func TestANewerCLIRewritesTheRouterSkill(t *testing.T) {
	repo := startedRepo(t)
	old := strings.Replace(agentfiles.RouterSkill("0.0.0-alpha"), "Compose, extend", "OLD Compose, extend", 1)
	repo.WriteFile(t, ".claude/skills/automap/SKILL.md", old)

	run := repo.Automap(t, "find", "--map", "map", "--json")

	run.ExpectExit(t, 0)
	if got := run.ReadFile(t, ".claude/skills/automap/SKILL.md"); got != agentfiles.RouterSkill("0.0.0-dev") {
		t.Errorf("the skill wasn't rewritten at the CLI's version:\n%s", got)
	}
	// The guides beside it are the skill's too: a composition an older CLI
	// wrote gets them with the newer skill.
	want := []string{".claude/skills/automap/SKILL.md", ".claude/skills/automap/game.md", ".claude/skills/automap/private-maps.md"}
	if strings.Join(run.Result.FilesWritten, " ") != strings.Join(want, " ") {
		t.Errorf("files_written %v, want the skill's files %v", run.Result.FilesWritten, want)
	}
}

// The router skill routes to two guides the CLI writes beside it: what a
// testbed needs to become a game, and how to write a private map.
func TestTheRouterSkillComesWithItsGuides(t *testing.T) {
	run := agentFilesRepo(t).Automap(t, startTestbed...)

	run.ExpectExit(t, 0)
	skill := run.ReadFile(t, ".claude/skills/automap/SKILL.md")
	for guide, wants := range map[string][]string{
		"game.md":         {`"executable"`, "--frames <n>", "--probe <file>", "am::probe::write_results()", "POST_BUILD", "CMAKE_BUILD_TYPE=Release", "What to ship"},
		"private-maps.md": {`"private_maps"`, "automap show recipe", "build-fragment", "automap verify --seed", "automap resolve --json"},
	} {
		if !strings.Contains(skill, "("+guide+")") {
			t.Errorf("SKILL.md doesn't link %s", guide)
		}
		for _, dir := range []string{".claude/skills/automap/", ".agents/skills/automap/"} {
			text := run.ReadFile(t, dir+guide)
			for _, want := range wants {
				if !strings.Contains(text, want) {
					t.Errorf("%s%s lacks %q", dir, guide, want)
				}
			}
		}
	}
	for _, want := range []string{"## Changing a composition", "automap remove <library> --json", "automap resolve --json", "automap vendor --json", "actions/attest-build-provenance"} {
		if !strings.Contains(skill, want) {
			t.Errorf("SKILL.md lacks %q", want)
		}
	}
}

func TestAnOlderCLILeavesANewerRouterSkillAlone(t *testing.T) {
	repo := startedRepo(t)
	newer := agentfiles.RouterSkill("9.0.0")
	repo.WriteFile(t, ".claude/skills/automap/SKILL.md", newer)

	run := repo.Automap(t, "find", "--map", "map", "--json")

	if got := run.ReadFile(t, ".claude/skills/automap/SKILL.md"); got != newer {
		t.Error("an older CLI rewrote a newer skill")
	}
}

func TestTheLockRefusalQuotesTheSameNeverAsTheSkill(t *testing.T) {
	repo := startedRepo(t)
	repo.EditJSON(t, "automap.lock.json", set("pins/stub-ui/version", "1.0.1"))
	skill := clitest.Read(t, repo, ".claude/skills/automap/SKILL.md")

	run := repo.Automap(t, "verify", "--map", "map", "--json")

	run.ExpectExit(t, 3)
	never := agentfiles.Never(1)
	run.ExpectFailure(t, "LOCK_EDITED", never)
	if !strings.Contains(skill, "1. "+never) {
		t.Errorf("the skill's first never isn't %q", never)
	}
}
