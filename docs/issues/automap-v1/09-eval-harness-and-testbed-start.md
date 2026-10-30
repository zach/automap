# 09: Eval harness and the testbed-start eval

**What to build:** The eval suite can run a fresh agent session against a task: Claude Code in headless mode on a GitHub-hosted runner, in a clean environment holding only the platform's build toolchain, the CLI and the task's starting repo, with a one-line request, no human input, and a grader that reads the outcome through the CLI's JSON output (ADR-0081). The first task is journey 1's start.

**Blocked by:** [06: Agent files: AGENTS.md section and router skill](06-agent-files.md), [08: SDL3 and Dear ImGui in the map](08-sdl3-and-dear-imgui-in-the-map.md)

**Status:** done

- [x] The `testbed-start` eval passes: from an empty repo and "Start a C++ testbed with SDL3 and Dear ImGui.", the session ends with a running testbed whose verification reports probed, with no human input.
- [x] The runner records pass or fail, duration, and warm- or cold-cache, per task, in a machine-readable report (ADR-0063).
- [x] Each report names the Claude model used; the API key is a repository secret available only to the eval workflow.
- [x] A run stopped by the key's spending cap is reported as skipped, not failed.
- [x] Evals can be run on demand in CI.
- [x] Before this ticket can pass, the product owner adds the API key as a repository secret and sets its monthly cap.

## Comments

- The `testbed-start` eval passed on its first run, on demand from `main` (Actions run 41097192702): a fresh Claude Code session with the CLI on its PATH and the router skill at user scope started the testbed, and the grader's `automap verify --json` reported probed, in 34 seconds with a warm vcpkg binary cache.
- The eval workflow runs on demand only, at the product owner's request, so the key's spend happens only when someone asks for a run. The runner prints each task's result, model, level and cost to the job log; the report and session transcripts are an artifact of each run.
