package mapmodel

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

// ContentHash is the content hash of a directory: the SHA-256 of each
// file's path and SHA-256 in path order, so it changes whenever a file in
// it does. It skips .git and the top-level directories named. With
// normalize, text (a file with no NUL byte) is hashed with CRLF line
// endings read as LF, so a checkout that converts line endings, as git
// does on Windows, keeps the hash.
func ContentHash(dir string, normalize bool, skip ...string) (string, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("not a directory")
	}
	var lines []string
	err = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		if d.IsDir() {
			if d.Name() == ".git" || slices.Contains(skip, filepath.ToSlash(rel)) {
				return filepath.SkipDir
			}
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if normalize && !bytes.Contains(data, []byte{0}) {
			data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
		}
		sum := sha256.Sum256(data)
		lines = append(lines, filepath.ToSlash(rel)+"\x00"+hex.EncodeToString(sum[:]))
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(lines)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// cleanCommit is the commit that last changed a directory in its git
// repository, when the directory has no uncommitted changes: the commit
// whose content is the directory's. It is "" when the directory isn't in a
// repository, was never committed, or has changed since.
func cleanCommit(dir string) string {
	status, err := exec.Command("git", "-C", dir, "status", "--porcelain", "--untracked-files=all", "--", ".").Output()
	if err != nil || len(bytes.TrimSpace(status)) > 0 {
		return ""
	}
	return lastCommit(dir)
}
