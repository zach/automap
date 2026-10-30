package agentfiles_test

import (
	"strings"
	"testing"

	"github.com/zach/automap/internal/agentfiles"
)

const section = "<!-- automap:begin -->\nnew section\n<!-- automap:end -->\n"

func TestUpdateAgentsMD(t *testing.T) {
	for _, c := range []struct{ name, existing, want string }{
		{"no file", "", section},
		{"appends after the developer's content", "# Rules\n\nMine.\n", "# Rules\n\nMine.\n\n" + section},
		{"appends after content with no final newline", "# Rules", "# Rules\n\n" + section},
		{"replaces the section, keeping both sides", "# Before\n\n<!-- automap:begin -->\nold\n<!-- automap:end -->\n\n# After\n", "# Before\n\n" + section + "\n# After\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := agentfiles.UpdateAgentsMD(c.existing, section); got != c.want {
				t.Errorf("got\n%q\nwant\n%q", got, c.want)
			}
		})
	}
}

func TestAddClaudeImport(t *testing.T) {
	for _, c := range []struct {
		name, existing, want string
		changed              bool
	}{
		{"adds the line", "My notes.\n", "My notes.\n@AGENTS.md\n", true},
		{"adds a newline first", "My notes.", "My notes.\n@AGENTS.md\n", true},
		{"only once", "@AGENTS.md\nMy notes.\n", "@AGENTS.md\nMy notes.\n", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, changed := agentfiles.AddClaudeImport(c.existing)
			if got != c.want || changed != c.changed {
				t.Errorf("got %q, %v; want %q, %v", got, changed, c.want, c.changed)
			}
		})
	}
}

func TestTheRouterSkillCarriesTheVersionAndTheFourNevers(t *testing.T) {
	skill := agentfiles.RouterSkill("1.2.3")
	if !strings.Contains(skill, `automap-cli-version: "1.2.3"`) {
		t.Error("the skill doesn't carry the CLI version")
	}
	for i, n := range agentfiles.Nevers {
		if !strings.Contains(skill, string(rune('1'+i))+". "+n+"\n") {
			t.Errorf("the skill lacks never %d: %s", i+1, n)
		}
	}
	if len(agentfiles.Nevers) != 4 {
		t.Errorf("%d nevers, want four (ADR-0046)", len(agentfiles.Nevers))
	}
}
