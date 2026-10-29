---
status: accepted
date: 2026-10-29
---

# Slots are roles with no Automap-defined API

## Context

With contracts owned by Automap per seam ([ADR-0011](0011-automap-owns-one-contract-per-seam.md)), a slot could either be a plain role (windowing, physics, audio) or define an engine-facing API that every library filling it must implement, so libraries could be swapped behind it. The brief describes the custom-integration tier as "swappable implementations behind contracts". The persona left big engines for control and a small surface ([05](../research/05-developer-needs-2026.md)). The research places the hard part of composition at the seams, not in the libraries' functional APIs ([SUMMARY](../research/SUMMARY.md)). The coding agent is the primary steward of the engine's code ([ADR-0005](0005-agent-is-primary-steward.md)).

## Options

1. **Slots define an engine-facing API:** every library filling a slot implements Automap's API for that role (for example a physics API), so libraries swap behind it.
2. **Slots are roles with no Automap-defined API:** the engine calls the filling library's own API; swapping a library means the steward rewrites those call sites.

## Decision

Option 2. A slot is a role with no Automap-defined engine-facing API. The engine's code calls the filling library's own API directly, such as Jolt's or Box2D's. Swapping a library means the steward rewrites those call sites. A developer may put their own interface in their glue; Automap doesn't require one.

## Consequences

- Automap does not become an engine framework and defines no lowest-common-denominator APIs for physics, rendering, audio or any other role.
- Libraries are swappable at the seam contracts, not behind an Automap API.
