package mapmodel_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zach/automap/internal/mapmodel"
)

// A private map's content hash leaves out its evidence, which is about
// the map and not of it: adding evidence doesn't change what the lock
// pins.
func TestContentHashSkipsTheDirectoriesNamed(t *testing.T) {
	dir := t.TempDir()
	write := func(name, data string) {
		p := filepath.Join(dir, filepath.FromSlash(name))
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("map.json", "{}")
	write("recipes/library/a.json", "{}")
	before, err := mapmodel.ContentHash(dir, true, "evidence")
	if err != nil {
		t.Fatal(err)
	}
	write("evidence/binding/a/linux.json", "{}")
	if after, _ := mapmodel.ContentHash(dir, true, "evidence"); after != before {
		t.Errorf("the hash changed when evidence was added")
	}
	write("recipes/evidence/x.json", "{}")
	if after, _ := mapmodel.ContentHash(dir, true, "evidence"); after == before {
		t.Errorf("a directory named evidence below the top level was skipped too")
	}
}
