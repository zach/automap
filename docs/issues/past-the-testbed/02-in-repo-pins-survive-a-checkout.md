# 02: In-repo library pins survive a checkout

**What to build:** An in-repo library's content hash is the same on a Windows checkout with CRLF line endings as on the Linux machine that wrote the lock. Field notes [39].

**Blocked by:** None

**Status:** done

- [x] `SourceHash` gives one hash for a directory whose text files have LF endings and for the same directory with CRLF endings; a binary file (one with a NUL byte) is hashed as it is.
- [x] A lock whose pin is the 0.1.0 raw hash still checks, and the next command that writes the lock writes the normalized hash.
- [x] A changed file still fails with `LOCK_MISMATCH(<library>)`.
- [x] SPEC says what an in-repo pin hashes.

## Comments

- `composition.SourceHash` normalizes text; `composition.SourceMatches`, which `Check` uses, also accepts the hash of the raw bytes, the pin a 0.1.0 CLI wrote, and `Resolve` writes the normalized hash whenever it writes the lock.
