---
status: accepted
date: 2026-10-31
---

# NDA platforms are deferred indefinitely

## Context

Each NDA platform's content was to live in its own private map ([ADR-0048](0048-one-private-map-per-nda-platform.md)), with evidence signed by the NDA holder's own key ([ADR-0049](0049-nda-evidence-from-nda-holders-only.md)). Building that needs a signature format, key configuration and a way to build for a platform other than the host. Automap has no users yet, and supporting NDA platforms was always meant as a possibility the design leaves room for, not a requirement.

## Options

1. **Build NDA handling in v1:** NDA maps, NDA signing and the `custom-nda` eval.
2. **Defer it indefinitely:** leave room for it, and design and build none of it.

## Decision

Option 2. NDA handling is deferred indefinitely: no design or code for NDA maps, NDA evidence signing or the stand-in NDA map. ADR-0048 and ADR-0049 stand as the direction if it is ever taken up.

## Consequences

- v1 targets the three desktop platforms only ([ADR-0056](0056-v1-platforms-desktop-only.md)).
- What already exists stays: per-platform variant choices and pins ([ADR-0075](0075-variants-per-platform.md)), private maps' `platforms`, and the spec's Nebula examples.
- The `custom-nda` eval is dropped from the eval suite.
