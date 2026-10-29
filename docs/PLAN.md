# Exploration plan

Source: `docs/EXPLORATION_BRIEF.md`. This file tracks where the exploration stands. Update the status column as work lands.

## Phases

| Phase | What | Output | Status |
| --- | --- | --- | --- |
| 0. Orientation | Install skills, seed glossary, plan, start research | `CLAUDE.md`, `CONTEXT.md`, `docs/agents/`, this file | Done 2026-10-29 |
| 1. Research | Eight parallel research agents (topic 2 split three ways) | `docs/research/*.md`, `SUMMARY.md`, top-5 walkthrough | Done 2026-10-29 (follow-ups listed in `SUMMARY.md`) |
| 2. Grilling | Decision tree below, one question at a time | ADRs, `CONTEXT.md`, `OPEN_QUESTIONS.md` | In progress: branch 11 |
| 3. Spikes | Only for questions words can't settle; plus the planned testbed spike | `spikes/<name>/` → ADR | Not started |
| 4. Synthesis | `SPEC.md`, `/to-spec`, `/to-tickets`, definition-of-done review | `docs/spec/SPEC.md`, `docs/issues/` | Not started |

## Phase 1 research topics

| File | Topic |
| --- | --- |
| `01-package-managers.md` | vcpkg, Conan 2, xmake/xrepo, CPM/FetchContent: what they can express; which to build on |
| `02a-catalog-platform-graphics-ui-audio.md` | SDL3/SDL_GPU, raylib, sokol, bgfx, WebGPU; shader toolchains; Dear ImGui and friends; audio incl. FMOD/Wwise |
| `02b-catalog-assets-and-data.md` | Asset libraries, serialization and reflection, asset-pipeline patterns and reference systems |
| `02c-catalog-simulation-runtime-glue.md` | Physics, navigation, ECS, core runtime, math, profiling, scripting, networking; seams catalog |
| `03-prior-art.md` | Bevy/cargo, Godot GDExtension/LibGodot, sokol/fips, Magnum, Cute, O3DE Gems, Unreal modules, Our Machinery, others |
| `04-agent-ecosystem.md` | Agent Skills, routers, MCP, AGENTS.md/CLAUDE.md, game-dev skill packs, agents writing C++ |
| `05-developer-needs-2026.md` | Engine landscape, platforms and consoles, iteration, determinism, licensing, agent trust |
| `06-verification.md` | CI matrices and cost, headless graphics, golden images, pruning N×M, evidence records, staleness |

## Phase 2 decision tree

Arrows point from a branch to the branches that depend on it. Work in dependency order; within a branch, one question at a time.

```
 1 Who & why
 └─► 2 Product shape
     ├─► 3 Domain model
     │   ├─► 4 Contracts ─────────────► 10 Data & asset slots (needs 3, 4)
     │   │   └─► 5 Compatibility & trust
     │   │       └─► 11 Freshness & governance
     │   ├─► 7 Growth & eject (needs 3, 6)
     │   └─► 8 Agent interface (needs 2, 3, 6)
     ├─► 6 Build integration
     │   └─► 9 Extensibility & privacy (needs 3, 5, 6)
     └─► 12 Product & distribution (needs 1, 2, 11)
 13 Scope of v1 (needs all of the above)
 ├─► 14 Success measures (needs 1, 13)
 └─► 15 Agent-driven build plan (needs 13)
```

| # | Branch | Depends on | Status | ADRs |
| --- | --- | --- | --- | --- |
| 1 | Who and why: personas, jobs-to-be-done, primary persona | — | Done | 0001–0005 |
| 2 | Product shape: map, scaffolder, package-manager layer, maintenance agent; day-one minimum | 1 | Done | 0006–0010 |
| 3 | Domain model: schema-ready definitions | 2 | Done | 0011–0020 |
| 4 | Contracts: per-seam expression; machine-checkable vs. guidance | 3 | Done | 0021–0027 |
| 5 | Compatibility and trust: declaring, verifying, versioning, retiring pairings | 3, 4 | Done | 0028–0033 |
| 6 | Build integration: backend(s), composition → build, pinning, supply chain | 2 | Done | 0034–0039 |
| 7 | Growth and eject: testbed to large composition without a rewrite; leaving Automap | 3, 6 | Done | 0040 |
| 8 | Agent interface: skills, router, manifest, CLI, MCP, generated AGENTS.md | 2, 3, 6 | Done | 0041–0046 |
| 9 | Extensibility and privacy: private libraries, private maps, NDA platforms, offline | 3, 5, 6 | Done | 0038, 0047–0049 |
| 10 | Data and asset slots: first-class slots with contracts? | 3, 4 | Done | 0050 |
| 11 | Freshness and governance: curation, contributions, staying current | 5 | In progress | 0051, 0052 |
| 12 | Product and distribution: open source, commercial, hybrid; earning trust | 1, 2, 11 | Not started | |
| 13 | Scope of v1: slots, libraries, platforms, worked example journeys; first vertical slice | all | Not started | |
| 14 | Success measures | 1, 13 | Not started | |
| 15 | Agent-driven build plan: how Automap itself is built by agents (spec section 10): repo conventions, instruction files, build-team skills, test strategy, human checkpoints | 13 | Not started | |

## Planned spike

From a draft manifest and recipes, a fresh agent session generates an SDL3 + SDL_GPU + Dear ImGui + Jolt testbed; record where it went wrong.
