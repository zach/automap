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
