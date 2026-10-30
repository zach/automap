package schemas_test

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/zach/automap/schemas"
)

func TestEveryPublishedSchemaIsEmbeddedAndCompiles(t *testing.T) {
	published, err := filepath.Glob("*.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var want []string
	for _, p := range published {
		want = append(want, strings.TrimSuffix(p, ".schema.json"))
	}
	if got := schemas.Names(); !slices.Equal(got, want) {
		t.Fatalf("embedded %v, published %v", got, want)
	}
	for _, name := range want {
		if _, err := schemas.Compile(name); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}
