---
status: accepted
date: 2026-10-30
---

# Evidence results have four statuses, and evidence records its run conditions

## Context

Known-good means exactly what was checked ([ADR-0022](0022-rules-are-probed-or-guidance.md), [ADR-0028](0028-evidence-is-graded-five-levels.md)). In the spike ([07](../research/07-spike-testbed-agent.md)), evidence could only record pass or fail. It couldn't show that a rule didn't apply (Dear ImGui has no runtime log hook) or couldn't really be checked (SDL built without allocation counting), and it couldn't record the conditions of the run, which matter: Jolt's assert hook exists only in Debug builds, and the clean ThreadSanitizer run depended on one driver setting.

## Options

1. **Pass or fail only.**
2. **Four statuses and recorded run conditions.**

## Decision

Option 2. Every result in evidence carries one of four statuses:

- **pass**;
- **fail**;
- **not-applicable**, with a reason;
- **unverifiable**, with a reason.

Every evidence record carries its **run conditions**: build type, compiler version, environment variables and sanitizer options. The **probed** level requires every applicable probed rule to pass, with none unverifiable.

## Consequences

- A probe that can't check its rule in a given build blocks the probed level, which pushes Automap to make probes checkable (for example, a probe build of SDL with allocation counting).
- Evidence is reproducible from its own record.
