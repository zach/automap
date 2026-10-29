---
status: accepted
date: 2026-10-29
---

# Automap ships no maintenance agent of its own

## Context

Upgrading libraries is the persona's second job ([ADR-0002](0002-core-job-is-extend.md)). Someone has to notice new versions and carry engines forward. The developer's coding agent is the primary steward of the engine's code ([ADR-0005](0005-agent-is-primary-steward.md)), and the persona wants no forced upgrades ([05](../research/05-developer-needs-2026.md)). Automap's shape is the map, the CLI and one router skill ([ADR-0006](0006-product-shape-map-cli-router-skill.md)).

## Options

1. **An Automap maintenance agent:** a hosted bot that watches upstream releases and upgrades developers' engines, for example by opening pull requests.
2. **No Automap maintenance agent:** the developer's own coding agent, or their CI, runs the CLI to upgrade when and as often as they choose.

## Decision

Option 2. Automap ships no maintenance agent of its own. Upgrades in a developer's engine are done by the developer's own coding agent or CI, running the CLI on their schedule.

## Consequences

- Automap never needs write access to developers' repositories.
- Keeping the map and its evidence current is Automap's own work and is decided in branch 11.
