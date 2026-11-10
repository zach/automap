package composition_test

import (
	"os"
	"strings"
	"testing"

	"github.com/zach/automap/internal/composition"
)

const exampleLock = "../../docs/spec/examples/composition/testbed-jolt/automap.lock.json"

func TestTheSpecsExampleLockChecks(t *testing.T) {
	raw, err := os.ReadFile(exampleLock)
	if err != nil {
		t.Fatal(err)
	}
	l, err := composition.CheckLock(raw)
	if err != nil {
		t.Fatal(err)
	}
	if l.Pins["jolt"].Version != "5.6.0" {
		t.Errorf("jolt pinned at %s, want 5.6.0", l.Pins["jolt"].Version)
	}
}

func TestEncodingTheExampleLockReproducesItByteForByte(t *testing.T) {
	raw, _ := os.ReadFile(exampleLock)
	l, err := composition.CheckLock(raw)
	if err != nil {
		t.Fatal(err)
	}
	got, err := composition.EncodeLock(*l)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(raw) {
		t.Errorf("got\n%s\nwant\n%s", got, raw)
	}
}

func TestAHandEditedLockFails(t *testing.T) {
	raw, _ := os.ReadFile(exampleLock)
	for name, edited := range map[string]string{
		"changed pin":    strings.Replace(string(raw), `"version":"5.6.0"`, `"version":"5.7.0"`, 1),
		"pretty-printed": strings.Replace(string(raw), `{"contracts"`, "{\n\"contracts\"", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := composition.CheckLock([]byte(edited)); err == nil {
				t.Error("an edited lock checked")
			}
		})
	}
}

// A private map still matches its pin when its content does; a pin a CLI
// before 0.2.0 wrote, a commit alone, matches by commit.
func TestAPrivateMapPinMatchesByContent(t *testing.T) {
	for _, c := range []struct {
		name            string
		pin             composition.PrivateMapPin
		content, commit string
		want            bool
	}{
		{"same content, another commit", composition.PrivateMapPin{ContentHash: "sha256:aa", Commit: "c1"}, "sha256:aa", "c2", true},
		{"same content, no commit now", composition.PrivateMapPin{ContentHash: "sha256:aa", Commit: "c1"}, "sha256:aa", "", true},
		{"changed content, same commit", composition.PrivateMapPin{ContentHash: "sha256:aa", Commit: "c1"}, "sha256:bb", "c1", false},
		{"a commit alone, the same", composition.PrivateMapPin{Commit: "c1"}, "sha256:bb", "c1", true},
		{"a commit alone, another", composition.PrivateMapPin{Commit: "c1"}, "sha256:bb", "c2", false},
	} {
		if got := c.pin.Matches(c.content, c.commit); got != c.want {
			t.Errorf("%s: Matches is %v, want %v", c.name, got, c.want)
		}
	}
}
