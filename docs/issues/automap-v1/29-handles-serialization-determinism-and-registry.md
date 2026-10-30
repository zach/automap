# 29: Handles, serialization and determinism; Jolt's variants through the Automap registry

**What to build:** The handles, serialization and determinism contracts arrive. The Automap registry is published on its own branch and supplies Jolt's double-precision and deterministic variants, which vcpkg's additive features can't express.

**Blocked by:** [12: Jolt in the map](12-jolt-in-the-map.md), [15: Automap's CI verifies declared edges and signs evidence into the map](15-automap-ci-signs-evidence-into-the-map.md)

**Status:** ready-for-agent

- [ ] The registry branch is a valid vcpkg git registry, and compositions that use it record it at a commit in their pins.
- [ ] Jolt's double-precision and deterministic variants reach probed on all three platforms; the deterministic variant passes a replay-hash check under the determinism profile.
- [ ] Handles and serialization probes fail their seeded violations for Jolt (and flecs once present).
