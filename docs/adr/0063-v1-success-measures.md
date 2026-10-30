---
status: accepted
date: 2026-10-30
---

# v1 is judged on four measures from the eval suite and the public map

## Context

Success is measured without telemetry, by an eval suite in Automap's CI and from the public map ([ADR-0062](0062-no-telemetry-eval-suite-measures-success.md)). The brief suggested two measures: time from request to running testbed, and the share of agent-generated compositions that build first time. Automap has made four promises the measures should track: day-one usefulness ([ADR-0010](0010-day-one-minimum-start-then-extend-once.md)), evidence in place of human review ([ADR-0005](0005-agent-is-primary-steward.md)), known-good meaning current ([ADR-0031](0031-evidence-states-current-stale-broken.md)), and no lock-in ([ADR-0040](0040-eject-is-one-cli-step.md)). The targets are judgment calls, not research findings.

## Options

1. **The brief's two measures only.**
2. **Four measures, one per promise.**

## Decision

Option 2. v1 is judged on four measures:

1. **Request to running testbed:** a fresh agent session goes from a one-line request to a running testbed with evidence at the probed level in **15 minutes or less** with a warm vcpkg binary cache. Cold-cache time is reported alongside.
2. **First-attempt success:** at least **90%** of eval-suite tasks, across all five jobs and three journeys, end with every probe passing and no human input.
3. **Freshness:** at least **90%** of catalog libraries have their latest upstream release in the map within **14 days** of release, the 7-day cooldown included; at least **95%** of the map's evidence is current at any time.
4. **Eject:** **100%** of ejected journey compositions build and pass their smoke test without Automap installed.

## Consequences

- The eval suite must record timings, per-task outcomes and whether any human input occurred.
- The watcher must record upstream release dates so freshness can be computed.
