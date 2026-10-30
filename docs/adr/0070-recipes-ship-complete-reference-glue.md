---
status: accepted
date: 2026-10-30
---

# Recipes ship complete reference glue, including build fragments and setup glue

## Context

A recipe is reference glue, probes and adaptation notes ([ADR-0016](0016-recipes-are-reference-glue-plus-probes.md)), and Automap's CI verifies reference glue ([ADR-0033](0033-automap-ci-verifies-declared-edges-latest-and-previous.md)). In the spike ([07](../research/07-spike-testbed-agent.md)), most failures were gaps in the reference glue itself: no build-file fragment for adding a library, no setup of Jolt's physics system and layers, no setup of the shared scheduler from the manifest's choices, and a probe harness that no step copied. CI can only catch errors in glue that exists. Slots have no Automap API ([ADR-0012](0012-slots-have-no-automap-api.md)), so anything Automap ships must stay library wiring, not game code.

## Options

1. **Bindings only:** reference glue covers the hooks at each seam; setup and build files are left to the agent.
2. **Complete reference glue:** each library recipe also ships a build fragment and setup glue, and the probe harness is always included.

## Decision

Option 2. Every library recipe ships complete reference glue:

- a **build fragment**: the part of the build file that adds the library (how it is fetched, its options, its targets, and any flags its consumers need);
- **setup glue** for the library: the minimal objects the engine needs to use it at all, such as Jolt's physics system and its layer setup, but no game logic;
- **setup glue** for each shared interface it uses, created from the manifest's choices, such as the scheduler from the jobs contract's `workers` choice.

The **probe harness** is a shared interface included in every composition.

## Consequences

- The probe harness is the one shared interface not defined by a contract.
- CI verification of reference glue covers build fragments and setup glue too.
- The extend job always updates the build file through the new library's build fragment.
