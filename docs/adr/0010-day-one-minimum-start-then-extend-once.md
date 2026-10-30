---
status: accepted
date: 2026-10-29
---

# Day-one minimum: start, then extend once

## Context

The product is the map, the CLI and one router skill ([ADR-0006](0006-product-shape-map-cli-router-skill.md)), and the core job is extending an engine, with starting as its first instance ([ADR-0002](0002-core-job-is-extend.md)). The smallest valuable release has to exercise that job end to end. Research found that templates are snapshots that go stale with no signal ([03](../research/03-prior-art.md)), and adoption depends on start and extend sharing one mechanism ([ADR-0003](0003-adopting-existing-engines-is-a-design-constraint.md)).

## Options

1. **A testbed only:** the CLI starts a testbed composition.
2. **Start, then extend once:** the CLI starts a testbed composition and then extends it with one more library.
3. **Start, extend and upgrade:** include the second job on day one.

## Decision

Option 2. On day one, a developer's coding agent uses the CLI to start a testbed composition and then extend it with one more library. Both steps put glue in the developer's repo and produce evidence that each resulting pairing works. Upgrade and adopt follow, in that order.

## Consequences

- The first vertical slice must include an extend step after start, through the same mechanism.
- The exact slots, libraries and platforms on day one are decided in branch 13.
