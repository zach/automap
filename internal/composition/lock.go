package composition

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/zach/automap/internal/canonical"
)

// EncodeLock returns the lock in canonical form (RFC 8785) with its
// lock_hash set: the SHA-256 of the canonical lock without lock_hash
// (ADR-0042).
func EncodeLock(l Lock) ([]byte, error) {
	l.LockHash = ""
	body, err := canonical.Marshal(l)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(body)
	l.LockHash = "sha256:" + hex.EncodeToString(sum[:])
	return canonical.Marshal(l)
}

// CheckLock decodes a lock file and checks that it is in canonical form and
// its lock_hash matches, which together mean only the CLI wrote it
// (ADR-0046).
func CheckLock(raw []byte) (*Lock, error) {
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("the lock is not valid JSON: %v", err)
	}
	stated, _ := doc["lock_hash"].(string)
	delete(doc, "lock_hash")
	body, err := canonical.Marshal(doc)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(body)
	if want := "sha256:" + hex.EncodeToString(sum[:]); stated != want {
		return nil, fmt.Errorf("its lock_hash is %s, but the lock hashes to %s", stated, want)
	}
	if canon, err := canonical.Transform(raw); err != nil || string(canon) != string(raw) {
		return nil, fmt.Errorf("it is not in canonical form")
	}
	var l Lock
	if err := json.Unmarshal(raw, &l); err != nil {
		return nil, err
	}
	return &l, nil
}
