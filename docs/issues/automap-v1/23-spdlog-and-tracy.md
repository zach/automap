# 23: spdlog and Tracy as diagnostics backends

**What to build:** spdlog fills the logging slot and Tracy the profiling slot; extending with them sets `diagnostics.log` and `diagnostics.profiler`, and the diagnostics facade forwards to them. Tracy's link-time option check and which fmt spdlog uses are recorded as facts.

**Blocked by:** [15: Automap's CI verifies declared edges and signs evidence into the map](15-automap-ci-signs-evidence-into-the-map.md)

**Status:** done

- [x] A composition with both reaches probed on all three platforms in Automap's CI.
- [x] Profile hooks of libraries already present (such as Jolt's external profile hooks) reach Tracy through the facade.
- [x] A seeded Tracy option mismatch fails the build-flags rule.

## Comments

- spdlog 1.17.0 and Tracy 0.14.1 are the curated ports at the map's baseline (port version 1 each). The port builds spdlog with vcpkg's fmt, not its bundled copy, and records `SPDLOG_FMT_EXTERNAL` in `tweakme.h`; the SPEC's catalog says so.
- spdlog's binding writes the facade's messages to one never-destroyed logger, `am_spdlog_logger()`, whose sink is a `dist_sink_mt`; its probe adds a callback sink and logs through the facade.
- Tracy's binding opens a `tracy::ScopedZone` per facade zone. Its probe checks that a facade zone's events reach the calling thread's queue, which the client holds while no profiler is connected.
- The profile hooks already present that reach Tracy are enkiTS's wait callbacks. Jolt's external profile hooks need a Jolt built with `JPH_USE_EXTERNAL_PROFILE`, which the curated port doesn't offer; ticket 29 carries them with the Automap registry.
- verify reads Tracy's link-time option check from a failed link: a missing `<ns>::<name>_CFG_...` symbol is `ABI_DEFINE_MISMATCH(<ns>)`.
- Fixed on the way: `vcpkg.json` overrides now carry each port's `port-version`; without it vcpkg built tracy 0.14.1#0, which leaves `TRACY_ENABLE` off.
