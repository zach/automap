---
status: accepted
date: 2026-10-31
---

# NDA platforms are out of scope

## Context

Each NDA platform's content was to live in its own private map ([ADR-0048](0048-one-private-map-per-nda-platform.md)), with evidence signed by the NDA holder's own key ([ADR-0049](0049-nda-evidence-from-nda-holders-only.md)). Building that needs a signature format, key configuration and a way to build for a platform other than the host: a change to how evidence is made, signed and read, not an addition. Automap has no users asking for it, and supporting NDA platforms was always a possibility the design left room for, not a requirement.

## Options

1. **Build NDA handling in v1:** NDA maps, NDA signing and the `custom-nda` eval.
2. **Defer it:** leave room for it, and design and build none of it until someone needs it.
3. **Rule it out:** NDA platforms are not something Automap supports; taking them up would be a new product decision, not a resumed one.

## Decision

Option 3. NDA platforms are out of scope: Automap designs and builds nothing for NDA maps, NDA evidence signing or a stand-in NDA map, and keeps no open question about them. ADR-0048 and ADR-0049 describe the shape such support would have had; nothing is built on them.

## Consequences

- Automap targets the three desktop platforms only ([ADR-0056](0056-v1-platforms-desktop-only.md)).
- What already exists stays: per-platform variant choices and pins ([ADR-0075](0075-variants-per-platform.md)), private maps' `platforms`, and the spec's Nebula examples.
- The `custom-nda` eval is dropped from the eval suite.
