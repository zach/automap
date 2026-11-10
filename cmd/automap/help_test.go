package main_test

import (
	"strings"
	"testing"

	"github.com/zach/automap/internal/clitest"
)

// --version and version say which CLI this is, without a composition or a
// map, and exit 0.
func TestVersionSaysWhichCLIThisIs(t *testing.T) {
	for _, args := range [][]string{{"--version"}, {"version"}} {
		run := clitest.Automap(t, "no-manifest", args...)
		run.ExpectExit(t, 0)
		if out := string(run.Stdout); !strings.HasPrefix(out, "automap ") || strings.Count(out, "\n") != 1 {
			t.Errorf("automap %s printed %q, want one line: automap <version>", strings.Join(args, " "), out)
		}
	}
	run := clitest.Automap(t, "no-manifest", "version", "--json")
	run.ExpectExit(t, 0)
	if !run.Result.OK || run.Result.Command != "version" || run.Result.CLIVersion == "" {
		t.Errorf("version --json is %+v, want ok with command version and a cli_version", run.Result)
	}
}

// Every command the CLI runs is in its usage: the text is written from
// the command table.
func TestTheUsageListsEveryCommand(t *testing.T) {
	run := clitest.Automap(t, "no-manifest", "help")
	run.ExpectExit(t, 0)
	usage := string(run.Stdout)
	for _, cmd := range []string{"start", "adopt", "extend", "upgrade", "verify", "show", "find", "fetch", "vendor", "eject", "version", "help"} {
		if !strings.Contains(usage, "\n  "+cmd) {
			t.Errorf("the usage doesn't list %s:\n%s", cmd, usage)
		}
	}
	for _, args := range [][]string{{"--help"}, {"-h"}} {
		if got := clitest.Automap(t, "no-manifest", args...); got.Exit != 0 || string(got.Stdout) != usage {
			t.Errorf("automap %s exits %d and prints\n%s\nwant the usage and 0", args[0], got.Exit, got.Stdout)
		}
	}
}

// With no arguments the CLI prints the usage on stderr and exits 2; an
// unknown command says so first.
func TestNoCommandIsAUsageError(t *testing.T) {
	usage := string(clitest.Automap(t, "no-manifest", "help").Stdout)
	run := clitest.Automap(t, "no-manifest")
	run.ExpectExit(t, 2)
	if string(run.Stderr) != usage {
		t.Errorf("automap with no arguments printed on stderr\n%s\nwant the usage", run.Stderr)
	}
	run = clitest.Automap(t, "no-manifest", "frobnicate")
	run.ExpectExit(t, 2)
	if !strings.HasPrefix(string(run.Stderr), `automap: unknown command "frobnicate"`) {
		t.Errorf("an unknown command printed %q", run.Stderr)
	}
}

// help <command> and <command> --help describe one command: its synopsis,
// what it does and its flags.
func TestHelpDescribesACommand(t *testing.T) {
	run := clitest.Automap(t, "no-manifest", "help", "upgrade")
	run.ExpectExit(t, 0)
	help := string(run.Stdout)
	for _, want := range []string{"usage: automap upgrade [<library>...]", "--map-release <release>", "--no-verify", "--json"} {
		if !strings.Contains(help, want) {
			t.Errorf("help upgrade lacks %q:\n%s", want, help)
		}
	}
	if got := clitest.Automap(t, "no-manifest", "upgrade", "--help"); got.Exit != 0 || string(got.Stdout) != help {
		t.Errorf("upgrade --help exits %d and prints\n%s\nwant what help upgrade prints", got.Exit, got.Stdout)
	}
	run = clitest.Automap(t, "no-manifest", "help", "frobnicate")
	run.ExpectExit(t, 2)
}
