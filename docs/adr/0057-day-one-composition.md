---
status: accepted
date: 2026-10-30
---

# Day one starts an SDL3 and Dear ImGui testbed, then extends it with Jolt

## Context

The day-one minimum is start, then extend once ([ADR-0010](0010-day-one-minimum-start-then-extend-once.md)), on three desktop platforms ([ADR-0056](0056-v1-platforms-desktop-only.md)). The brief's planned spike uses SDL3, SDL_GPU, Dear ImGui and Jolt. The research rated SDL3 with SDL_GPU and Dear ImGui (docking branch, with its first-party SDL_GPU backend) as the strongest desktop testbed ([02a](../research/02a-catalog-platform-graphics-ui-audio.md)). Dear ImGui's SDL_GPU backend embeds its own precompiled shaders. SDL_shadercross, the usual shader path for SDL_GPU, is untagged and bundles its own compilers. Jolt touches allocator hooks, a job system, a strict init order and ABI-affecting defines that it checks itself at runtime ([02c](../research/02c-catalog-simulation-runtime-glue.md)).

## Options

1. **SDL3 and Dear ImGui, extended with Jolt.**
2. **raylib, extended with Box2D.**
3. **SDL3 with a custom renderer and shader toolchain from day one.**

## Decision

Option 1.

- **Start:** a testbed with SDL3 (windowing, input, and GPU through SDL_GPU) and the docking branch of Dear ImGui with its SDL3 and SDL_GPU backends. No custom shaders: Dear ImGui's SDL_GPU backend ships its own.
- **Extend once:** add Jolt for 3D physics. The jobs interface is implemented by a small thread pool in reference glue, so the extend step adds exactly one library.

## Consequences

- The extend step exercises allocation, jobs, lifecycle and build-flag contracts, so it proves the contracts rather than just a build.
- No shader toolchain is needed on day one.
