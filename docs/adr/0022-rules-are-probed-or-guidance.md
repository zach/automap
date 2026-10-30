---
status: accepted
date: 2026-10-29
---

# Every rule is either probed or guidance

## Context

Contracts consist of rules ([ADR-0011](0011-automap-owns-one-contract-per-seam.md)), and success is judged by probes ([ADR-0016](0016-recipes-are-reference-glue-plus-probes.md)). Some real rules resist probing, such as "never cache a raw PhysX pointer across destroy" or "lock shared data touched from audio callbacks" ([02a](../research/02a-catalog-platform-graphics-ui-audio.md), [02c](../research/02c-catalog-simulation-runtime-glue.md)), yet still matter to the steward. Research found that contracts need executable probes and that prose guidance adds little on its own ([04](../research/04-agent-ecosystem.md), [06](../research/06-verification.md)).

## Options

1. **Probes only:** a rule that can't be probed isn't part of a contract.
2. **Probed or guidance:** every rule is one or the other; only probed rules count toward evidence.
3. **Prose contracts:** rules are prose; probes are optional extras.

## Decision

Option 2. Every rule is one of two kinds:

- **Probed:** it has a probe, and passing the probe counts as evidence.
- **Guidance:** it is labeled as such and never counts toward evidence.

A guidance rule becomes a probed rule as soon as a probe exists for it.

## Consequences

- A binding's or pairing's evidence covers only its probed rules, so "known-good" means exactly what was checked.
- Hazards that can't be probed are still visible to the steward.
