# Open questions and parked branches

Running list for the exploration. Each item names the Phase 2 branch (see `docs/PLAN.md`) that should resolve it. When an item is resolved, move it to **Resolved** with a link to the ADR or `CONTEXT.md` entry. By the end of the exploration only **Deferred past v1** may contain items.

## Open

| ID | Question | Branch | Raised |
| --- | --- | --- | --- |
| OQ-001 | Who owns a contract: the slot, the library, or the pairing? The seed definition hangs it on the slot, but per-library requirements (e.g. "init me before X") also exist. | 3, 4 | 2026-10-29 |
| OQ-002 | Is a pairing strictly two libraries, or can it be N-ary (e.g. SDL3 + SDL_GPU + Dear ImGui's SDL_GPU backend)? | 3, 5 | 2026-10-29 |
| OQ-003 | Does the developer's own code count as filling a slot, and if so is it modeled as a library (it would need contracts too)? | 3, 9 | 2026-10-29 |
| OQ-004 | Is a tier a property of a whole composition, or can one composition mix tiers (a starter engine with one custom-integration slot)? | 3, 7 | 2026-10-29 |
| OQ-005 | Is "layer" (platform, core runtime, engine systems, world and tools, game) a domain term with rules (e.g. dependency direction), or only a presentation of slots? | 3 | 2026-10-29 |
| OQ-006 | `CLAUDE.md` vs `AGENTS.md`: which this repo's build team uses, and which Automap generates for users' projects. This repo uses `CLAUDE.md` for now. Research: AGENTS.md is the cross-tool standard and Claude Code reads it natively; vcpkg's pattern is `CLAUDE.md` = `@AGENTS.md` ([04](research/04-agent-ecosystem.md)). | 8 | 2026-10-29 |
| OQ-007 | Does Automap ship any runtime code (e.g. a neutral scheduler contract, a generated init/shutdown graph), or only the map plus glue generated into the user's repo? ([SUMMARY](research/SUMMARY.md) finding 2) | 2, 4 | 2026-10-29 |
| OQ-008 | Is a web-capable testbed (sokol, raylib, or SDL + WebGPU) in v1, given SDL_GPU has no web backend? ([SUMMARY](research/SUMMARY.md) finding 5) | 7, 13 | 2026-10-29 |
| OQ-010 | How are console compositions verified without NDA access: self-hosted devkits, partner-run verifiers that sign evidence, or unverifiable claims? | 5, 9 | 2026-10-29 |
| OQ-011 | What canonical identity does a composition lock have, so the same composition built through different backends shares one evidence subject? | 5, 6 | 2026-10-29 |

## Parked

_None yet._

## Deferred past v1

_None yet._

## Resolved

| ID | Question | Resolution |
| --- | --- | --- |
| OQ-000 | Glossary file name: `CONTEXT.md` (brief) or `GLOSSARY.md` (current upstream skills)? | `CONTEXT.md`, as the brief requests. A rename later is a `git mv`. |
| OQ-009 | Are non-C++ consumers of the same C libraries (C#, Odin, Zig, Rust bindings) in scope? | v1 non-goal: [ADR-0004](adr/0004-v1-non-goals.md). |
