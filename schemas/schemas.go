// Package schemas embeds Automap's published JSON Schemas, so the CLI
// validates against exactly the schemas it ships with.
package schemas

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
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

var compiled sync.Map // name -> *jsonschema.Schema

// Validate checks doc, a value decoded from JSON, against the named
// published schema. It returns one message per violation, each
// "<instance location>: <reason>", or nil when doc is valid.
func Validate(name string, doc any) []string {
	s, ok := compiled.Load(name)
	if !ok {
		c, err := Compile(name)
		if err != nil {
			panic(fmt.Sprintf("embedded schema %s: %v", name, err))
		}
		s, _ = compiled.LoadOrStore(name, c)
	}
	return violations(s.(*jsonschema.Schema), doc)
}

// ValidateAgainst checks doc against an inline JSON Schema, such as a
// contract's schema for one fact. It returns an error when the schema
// itself doesn't compile.
func ValidateAgainst(schema any, doc any) ([]string, error) {
	c := jsonschema.NewCompiler()
	if err := c.AddResource("inline.json", schema); err != nil {
		return nil, err
	}
	s, err := c.Compile("inline.json")
	if err != nil {
		return nil, err
	}
	return violations(s, doc), nil
}

var printer = message.NewPrinter(language.English)

func violations(s *jsonschema.Schema, doc any) []string {
	err := s.Validate(doc)
	if err == nil {
		return nil
	}
	ve, ok := err.(*jsonschema.ValidationError)
	if !ok {
		return []string{err.Error()}
	}
	var out []string
	var walk func(e *jsonschema.ValidationError)
	walk = func(e *jsonschema.ValidationError) {
		if len(e.Causes) == 0 {
			loc := "<root>"
			if len(e.InstanceLocation) > 0 {
				loc = strings.Join(e.InstanceLocation, "/")
			}
			out = append(out, loc+": "+e.ErrorKind.LocalizedString(printer))
		}
		for _, c := range e.Causes {
			walk(c)
		}
	}
	walk(ve)
	return out
}
