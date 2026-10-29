---
status: accepted
date: 2026-10-29
---

# Product shape: the map, a CLI, and one router skill

## Context

Automap could take several shapes: a knowledge map that agents read and act on by themselves, a scaffolder, a layer over package managers, a maintenance agent, or a combination. Starting a new engine is the first instance of extending one ([ADR-0002](0002-core-job-is-extend.md)), so a separate scaffolder isn't needed. The coding agent is the primary steward of the engine's code ([ADR-0005](0005-agent-is-primary-steward.md)), so the agent is the main consumer of whatever Automap exposes.

Research ([04](../research/04-agent-ecosystem.md), [05](../research/05-developer-needs-2026.md), [06](../research/06-verification.md)):

- First-party agent guidance prefers a CLI where one exists, with MCP for hosts without a shell. The most mature game-dev skill packs put a deterministic index or router behind a thin skill.
- Skill triggering is description-only and budgeted; hundreds of installed skills don't trigger reliably.
- Evidence that a pairing works has to come from repeatable operations.

## Options

1. **Map plus skills only:** versioned data and per-library skills; the agent reasons its way through every operation.
2. **Map, CLI and one router skill:** the CLI is the deterministic way to query the map and act on a composition; one skill teaches agents to drive it.
3. **A hosted service at the center:** a web API or MCP server that owns the map and performs operations remotely.

## Decision

Option 2. Automap consists of:

1. **The map:** versioned data describing slots, libraries, contracts, pairings and recipes, plus the evidence records behind them.
2. **The CLI:** the one deterministic way to query the map and act on a composition (resolve, apply recipes, check contracts, run verification), designed for agents first.
3. **The router skill:** one agent skill that teaches agents to drive the CLI.

Automap's own CI produces the published evidence. Some functions may rely on a remote hosted registry or package repository.

## Consequences

- Agents reach library knowledge through the CLI and the router skill.
- Every operation whose result counts as evidence is a deterministic CLI operation.
- Which functions may rely on a remote hosted registry or package repository, and which must work offline, is open (OQ-013).
