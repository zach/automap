---
status: accepted
date: 2026-10-30
---

# A composition can choose a library's variant per platform

## Context

A composition names one variant per library ([ADR-0020](0020-build-variants-are-first-class.md)), and a pin always names its variant ([ADR-0036](0036-what-a-pin-contains.md)). Each NDA platform's content lives in its own private map, and the public map says only that a console variant exists and who distributes it ([ADR-0048](0048-one-private-map-per-nda-platform.md)). SDL keeps one private repo per console ([02a](../research/02a-catalog-platform-graphics-ui-audio.md)), so a studio shipping on desktop and a console needs SDL's public source on desktop and the console port on the console. Working the custom-integration journey ([ADR-0060](0060-three-worked-journeys.md)) showed that one variant per library across every platform can't express this. Libraries that need no console variant, such as Jolt, still have to be declared as supporting the console before the platform rule ([ADR-0027](0027-v1-probe-floor-and-classification.md)) can pass, and that declaration can't appear in the public map.

## Options

1. **One variant per library:** a studio keeps a separate composition for each NDA platform.
2. **Variants per platform:** the manifest may choose a library's variant per platform, and an NDA platform's private map extends libraries from the maps beneath it.

## Decision

Option 2.

- The manifest chooses each library's variant either once for every platform or per platform, with a default.
- The lock pins each platform whose variant differs from the default.
- An NDA platform's private map may add console variants, each with its own pin, and may add its platform to a library's supported targets. It changes nothing else in the libraries it extends.

## Consequences

- Each engine has one composition, whatever platforms it ships on.
- NDA content, including which libraries support the console, appears only in that console's private map.
