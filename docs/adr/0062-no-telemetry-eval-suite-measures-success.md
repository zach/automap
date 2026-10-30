---
status: accepted
date: 2026-10-30
---

# No telemetry; an agent eval suite in Automap's CI measures success

## Context

Only fetching needs the network ([ADR-0038](0038-only-fetching-needs-the-network.md)), private maps are never uploaded ([ADR-0047](0047-private-maps.md)), and Automap is open source and non-commercial, pitched on nothing being taken away ([ADR-0054](0054-open-source-non-commercial.md)). Telemetry would contradict all three. What matters is whether an agent with only the CLI and router skill succeeds at the jobs. Scripted agent tasks verified by builds and tests are how SWE-bench-style benchmarks and Summer Engine's routing evals already measure that ([04](../research/04-agent-ecosystem.md)).

## Options

1. **Opt-in telemetry from developers' CLIs.**
2. **No telemetry; an eval suite in Automap's CI.**

## Decision

Option 2. Automap collects no telemetry. Success is measured by an **eval suite** in Automap's CI: fresh agent sessions carry out the five v1 jobs along the three worked journeys ([ADR-0059](0059-v1-jobs.md), [ADR-0060](0060-three-worked-journeys.md)) on every map release.

## Consequences

- The CLI never sends anything about a developer or their composition anywhere.
- Every success measure must be computable from the eval suite or from the public map.
