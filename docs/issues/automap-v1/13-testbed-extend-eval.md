# 13: The testbed-extend eval

**What to build:** Journey 1's extend step as an eval task.

**Blocked by:** [09: Eval harness and the testbed-start eval](09-eval-harness-and-testbed-start.md), [12: Jolt in the map](12-jolt-in-the-map.md)

**Status:** done

- [x] From the testbed and "Add 3D physics.", a fresh session ends with Jolt in physics-3d and verification at probed, with no human input.

## Comments

- An eval task can set up its starting point with `automap` invocations before the session (`testbed-extend` starts the SDL3 and Dear ImGui testbed) and can require a library in a named slot, read from the manifest after grading through `automap verify --json`.
- The Evals workflow's run 41134851896 (2026-10-30) passed `testbed-extend` at probed with Jolt in `physics-3d`, in 60 seconds on a warm vcpkg binary cache.
