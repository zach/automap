package composition

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, data := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// Text hashes the same with LF and CRLF line endings; a binary file, one
// with a NUL byte, is hashed as it is.
func TestSourceHashNormalizesTextLineEndings(t *testing.T) {
	lf := writeTree(t, map[string]string{"a.cpp": "int a;\nint b;\n", "include/a.h": "#pragma once\n", "logo.bin": "\x00\x01\r\n\x02"})
	crlf := writeTree(t, map[string]string{"a.cpp": "int a;\r\nint b;\r\n", "include/a.h": "#pragma once\r\n", "logo.bin": "\x00\x01\r\n\x02"})
	changedBinary := writeTree(t, map[string]string{"a.cpp": "int a;\nint b;\n", "include/a.h": "#pragma once\n", "logo.bin": "\x00\x01\n\x02"})

	a, err := SourceHash(lf)
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := SourceHash(crlf); a != b {
		t.Errorf("CRLF text hashes to %s, LF text to %s; want the same", b, a)
	}
	if c, _ := SourceHash(changedBinary); a == c {
		t.Errorf("a binary file whose CRLF became LF hashes the same; want it hashed as it is")
	}
}

// A pin a CLI wrote before text was normalized, the hash of the raw
// bytes, still matches its source.
func TestSourceMatchesAPinOfRawBytes(t *testing.T) {
	dir := writeTree(t, map[string]string{"a.cpp": "int a;\r\n"})
	raw, err := sourceHash(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	normalized, _ := SourceHash(dir)
	if raw == normalized {
		t.Fatal("the raw and normalized hashes of CRLF text are the same")
	}
	if !SourceMatches(dir, raw) || !SourceMatches(dir, normalized) {
		t.Errorf("SourceMatches rejects the raw or the normalized hash")
	}
	if SourceMatches(dir, "sha256:0000") {
		t.Errorf("SourceMatches accepts a hash that is neither")
	}
}
