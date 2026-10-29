---
status: accepted
date: 2026-10-29
---

# A composition is a manifest and a lock in the developer's repo

## Context

A composition needs a concrete form the agent can edit and the CLI can act on. Adoption needs a place to declare where existing glue lives ([ADR-0003](0003-adopting-existing-engines-is-a-design-constraint.md)). Research found that evidence has to be tied to one exact resolution on one platform ([06](../research/06-verification.md)). Package managers separate what is requested from what was resolved: vcpkg's manifest and baseline, Conan's lockfiles ([01](../research/01-package-managers.md)).

## Options

1. **One file:** a single file holds both intent and exact resolution.
2. **A manifest and a lock:** the agent edits a manifest of intent; the CLI writes a lock with the exact resolution.
3. **No file:** the composition is inferred from the developer's build files each time.

## Decision

Option 2. A composition is two files in the developer's repo:

1. **A manifest** that the agent edits. For each slot it names the library filling it, and for each binding and pairing it says where that glue lives.
2. **A lock** that the CLI writes, holding the exact resolved pins.

## Consequences

- Evidence is tied to a lock.
- The file format and what a pin contains (for example version or commit, build variant, build backend) are decided in later branches.
