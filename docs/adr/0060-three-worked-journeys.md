---
status: accepted
date: 2026-10-30
---

# The spec's three worked journeys

## Context

The domain model has no tiers; three worked example journeys meet the brief's scalability check and definition of done instead ([ADR-0018](0018-no-tiers-in-the-domain-model.md)). The definition of done also requires private and proprietary libraries to be demonstrated "with a hypothetical in-house library as the example, without special cases". v1 ships five jobs: start, extend, upgrade, adopt and eject ([ADR-0059](0059-v1-jobs.md)). Private content lives in private maps, one per NDA platform ([ADR-0047](0047-private-maps.md), [ADR-0048](0048-one-private-map-per-nda-platform.md)).

## Options

1. **Three journeys that together cover all five v1 jobs and every kind of library.**
2. **Three independent greenfield journeys** of increasing size.

## Decision

Option 1. The spec's three worked journeys are:

1. **Testbed:** start with SDL3 and Dear ImGui, then extend with Jolt ([ADR-0057](0057-day-one-composition.md)), on all three v1 platforms.
2. **Starter engine:** grow that testbed through the rest of the v1 catalog ([ADR-0058](0058-v1-catalog.md)) into a layered engine, then upgrade one library across a real new release.
3. **Custom integration:** adopt a hypothetical studio's hand-built engine whose in-house job system and renderer are libraries described in a private map; add a patched SDL3 variant and a hypothetical NDA console's private map, with evidence signed by the studio; then eject.

## Consequences

- Together the journeys cover all five v1 jobs and every kind of library: public, in-house, patched and console.
- Writing them up tests the design for gaps.
