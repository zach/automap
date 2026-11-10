# 03: Upgrade carries a private map's changes

**What to build:** `upgrade` compares against each private map as the lock pinned it, so a private recipe's changed reference glue reaches the repo where the developer's copy still matches the old reference; and each lock problem is reported once. Field notes [29].

**Blocked by:** None

**Status:** ready-for-agent

- [ ] With a private map whose recipe glue changed since the lock's commit, `upgrade <library>` updates the repo's copy that matched the old reference and counts it; a copy the developer changed is left with `GLUE_DIVERGED` and the new reference beside it.
- [ ] A private map outside a git repository, or one whose pinned commit is gone, is read from its working tree for both sides, as in 0.1.0.
- [ ] A pin that differs from the map is one `LOCK_MISMATCH(<library>)` failure, however many platforms the composition has.
- [ ] When the composition as locked can't be regenerated, the summary says the lock and the maps disagree and names the map the failure comes from, not the public release.
- [ ] Golden test `upgrade-private-glue`.
