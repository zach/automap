---
status: accepted
date: 2026-10-29
---

# Glue and generated build files target CMake only in v1

## Context

vcpkg is v1's only build backend ([ADR-0034](0034-vcpkg-is-the-v1-build-backend.md)) and integrates through a CMake toolchain file. The research found CMake package names and imported targets are the one contract shared across vcpkg, Conan and CPM, so glue written against them survives a change of build backend ([01](../research/01-package-managers.md)). The developer's own code is a library like any other ([ADR-0013](0013-developer-code-is-a-library.md)).

## Options

1. **CMake only,** with glue referring to libraries by CMake package and imported-target names.
2. **CMake plus other build systems** (Premake, xmake, Meson).
3. **CMake with vcpkg-specific variables** in glue.

## Decision

Option 1. In v1, glue and generated build files target CMake only. Glue refers to libraries only by their CMake package and imported-target names, such as `find_package(SDL3)` and `SDL3::SDL3`, never by vcpkg-specific variables. The developer's own libraries join the build as ordinary CMake targets in their repo.

## Consequences

- Glue survives a later Conan backend without a rewrite.
- Premake, xmake, Meson and other build systems are out of scope for v1 rather than half-supported.
