---
status: accepted
date: 2026-10-29
---

# Primary persona: the framework-first indie

## Context

Automap could be aimed at several kinds of developer, and the primary persona decides what "immediately useful" means, what v1 ships and what the first vertical slice proves. Phase 1 research ([SUMMARY](../research/SUMMARY.md), [05](../research/05-developer-needs-2026.md)) found:

- The pure C++ custom-engine audience is narrow: 40% of game developers used C# in the past year against 29% for C++, and custom engines made about 13% of 2024 Steam releases.
- Developers who ship games on their own library-built engines exist and state what they want: control, a small surface, no forced upgrades, vendored code. Managing third-party libraries is C++ developers' top pain (about 47% call it major).
- Jam developers and learners mostly use Godot, raylib or LÖVE, and jams effectively require browser builds, which an SDL_GPU stack cannot produce.
- Console-bound studios need private, per-platform overlays that depend on platform-holder and vendor cooperation.

## Options

1. **Framework-first indie** (solo or small team building and shipping games on their own engine assembled from libraries, with a coding agent).
2. **Jam developer or learner** wanting a testbed in minutes, often for the web.
3. **Small studio with custom tech going to consoles.**
4. **Engine extender** working inside Godot (GDExtension, LibGodot).
5. **Agent-only builder** who directs an agent but doesn't read the C++ it writes.

## Decision

The primary persona is the **framework-first indie** (option 1): a solo developer or small team who builds and ships games on their own engine assembled from libraries. Their coding agent is the primary steward of the engine's code and the human sets direction; reading the C++ is optional ([ADR-0005](0005-agent-is-primary-steward.md)). Agent-only builders (option 5) are therefore inside the primary persona ([ADR-0004](0004-v1-non-goals.md)).

Jam developers and learners are served by testbeds as an on-ramp, not designed around. Console-bound studios are designed for (private maps must work for them) but are not the primary persona for v1. Engine extenders are a v1 non-goal ([ADR-0004](0004-v1-non-goals.md)).

## Consequences

- "Immediately useful" is judged against this persona's jobs, not a jam developer's: known-good pairings, seam glue that lives in their repo under their control, reproducible and vendoring-friendly output, and safe upgrades.
- Web builds are not a v1 requirement by default; a web-capable testbed remains an open question (OQ-008).
- Private maps (branch 9) must accommodate console studios even though they ship later.
- The coding agent is the persona's primary way of working, so the agent interface (branch 8) is a primary surface.
