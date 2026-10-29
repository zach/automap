---
status: accepted
date: 2026-10-29
---

# The public map ships as immutable, dated releases recorded in every lock

## Context

The public map changes continuously as curation agents and contributors update it ([ADR-0051](0051-agents-curate-the-public-map.md), [ADR-0052](0052-open-contributions-same-evidence-gates.md)). vcpkg publishes dated registry releases (such as 2026.10.16) and projects pin a baseline commit, which makes builds reproducible months later ([01](../research/01-package-managers.md)). Automap ships no maintenance agent, so nothing should change a developer's engine until their own agent upgrades it ([ADR-0009](0009-no-automap-maintenance-agent.md)).

## Options

1. **A rolling map:** the CLI always reads the latest map.
2. **Immutable, dated releases,** with each lock recording the release it was resolved against.

## Decision

Option 2. The public map is published as immutable, dated **map releases**. Every lock records which map release it was resolved against. Moving to a newer map release is part of the upgrade job.

## Consequences

- A lock can always be resolved again exactly, even after newer releases change facts or retire versions.
- The developer's agent can tell precisely what knowledge a composition was built on.
- A newer map release never changes anyone's engine until their own agent upgrades.
