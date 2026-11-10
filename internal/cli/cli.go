// Package cli implements the automap commands.
package cli

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/zach/automap/internal/agentfiles"
	"github.com/zach/automap/internal/composition"
)

const manifestFile = composition.ManifestFile

// flag is one of a command's flags. A flag with a value names it, such as
// "<dir>"; a flag without one is a switch.
type flag struct {
	name, value, does string
}

// command is one automap command.
type command struct {
	name     string
	synopsis string // its arguments, as the usage shows them
	does     string
	flags    []flag
	run      func(a *args, r *result) int
}

var (
	jsonFlag     = flag{"json", "", "print a CLI result as JSON"}
	mapFlag      = flag{"map", "<dir>", "read the public map from a directory (or set AUTOMAP_MAP)"}
	noVerifyFlag = flag{"no-verify", "", "write the files without building and verifying"}
	platformFlag = flag{"platform", "<id>", "a platform the composition targets; repeat for more (default: all three)"}
)

// commandList is every command, in the order the usage lists them. The
// usage and each command's help are written from it, so neither can leave
// a command or a flag out.
var commandList = []command{
	{name: "start", synopsis: "--with <library>[:<variant>]... [--slot <slot>=<library>]... [--platform <id>]... [--no-verify]",
		does: "create a composition in the current directory, then verify it",
		flags: []flag{{"with", "<library>[:<variant>]", "a library to compose, in the variant named or its first; repeat for more"},
			{"slot", "<slot>=<library>", "the library that fills a slot, when more than one could"}, platformFlag, noVerifyFlag, mapFlag, jsonFlag},
		run: start},
	{name: "adopt", synopsis: "[--private-map <dir>]... [--target <executable>] [--platform <id>]... [--no-verify]",
		does: "bring an existing CMake and vcpkg project under Automap",
		flags: []flag{{"private-map", "<dir>", "a private map that describes the project's own libraries; repeat for more"},
			{"target", "<executable>", "the CMake target of the engine's executable"}, platformFlag, noVerifyFlag, mapFlag, jsonFlag},
		run: adopt},
	{name: "extend", synopsis: "<library>[:<variant>] [--slot <slot>]... [--no-verify]",
		does:  "add a library to the composition, then verify it",
		flags: []flag{{"slot", "<slot>", "a slot the library fills, instead of every slot it can; repeat for more"}, noVerifyFlag, mapFlag, jsonFlag},
		run:   extend},
	{name: "upgrade", synopsis: "[<library>...] [--map-release <release>] [--no-verify]",
		does:  "move the composition to a newer map release, re-resolve the libraries named (all by default),\nupdate the glue the developer hasn't changed, then verify it",
		flags: []flag{{"map-release", "<release>", "the map release to move to (default: the newest)"}, noVerifyFlag, mapFlag, jsonFlag},
		run:   upgrade},
	{name: "verify", synopsis: "[--seed <rule>[:<library>]]",
		does: "build the composition, run its smoke test and probes, write evidence;\nwith --seed, check a probe catches its seeded violation",
		flags: []flag{{"seed", "<rule>[:<library>]", "apply the rule's seeded violation in a scratch copy and check its probe fails"},
			{"platform", "<id>", "the platform to verify, which must be this machine's"}, mapFlag, jsonFlag},
		run: verifyCmd},
	{name: "show", synopsis: "<library|slot|contract|recipe> <id>", does: "show one map entry and its facts",
		flags: []flag{mapFlag, jsonFlag}, run: show},
	{name: "find", synopsis: "[<query>] [--slot <slot>]", does: "search the map's libraries",
		flags: []flag{{"slot", "<slot>", "only libraries that fill this slot"}, mapFlag, jsonFlag}, run: find},
	{name: "fetch", does: "fill the caches the composition needs to build and verify offline",
		flags: []flag{mapFlag, jsonFlag}, run: fetchCmd},
	{name: "vendor", does: "copy the composition's ports, sources, map release and evidence\ninto vendor/automap/, for builds with no network and no caches",
		flags: []flag{mapFlag, jsonFlag}, run: vendorCmd},
	{name: "eject", does: "leave Automap: copy the Automap registry's ports into ports/,\npoint vcpkg at them, and remove the manifest and lock",
		flags: []flag{mapFlag, jsonFlag}, run: eject},
	{name: "version", does: "print the CLI's version (also --version)", flags: []flag{jsonFlag}},
	{name: "help", synopsis: "[<command>]", does: "print this usage, or one command's (also <command> --help)"},
}

// commandNamed finds a command by name.
func commandNamed(name string) (command, bool) {
	for _, c := range commandList {
		if c.name == name {
			return c, true
		}
	}
	return command{}, false
}

