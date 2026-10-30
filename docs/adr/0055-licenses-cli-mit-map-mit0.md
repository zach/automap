---
status: accepted
date: 2026-10-29
---

# The CLI is MIT; everything in the map is MIT-0

## Context

Automap is fully open source ([ADR-0054](0054-open-source-non-commercial.md)), and the repo is already MIT. Reference glue and shared interfaces are copied into developers' engines and ship inside their games; glue becomes the developer's own code ([ADR-0002](0002-core-job-is-extend.md)). MIT requires its notice to travel with every copy, a chore for code that becomes the developer's. The single-file libraries this persona already uses avoid that: miniaudio offers public domain or MIT-0, and stb offers public domain or MIT ([02a](../research/02a-catalog-platform-graphics-ui-audio.md), [02b](../research/02b-catalog-assets-and-data.md)). The CLI is never shipped inside a game.

## Options

1. **MIT for everything.**
2. **MIT for the CLI; MIT-0 for everything in the map.**
3. **Public domain (or CC0) for the map.**

## Decision

Option 2. The CLI stays MIT, matching the repo. Everything in the map (facts, recipes, reference glue, probes and shared interfaces) is MIT-0, so anything copied from the map can ship inside a game with no attribution.

## Consequences

- Glue carries no one else's license obligations into the developer's game.
- Contributions to the map must be made under MIT-0.
