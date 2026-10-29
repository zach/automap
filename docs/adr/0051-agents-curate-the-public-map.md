---
status: accepted
date: 2026-10-29
---

# Agents curate the public map under evidence gates; humans decide coverage

## Context

The map has to stay current as dozens of libraries ship. Hand-maintained compatibility metadata decays: O3DE's gem metadata is mostly empty or wrong in its own canonical repo, and Bevy's README compatibility tables lag each release ([03](../research/03-prior-art.md)). Renovate has no vcpkg support, so Automap needs its own watcher for upstream releases ([06](../research/06-verification.md)). Much of a library's seam data can be extracted from headers ([02c](../research/02c-catalog-simulation-runtime-glue.md)). Agents can be primary stewards of code given evidence ([ADR-0005](0005-agent-is-primary-steward.md)), and nothing enters the map without known-good evidence after a cooldown ([ADR-0028](0028-evidence-is-graded-five-levels.md), [ADR-0037](0037-seven-day-cooldown-plus-evidence.md)).

## Options

1. **Human curators** maintain the map by hand.
2. **Agents curate under evidence gates;** humans decide coverage and breaking contract changes.
3. **Fully autonomous agents,** including decisions about coverage and contracts.

## Decision

Option 2. The public map is curated primarily by agents:

- A **watcher** notices upstream releases, tags and commits.
- **Curation agents** update facts, recipes and probes, extracting facts from headers where possible.
- CI evidence and the cooldown alone decide what enters the map.
- Humans decide only what the map covers: which slots, which libraries, and when a contract gets a new major version.

## Consequences

- Nothing enters the map on a curation agent's say-so alone.
- Human time goes to direction, not upkeep.
- The watcher also keeps each library's contribution policy current ([ADR-0046](0046-four-agent-nevers.md)).
