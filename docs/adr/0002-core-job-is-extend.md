---
status: accepted
date: 2026-10-29
---

# The core job is extending an engine with verified seam glue

## Context

With the framework-first indie as the primary persona ([ADR-0001](0001-primary-persona-framework-first-indie.md)), the order of their jobs decides what the first vertical slice proves and what "immediately useful" means. The candidate jobs were: start a new engine, extend it with a library, upgrade libraries, audit what ships (licenses, provenance, versions), and port to a new platform.

Research ([SUMMARY](../research/SUMMARY.md)): package managers already handle fetching and build-level fit, but nothing handles the seams (allocators, jobs, math types, init order, build flags). Version churn breaks glue regularly; in 2026 alone, Box2D 3.2 changed its callback signatures and Lua 5.5 changed `lua_newstate`. The persona wants known-good pairings, readable glue they own, and no forced upgrades.

## Options

1. **Start first:** optimize for a new engine or testbed from a one-line request; treat later changes as secondary.
2. **Extend first:** optimize for adding or swapping a library with correct seam glue and evidence that the pairing works; starting is the same job applied to the first slots.
3. **Upgrade first:** optimize for moving existing engines to new library versions safely.
4. **Audit first:** lead with license, provenance and version reporting.

## Decision

The core job is **extend**: "add a library to my engine, with seam glue I own and evidence that the pairing works." **Start** is the first instance of that job, not a separate product. **Upgrade** is the second job. **Audit** and **port** come later.

## Consequences

- The first vertical slice still produces a testbed, but it must be built by the same mechanism that later adds a library to an existing composition, not by a separate template path.
- Every library addition produces glue in user-owned code plus evidence of the pairing; a template that can't be extended afterwards doesn't satisfy the job.
- Upgrade support shapes the design early (pins, evidence that expires, glue that can be regenerated or migrated) even though it ships second.
- Audit (licenses, SBOM, provenance) and port (web, consoles) are not v1 drivers.
