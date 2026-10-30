---
status: accepted
date: 2026-10-29
---

# Adopting an existing engine is a design constraint from v1

## Context

The core job is extending an engine with verified seam glue ([ADR-0002](0002-core-job-is-extend.md)). Many framework-first indies ([ADR-0001](0001-primary-persona-framework-first-indie.md)) already reuse a personal framework across games ([05](../research/05-developer-needs-2026.md): Celeste's developer, Mewgenics). If Automap could only extend engines it generated, it would give them nothing until their next engine.

## Options

1. **Automap-born only:** only compositions Automap created can be extended; existing engines must start over.
2. **Adoption as a design constraint, shipped second:** it must be possible to describe an engine Automap didn't generate as a composition and extend it with the same mechanism; the new-engine path ships and is proven first.
3. **Adoption first:** lead with describing and extending existing engines.

## Decision

Option 2. Adoption is a v1 design constraint and ships after the new-engine path.

## Consequences

- No design may assume Automap generated every file. In particular, recipes cannot assume glue lives where Automap would have put it; a composition must be able to say where the existing glue for a seam is.
- Describing an existing engine must be possible by hand or with an agent, without regenerating the engine.
- The first vertical slice still proves the new-engine path; an adoption slice follows.