// usage is the CLI's usage: every command with its arguments and what it
// does.
func usage() string {
	var b strings.Builder
	b.WriteString("usage: automap <command> [arguments] [--json]\n\ncommands:\n")
	for _, c := range commandList {
		fmt.Fprintf(&b, "  %s\n", strings.TrimSpace(c.name+" "+c.synopsis))
		for _, line := range strings.Split(c.does, "\n") {
			fmt.Fprintf(&b, "      %s\n", line)
		}
	}
	b.WriteString("\nflags every command takes:\n  --json        print a CLI result as JSON\n  --map <dir>   read the public map from a directory (or set AUTOMAP_MAP)\n\nautomap help <command> lists a command's flags.\n")
	return b.String()
}

// help is one command's help: its synopsis, what it does and its flags.
func help(c command) string {
	var b strings.Builder
	fmt.Fprintf(&b, "usage: automap %s\n\n", strings.TrimSpace(c.name+" "+c.synopsis))
	for _, line := range strings.Split(c.does, "\n") {
		fmt.Fprintf(&b, "%s\n", line)
	}
	if len(c.flags) > 0 {
		b.WriteString("\nflags:\n")
		width := 0
		for _, f := range c.flags {
			width = max(width, len(strings.TrimSpace(f.name+" "+f.value)))
		}
		for _, f := range c.flags {
			fmt.Fprintf(&b, "  --%-*s  %s\n", width, strings.TrimSpace(f.name+" "+f.value), f.does)
		}
	}
	return b.String()
}

// Run runs the CLI with args (without the program name) in the current
// directory and returns the exit status.
func Run(argv []string, stdout, stderr io.Writer) int {
	if len(argv) == 0 {
		fmt.Fprint(stderr, usage())
		return exitUsage
	}
	// Help and the version need no composition and no map, and asking for
	// them is never an error.
	name, rest := argv[0], argv[1:]
	switch name {
	case "--help", "-h":
		name = "help"
	case "--version":
		name = "version"
	}
	cmd, ok := commandNamed(name)
	if !ok {
		fmt.Fprintf(stderr, "automap: unknown command %q\n\n%s", argv[0], usage())
		return exitUsage
	}
	switch {
	case name == "help" && len(rest) == 0:
		fmt.Fprint(stdout, usage())
		return exitOK
	case name == "help":
		about, ok := commandNamed(rest[0])
		if !ok || len(rest) > 1 {
			fmt.Fprintf(stderr, "automap: unknown command %q\n\n%s", strings.Join(rest, " "), usage())
			return exitUsage
		}
		fmt.Fprint(stdout, help(about))
		return exitOK
	case slices.Contains(rest, "--help") || slices.Contains(rest, "-h"):
		fmt.Fprint(stdout, help(cmd))
		return exitOK
	}
	r := newResult(name)
	var bools, values []string
	for _, f := range cmd.flags {
		if f.value == "" {
			bools = append(bools, f.name)
		} else {
			values = append(values, f.name)
		}
	}
	a, err := parseArgs(rest, bools, values)
	if err != nil {
		r.Summary = "Usage error: " + err.Error() + "."
		r.fail("USAGE", err.Error()+"; automap help "+name+" lists the flags")
		return finish(r, exitUsage, hasJSONFlag(argv), stdout, stderr)
	}
	if name == "version" {
		if len(a.positional) > 0 {
			r.Summary = "Usage error: version takes no arguments."
			r.fail("USAGE", "automap version [--json]")
			return finish(r, exitUsage, a.bools["json"], stdout, stderr)
		}
		r.OK = true
		r.Summary = "automap " + Version
		return finish(r, exitOK, a.bools["json"], stdout, stderr)
	}
	if status := checkLock(r); status != exitOK {
		return finish(r, status, a.bools["json"], stdout, stderr)
	}
	if stale := agentfiles.StaleSkills(".", Version); len(stale) > 0 {
		written, err := writeFiles(stale)
		if err != nil {
			fmt.Fprintf(stderr, "automap: rewriting the router skill: %v\n", err)
		}
		r.FilesWritten = append(r.FilesWritten, written...)
	}
	status := cmd.run(a, r)
	return finish(r, status, a.bools["json"], stdout, stderr)
}

func hasJSONFlag(argv []string) bool {
	for _, a := range argv {
		if a == "--json" {
			return true
		}
	}
	return false
}

// finish prints r and returns status, or reports on stderr and returns
// exitFailed when r can't be printed.
func finish(r *result, status int, asJSON bool, stdout, stderr io.Writer) int {
	if err := r.print(stdout, asJSON); err != nil {
		fmt.Fprintf(stderr, "automap: %v\n", err)
		return exitFailed
	}
	return status
}

// checkLock refuses to act on a lock in the current directory whose hash
// doesn't match: only the CLI writes the lock (ADR-0046).
func checkLock(r *result) int {
	raw, err := os.ReadFile(composition.LockFile)
	if err != nil {
		return exitOK
	}
	if _, err := composition.CheckLock(raw); err != nil {
		r.Summary = "Refusing to act on a hand-edited lock."
		r.fail("LOCK_EDITED", composition.LockFile+" was edited by hand: "+err.Error()+". "+agentfiles.Never(1)+" Restore it from version control and make the change through the CLI.")
		return exitRefused
	}
	return exitOK
}
