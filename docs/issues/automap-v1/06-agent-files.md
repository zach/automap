# 06: Agent files: AGENTS.md section and router skill

**What to build:** `start` writes, and every later command that changes the composition updates, the Automap section of `AGENTS.md` between markers, adds an `@AGENTS.md` import to an existing `CLAUDE.md`, and writes the router skill into each agent skill directory. The router skill's first version covers `start`, `verify`, reading failure codes and the four nevers.

**Blocked by:** [03: Start writes a composition](03-start-writes-a-composition.md)

**Status:** done

- [x] Content outside the markers is byte-identical before and after; the section holds the build, test and run commands, where the manifest and lock live, the slots in use, and the instruction to go through the router skill and CLI before changing glue (ADR-0043).
- [x] The `CLAUDE.md` import line is added only if `CLAUDE.md` exists, and only once.
- [x] The router skill carries the CLI's version and is rewritten when a newer CLI runs any command (ADR-0044).
- [x] The four nevers come from one source shared by the router skill and the CLI's refusal messages (ADR-0046, ADR-0074).

## Comments

- The four nevers live in `internal/agentfiles` as one list; the router skill renders it and the `LOCK_EDITED` refusal quotes the first. The other three have no refusal yet: the CLI can't detect them until later tickets give it something to check (evidence signing, probe changes).
- The section is replaced between its markers, or appended after a blank line; the developer's bytes before and after are kept. `start` writes the agent files now; `extend`, `upgrade` and `adopt` will call the same `agentfiles.Files`.
