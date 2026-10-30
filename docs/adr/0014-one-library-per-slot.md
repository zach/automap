---
status: accepted
date: 2026-10-29
---

# Each slot holds at most one library; a library may fill several slots

## Context

A slot is a role with no Automap-defined API ([ADR-0012](0012-slots-have-no-automap-api.md)), and libraries are bound to Automap's per-seam contracts ([ADR-0011](0011-automap-owns-one-contract-per-seam.md)). Some libraries cover many roles: SDL3 provides windowing, input, GPU and audio ([02a](../research/02a-catalog-platform-graphics-ui-audio.md)). Some roles invite several libraries side by side, such as Dear ImGui with add-ons like ImGuizmo and ImPlot, or 2D and 3D physics in one engine.

## Options

1. **Many libraries per slot:** a slot may hold several libraries at once.
2. **At most one library per slot, many slots per library:** where two libraries seem to share a role, the map splits it into finer slots.

## Decision

Option 2. In a composition each slot holds at most one library, and one library may fill several slots. Where two libraries seem to share a role, the map splits it into finer slots, for example physics-2d and physics-3d, or a gizmo slot for ImGuizmo next to Dear ImGui's tools-UI slot.

## Consequences

- For any slot, exactly one library's seam behavior (allocator hook, job model, init steps) is the one that gets bound.
- Slot granularity is a curation decision in the map.
