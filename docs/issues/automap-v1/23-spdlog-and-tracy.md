# 23: spdlog and Tracy as diagnostics backends

**What to build:** spdlog fills the logging slot and Tracy the profiling slot; extending with them sets `diagnostics.log` and `diagnostics.profiler`, and the diagnostics facade forwards to them. Tracy's link-time option check and which fmt spdlog uses are recorded as facts.

**Blocked by:** [15: Automap's CI verifies declared edges and signs evidence into the map](15-automap-ci-signs-evidence-into-the-map.md)

**Status:** ready-for-agent

- [ ] A composition with both reaches probed on all three platforms in Automap's CI.
- [ ] Profile hooks of libraries already present (such as Jolt's external profile hooks) reach Tracy through the facade.
- [ ] A seeded Tracy option mismatch fails the build-flags rule.
