---
status: accepted
date: 2026-10-30
---

# Seven format fixes from the spike

## Context

The spike ([07](../research/07-spike-testbed-agent.md)) found smaller gaps that fit the existing model without new concepts: facts recorded once per library although they change between versions; manifest glue entries that could name only one file and had no way to say a binding has no glue; lifecycle `after` dependencies declared but never checked; a cross-library ordering (`SDL_WaitForGPUIdle` before Dear ImGui's backend shutdown) that couldn't be expressed; recipe entry points keyed by names that aren't phases; a manifest choice (the scheduler's worker count) copied into glue by hand; and two different lists of the agent's "nevers".

## Options

1. **Adopt all seven fixes as spec requirements.**
2. **Leave them to the build team.**

## Decision

Option 1. The spec requires:

1. Facts are recorded per library version, and per variant where they differ.
2. A manifest glue entry is a list of files, or "not-applicable" with a reason.
3. Lifecycle `after` dependencies are passed to the lifecycle interface and checked by its probe.
4. Pairings can declare cross-library ordering constraints, checked by the same probe.
5. Recipe entry points name a phase and whether they run at init or shutdown.
6. Manifest choices reach glue only through a **choices header** the CLI writes from the manifest, committed like other build files.
7. The router skill and the CLI share one list of the four nevers ([ADR-0046](0046-four-agent-nevers.md)).

## Consequences

- The first agent to build from the spec doesn't meet the drafts' ambiguities.
- Glue never hard-codes a composition choice.
