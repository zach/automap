---
status: accepted
date: 2026-10-30
---

# The first three vertical slices

## Context

The brief requires the first slice to be a thin end-to-end path, such as one testbed composed, built and verified in CI. Day one is an SDL3 and Dear ImGui testbed extended with Jolt ([ADR-0057](0057-day-one-composition.md)), on three desktop platforms ([ADR-0056](0056-v1-platforms-desktop-only.md)). Headless graphics is most mature on Linux, through SDL's offscreen driver and Mesa's software Vulkan driver (lavapipe) ([06](../research/06-verification.md)). Published evidence is signed and stored in the map, and the public map ships as dated releases recorded in locks ([ADR-0030](0030-evidence-is-signed-and-stored-in-the-map.md), [ADR-0053](0053-map-releases-are-immutable-and-recorded-in-locks.md)).

## Options

1. **Linux first, then extend, then the platform matrix with published evidence.**
2. **All three platforms from slice 1.**
3. **The whole day-one path (start and extend) in slice 1.**

## Decision

Option 1. The first three vertical slices are:

1. **Start a testbed, on Linux x64.** The map contains just SDL3 and Dear ImGui. The CLI's `start` writes the manifest, lock, `vcpkg.json`, CMake files and reference glue, plus the `AGENTS.md` section and the router skill. The testbed builds and runs headless for a set number of frames, its probes pass, and the CLI writes evidence at the probed level. **Acceptance:** a fresh agent session given a one-line request ends with a running testbed whose verification reports "probed".
2. **Extend with Jolt, on Linux.** Adds the jobs interface and its thread pool, and probes for allocation, lifecycle and build flags, including Jolt's `VerifyJoltVersionID()` and a ThreadSanitizer run.
3. **Three platforms and published evidence.** Windows and macOS join Automap's CI; evidence is signed and stored in the map; the first map release is published and locks record it.

Upgrade, adopt, eject, the rest of the catalog and private maps follow as later slices.

## Consequences

- Every later slice extends a path that already works end to end.
- Slice 2 proves the contracts on the hardest seams; slice 3 makes evidence trustworthy beyond the developer.
