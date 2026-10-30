---
status: accepted
date: 2026-10-29
---

# Build variants are first-class in the map

## Context

Some libraries ship in configurations that are effectively different builds. Jolt and Box2D both have double-precision modes that change their ABI; Jolt packs 12 build defines into `JPH_VERSION_ID` and checks them at runtime because mismatches fail at link time or at runtime ([02c](../research/02c-catalog-simulation-runtime-glue.md)). vcpkg requires features to be additive, so such variants need overlay ports ([01](../research/01-package-managers.md)). Evidence for one variant says nothing about another ([06](../research/06-verification.md)).

## Options

1. **Free-form build options:** a composition passes whatever options it likes to a library's build.
2. **First-class variants:** a variant is a named set of build options for a library, with its own bindings, pairings and evidence.

## Decision

Option 2. A **variant** is a named set of build options for a library, such as "Jolt, double precision". Each variant has its own bindings, pairings and evidence, and a pin always names its variant.

## Consequences

- Every library has at least one variant, so every pin can name one.
- A pin can't silently point at an untested build.
- Where a build backend can't express a variant natively, Automap delivers it through that backend's own mechanism, such as a vcpkg overlay port.
