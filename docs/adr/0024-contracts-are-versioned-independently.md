---
status: accepted
date: 2026-10-29
---

# Each contract is versioned independently

## Context

Contracts will evolve, and every binding and shared interface depends on one ([ADR-0011](0011-automap-owns-one-contract-per-seam.md), [ADR-0023](0023-contract-shape-facts-choices-rules-interface.md)). Research on churn ([03](../research/03-prior-art.md)): a single core version that breaks every few months makes the whole ecosystem re-release, and Bevy's plugins lag each release. Godot keeps old extensions working by checking compatibility against the specific thing that changed.

## Options

1. **One version for all contracts:** the whole contract vocabulary moves together.
2. **Independent semver per contract:** bindings and shared interfaces name the contract version they target.
3. **Unversioned contracts:** contracts change in place with the map.

## Decision

Option 2. Each contract is versioned independently with semver. Bindings and shared interfaces name the contract version they target.

## Consequences

- A breaking change to one contract invalidates only that contract's bindings; a change to the jobs contract doesn't touch allocation or math bindings.
- Moving a composition's glue to a new contract version is part of the upgrade job.
