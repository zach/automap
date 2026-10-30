// Package cli implements the automap commands.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

const manifestFile = "automap.manifest.json"

const usage = `usage: automap <command> [--json] [flags]

commands:
  verify    build the composition, run its smoke test and probes, write evidence
`

// Run runs the CLI with args (without the program name) in the current
// directory and returns the exit status.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return exitUsage
	}
	switch args[0] {
	case "verify":
		return verify(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "automap: unknown command %q\n\n%s", args[0], usage)
		return exitUsage
	}
}

func verify(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("verify", flag.ContinueOnError)
	flags.SetOutput(stderr)
	asJSON := flags.Bool("json", false, "print a CLI result as JSON")
	if err := flags.Parse(args); err != nil {
		return exitUsage
	}

	r := newResult("verify")
	status := exitFailed
	if _, err := os.Stat(manifestFile); errors.Is(err, os.ErrNotExist) {
		r.Summary = "Nothing to verify: no " + manifestFile + " in the current directory."
		r.fail("NO_MANIFEST", "No composition here. Run automap start to create one, or automap adopt to adopt an existing engine.")
		status = exitUsage
	} else {
		r.Summary = "Verify cannot build a composition yet."
		r.fail("NOT_IMPLEMENTED", "This CLI build does not build or probe compositions.")
	}
	return finish(r, status, *asJSON, stdout, stderr)
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
