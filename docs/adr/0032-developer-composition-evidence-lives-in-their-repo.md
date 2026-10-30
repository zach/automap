---
status: accepted
date: 2026-10-29
---

# The CLI produces evidence for the developer's own composition

## Context

Automap's CI can verify bindings, pairings and reference compositions, but it can't pre-verify every developer's unique lock. The day-one minimum promises evidence that each pairing in the developer's engine works ([ADR-0010](0010-day-one-minimum-start-then-extend-once.md)), and evidence is what replaces human review when an agent is the steward ([ADR-0005](0005-agent-is-primary-steward.md)). The CLI already runs probes deterministically ([ADR-0006](0006-product-shape-map-cli-router-skill.md)), and published evidence is signed and stored in the map ([ADR-0030](0030-evidence-is-signed-and-stored-in-the-map.md)).

## Options

1. **Automap's CI only:** developers rely on published binding and pairing evidence and get none for their own lock.
2. **The CLI also produces composition evidence for the developer,** stored in their repo.
3. **Developers submit their runs to Automap's map.**

## Decision

Option 2. The CLI produces evidence for the developer's own composition. It runs on the developer's machine or in their CI, at whatever evidence level their platforms allow, and stores the records in their repo beside the lock. Records produced in CI are signed by the developer's CI identity; records produced locally are marked as unsigned. The map's published evidence comes only from Automap's CI.

## Consequences

- A developer's evidence never depends on an Automap service.
- Whether community-submitted runs can ever enter the map is decided with contributions in branch 11.
