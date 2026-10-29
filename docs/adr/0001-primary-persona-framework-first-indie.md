---
status: accepted
date: 2026-10-29
---

# Primary persona: the framework-first indie

## Context

Automap could be aimed at several kinds of developer, and the primary persona decides what "immediately useful" means, what v1 ships and what the first vertical slice proves. Phase 1 research ([SUMMARY](../research/SUMMARY.md), [05](../research/05-developer-needs-2026.md)) found:

- The pure C++ custom-engine audience is narrow: 40% of game developers used C# in the past year against 29% for C++, and custom engines made about 13% of 2024 Steam releases.
- Developers who ship games on their own library-built engines exist and state what they want: control, a small surface, no forced upgrades, vendored and readable code. Managing third-party libraries is C++ developers' top pain (about 47% call it major).
- Jam developers and learners mostly use Godot, raylib or LÖVE, and jams effectively require browser builds, which an SDL_GPU stack cannot produce.
- Console-bound studios need private, per-platform overlays that depend on platform-holder and vendor cooperation.

## Options

1. **Framework-first indie** (solo or small team, C/C++ literate, building their own engine from libraries, working with a coding agent).
2. **Jam developer or learner** wanting a testbed in minutes, often for the web.
3. **Small studio with custom tech going to consoles.**
4. **Engine extender** working inside Godot (GDExtension, LibGodot).
5. **Agent-only builder** who directs an agent but doesn't read the C++ it writes.

## Decision

The primary persona is the **framework-first indie** (option 1). Jam developers and learners are served by the testbed tier as an on-ramp, not designed around. Console-bound studios are designed for (the private-overlay model must work for them) but are not the primary persona for v1. Engine extenders and agent-only builders are not primary; whether they are non-goals is decided separately.

## Consequences

- "Immediately useful" is judged against this persona's jobs, not a jam developer's: known-good pairings, seam glue they own and can read, reproducible and vendoring-friendly output, and safe upgrades.
- Web builds are not a v1 requirement by default; a web-capable testbed remains an open question (OQ-008).
- The private-overlay design (branch 9) must accommodate console studios even though they ship later.
- The persona works with a coding agent, so the agent interface (branch 8) is a primary surface, but the human must be able to read and own every line of glue.
