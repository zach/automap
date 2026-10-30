# 10: automap extend

**What to build:** `automap extend <library>[:<variant>] [--slot <slot>]...` adds a library to an existing composition: it fills the slots, writes the library's build fragment, bindings, setup glue and any pairing glue with libraries already present, adds any shared interface the library needs, updates the choices header, startup, lock and `AGENTS.md` section, and ends with `verify`.

**Blocked by:** [05: Probe harness and the probed level](05-probe-harness-and-probed-level.md), [06: Agent files: AGENTS.md section and router skill](06-agent-files.md)

**Status:** done

- [x] Golden tests on the fixture map cover extending a composition with a stub library, including pairing glue appearing when both of a pairing's libraries are present.
- [x] Adding a library whose license class is copyleft or proprietary emits `LICENSE_WARNING(<library>)` with the map's reason before writing files (ADR-0039).
- [x] A filled slot is `SLOT_TAKEN(<slot>)`.
- [x] The `AGENTS.md` section lists the new slots.

## Comments

- `composition.Extend` is a pure function with unit tests. `extend` rewrites only what the CLI generates (manifest, lock, vcpkg files, include file, choices header, startup, `AGENTS.md` section, router skill) and writes new glue where none exists; glue already in the repo is the developer's and is left alone, even where the reference changed. Glue paths come from the manifest, so moved glue is respected.
- Extending stub-window with stub-ui resolves the same lock and startup as starting with both.
- A library already in the composition is `ALREADY_PRESENT(<library>)`, pointing to `upgrade`.
