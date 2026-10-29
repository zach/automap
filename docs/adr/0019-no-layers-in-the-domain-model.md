---
status: accepted
date: 2026-10-29
---

# The domain model has no layers

## Context

The brief describes a typical custom-engine layering, bottom to top: platform and GPU, core runtime, engine systems, world and tools, game. A layer could be a domain term with rules, such as "lower layers never depend on higher ones", or just a way of presenting slots. Slots have no Automap-defined API and the engine calls libraries directly ([ADR-0012](0012-slots-have-no-automap-api.md)), so the only behavior a layer could have is a rule about how the developer structures their own code. Tiers were dropped for the same reason: no behavior behind the term ([ADR-0018](0018-no-tiers-in-the-domain-model.md)).

## Options

1. **Layer as a term with rules:** slots belong to layers and the CLI enforces dependency direction.
2. **No layer term:** slots carry no layer rules; layering appears only as an example architecture.

## Decision

Option 2. "Layer" is not a domain term, and slots carry no layer rules. The starter-engine journey in the spec shows a layered architecture as an example.

## Consequences

- Automap does not police the engine's architecture; its concern is the seams between libraries.
