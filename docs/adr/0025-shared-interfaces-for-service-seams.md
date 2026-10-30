---
status: accepted
date: 2026-10-29
---

# The service seams have shared interfaces

## Context

A contract may define a shared interface, delivered as glue source ([ADR-0023](0023-contract-shape-facts-choices-rules-interface.md), [ADR-0008](0008-no-automap-runtime-dependency.md)). Shared interfaces pay off where many libraries plug into one engine-wide service through hooks. The research counted at least eight allocator hook styles, four scheduling models (Jolt jobs and barriers, Box2D enqueue/finish, flecs task threads, PhysX dispatcher), and assert hooks that differ in signature and return value ([02c](../research/02c-catalog-simulation-runtime-glue.md)). Godot's production Jolt module routes Jolt's allocator, trace and assert hooks and its job system into Godot's own services. Asset libraries read files through injectable I/O hooks (cgltf, fastgltf, ufbx, libktx, PhysicsFS) ([02b](../research/02b-catalog-assets-and-data.md)). Math libraries are sets of value types, not services.

## Options

1. **No shared interfaces:** every binding adapts a library directly to whatever service the composition chose.
2. **Shared interfaces for the service seams:** allocation, jobs, diagnostics and assets get one each; the other seven contracts have none.
3. **Shared interfaces everywhere:** every contract defines one.

## Decision

Option 2. Exactly four contracts have a shared interface:

- **Allocation:** an allocator facade that every library's hooks route to.
- **Jobs:** a neutral job-system interface that each scheduler implements and each library's job hooks call.
- **Diagnostics:** a log, assert and profile facade.
- **Assets:** a file-I/O interface that every asset library's reads go through ([ADR-0050](0050-assets-is-a-seam-not-a-slot.md)).

The other seven contracts have none. Math types are handled by per-library adapters to the composition's chosen math library.

## Consequences

- Each library needs one adapter per service seam, rather than one per library–service pair.
- Choosing or swapping an allocator, scheduler, diagnostics backend or file system touches only its implementation of the shared interface.
