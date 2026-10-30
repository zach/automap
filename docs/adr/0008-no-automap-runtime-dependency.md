---
status: accepted
date: 2026-10-29
---

# Automap is never a runtime dependency of the engine

## Context

Research found that in C++ no library owns the seams, so Automap acts as a "virtual core" whose contract vocabulary is the product ([SUMMARY](../research/SUMMARY.md) finding 2). Some seams could be served by shared code: job systems, for example, assume different scheduling models, so bridging every scheduler to every library is N×M unless a neutral interface makes it N+M ([02c](../research/02c-catalog-simulation-runtime-glue.md)). The persona wants no lock-in and vendored code ([05](../research/05-developer-needs-2026.md)); glue lives in the developer's repo ([ADR-0002](0002-core-job-is-extend.md)); and adopted engines bring their own job systems and allocators ([ADR-0003](0003-adopting-existing-engines-is-a-design-constraint.md)).

## Options

1. **An Automap runtime library:** shared seam code ships as a library the engine links; recipes target it.
2. **No runtime dependency:** anything shared across recipes is delivered as glue source in the developer's repo.

## Decision

Option 2. Automap is never a runtime dependency of the engine. Anything shared across recipes, such as a neutral job-system interface if branch 4 calls for one, is delivered as glue source in the developer's repo, which the developer owns.

## Consequences

- An engine built with Automap builds and runs without any Automap library.
- Shared interfaces still give the N+M reduction for seams like job systems, without a dependency.
