---
status: accepted
date: 2026-10-29
---

# v1 non-goals: Godot extenders and non-C++ consumers

## Context

With the framework-first indie as the primary persona ([ADR-0001](0001-primary-persona-framework-first-indie.md)), three adjacent groups needed an explicit call: developers extending Godot (GDExtension, LibGodot), developers consuming the same C libraries from other languages (C# over SDL3, Odin, Zig, Rust bindings), and agent-only builders who direct an agent and don't read the C++ it writes.

## Options

For each group: a v1 non-goal, or in scope.

## Decision

- **Godot and LibGodot extenders are a v1 non-goal.** Godot's core already owns the seams for its users ([03](../research/03-prior-art.md)), so Automap's contract vocabulary adds little there.
- **Consumers in languages other than C++ are a v1 non-goal.** Contracts describe libraries' C/C++ APIs and glue recipes target C++ only; supporting more languages would multiply recipes and verification before C++ is proven.
- **Agent-only builders are in scope, not a non-goal.** Coding agents can be the primary steward of an engine's code, so the persona does not depend on reading C++ ([ADR-0005](0005-agent-is-primary-steward.md)). Agent-only builders are part of the primary persona.

## Consequences

- Godot-specific and non-C++ recipes, contracts and verification are out of v1. Nothing is designed to preclude them later, but nothing is added for them either.
- Nothing in the design may assume a human reads or reviews generated C++ line by line.
