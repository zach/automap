---
status: accepted
date: 2026-10-29
---

# v1 supports three desktop platforms; web is deferred past v1

## Context

GDC 2026 puts PC at 83% of developers' most recent projects, and Steam Deck at 28%, reached through Proton ([05](../research/05-developer-needs-2026.md)). Windows, Linux and macOS (including a paravirtual Metal device on macOS arm64) all run on hosted CI runners ([06](../research/06-verification.md)). SDL_GPU has no web backend, so a web testbed would need a second graphics stack (sokol, raylib, or SDL with WebGPU through Dawn) with its own bindings, pairings and evidence ([02a](../research/02a-catalog-platform-graphics-ui-audio.md)). Web builds are not a requirement for the primary persona ([ADR-0001](0001-primary-persona-framework-first-indie.md)). A platform is an OS, a CPU architecture and a toolchain ([ADR-0029](0029-evidence-subjects-and-platform.md)).

## Options

1. **Desktop only:** Windows, Linux and macOS.
2. **Desktop plus web,** through a second graphics stack.
3. **Desktop plus Linux arm64 and other Steam Deck paths.**

## Decision

Option 1. v1 supports exactly three platforms:

1. Windows x64 with MSVC.
2. Linux x64 with Clang.
3. macOS arm64 with Apple Clang.

A web-capable testbed is deferred past v1.

## Consequences

- v1's evidence matrix has one graphics stack and three platforms, all verifiable on hosted runners.
- Linux arm64 and other platforms can join once the matrix is proven.
