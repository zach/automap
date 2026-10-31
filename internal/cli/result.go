package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

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
	Command      string    `json:"command"`
	CLIVersion   string    `json:"cli_version"`
	OK           bool      `json:"ok"`
	Summary      string    `json:"summary"`
	Level        string    `json:"level,omitempty"`
	Warnings     []warning `json:"warnings"`
	Failures     []failure `json:"failures"`
	FilesWritten []string  `json:"files_written,omitempty"`
	EvidenceFile string    `json:"evidence_file,omitempty"`
	Entries      []entry   `json:"entries,omitempty"`
}

// entry is one map entry in a find or show result.
type entry map[string]any

// warn records a warning.
func (r *result) warn(code, message string) {
	r.Warnings = append(r.Warnings, warning{Code: code, Message: message})
}

type warning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type failure struct {
	Code        string `json:"code"`
	Rule        string `json:"rule,omitempty"`
	Library     string `json:"library,omitempty"`
	Message     string `json:"message"`
	DetailsFile string `json:"details_file,omitempty"`
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
// published contract is a CLI bug, reported on stderr instead of printed.
func (r *result) print(w io.Writer, asJSON bool) error {
	if !asJSON {
		var text bytes.Buffer
		fmt.Fprintln(&text, r.Summary)
		for _, f := range r.Failures {
			fmt.Fprintf(&text, "%s: %s\n", f.Code, f.Message)
		}
		_, err := w.Write(text.Bytes())
		return err
	}
	out, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	var doc any
	if err := json.Unmarshal(out, &doc); err != nil {
		return err
	}
	if v := schemas.Validate("cli-result", doc); v != nil {
		return fmt.Errorf("CLI result breaks the cli-result schema: %s\n%s", strings.Join(v, "; "), out)
	}
	_, err = w.Write(append(out, '\n'))
	return err
}
