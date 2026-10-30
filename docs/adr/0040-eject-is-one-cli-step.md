---
status: accepted
date: 2026-10-29
---

# Ejecting is one CLI step, and build files are always committed

## Context

The persona's first demand is no lock-in ([05](../research/05-developer-needs-2026.md)). The research's cautionary tale is Our Machinery, whose license termination in 2022 left its users with nothing ([03](../research/03-prior-art.md)). Automap is never a runtime dependency ([ADR-0008](0008-no-automap-runtime-dependency.md)), glue lives in the developer's repo ([ADR-0002](0002-core-job-is-extend.md)), and builds are CMake on vcpkg ([ADR-0034](0034-vcpkg-is-the-v1-build-backend.md), [ADR-0035](0035-glue-targets-cmake-only.md)). The remaining dependency is the Automap registry's overlay ports.

## Options

1. **No eject path:** leaving Automap is the developer's problem.
2. **Documented manual steps.**
3. **One CLI step,** with build files kept as ordinary committed files so ejecting is always possible.

## Decision

Option 3. Ejecting is a single CLI step: it vendors any Automap registry ports into the developer's repo and removes the manifest and lock, leaving a plain CMake and vcpkg project that builds and runs without Automap. The build files the CLI produces, `vcpkg.json` and the CMake files, are ordinary files committed in the developer's repo, never generated on the fly.

## Consequences

- Leaving Automap never requires a rewrite.
- The developer's agent and CI can always read exactly what gets built without running Automap.
