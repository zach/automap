---
status: accepted
date: 2026-10-29
---

# Evidence is about a binding, a pairing or a composition, on one platform

## Context

Evidence is graded ([ADR-0028](0028-evidence-is-graded-five-levels.md)), but it also has to say what it is about. Bindings connect library variants to contracts ([ADR-0011](0011-automap-owns-one-contract-per-seam.md), [ADR-0020](0020-build-variants-are-first-class.md)); pairings are exactly two libraries, and interactions among three or more were left to evidence about whole compositions ([ADR-0015](0015-pairings-are-exactly-two-libraries.md)); a composition's exact resolution is its lock ([ADR-0017](0017-composition-is-a-manifest-and-a-lock.md)). The research found the toolchain is part of what changes a built library: vcpkg's ABI hash includes the compiler, and Jolt's cross-platform determinism holds only under matching compiler settings ([06](../research/06-verification.md), [02c](../research/02c-catalog-simulation-runtime-glue.md)).

## Options

1. **Composition evidence only.**
2. **Binding and pairing evidence only.**
3. **Three subjects:** bindings, pairings and compositions.

And for platform: OS and architecture only, or OS, architecture and toolchain.

## Decision

Option 3. Evidence attaches to exactly three kinds of subject, each on exactly one platform:

- **A binding:** one library variant against one contract version.
- **A pairing:** two library variants.
- **A composition:** one lock.

A **platform** is an OS, a CPU architecture and a toolchain, such as Windows x64 with MSVC 2026.

## Consequences

- Binding and pairing evidence lets the map say a library works before anyone has composed it.
- Composition evidence catches interactions among three or more libraries that pairwise checks can't see.
- A toolchain change is a platform change, so it needs its own evidence.
