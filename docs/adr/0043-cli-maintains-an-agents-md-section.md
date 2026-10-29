---
status: accepted
date: 2026-10-29
---

# The CLI maintains one short Automap section in the developer's AGENTS.md

## Context

The developer's coding agent needs to know, every session, how to build and test the engine and that integrations go through Automap. AGENTS.md is the cross-tool standard, governed under the Linux Foundation's Agentic AI Foundation and read by Codex, Copilot and Cursor; Claude Code reads it natively when a repo has no `CLAUDE.md` ([04](../research/04-agent-ecosystem.md)). vcpkg's own repo uses a `CLAUDE.md` that just imports `@AGENTS.md`. Claude Code's guidance keeps instruction files under 200 lines and puts multi-step procedures in skills.

## Options

1. **No instruction file:** the agent learns about Automap only from the router skill.
2. **An Automap-owned `AGENTS.md`** that replaces the developer's.
3. **One delimited Automap section in the developer's `AGENTS.md`.**

## Decision

Option 3. The CLI maintains one short Automap section in the developer's `AGENTS.md`:

- inside clearly delimited markers, so the rest of the file stays the developer's;
- under 200 lines;
- giving the build, test and run commands, where the manifest and lock live, and the slots in use;
- telling the agent to go through the router skill and CLI before changing any glue.

If the repo already has a `CLAUDE.md`, the CLI adds an `@AGENTS.md` import line to it.

## Consequences

- One file reaches every major coding agent.
- Automap never overwrites the developer's own instructions.
