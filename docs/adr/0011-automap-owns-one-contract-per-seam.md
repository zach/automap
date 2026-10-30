---
status: accepted
date: 2026-10-29
---

# Automap owns one contract per seam

## Context

Every composition ecosystem that thrived has one core that owns the seams: Bevy's plugin lifecycle, Godot's initialization levels and ABI check, Unreal's loading phases and host types, Our Machinery's API registry ([03](../research/03-prior-art.md)). Third parties then target one versioned thing, and compatibility becomes "works with core version N" rather than a matrix of pairs. In C++ no library plays that role: each brings its own allocator hooks, job model, math types, handle layout, init order and error model ([02c](../research/02c-catalog-simulation-runtime-glue.md)). Bridging every job scheduler to every library is N×M without a neutral interface. Automap is never a runtime dependency, but shared interfaces can be delivered as glue source ([ADR-0008](0008-no-automap-runtime-dependency.md)).

## Options

1. **Contracts owned by slots:** each slot states what its filler must provide.
2. **Contracts owned by libraries:** each library states what it needs from its surroundings; compatibility is checked pair by pair.
3. **Contracts owned by Automap, one per seam:** Automap defines the vocabulary for each seam; recipes bind each library to those contracts.

## Decision

Option 3. Automap defines one contract per seam (for example allocation, jobs, math types, handles, lifecycle, errors, serialization, build flags, platform), acting as the "virtual core" that no C++ library provides. Recipes bind each library to the contracts. Most compatibility is checked library-to-contract (N+M). Pairings cover direct library-to-library integrations that don't go through a contract, such as Dear ImGui's SDL3 backend.

## Consequences

- The contract vocabulary is the core of Automap's content and must be versioned.
- Evidence comes in two kinds: a library bound to a contract, and a direct pairing of libraries.
- The final list of seams and how each contract is expressed are decided in branch 4.
