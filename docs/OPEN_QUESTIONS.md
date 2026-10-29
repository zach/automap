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
| OQ-006 | `CLAUDE.md` vs `AGENTS.md`: which this repo's build team uses, and which Automap generates for users' projects. This repo uses `CLAUDE.md` for now. | 8 | 2026-10-29 |

## Parked

_None yet._

## Deferred past v1

_None yet._

## Resolved

| ID | Question | Resolution |
| --- | --- | --- |
| OQ-000 | Glossary file name: `CONTEXT.md` (brief) or `GLOSSARY.md` (current upstream skills)? | `CONTEXT.md`, as the brief requests. A rename later is a `git mv`. |
