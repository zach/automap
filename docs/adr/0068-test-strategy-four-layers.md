---
status: accepted
date: 2026-10-30
---

# The build's test strategy has four layers

## Context

Automap has three kinds of thing to test: Go code ([ADR-0064](0064-cli-is-written-in-go.md)), map content, and whether agents succeed end to end. The CLI's JSON output and failure codes are its agent-facing interface ([ADR-0041](0041-cli-json-output-and-failure-codes.md)). Map content is verified by probes on declared edges ([ADR-0033](0033-automap-ci-verifies-declared-edges-latest-and-previous.md)). An eval suite of fresh agent sessions measures success ([ADR-0062](0062-no-telemetry-eval-suite-measures-success.md)). The build team uses `tdd` ([ADR-0067](0067-build-team-skills.md)).

## Options

1. **Unit tests only.**
2. **Four layers, each testing at the level where things fail.**

## Decision

Option 2. The test strategy has four layers:

1. **Test-driven Go unit tests** for CLI logic: resolution, schema validation, lock canonicalization, failure codes.
2. **Golden tests** of the CLI's JSON output and failure codes against the published schemas.
3. **The map tests itself:** every recipe's probes run in Automap's CI on the declared edges, so map content is verified by evidence rather than unit tests.
4. **The eval suite as acceptance:** each vertical slice's acceptance criteria are eval tasks a fresh agent session must pass.

## Consequences

- The CLI's agent-facing interface is a tested contract.
- A vertical slice can't be declared done on anyone's say-so alone.
