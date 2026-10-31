# 29: Handles, serialization and determinism; Jolt's variants through the Automap registry

**What to build:** The handles, serialization and determinism contracts arrive. The Automap registry is published on its own branch and supplies Jolt's double-precision and deterministic variants, which vcpkg's additive features can't express.

**Blocked by:** [12: Jolt in the map](12-jolt-in-the-map.md), [15: Automap's CI verifies declared edges and signs evidence into the map](15-automap-ci-signs-evidence-into-the-map.md)

**Status:** done

- [x] The registry branch is a valid vcpkg git registry, and compositions that use it record it at a commit in their pins.
- [x] Jolt's double-precision and deterministic variants reach probed on all three platforms; the deterministic variant passes a replay-hash check under the determinism profile.
- [x] Jolt's builds from the registry use Jolt's external profile hooks (`JPH_USE_EXTERNAL_PROFILE`), and Jolt's diagnostics binding routes them to the facade's zones, so they reach the profiling slot; the curated joltphysics port builds Jolt's own profiler instead.
- [x] Handles and serialization probes fail their seeded violations for Jolt (and flecs once present).

## Comments

- The registry's ports live on main in `ci/registry/ports/`; `ci/cmd/registry` copies them onto the `registry` branch and records each port's git tree in vcpkg's versions database, refusing a changed port whose version wasn't raised. The `Publish registry` workflow runs it on push to main and points `map/map.json` at the new commit.
- The registry's joltphysics port is vcpkg's with `JPH_USE_EXTERNAL_PROFILE` always on and `double-precision` and `deterministic` features; Jolt's two registry variants select them. A registry variant takes the registry's port version, not the curated port's.
- Jolt's diagnostics binding opens a facade zone per external profile scope: through `JPH::ProfileStartMeasurement` in a shared Jolt, by defining `ExternalProfileMeasurement`'s constructor and destructor in a static one. Its probe checks that a `JPH_PROFILE` scope reaches the glue.
- The handles, serialization and determinism contracts are 1.0.0. Rules may name the choices they apply under (`when`); both determinism rules apply only under `determinism.profile` on, and `determinism.flags-set` is computed from each library's determinism facts in its variant. Its seed turns the profile on where Jolt isn't deterministic, as the precision seed does.
- The replay hash (`0x0a9c29e881ff7288`, 16 boxes for 120 steps) matched on Linux, Windows and macOS, in Debug and Release.
- `automap.cmake` defines `AUTOMAP_VARIANT_<ID>` for rebuild keys and turns off floating-point contraction under the determinism profile; `verify` rewrites the generated files first, so an edited choice takes effect.
- Automap CI's reference compositions take no choices, so it publishes no evidence under the determinism profile; the public-map test covers it.
- A composition whose every library comes from the Automap or a private registry gets no default curated registry, so registry ports can't find vcpkg-cmake; every composition today includes SDL3 from the curated registry.
- The public-map CI job is split into three groups per platform.
