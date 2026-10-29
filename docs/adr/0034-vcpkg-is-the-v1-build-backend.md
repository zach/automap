---
status: accepted
date: 2026-10-29
---

# vcpkg is v1's only build backend, extended by an Automap registry

## Context

Automap sits above existing package managers ([ADR-0007](0007-automap-sits-above-build-backends.md)). The research scored vcpkg and Conan as roughly tied but winning on different axes ([01](../research/01-package-managers.md)):

- **vcpkg** carries the most current catalog for the target libraries (21 of 23 carried, 18 at the latest release, against Conan Center's 9), has a JSON manifest with a published schema, and integrates natively with CMake. Its features must be additive, so non-additive variants need overlay ports; some libraries are missing (ufbx, ozz-animation) or stale (sokol from 2023, bgfx).
- **Conan** is strongest for studios with private Artifactory remotes and prebuilt binaries, non-additive options and signing.

Every additional backend multiplies the verification matrix, because a variant can exist in one backend and not another.

## Options

1. **vcpkg only.**
2. **Conan only.**
3. **vcpkg and Conan from v1.**

## Decision

Option 1. v1 supports exactly one build backend, vcpkg. Automap runs its own registry of overlay ports for:

- variants vcpkg's features can't express, such as Jolt double precision;
- libraries vcpkg lacks, such as ufbx and ozz-animation;
- fresh commits of libraries vcpkg packages stale, such as sokol and bgfx.

Conan support is considered after v1.

## Consequences

- v1's evidence matrix has one build-backend dimension.
- Automap extends vcpkg through its own mechanism rather than working around it.
- A canonical lock identity shared across build backends (OQ-011) is not needed in v1.
