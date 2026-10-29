# Open questions and parked branches

Running list for the exploration. Each item names the Phase 2 branch (see `docs/PLAN.md`) that should resolve it. A resolved item is removed; its answer lives in an ADR or `CONTEXT.md`. By the end of the exploration only **Deferred past v1** may contain items.

## Open

| ID | Question | Branch | Raised |
| --- | --- | --- | --- |
| OQ-006 | `CLAUDE.md` vs `AGENTS.md`: which this repo's build team uses, and which Automap generates for users' projects. This repo uses `CLAUDE.md`. Research: AGENTS.md is the cross-tool standard and Claude Code reads it natively; vcpkg's pattern is `CLAUDE.md` = `@AGENTS.md` ([04](research/04-agent-ecosystem.md)). | 8 | 2026-10-29 |
| OQ-008 | Is a web-capable testbed (sokol, raylib, or SDL + WebGPU) in v1, given SDL_GPU has no web backend? ([SUMMARY](research/SUMMARY.md) finding 5) | 7, 13 | 2026-10-29 |
| OQ-010 | How are console compositions verified without NDA access: self-hosted devkits, partner-run verifiers that sign evidence, or unverifiable claims? | 5, 9 | 2026-10-29 |
| OQ-012 | Should the agent be barred from opening issues or PRs on upstream projects whose policies ban AI-generated contributions (SDL, SDL_shadercross, sokol, stb, PhysicsFS, Godot)? Policy, not capability ([SUMMARY](research/SUMMARY.md) finding 4). | 8 | 2026-10-29 |
| OQ-013 | Which functions may rely on a remote hosted registry or package repository, and which must work offline? Research lists reproducible offline builds and vendoring-friendly output as must-haves for the persona ([05](research/05-developer-needs-2026.md)). | 6, 9 | 2026-10-29 |

## Parked

_None yet._

## Deferred past v1

| ID | Question | Why deferred | Deferred |
| --- | --- | --- | --- |
| OQ-011 | What canonical identity does a composition lock have, so the same composition built through different build backends shares one evidence subject? | v1 has one build backend ([ADR-0034](adr/0034-vcpkg-is-the-v1-build-backend.md)). | 2026-10-29 |
