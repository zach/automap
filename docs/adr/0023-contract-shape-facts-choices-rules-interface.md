---
status: accepted
date: 2026-10-29
---

# Every contract has the same four parts

## Context

Automap owns one contract per seam, across ten seams ([ADR-0011](0011-automap-owns-one-contract-per-seam.md), [ADR-0021](0021-the-ten-seams.md)), and every rule is probed or guidance ([ADR-0022](0022-rules-are-probed-or-guidance.md)). The research concluded that the map should record each library's hooks as structured data (allocator hook style and signature, job interface kind, handle layout, init steps, ABI-affecting defines), much of which can be extracted semi-automatically from headers ([02c](../research/02c-catalog-simulation-runtime-glue.md)). Shared interfaces, such as a neutral job-system interface, are delivered as glue source ([ADR-0008](0008-no-automap-runtime-dependency.md)).

## Options

1. **A uniform shape for every contract.**
2. **A bespoke shape per seam.**

## Decision

Option 1. Every contract has the same four parts:

1. **Facts:** what each library declares about the seam, such as its allocator hook style and signature, or its ABI-affecting defines. Facts live in the map.
2. **Choices:** what the composition decides at the seam, such as which allocator, which engine math library, or whether determinism is on. Choices live in the manifest.
3. **Rules:** probed or guidance, stated in terms of the facts and choices.
4. **A shared interface (optional):** delivered as glue source.

## Consequences

- The CLI handles all ten seams the same way.
- Probes can be generated from facts and choices rather than hand-written; for example, a define-mismatch probe follows from a library's declared ABI defines and the composition's chosen variant.
- The manifest holds the composition's choices alongside its slots and glue locations.
