# 21: enkiTS and the jobs slot

**What to build:** enkiTS enters the map and fills the jobs slot. Extending with it sets the `jobs.scheduler` choice to `jobs-slot`, writes enkiTS's implementation of the jobs interface, and removes the reference thread pool; Jolt's jobs binding doesn't change.

**Blocked by:** [12: Jolt in the map](12-jolt-in-the-map.md), [15: Automap's CI verifies declared edges and signs evidence into the map](15-automap-ci-signs-evidence-into-the-map.md)

**Status:** ready-for-agent

- [ ] `automap extend enkits` on testbed + Jolt ends probed on all three platforms in Automap's CI.
- [ ] enkiTS's allocation and profiler-callback bindings route to the facades.
- [ ] Every new probe fails its seeded violation.
