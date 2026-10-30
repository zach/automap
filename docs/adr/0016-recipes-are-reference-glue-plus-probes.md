---
status: accepted
date: 2026-10-29
---

# A recipe is reference glue, probes, and adaptation notes

## Context

Recipes bind libraries to contracts and wire pairings together ([ADR-0011](0011-automap-owns-one-contract-per-seam.md), [ADR-0015](0015-pairings-are-exactly-two-libraries.md)). No design may assume glue lives where Automap would have put it ([ADR-0003](0003-adopting-existing-engines-is-a-design-constraint.md)), so a recipe can't be only a file template. The coding agent is the primary steward and may restructure glue freely ([ADR-0005](0005-agent-is-primary-steward.md)). Research ([04](../research/04-agent-ecosystem.md)): generic skill guidance added little on average, and version-mismatched guidance made results worse; only verified, version-specific material carries weight.

## Options

1. **A template:** the CLI generates glue files from a parameterized template.
2. **Instructions:** prose and examples for the agent to apply in context.
3. **Reference glue, probes and adaptation notes:** verified glue the CLI can write, executable checks that judge any glue doing the same job, and notes on adapting it.

## Decision

Option 3. A recipe is:

1. **Reference glue** that Automap has verified.
2. **Probes** that check any glue doing the same job, such as "allocator hook installed before first allocation" or "`JPH_*` defines match".
3. **Adaptation notes** for applying the reference glue where glue already exists.

For a new composition, the CLI writes the reference glue into the developer's repo. Where glue already exists, as in an adopted engine, the agent adapts it. Either way, success is judged by the probes passing, not by matching the reference.

## Consequences

- One recipe serves both new and adopted engines.
- The agent can restructure glue without losing evidence, as long as the probes still pass.
- Every recipe must ship probes; a recipe without them can't produce evidence.
