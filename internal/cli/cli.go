// Package cli implements the automap commands.
package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/zach/automap/internal/agentfiles"
	"github.com/zach/automap/internal/composition"
)

const manifestFile = composition.ManifestFile

const usage = `usage: automap <command> [arguments] [--json]

commands:
  start --with <library>[:<variant>]... [--slot <slot>=<library>]... [--platform <id>]... [--no-verify]
                                             create a composition in the current directory, then verify it
  adopt [--private-map <dir>]... [--target <executable>] [--platform <id>]... [--no-verify]
                                             bring an existing CMake and vcpkg project under Automap
  show <library|slot|contract|recipe> <id>   show one map entry and its facts
  find [<query>] [--slot <slot>]             search the map's libraries
  extend <library>[:<variant>] [--slot <slot>]... [--no-verify]
                                             add a library to the composition, then verify it
  verify [--seed <rule>[:<library>]]         build the composition, run its smoke test and probes, write evidence;
                                             with --seed, check a probe catches its seeded violation
  fetch                                      fill the caches the composition needs to build and verify offline
  vendor                                     copy the composition's ports, sources, map release and evidence
                                             into vendor/automap/, for builds with no network and no caches

flags:
  --map <dir>   read the public map from a directory (or set AUTOMAP_MAP)
  --json        print a CLI result as JSON
`

// command is one automap command.
type command struct {
	bools  []string
	values []string
	run    func(a *args, r *result) int
}

var commands = map[string]command{
	"start":   {bools: []string{"json", "no-verify"}, values: []string{"map", "with", "slot", "platform"}, run: start},
	"extend":  {bools: []string{"json", "no-verify"}, values: []string{"map", "slot"}, run: extend},
	"adopt":   {bools: []string{"json", "no-verify"}, values: []string{"map", "private-map", "platform", "target"}, run: adopt},
	"upgrade": {bools: []string{"json", "no-verify"}, values: []string{"map", "map-release"}, run: upgrade},
	"verify":  {bools: []string{"json"}, values: []string{"map", "platform", "seed"}, run: verifyCmd},
	"show":    {bools: []string{"json"}, values: []string{"map"}, run: show},
	"find":    {bools: []string{"json"}, values: []string{"map", "slot"}, run: find},
	"fetch":   {bools: []string{"json"}, values: []string{"map"}, run: fetchCmd},
	"vendor":  {bools: []string{"json"}, values: []string{"map"}, run: vendorCmd},
}

// Run runs the CLI with args (without the program name) in the current
// directory and returns the exit status.
func Run(argv []string, stdout, stderr io.Writer) int {
	if len(argv) == 0 {
		fmt.Fprint(stderr, usage)
		return exitUsage
	}
	cmd, ok := commands[argv[0]]
	if !ok {
		fmt.Fprintf(stderr, "automap: unknown command %q\n\n%s", argv[0], usage)
		return exitUsage
	}
	r := newResult(argv[0])
	a, err := parseArgs(argv[1:], cmd.bools, cmd.values)
	if err != nil {
		r.Summary = "Usage error: " + err.Error() + "."
		r.fail("USAGE", err.Error())
		return finish(r, exitUsage, hasJSONFlag(argv), stdout, stderr)
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
