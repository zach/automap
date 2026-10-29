---
status: accepted
date: 2026-10-29
---

# The developer's own code is a library like any other

## Context

Slots can be filled by third-party libraries or by the developer's own code, such as an in-house renderer or job system. The brief requires private and in-house libraries to be supported "without special cases". Adopted engines bring their own job systems and allocators, which must bind to the same contracts as everything else ([ADR-0003](0003-adopting-existing-engines-is-a-design-constraint.md), [ADR-0011](0011-automap-owns-one-contract-per-seam.md)).

## Options

1. **A separate kind:** the developer's own code is modeled differently from libraries, with its own schema.
2. **A library:** the developer's own code is a library with the same schema as any third-party library.

## Decision

Option 2. The developer's own code is modeled as a library, with the same schema as any third-party library. It fills slots and is bound to contracts exactly like Jolt or SDL3. Its source may simply be a directory in the developer's repo rather than something a build backend fetches.

## Consequences

- Every CLI check applies uniformly to third-party, proprietary and in-house libraries.
- Where a library's source comes from is a property of the library, not a different kind of entity.
