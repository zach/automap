---
status: accepted
date: 2026-10-29
---

# Automap is fully open source and non-commercial

## Context

Trust in licenses is fragile for this persona: Unity's runtime-fee episode still drives engine choices, and Slay the Spire II moved to Godot partly because of it ([05](../research/05-developer-needs-2026.md)). Our Machinery's closed SDK left its users nothing when its license ended ([03](../research/03-prior-art.md)). Open contributions ([ADR-0052](0052-open-contributions-same-evidence-gates.md)) and the goodwill of upstream maintainers depend on openness. No decision requires a paid service: builds, evidence and private maps all work without Automap's servers ([ADR-0038](0038-only-fetching-needs-the-network.md)).

## Options

1. **Fully open source, non-commercial.**
2. **Open core:** open CLI and map, with paid services such as private-map hosting or verification for studios.
3. **Commercial.**

## Decision

Option 1. Automap is fully open source: the CLI, the public map and the Automap registry. It is not a commercial project; there is no paid tier and no plan for commercial services.

## Consequences

- "Can this be taken away from me?" is answered before a developer asks.
- No part of the design needs to reserve anything for a paid offering.
