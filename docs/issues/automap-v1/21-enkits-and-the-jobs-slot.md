# 21: enkiTS and the jobs slot

**What to build:** enkiTS enters the map and fills the jobs slot. Extending with it sets the `jobs.scheduler` choice to `jobs-slot`, writes enkiTS's implementation of the jobs interface, and removes the reference thread pool; Jolt's jobs binding doesn't change.

**Blocked by:** [12: Jolt in the map](12-jolt-in-the-map.md), [15: Automap's CI verifies declared edges and signs evidence into the map](15-automap-ci-signs-evidence-into-the-map.md)

**Status:** done

- [x] `automap extend enkits` on testbed + Jolt ends probed on all three platforms in Automap's CI.
- [x] enkiTS's allocation and profiler-callback bindings route to the facades.
- [x] Every new probe fails its seeded violation.

## Comments

- Filling a slot sets every contract choice whose enum holds `<slot>-slot`, so `jobs.scheduler` becomes `jobs-slot` without a rule naming enkiTS.
- The three bindings share one `enki::TaskScheduler` and its config (`am_enkits_config()`, `am_enkits_scheduler()`), so allocation and profiler callbacks are set before the scheduler starts.
- The profiler callbacks log thread start and stop through diagnostics and open zones for waits.
- The jobs binding registers up to 8 external threads, enkiTS's configured limit.
- The TSan seed runs only in reference-composition jobs, where the composition puts the jobs slot under test.
