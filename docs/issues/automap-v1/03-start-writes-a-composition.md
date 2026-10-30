# 03: Start writes a composition

**What to build:** `automap start --with <library>[:<variant>]...` on the fixture map writes a complete composition into the current directory: manifest, canonical lock, vcpkg manifest and configuration, one build fragment per library and the include file the developer's CMake project includes, the shared interfaces and probe harness the map ships, each recipe's reference glue and setup glue, the choices header, and the startup that calls every entry point in phase order with the context (ADR-0080). It doesn't build yet.

**Blocked by:** [02: Map model on the fixture map](02-map-model-on-fixture-map.md)

**Status:** done

- [x] In a repo with no CMake project, `start` writes the top-level CMake file and `main.cpp` from the map's testbed template and records the executable in the manifest; the fixture map has a testbed template.
- [x] A library fills every slot it can unless `--slot <slot>=<library>` assigns slots; a slot already filled is `SLOT_TAKEN(<slot>)`.
- [x] Resolution and lock canonicalization are pure functions with unit tests; the lock is in RFC 8785 canonical form with a correct `lock_hash` (ADR-0042).
- [x] `map_release` records `git:<commit>` for a map read from a clean checkout and `unreleased` otherwise (ADR-0079).
- [x] Every written file matches golden files; `start` in a directory that already holds a composition refuses with `COMPOSITION_EXISTS`.
- [x] Any command run against a lock whose hash doesn't match refuses with `LOCK_EDITED` and exit status 3 (ADR-0046).
- [x] The startup calls every init entry point in phase order, respecting `after` lists and ordering constraints, and every context field read is provided earlier; otherwise `start` reports `CONTEXT_UNPROVIDED(<field>)`.
- [x] The choices header carries every choice from the manifest, with contract defaults filled in, and glue reads choices only through it (ADR-0074).
- [x] Glue and build files refer to libraries only by CMake package and imported target (ADR-0035).

## Comments

- `start` writes the startup as `glue/automap/context.h`, `startup.h` and `startup.cpp`, listed in the manifest's `glue.startup`. Context fields of pointer type are forward-declared and qualified, so `context.h` needs no library headers.
- `vcpkg.json` carries the curated registry's baseline as `builtin-baseline`; `vcpkg-configuration.json` lists the Automap registry (its `registry` branch, at the map's commit) for the ports it supplies. Ticket 04 checks both against a real vcpkg.
- The composition `start` writes from the fixture map passes the spec's `check-examples.py` composition rules.
