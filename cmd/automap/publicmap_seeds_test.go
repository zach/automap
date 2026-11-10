package main_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/zach/automap/internal/clitest"
	"github.com/zach/automap/internal/composition"
	"github.com/zach/automap/internal/mapmodel"
	"github.com/zach/automap/internal/verify"
)

// Every seeded violation in the public map still applies to the reference
// glue it breaks: a patch written against glue that has since changed
// would leave its probe unproven (ADR-0071). Each reference composition is
// written without building, and each of its seeds applied to a copy.
func TestEveryPublicSeedAppliesToItsReferenceGlue(t *testing.T) {
	var refs []struct {
		ID   string
		With []string
	}
	raw, err := os.ReadFile(filepath.Join(clitest.Root(t), "ci", "compositions.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &refs); err != nil {
		t.Fatal(err)
	}
	applied := map[string]bool{}
	for _, ref := range refs {
		repo := testbedRepo(t)
		args := []string{"start", "--map", "map", "--no-verify", "--json"}
		for _, w := range ref.With {
			args = append(args, "--with", w)
		}
		repo.Automap(t, args...).ExpectExit(t, 0)
		m, problems := mapmodel.Load(filepath.Join(repo.Dir, "map"), nil)
		if len(problems) > 0 {
			t.Fatalf("%s: the public map: %v", ref.ID, problems)
		}
		var man composition.Manifest
		json.Unmarshal([]byte(clitest.Read(t, repo, "automap.manifest.json")), &man)
		for _, seed := range verify.Seeds(m, &man) {
			name := seed.Owner + " " + seed.Rule
			if applied[name] {
				continue
			}
			scratch := t.TempDir()
			if err := os.CopyFS(scratch, os.DirFS(repo.Dir)); err != nil {
				t.Fatal(err)
			}
			if err := verify.ApplySeed(scratch, seed); err != nil {
				t.Errorf("%s: the seeded violation of %s in %s doesn't apply: %v", ref.ID, seed.Rule, seed.Owner, err)
			}
			applied[name] = true
		}
	}
	if len(applied) < 30 {
		t.Errorf("%d seeds applied; the public map has more than that", len(applied))
	}
}
