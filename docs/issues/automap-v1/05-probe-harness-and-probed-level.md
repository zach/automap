# 05: Probe harness and the probed level

**What to build:** The composition's executable supports `--probe`: after startup and the smoke frames it runs every probe, records rule results and each lifecycle step through the probe harness, and writes a results file. `verify` reads it, checks lifecycle order and pairing ordering constraints, assigns each rule a status, and reaches the probed level when every applicable probed rule passes.

**Blocked by:** [04: Verify reaches smoke](04-verify-reaches-smoke.md)

**Status:** done

- [x] The fixture composition's probes pass and `verify` reports probed.
- [x] `automap verify --seed <rule>[:<library>]` applies a seeded violation in a scratch copy, writes no evidence, and reports whether the probe caught it; for every seed in the fixture map, the probe does (ADR-0071).
- [x] Not-applicable and unverifiable results carry reasons; an unverifiable applicable rule blocks probed (ADR-0072).
- [x] Lifecycle order and pairing ordering constraints are checked from the recorded steps; violations report `INIT_ORDER(...)` or `SHUTDOWN_ORDER(...)` (ADR-0026, ADR-0074).
- [x] A platform-limited rule is not-applicable on other platforms (ADR-0076).
- [x] An init entry point that fails stops startup with `INIT_FAILED(<call>)`.
- [x] Level computation is a pure function with unit tests.

## Comments

- Probe glue registers each probe with the harness (`am::probe::Registration`), to run after the frames or after shutdown. Late allocator hooks only show once a library frees what it allocated, so the allocation probes check `allocation.hooks-before-first-allocation` after shutdown. The glue target is an object library so registrations aren't dropped by the linker.
- `verify` computes `lifecycle.init-order`, `lifecycle.shutdown-reverse` and `platform.targets-supported` itself, from the recorded steps and the manifest; every other rule's result comes from the harness. A rule with no result is unverifiable, which blocks probed and is reported as an `UNVERIFIABLE(<rule>[,<library>])` warning.
- `verify --seed` applies a unified diff by mapping each reference file to the composition's file through the manifest, or a JSON Patch to the manifest, in a scratch copy without `build/` or `evidence/`.
- The fixture map's twelve seeds are each tested; a test fails if the map gains a seed the test doesn't cover.
