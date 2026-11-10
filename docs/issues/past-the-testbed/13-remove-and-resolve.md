# 13: automap remove and automap resolve

**What to build:** `automap remove <library>` takes a library out of the composition and deletes its files the developer hasn't changed; `automap resolve` re-resolves against the lock's own map release after the manifest or a private map changed. Field notes [29, 44].

**Blocked by:** 03

**Status:** ready-for-human

- [ ] The product owner agrees an ADR for the two commands.
- [ ] `remove` empties the library's slots, drops its variant and glue entries, resets choices that named its slots, re-resolves, deletes its glue and build fragment where they still match the reference, warns `GLUE_LEFT(<file>)` for the rest, and verifies.
- [ ] `resolve` keeps the map release, re-resolves, updates unchanged glue from a changed private map, and verifies.
- [ ] The usage, SPEC §5.4 and the router skill name both.
