---
status: accepted
date: 2026-10-29
---

# A pin names library, version or commit, variant, registry and content hash

## Context

A lock holds exact pins ([ADR-0017](0017-composition-is-a-manifest-and-a-lock.md)), and evidence names exact pins ([ADR-0030](0030-evidence-is-signed-and-stored-in-the-map.md)), so a pin must identify exactly what was built. Several key libraries have no release tags: sokol, bgfx and SDL_shadercross ([02a](../research/02a-catalog-platform-graphics-ui-audio.md)). vcpkg makes builds reproducible through a registry baseline commit and pins every download by SHA512 ([01](../research/01-package-managers.md)). Variants are first-class ([ADR-0020](0020-build-variants-are-first-class.md)), and the Automap registry supplies overlay ports alongside vcpkg's curated registry ([ADR-0034](0034-vcpkg-is-the-v1-build-backend.md)).

## Options

1. **Library and version only.**
2. **Library, version or commit, variant, registry at a commit, and a content hash.**

## Decision

Option 2. A pin consists of:

1. the library;
2. its version, or the exact commit for libraries without release tags;
3. its variant;
4. the registry it came from: vcpkg's curated registry at a baseline commit, or the Automap registry at a commit;
5. a content hash of the source.

## Consequences

- Two developers holding the same pin build the same code, so evidence about one pin can't vouch for a different build.
- Libraries released by commit are pinned honestly, not by an invented version.
