---
status: accepted
date: 2026-10-29
---

# The coding agent is the primary steward of the engine's code

## Context

ADR-0001 as first written assumed the human reads and owns every line of glue, and defined the persona as C/C++ literate. That assumption was never agreed. The product owner rejected it: coding agents are capable of being the primary steward of an engine's code, and capability moves fast enough that benchmark figures go stale within months. The figure the assumption leaned on (GameEngineBench, best agent 55.5% pass@1 on UE5 C++ tasks, July 2026, [05](../research/05-developer-needs-2026.md)) did not include current frontier agents.

The research's trust findings ([SUMMARY](../research/SUMMARY.md) finding 1) are about developers' willingness to adopt agent output, not about agent ability. Developers trust output they can verify.

## Options

1. **Human steward, agent assists:** the human reads and owns every line; the agent drafts.
2. **Agent writes, human reviews everything:** every change needs line-by-line human review.
3. **Agent is the primary steward, human sets direction:** the agent writes and maintains the engine's code; reading the C++ is optional for the human.

## Decision

Option 3. The persona is defined by what they want, not by whether they read C++ ([ADR-0001](0001-primary-persona-framework-first-indie.md), amended).

## Consequences

- Agent-only builders are inside the primary persona ([ADR-0004](0004-v1-non-goals.md)).
- Verification evidence, not human line-by-line review, is what makes a change trustworthy. This raises the weight of research finding 1 (graded, signed, time-boxed evidence) in later branches.
- "Glue the developer owns" ([ADR-0002](0002-core-job-is-extend.md)) means glue that lives in the developer's repo under their control, whoever wrote it.
- No design may assume a human reads or reviews generated C++.
