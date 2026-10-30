---
status: accepted
date: 2026-10-30
---

# Every probe must prove it can fail before its rule counts as probed

## Context

In the spike ([07](../research/07-spike-testbed-agent.md)), SDL's "hooks before first allocation" probe passed even with the exact mistake it guards against deliberately added: `SDL_GetNumAllocations()` returns -1 unless SDL is built with allocation counting, and the glue treated -1 as a pass. A probe that can't fail produces evidence while checking nothing. Anthropic's parallel-agent C compiler project found that verifier quality decided whether agents made progress ([06](../research/06-verification.md)). Known-good means evidence at the probed level or higher ([ADR-0028](0028-evidence-is-graded-five-levels.md)), and only probed rules count toward evidence ([ADR-0022](0022-rules-are-probed-or-guidance.md)).

## Options

1. **Trust probes as written.**
2. **Review probes by hand.**
3. **Seeded violations in CI:** each probe must report failure on a known violation before its rule counts as probed.

## Decision

Option 3. Every probe must prove it can fail. Automap's CI seeds a known violation for each probe (for example installing hooks late, running init steps out of order, or mismatching a define) and checks that the probe reports failure. A probe that passes its seeded violation is treated as guidance until it is fixed.

## Consequences

- Every probe ships with at least one seeded violation.
- Seeded-violation runs repeat whenever a probe or its library changes.
- "Probed" means a check known to catch the mistake ran, not just that a check ran.
