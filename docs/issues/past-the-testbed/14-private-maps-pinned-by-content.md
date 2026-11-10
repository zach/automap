# 14: Private maps pinned by content

**What to build:** The lock pins each private map by the content of its directory, with the commit beside it when the directory is committed and clean. A private map with uncommitted changes composes, with a warning and unsigned evidence, and a lock checks the same in a shallow clone as in a full one. Field notes [27].

**Blocked by:** 03

**Status:** done

- [x] The lock's `private_maps.<id>` records `content_hash` (the hash an in-repo library gets, over the map's directory) and, when the directory has no uncommitted changes, `commit`.
- [x] A private map with no commit, or with uncommitted changes, composes with a `PRIVATE_MAP_UNCOMMITTED(<id>)` warning, and evidence produced from it is unsigned.
- [x] A lock checks by content: in a shallow clone of the repo, a composition with a private variant passes the lock check. Any change to a pinned private map, committed or not, is `LOCK_MISMATCH(private map <id>)` until `automap resolve`.
- [x] A private variant's pin names the private map and its content hash; a lock a 0.1.0 CLI wrote, with commits only, still checks where the commit matches, and is rewritten with the content hash by the next command that writes the lock.
- [x] The lock schema, its examples and SPEC §4.2 say what the pin is.

## Comments

- How the lock records a private map isn't in an ADR ([ADR-0047](../../adr/0047-private-maps.md) says where private content lives; [ADR-0036](../../adr/0036-what-a-pin-contains.md) is about library pins), so it is the spec's detail to change.
- 0.1.0 read a private map's working tree and pinned the last commit that changed its directory, so uncommitted changes were built but not pinned, and in a shallow clone, where that commit is the clone's one commit, a private variant's pin failed with `LOCK_MISMATCH`.
- The content hash leaves out the map's `evidence/` directory: evidence is about the map, and adding it mustn't change what the lock pins.
- 0.1.0 checked a private map only through the pins of private variants, so a changed private recipe for a library from another registry went unnoticed; every pinned private map is now checked.
