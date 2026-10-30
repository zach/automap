# Handoff: exploration complete, implementation starts with ticket 01

Date: 2026-10-30. Branch: `exploration` on `zach/automap`, pushed.

## Where things stand

- **Phases 0–3 are done.** Research is in `docs/research/` (start with `SUMMARY.md`); the spike's findings are in `docs/research/07-spike-testbed-agent.md`, its code archived on the `spike/testbed-agent` branch.
- **Decisions** are ADR-0001 to ADR-0082 in `docs/adr/`; terms are in `CONTEXT.md`. `docs/OPEN_QUESTIONS.md` holds only OQ-008 and OQ-011, both deferred past v1.
- **The spec** is `docs/spec/SPEC.md` (the brief's twelve sections), with nine JSON Schemas in `docs/spec/schemas/`, 45 example files in `docs/spec/examples/`, and `docs/spec/check-examples.py`, which validates them all. SPEC §12.3 lists the details no ADR records; the product owner confirmed them as spec defaults.
- **The feature spec and tickets** are in `docs/issues/automap-v1/`: `spec.md` and 38 vertical-slice tickets, each with blocking edges and acceptance criteria, all `ready-for-agent`.
- **The definition-of-done review** is `docs/spec/definition-of-done.md`. All eight items are done, including two fresh-agent tests; the second found nothing foundational. The product owner signed it off on 2026-10-30.

## What happens next

1. **`exploration` merges to `main`** through a pull request, with a merge commit; `spike/testbed-agent` stays unmerged ([ADR-0078](../adr/0078-exploration-merges-to-main-after-review.md)).
2. **Ticket 01 starts on `main`.** It switches the instruction files to `AGENTS.md` ([ADR-0066](../adr/0066-build-repo-uses-agents-md.md)), which ends the exploration's "no product code" rule. Work the frontier of `docs/issues/automap-v1/` from there, with the build-team skills ([ADR-0067](../adr/0067-build-team-skills.md)).

## Prerequisites the build will hit

- **Ticket 09** needs the product owner to add an Anthropic API key as a repository secret and set its monthly cap ([ADR-0081](../adr/0081-eval-suite-runs-claude-code-headless.md)).
- **Real builds (ticket 04 onward)** need vcpkg to bootstrap and fetch sources. In this cloud container, the proxy returns 403 for GitHub archive and release downloads, so either run those on GitHub-hosted runners or allow those downloads in the environment's network settings. Go 1.24, Clang 18, CMake 3.28 and Mesa lavapipe are present here.

## Process rules that matter (all in `CLAUDE.md`)

- Documents state the current version only; git history is the record of changes.
- One question at a time, with a recommendation; say when a question needs the product owner's judgment.
- An ADR records only what the product owner agreed; ask about any new consequence.
- Don't build in assumptions that agents are weaker stewards of code than humans.
- Commit and push to the working branch; open a pull request only when asked.
