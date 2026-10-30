---
status: accepted
date: 2026-10-30
---

# The CLI generates the composition's startup from its recipes

## Context

Each recipe names entry points, each with a phase and whether it runs at init or shutdown ([ADR-0074](0074-spike-format-fixes.md)), and init and shutdown run through six fixed phases ([ADR-0026](0026-lifecycle-has-six-fixed-phases.md)). Nothing said what calls the entry points, and some calls took objects, such as the window and GPU device, that nothing declared. In the spike ([07](../research/07-spike-testbed-agent.md)), the agent hand-wrote `main` and had to get the order right itself. Manifest choices already reach glue through a choices header the CLI writes and the developer commits ([ADR-0074](0074-spike-format-fixes.md)). Adopted engines bring startup code of their own ([ADR-0003](0003-adopting-existing-engines-is-a-design-constraint.md)).

## Options

1. **Hand-written startup:** the agent writes `main` from the recipes' entry points.
2. **Generated startup:** the CLI writes a startup source from the recipes.

## Decision

Option 2.

- The CLI writes a startup source, committed and regenerated whenever the composition changes. It provides a startup function and a shutdown function that call every entry point in phase order, respecting each step's `after` list and the pairings' ordering constraints, and in `--probe` mode also run the probes.
- Every entry point takes one shared **context**. Each recipe declares the context fields its entry points provide, such as SDL3's window and GPU device, and the fields they read. The CLI generates the context and checks that every field read is provided by an entry point that runs earlier.
- The developer's `main` calls startup once and shutdown once: the testbed's `main` that `start` writes, or an adopted engine's own.
- An adopted engine may keep its own startup code instead, marked in the manifest; the lifecycle probe judges its order the same way.

## Consequences

- No agent has to work out init order by hand; the probes still check it.
- Recipes must declare what their entry points provide and read.
