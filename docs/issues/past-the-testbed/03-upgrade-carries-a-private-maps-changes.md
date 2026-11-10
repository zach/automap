# 03: Upgrade carries a private map's changes

**What to build:** `upgrade` compares against each private map as the lock pinned it, so a private recipe's changed reference glue reaches the repo where the developer's copy still matches the old reference; and each lock problem is reported once. Field notes [29].

**Blocked by:** None

**Status:** done

- [x] With a private map whose recipe glue changed since the lock's commit, `upgrade <library>` updates the repo's copy that matched the old reference and counts it; a copy the developer changed is left with `GLUE_DIVERGED` and the new reference beside it.
- [x] A private map whose pinned commit can't be read, as in a shallow clone, is read from its working tree for both sides, as in 0.1.0.
- [x] A pin that differs from the map is one `LOCK_MISMATCH(<library>)` failure, however many platforms the composition has.
- [x] When the composition as locked can't be regenerated, the summary names the release and private maps the lock pinned, and `NO_BUILD_FRAGMENT(<library>)` names the map that defines the library.
- [x] Process-level tests cover the carried change, the diverged copy, the shallow clone and the single failure.

## Comments

- `upgrade` extracts each private map at the lock's commit with `git archive` into a scratch directory and computes the old files against it; when the map can't be read there, or the composition can't be planned against it, it falls back to the working tree, which is what 0.1.0 always did.
- `Check` reported a pin's problem once per platform of the composition; it now says each once.
- Found on the way: in a shallow clone, which is how CI checks a repo out, `git log -1 -- <private map>` names the clone's one commit, so a private variant's pin fails with `LOCK_MISMATCH` there. Ticket 14 fixes it by pinning a private map by its content.
