# 05: Probe harness and the probed level

**What to build:** The composition's executable supports `--probe`: after startup and the smoke frames it runs every probe, records rule results and each lifecycle step through the probe harness, and writes a results file. `verify` reads it, checks lifecycle order and pairing ordering constraints, assigns each rule a status, and reaches the probed level when every applicable probed rule passes.

**Blocked by:** [04: Verify reaches smoke](04-verify-reaches-smoke.md)

**Status:** ready-for-agent

- [ ] The fixture composition's probes pass and `verify` reports probed.
- [ ] `automap verify --seed <rule>[:<library>]` applies a seeded violation in a scratch copy, writes no evidence, and reports whether the probe caught it; for every seed in the fixture map, the probe does (ADR-0071).
- [ ] Not-applicable and unverifiable results carry reasons; an unverifiable applicable rule blocks probed (ADR-0072).
- [ ] Lifecycle order and pairing ordering constraints are checked from the recorded steps; violations report `INIT_ORDER(...)` or `SHUTDOWN_ORDER(...)` (ADR-0026, ADR-0074).
- [ ] A platform-limited rule is not-applicable on other platforms (ADR-0076).
- [ ] An init entry point that fails stops startup with `INIT_FAILED(<call>)`.
- [ ] Level computation is a pure function with unit tests.
