# 10: automap extend

**What to build:** `automap extend <library>[:<variant>] [--slot <slot>]...` adds a library to an existing composition: it fills the slots, writes the library's build fragment, bindings, setup glue and any pairing glue with libraries already present, adds any shared interface the library needs, updates the choices header, startup, lock and `AGENTS.md` section, and ends with `verify`.

**Blocked by:** [05: Probe harness and the probed level](05-probe-harness-and-probed-level.md), [06: Agent files: AGENTS.md section and router skill](06-agent-files.md)

**Status:** ready-for-agent

- [ ] Golden tests on the fixture map cover extending a composition with a stub library, including pairing glue appearing when both of a pairing's libraries are present.
- [ ] Adding a library whose license class is copyleft or proprietary emits `LICENSE_WARNING(<library>)` with the map's reason before writing files (ADR-0039).
- [ ] A filled slot is `SLOT_TAKEN(<slot>)`.
- [ ] The `AGENTS.md` section lists the new slots.
