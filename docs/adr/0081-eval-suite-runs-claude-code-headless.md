---
status: accepted
date: 2026-10-30
---

# The eval suite runs Claude Code headless, under a spending cap

## Context

Success is measured by an eval suite in which fresh agent sessions carry out the five v1 jobs along the three journeys on every map release ([ADR-0062](0062-no-telemetry-eval-suite-measures-success.md)), and v1 is judged on measures computed from it ([ADR-0063](0063-v1-success-measures.md)). Nothing named the agent runtime, the model, the credentials or a budget, and every run is an ongoing cost to the product owner. Automap targets agents through `AGENTS.md` and a router skill written where Claude Code reads skills ([ADR-0043](0043-cli-maintains-an-agents-md-section.md), [ADR-0044](0044-router-skill-ships-inside-the-cli.md)).

## Options

1. **Claude Code only, headless, under a spending cap.**
2. **Several agent runtimes from v1.**
3. **Runs by hand on a developer's machine.**

## Decision

Option 1.

- **Runtime:** Claude Code in headless mode on GitHub-hosted runners. Each task runs in a fresh environment holding only the platform's build toolchain, the CLI and the task's starting repo.
- **Model:** the newest Claude model at the time of the run, recorded in the eval report.
- **Credentials:** an Anthropic API key stored as a repository secret, used only by the eval workflow.
- **Cadence:** each task runs once per map release and on demand. The first-attempt success measure is computed over those single runs.
- **Budget:** a monthly spending cap the product owner sets on the key. A run the cap stops is reported as skipped, not failed.

## Consequences

- Eval results are comparable across runs because each report names its model.
- Other agent runtimes can join the suite after v1.
