---
status: accepted
date: 2026-10-30
---

# v1's catalog: twelve libraries filling fifteen slots

## Context

Beyond the day-one composition ([ADR-0057](0057-day-one-composition.md)), v1 needs enough slots for the starter-engine journey and enough libraries to exercise all eleven contracts ([ADR-0021](0021-the-seams.md)). The research's per-slot comparisons name a default pick for each slot ([02a](../research/02a-catalog-platform-graphics-ui-audio.md), [02b](../research/02b-catalog-assets-and-data.md), [02c](../research/02c-catalog-simulation-runtime-glue.md)), and its verification cost estimate assumed 25 libraries ([06](../research/06-verification.md)). A check of vcpkg's port tree at commit `d186940` (2026-10-29), made on 2026-10-30, found every library below: SDL3 3.4.18, Dear ImGui 1.93.0, Jolt 5.6.0, enkiTS 1.12, mimalloc 3.5.3, Tracy 0.14.1, spdlog 1.17.0, GLM 1.0.3, miniaudio 0.11.25, flecs 4.1.6, fastgltf 0.9.1, stb 2026-08-02 and SDL_shadercross 3.0.0-preview2 (a pinned commit).

## Options

1. **A minimal catalog:** only the day-one libraries.
2. **The research's default pick per slot,** enough to cover every contract.
3. **A broad catalog** near the research's 25-library estimate.

## Decision

Option 2. v1's catalog is twelve libraries filling fifteen slots:

| Slots | Library |
| --- | --- |
| windowing, input, GPU | SDL3 (SDL_GPU) |
| tools-UI | Dear ImGui (docking) |
| physics-3d | Jolt |
| jobs | enkiTS |
| allocator | mimalloc |
| profiling | Tracy |
| logging | spdlog |
| math | GLM |
| audio | miniaudio |
| ECS | flecs |
| model import | fastgltf |
| image decode | stb_image |
| shader compiler | SDL_shadercross |

## Consequences

- Every contract is exercised: allocation (mimalloc); jobs (enkiTS, Jolt, flecs); math types (GLM and Jolt); handles (flecs entities and Jolt bodies); diagnostics (spdlog, Tracy); build flags (Jolt's and Tracy's self-checks); determinism (Jolt's deterministic variant); assets (fastgltf, stb_image); lifecycle, serialization and platform across all of them.
- All are permissively licensed and carried by vcpkg; SDL_shadercross is pinned by commit.
