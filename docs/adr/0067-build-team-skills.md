---
status: accepted
date: 2026-10-30
---

# The build team uses the pack's implementation skills and Automap's own router skill

## Context

The exploration used skills from mattpocock/skills, vendored into this repo at commit `a496f55` (v1.3.2). The same pack has implementation skills: `tdd`; `implement`, for one ticket at a time; `implement-spec`, which works the ticket graph in parallel worktrees, lands everything on one integration branch and finishes with `code-review`; `code-review`; and `diagnosing-bugs`. Slice 1 produces Automap's router skill ([ADR-0061](0061-first-three-vertical-slices.md)), and Automap maintains reference compositions and an eval suite ([ADR-0033](0033-automap-ci-verifies-declared-edges-latest-and-previous.md), [ADR-0062](0062-no-telemetry-eval-suite-measures-success.md)).

## Options

1. **The pack's implementation skills, vendored at the same commit,** plus Automap's own router skill once it exists.
2. **A different skill set** for implementation.
3. **No skills** for the build.

## Decision

Option 1. The build team uses `tdd`, `implement`, `implement-spec`, `code-review` and `diagnosing-bugs`, vendored at the same pinned commit as the exploration skills, alongside `domain-modeling`, `to-tickets` and `handoff`. It starts using Automap's own router skill as soon as slice 1 produces it, for maintaining the eval suite and the reference compositions.

## Consequences

- Tickets from `to-tickets` feed straight into `implement` or `implement-spec`.
- Automap's builders meet gaps in the router skill before users do.
