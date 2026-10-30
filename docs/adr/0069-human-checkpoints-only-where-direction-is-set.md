---
status: accepted
date: 2026-10-30
---

# Human checkpoints exist only where direction is set

## Context

Agents steward code and evidence replaces line-by-line review ([ADR-0005](0005-agent-is-primary-steward.md)). For the public map, humans decide only coverage and contract major versions ([ADR-0051](0051-agents-curate-the-public-map.md)). Vertical slices are accepted by eval tasks ([ADR-0068](0068-test-strategy-four-layers.md)). ADRs are the record of every decision with real alternatives.

## Options

1. **Human review of every merge.**
2. **Checkpoints at milestones** (each slice, each release).
3. **Checkpoints only where direction is set.**

## Decision

Option 3. While agents build Automap, human checkpoints exist only for:

- any new or changed ADR;
- what the map covers, and new contract major versions;
- the v1 release itself.

Everything else (merging slices, map releases, CLI point releases) proceeds once its evidence and evals pass.

## Consequences

- Every change of direction reaches a human, because it requires an ADR.
- Routine work is gated by evidence, not by human availability.
