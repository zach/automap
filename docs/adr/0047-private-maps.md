---
status: accepted
date: 2026-10-29
---

# Private content lives in private maps with the public map's schema

## Context

The developer's own code is a library like any other ([ADR-0013](0013-developer-code-is-a-library.md)), but an in-house library's facts, bindings and evidence can't live in the public map; neither can a studio's proprietary libraries, patched libraries or console content. The brief requires private and proprietary libraries to be supported "without special cases". vcpkg handles private registries with the same format as its curated registry, in a different location ([01](../research/01-package-managers.md)). Variants are first-class ([ADR-0020](0020-build-variants-are-first-class.md)) and are delivered through overlay ports where vcpkg can't express them ([ADR-0034](0034-vcpkg-is-the-v1-build-backend.md)).

## Options

1. **Private content in the public map,** hidden by access control.
2. **Special-purpose private formats** for in-house libraries, patches and consoles.
3. **Private maps:** the same schema as the public map, held where the developer controls it, layered over the public map by the CLI.

## Decision

Option 3. Private content lives in **private maps**. A private map uses exactly the same schema as the public map, sits in the developer's repo or another repo they control, and is layered over the public map by the CLI. It holds:

- in-house and proprietary libraries;
- private variants, including patched libraries;
- recipes;
- evidence.

The CLI never uploads a private map's content anywhere.

## Consequences

- A patched library is a private variant with its own pin and its own evidence; there is no separate patch concept.
- Private variants that need different builds are delivered through vcpkg's own mechanisms, in overlay ports the developer controls.
- Every CLI check works the same on public and private content.
