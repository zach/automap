# 03: Start writes a composition

**What to build:** `automap start --with <library>[:<variant>]...` on the fixture map writes a complete composition into the current directory: manifest, canonical lock, vcpkg manifest and configuration, one build fragment per library and the include file the developer's CMake project includes, the shared interfaces and probe harness the map ships, each recipe's reference glue and setup glue, and the choices header. It doesn't build yet.

**Blocked by:** [02: Map model on the fixture map](02-map-model-on-fixture-map.md)

**Status:** ready-for-agent

- [ ] A library fills every slot it can unless `--slot` names some; a slot already filled is `SLOT_TAKEN(<slot>)`.
- [ ] Resolution and lock canonicalization are pure functions with unit tests; the lock is in RFC 8785 canonical form with a correct `lock_hash` (ADR-0042).
- [ ] `map_release` records `git:<commit>` for a map read from a clean checkout and `unreleased` otherwise (ADR-0079).
- [ ] Every written file matches golden files; `start` in a directory that already holds a composition refuses with `COMPOSITION_EXISTS`.
- [ ] Any command run against a lock whose hash doesn't match refuses with `LOCK_EDITED` and exit status 3 (ADR-0046).
- [ ] The choices header carries every choice from the manifest, with contract defaults filled in, and glue reads choices only through it (ADR-0074).
- [ ] Glue and build files refer to libraries only by CMake package and imported target (ADR-0035).
