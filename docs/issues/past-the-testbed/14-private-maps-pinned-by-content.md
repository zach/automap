# 14: Private maps pinned by content

**What to build:** The lock pins each private map by the content of its directory, with the commit beside it when the directory is committed and clean. A private map with uncommitted changes composes, with a warning and unsigned evidence, and a lock checks the same in a shallow clone as in a full one. Field notes [27].

**Blocked by:** 03

**Status:** ready-for-agent

- [ ] The lock's `private_maps.<id>` records `content_hash` (the hash an in-repo library gets, over the map's directory) and, when the directory has no uncommitted changes, `commit`.
- [ ] A private map with no commit, or with uncommitted changes, composes with a `PRIVATE_MAP_UNCOMMITTED(<id>)` warning, and evidence produced from it is unsigned.
- [ ] A lock checks by content: in a shallow clone of the repo, a composition with a private variant passes the lock check and verifies.
- [ ] A private variant's pin names the private map and its content hash; a lock a 0.1.0 CLI wrote, with commits only, still checks where the commit matches, and is rewritten with the content hash by the next command that writes the lock.
- [ ] The lock schema, its examples and SPEC §4.2 say what the pin is.

## Comments

- How the lock records a private map isn't in an ADR ([ADR-0047](../../adr/0047-private-maps.md) says where private content lives; [ADR-0036](../../adr/0036-what-a-pin-contains.md) is about library pins), so it is the spec's detail to change.
- 0.1.0 read a private map's working tree and pinned the last commit that changed its directory, so uncommitted changes were built but not pinned, and in a shallow clone, where that commit is the clone's one commit, a private variant's pin failed with `LOCK_MISMATCH`.
