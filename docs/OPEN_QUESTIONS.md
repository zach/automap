# Open questions and parked branches

Running list for the exploration. Each item names the Phase 2 branch (see `docs/PLAN.md`) that should resolve it. A resolved item is removed; its answer lives in an ADR or `CONTEXT.md`. By the end of the exploration only **Deferred past v1** may contain items.

## Open

| ID | Question | Branch | Raised |
| --- | --- | --- | --- |
| OQ-006 | Which agent instruction files (`CLAUDE.md`, `AGENTS.md`) the repo where Automap itself is built uses, and what they contain (spec section 10). This exploration repo uses `CLAUDE.md`. | 15 | 2026-10-29 |

## Parked

_None yet._

## Deferred past v1

| ID | Question | Why deferred | Deferred |
| --- | --- | --- | --- |
| OQ-011 | What canonical identity does a composition lock have, so the same composition built through different build backends shares one evidence subject? | v1 has one build backend ([ADR-0034](adr/0034-vcpkg-is-the-v1-build-backend.md)). | 2026-10-29 |
| OQ-008 | Is a web-capable testbed (sokol, raylib, or SDL + WebGPU) supported, given SDL_GPU has no web backend? | v1 is desktop only ([ADR-0056](adr/0056-v1-platforms-desktop-only.md)). | 2026-10-29 |
