# 14: Private maps from the working tree

**What to build:** A private map with uncommitted changes is read and marked: the lock records that its pin has uncommitted changes, results warn, and evidence is unsigned. Field notes [27].

**Blocked by:** None

**Status:** ready-for-human

- [ ] The product owner agrees how the lock records a private map with uncommitted changes (it changes [ADR-0047](../../adr/0047-private-maps.md)'s pin).
- [ ] A private map with no commit, or with uncommitted changes, composes with a `PRIVATE_MAP_UNCOMMITTED(<id>)` warning.
- [ ] The lock never names a commit whose content the CLI didn't read.
