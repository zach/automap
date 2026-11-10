---
status: accepted
date: 2026-11-10
---

# A composition has one set of slots for every build type

## Context

A composition is a manifest and a lock ([ADR-0017](0017-composition-is-a-manifest-and-a-lock.md)) with one library per slot ([ADR-0014](0014-one-library-per-slot.md)). Building Spintacular, the developer wanted Tracy in development builds and not in the release: the map's one Tracy variant recorded from startup and held everything until a profiler connected, which a player's copy can't have ([08](../research/08-spintacular-field-report.md)). Since then `automap remove <library>` and `automap extend <library>` take a library out and put it back in one command each, game code that goes through the diagnostics facade doesn't change either way, and Tracy has an `on-demand` variant that records only while a profiler is connected and can ship.

## Options

1. **One set of slots per composition,** as now: a library that shouldn't ship is removed before the release build, or composed in a variant that can ship.
2. **Libraries held in development builds only:** the manifest marks a slot's library as for development builds, and a release build links nothing from it.

## Decision

Option 1. A composition has one set of slots, variants and choices for every build type.

## Consequences

- The lock, the evidence and `verify` keep one meaning: they are about the one composition.
- A profiler in development and none in the release is two commands, `remove` and `extend`, or the on-demand variant throughout.
