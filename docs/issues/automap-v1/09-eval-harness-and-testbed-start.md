# 09: Eval harness and the testbed-start eval

**What to build:** The eval suite can run a fresh agent session against a task: Claude Code in headless mode on a GitHub-hosted runner, in a clean environment holding only the platform's build toolchain, the CLI and the task's starting repo, with a one-line request, no human input, and a grader that reads the outcome through the CLI's JSON output (ADR-0081). The first task is journey 1's start.

**Blocked by:** [06: Agent files: AGENTS.md section and router skill](06-agent-files.md), [08: SDL3 and Dear ImGui in the map](08-sdl3-and-dear-imgui-in-the-map.md)

**Status:** ready-for-agent

- [ ] The `testbed-start` eval passes: from an empty repo and "Start a C++ testbed with SDL3 and Dear ImGui.", the session ends with a running testbed whose verification reports probed, with no human input.
- [ ] The runner records pass or fail, duration, and warm- or cold-cache, per task, in a machine-readable report (ADR-0063).
- [ ] Each report names the Claude model used; the API key is a repository secret available only to the eval workflow.
- [ ] A run stopped by the key's spending cap is reported as skipped, not failed.
- [ ] Evals can be run on demand in CI.
- [ ] Before this ticket can pass, the product owner adds the API key as a repository secret and sets its monthly cap.
