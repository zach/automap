---
status: accepted
date: 2026-11-10
---

# The public map covers text and data files

## Context

v1's catalog is thirteen libraries filling fifteen slots ([ADR-0058](0058-v1-catalog.md)), chosen to exercise every contract. The first game built on it, Spintacular, needed two things the catalog doesn't have before it needed several it does: text drawn from fonts, and structured data files for balance tables, saves and settings. It filled both from a private map, with a `text` slot holding stb_truetype and a `data-files` slot holding nlohmann/json, and wrote stb_truetype's allocation binding, its probe and a seeded violation there; both reached probed on Windows and Linux ([08](../research/08-spintacular-field-report.md)). stb_truetype comes from the vcpkg port stb_image comes from. What the map covers is the product owner's decision ([ADR-0069](0069-human-checkpoints-only-where-direction-is-set.md)).

## Options

1. **Both:** the `text` slot with stb_truetype and the `data-files` slot with nlohmann/json, as in Spintacular's private map.
2. **Text only,** leaving data files to private maps.
3. **Neither:** the public map's coverage stays as it is.

## Decision

Option 1. The public map gains two slots and two libraries:

| Slot | Library |
| --- | --- |
| text | stb_truetype |
| data-files | nlohmann/json |

## Consequences

- The public map covers fifteen libraries filling seventeen slots. The starter engine stays v1's thirteen.
- stb_truetype is bound to the allocation contract, with a probe and a seeded violation, and verified on all three platforms like every other binding.
- stb_truetype and stb_image share the `stb` port in one composition.
- The watcher follows nlohmann/json's releases; stb has none and is pinned by commit, as stb_image is.
