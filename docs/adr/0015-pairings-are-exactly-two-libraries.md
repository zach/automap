---
status: accepted
date: 2026-10-29
---

# A pairing is exactly two libraries

## Context

Pairings cover direct library-to-library integrations that don't go through a contract ([ADR-0011](0011-automap-owns-one-contract-per-seam.md)). Integrations are pairwise in practice: Dear ImGui keeps one backend per platform library, and rlImGui keeps one branch per (raylib, Dear ImGui) version pair ([03](../research/03-prior-art.md)). Some integrations are provided by a third, bridging library, such as rlImGui between raylib and Dear ImGui.

## Options

1. **N-ary pairings:** a pairing may name any number of libraries.
2. **Binary pairings:** a pairing always names exactly two libraries; larger integrations are several pairings.

## Decision

Option 2. A pairing is always exactly two libraries. Anything involving more is several pairings. A bridge library such as rlImGui is an ordinary library with its own pairings to each side.

## Consequences

- A version change to a library invalidates exactly the pairings that name it.
- Interactions among three or more libraries are covered by evidence about whole compositions, decided in branch 5.
