---
status: accepted
date: 2026-10-29
---

# The lifecycle contract has six fixed phases

## Context

Every composition ecosystem that thrived owns a named phase vocabulary with reverse teardown: Unreal's loading phases and host types, Godot's four initialization levels, Bevy's plugin build and finish ([03](../research/03-prior-art.md)). Real library constraints fit a small set ([02a](../research/02a-catalog-platform-graphics-ui-audio.md), [02c](../research/02c-catalog-simulation-runtime-glue.md)): SDL wants `SDL_SetMemoryFunctions` called before anything else; GameNetworkingSockets takes its allocator once, before everything; Jolt needs its allocator, then trace and assert hooks, then its type factory, then its job system; Dear ImGui's backends need the GPU device.

## Options

1. **Free-form ordering:** each library declares dependencies on others and the order is solved per composition.
2. **Fixed phases with dependencies inside them:** a short fixed list of named phases; each library's init steps are placed in one, and declared dependencies order steps within a phase.
3. **Many fine-grained phases,** in the style of Unreal's loading phases.

## Decision

Option 2. The lifecycle contract defines six fixed phases, in order:

1. **memory**
2. **diagnostics**
3. **platform**
4. **jobs**
5. **systems**
6. **game**

Each library's init steps are placed in a phase and ordered within it by declared dependencies. Shutdown runs in exact reverse. A probe checks the actual order at startup.

## Consequences

- Every composition's startup is readable and comparable.
- Init order is evidence, not a comment in someone's main loop.
