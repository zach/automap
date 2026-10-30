# 05: Probe harness and the probed level

**What to build:** The composition's entry point supports `--probe`: after init and the smoke frames it runs every probe, records rule results and each lifecycle step through the probe harness, and writes a results file. `verify` reads it, checks lifecycle order and pairing ordering constraints, assigns each rule a status, and reaches the probed level when every applicable probed rule passes.

**Blocked by:** [04: Verify reaches smoke](04-verify-reaches-smoke.md)

**Status:** ready-for-agent

- [ ] The fixture composition's probes pass and `verify` reports probed.
- [ ] For every seeded-violation patch in the fixture map, a test applies it and `verify` reports that rule's failure code and a level below probed (ADR-0071).
- [ ] Not-applicable and unverifiable results carry reasons; an unverifiable applicable rule blocks probed (ADR-0072).
- [ ] Lifecycle order and pairing ordering constraints are checked from the recorded steps; violations report `INIT_ORDER(...)` or `SHUTDOWN_ORDER(...)` (ADR-0026, ADR-0074).
- [ ] A platform-limited rule is not-applicable on other platforms (ADR-0076).
- [ ] Level computation is a pure function with unit tests.
