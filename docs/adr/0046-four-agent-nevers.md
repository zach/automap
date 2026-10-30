---
status: accepted
date: 2026-10-29
---

# The agent is held to four "nevers"

## Context

Evidence is what replaces human review when the coding agent is the steward ([ADR-0005](0005-agent-is-primary-steward.md)), so anything that would let evidence be faked or weakened breaks trust in the whole product. Separately, several upstream projects adopted policies in 2026 banning AI-generated contributions: SDL, SDL_shadercross, sokol, stb, PhysicsFS and Godot ([02a](../research/02a-catalog-platform-graphics-ui-audio.md), [02b](../research/02b-catalog-assets-and-data.md), [05](../research/05-developer-needs-2026.md)). Those policies hold regardless of how capable the agent is. Everything else is the steward's judgment.

## Options

1. **No explicit prohibitions:** rely on the steward's judgment throughout.
2. **A short list of prohibitions** limited to protecting evidence and respecting others' rules, enforced by the CLI where it can detect a violation.
3. **A long rulebook** covering coding practice.

## Decision

Option 2. The router skill holds the agent to exactly four "nevers". Where the CLI can detect a violation, it refuses rather than relying on the skill.

1. **Never edit the lock by hand.** Only the CLI writes it; the CLI detects a hand edit because the lock's canonical hash no longer matches.
2. **Never create or claim evidence except by running the CLI.**
3. **Never weaken, skip or delete a probe to get a pass.** A failing probe is reported, or the glue is fixed.
4. **Never open an issue or pull request on an upstream project whose contribution policy bans AI-generated contributions.** The map records each library's contribution policy as a fact.

## Consequences

- Evidence can only come from the CLI running real probes.
- Each library's contribution policy must be kept current in the map.
