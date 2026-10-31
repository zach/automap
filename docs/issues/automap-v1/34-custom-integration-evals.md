# 34: Custom-integration evals

**What to build:** Journey 3 as eval tasks: adopt and a patched variant, using a fixture copy of Acme's engine.

**Blocked by:** [09: Eval harness and the testbed-start eval](09-eval-harness-and-testbed-start.md), [32: automap adopt](32-automap-adopt.md)

**Status:** done

- [x] `custom-adopt`: the session ends with a manifest, a private map describing the in-house libraries, and probed desktop evidence.
- [x] `custom-variant`: the session adds the patched SDL3 variant and ends probed.

## Comments

Both tasks passed against map release `map-2026.10.31.1` in the Evals workflow's run [41231960748](https://github.com/zach/automap/actions/runs/41231960748): `custom-adopt` reached probed in 275 s and `custom-variant` in 317 s, each with no human input.
