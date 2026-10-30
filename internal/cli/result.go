package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/zach/automap/schemas"
)

// Exit statuses (SPEC §5.4).
const (
	exitOK      = 0
	exitFailed  = 1 // a rule or build failed
	exitUsage   = 2
	exitRefused = 3
)

// Version is the CLI's version, set at release with
// -ldflags "-X github.com/zach/automap/internal/cli.Version=<semver>".
var Version = "0.0.0-dev"

// result is a CLI result, the JSON every command prints with --json
// (schemas/cli-result.schema.json).
type result struct {
	Command    string    `json:"command"`
	CLIVersion string    `json:"cli_version"`
	OK         bool      `json:"ok"`
	Summary    string    `json:"summary"`
	Warnings   []warning `json:"warnings"`
	Failures   []failure `json:"failures"`
}

type warning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type failure struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func newResult(command string) *result {
	return &result{Command: command, CLIVersion: Version, Warnings: []warning{}, Failures: []failure{}}
}

func (r *result) fail(code, message string) {
	r.OK = false
	r.Failures = append(r.Failures, failure{Code: code, Message: message})
}

// print writes r as JSON, or as text for a person. The JSON is checked
// against the embedded cli-result schema first: output that breaks the
// published contract is a CLI bug, never something to print.
func (r *result) print(w io.Writer, asJSON bool) {
	if !asJSON {
		fmt.Fprintln(w, r.Summary)
		for _, f := range r.Failures {
			fmt.Fprintf(w, "%s: %s\n", f.Code, f.Message)
		}
		return
	}
	out, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		panic(err)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(out))
	if err != nil {
		panic(err)
	}
	if err := cliResultSchema().Validate(doc); err != nil {
		panic(fmt.Sprintf("CLI result breaks the cli-result schema: %v\n%s", err, out))
	}
	w.Write(append(out, '\n'))
}

func cliResultSchema() *jsonschema.Schema {
	s, err := schemas.Compile("cli-result")
	if err != nil {
		panic(err)
	}
	return s
}
