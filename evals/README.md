# Eval suite

Fresh Claude Code sessions in headless mode doing Automap's jobs along the worked journeys (SPEC §8.7, ADR-0081). Each task runs in a clean repo with only the build toolchain, the `automap` CLI on the PATH and the router skill installed at user scope, given a one-line request and no human input. The runner grades the outcome through `automap verify --json` and writes a machine-readable report: each task's pass, fail or skipped, its duration, and whether the vcpkg binary cache was warm or cold, with the model that answered (ADR-0063).

- `tasks/<id>/task.json`: the request; the starting repo (a directory beside it, or none for an empty repo) and any `automap` invocations that bring it to the task's starting point, such as starting the testbed; the level verification must reach; and the library each named slot must hold.
- Tasks: `testbed-start` (Journey 1's first step, from an empty repo) and `testbed-extend` (its extend step: "Add 3D physics." on the started testbed, passing with Jolt in `physics-3d` at probed).
- `cmd/run`: the runner. `go run ./evals/cmd/run -task testbed-start -automap <cli> -map map -report report.json`.
- `.github/workflows/evals.yml` runs the suite on demand (Actions, "Evals", Run workflow). The Anthropic API key is the repository secret `ANTHROPIC_KEY`, read only by that workflow; a run its spending cap stops is reported as skipped, not failed.
