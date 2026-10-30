# 35: automap eject

**What to build:** `automap eject` vendors any Automap registry ports into the repo, points the vcpkg configuration at them, and removes the manifest and lock, leaving a plain CMake and vcpkg project (ADR-0040).

**Blocked by:** [16: Map releases, fetch and cache, and evidence states](16-map-releases-and-evidence-states.md), [29: Handles, serialization and determinism; Jolt's variants through the Automap registry](29-handles-serialization-determinism-and-registry.md)

**Status:** ready-for-agent

- [ ] After `eject`, the fixture composition and both reference compositions build and pass their smoke test with Automap uninstalled.
- [ ] `eject` refuses with `LOCK_EDITED` on a hand-edited lock.
- [ ] Glue, shared interfaces and build files are left in place unchanged.
