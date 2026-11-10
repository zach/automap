# 13: automap remove and automap resolve

**What to build:** `automap remove <library>` takes a library out of the composition and deletes its files the developer hasn't changed; `automap resolve` re-resolves against the lock's own map release after the manifest or a private map changed. Field notes [29, 44].

**Blocked by:** 03

**Status:** done

- [x] `remove` empties the library's slots, drops its variant and glue entries, resets choices that named its slots, drops the choices of contracts nothing else touches, re-resolves, deletes the files the composition no longer has where they still match the reference, warns `GLUE_LEFT(<file>)` for the rest, and verifies. Removing what `extend` just added leaves the repo as it was, file for file.
- [x] `resolve` keeps the map release and every library's version, re-resolves, updates unchanged glue from a changed private map, and verifies; `LOCK_MISMATCH` failures name it.
- [x] The usage, SPEC §5.4 and the router skill name both.

## Comments

- Command names beyond the five jobs are the spec's to fill in (SPEC §12.3, item 1), as `verify`, `find` and `show` were, so the two commands need no ADR: `remove` is the other half of extend's "add or swap", and `resolve` is upgrade without the move to a newer release.

- `upgrade`, `resolve` and `remove` share one path (`recompose` in `internal/cli/upgrade.go`): the composition's files as the lock pinned them, the new manifest resolved against the map, the copied files updated, left or reported, the files no longer in the composition deleted where unchanged, and verify. `composition.Remove` is the manifest's edit, a pure function.
