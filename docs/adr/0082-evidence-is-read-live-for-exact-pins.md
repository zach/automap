---
status: accepted
date: 2026-10-30
---

# Map releases freeze facts, recipes and probes; evidence is read live for exact pins

## Context

The public map ships as immutable, dated releases recorded in every lock ([ADR-0053](0053-map-releases-are-immutable-and-recorded-in-locks.md)), and published evidence is stored in the map ([ADR-0030](0030-evidence-is-signed-and-stored-in-the-map.md)). Evidence is current for 30 days, then stale ([ADR-0031](0031-evidence-states-current-stale-broken.md)); older versions retire by no longer being re-verified, so their evidence goes stale ([ADR-0033](0033-automap-ci-verifies-declared-edges-latest-and-previous.md)); and published evidence is one of the things the CLI fetches ([ADR-0038](0038-only-fetching-needs-the-network.md)). If evidence were frozen inside a release, a lock on an older release would show evidence growing stale even while Automap's CI kept re-verifying its pins.

## Options

1. **Evidence frozen in each release.**
2. **Evidence read live:** releases freeze facts, recipes and probes; the CLI fetches evidence for a lock's exact pins from the map's main branch.

## Decision

Option 2. A map release freezes facts, recipes and probes. Evidence names exact pins, so the CLI fetches the evidence for a lock's pins from the map's main branch, where Automap's CI keeps adding it, whatever release the lock records.

## Consequences

- A composition on an older map release still sees current evidence for its pins while Automap's CI re-verifies them.
- Versions retire by going stale once CI stops re-verifying them.
