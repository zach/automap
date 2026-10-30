---
status: accepted
date: 2026-10-30
---

# The build repo adopts the AGENTS.md convention it generates for users

## Context

This repo's agent instructions live in `CLAUDE.md`, holding exploration rules. Automap writes one `AGENTS.md` section into users' repos and adds an `@AGENTS.md` import to any existing `CLAUDE.md` ([ADR-0043](0043-cli-maintains-an-agents-md-section.md)). AGENTS.md is read by Codex, Copilot and Cursor, and natively by Claude Code when no `CLAUDE.md` exists ([04](../research/04-agent-ecosystem.md)).

## Options

1. **Keep `CLAUDE.md`** as the build repo's instruction file.
2. **Adopt the user convention:** `AGENTS.md` as the instruction file and `CLAUDE.md` as a single import line.

## Decision

Option 2. When implementation starts, this repo moves its instructions to the convention it generates for users:

- `AGENTS.md` is the one instruction file. It holds only what every session needs: build and test commands, the repo layout, and the working rules that still apply (documents state the current version only; ADRs record only what the product owner agreed; agents are not assumed to be weaker stewards than humans).
- `CLAUDE.md` is a single `@AGENTS.md` import line.
- Multi-step procedures live in skills.

## Consequences

- Automap's builders meet any rough edge in the convention before users do.
- Any coding agent, not only Claude Code, can work on Automap.
