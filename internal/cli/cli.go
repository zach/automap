// Package cli implements the automap commands.
package cli

import (
	"fmt"
	"io"
	"os"
)

const manifestFile = "automap.manifest.json"

const usage = `usage: automap <command> [arguments] [--json]

commands:
  show <library|slot|contract|recipe> <id>   show one map entry and its facts
  find [<query>] [--slot <slot>]             search the map's libraries
  verify                                     build the composition, run its smoke test and probes, write evidence

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
	"verify": {bools: []string{"json"}, run: verify},
	"show":   {bools: []string{"json"}, values: []string{"map"}, run: show},
	"find":   {bools: []string{"json"}, values: []string{"map", "slot"}, run: find},
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

func verify(a *args, r *result) int {
	if _, err := os.Stat(manifestFile); os.IsNotExist(err) {
		r.Summary = "Nothing to verify: no " + manifestFile + " in the current directory."
		r.fail("NO_MANIFEST", "No composition here. Run automap start to create one, or automap adopt to adopt an existing engine.")
		return exitUsage
	}
	r.Summary = "Verify cannot build a composition yet."
	r.fail("NOT_IMPLEMENTED", "This CLI build does not build or probe compositions.")
	return exitFailed
}
