// Package schemas embeds Automap's published JSON Schemas, so the CLI
// validates against exactly the schemas it ships with.
package schemas

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed *.schema.json
var files embed.FS

// BaseURL is the prefix of every schema's $id.
const BaseURL = "https://github.com/zach/automap/schemas/"

// Names lists the embedded schemas by name, such as "cli-result".
func Names() []string {
	entries, err := fs.Glob(files, "*.schema.json")
	if err != nil {
		panic(err)
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = strings.TrimSuffix(e, ".schema.json")
	}
	return names
}

// Compile compiles the named schema, resolving references between the
// embedded schemas by their $id.
func Compile(name string) (*jsonschema.Schema, error) {
	c := jsonschema.NewCompiler()
	c.AssertFormat()
	for _, n := range Names() {
		raw, err := files.ReadFile(n + ".schema.json")
		if err != nil {
			return nil, err
		}
		doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
		if err != nil {
			return nil, fmt.Errorf("schema %s: %w", n, err)
		}
		if err := c.AddResource(BaseURL+n+".schema.json", doc); err != nil {
			return nil, fmt.Errorf("schema %s: %w", n, err)
		}
	}
	return c.Compile(BaseURL + name + ".schema.json")
}
