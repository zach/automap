---
status: accepted
date: 2026-10-29
---

# The CLI maintains one Automap section in the developer's AGENTS.md

## Context

The developer's coding agent needs to know, every session, how to build and test the engine and that integrations go through Automap. AGENTS.md is the cross-tool standard, governed under the Linux Foundation's Agentic AI Foundation and read by Codex, Copilot and Cursor; Claude Code reads it natively when a repo has no `CLAUDE.md` ([04](../research/04-agent-ecosystem.md)). vcpkg's own repo uses a `CLAUDE.md` that just imports `@AGENTS.md`. Always-loaded instruction files are for facts and rules every session needs; multi-step procedures belong in skills. A fixed size cap would encode an assumption about today's models that goes stale.

## Options

1. **No instruction file:** the agent learns about Automap only from the router skill.
2. **An Automap-owned `AGENTS.md`** that replaces the developer's.
3. **One delimited Automap section in the developer's `AGENTS.md`.**

## Decision

Option 3. The CLI maintains one Automap section in the developer's `AGENTS.md`, inside clearly delimited markers so the rest of the file stays the developer's. The section holds only what every session needs:

- the build, test and run commands;
- where the manifest and lock live;
- the slots in use;
- the instruction to go through the router skill and CLI before changing any glue.

Everything else lives in the router skill. If the repo already has a `CLAUDE.md`, the CLI adds an `@AGENTS.md` import line to it.

## Consequences

- One file reaches every major coding agent.
- Automap never overwrites the developer's own instructions.
- The section stays small because its content is fixed, not because of a size limit.
