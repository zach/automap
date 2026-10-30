---
status: accepted
date: 2026-10-29
---

# Automap sits above existing package managers

## Context

Automap could fetch and build libraries itself, or delegate that to existing C++ package managers and add only what they lack. Research ([01](../research/01-package-managers.md)):

- vcpkg carries 21 of the 23 key libraries, 18 at their latest release, with private registries, binary caching and SBOMs. Conan is strongest on private remotes, binaries and non-additive options.
- Package managers already encode build-level fit: feature-to-dependency edges, platform `supports`, ABI-define propagation, file-conflict rules.
- None can express allocator, threading, math-type or init-order seams, or evidence that libraries were tested together.

## Options

1. **Replace package managers:** Automap fetches, builds and caches libraries itself.
2. **Sit above them:** Automap translates a composition into a package manager's own manifests and configuration, and leaves fetching and building to it.

## Decision

Option 2. Automap sits above existing package managers, which act as its **build backends**. Automap's own content is the layer they can't express: seams, contracts, pairings and glue.

## Consequences

- Automap does not fetch or build libraries itself; it produces inputs for a build backend (for example vcpkg features and overlay ports).
- Which build backend or backends Automap supports is decided in branch 6.
