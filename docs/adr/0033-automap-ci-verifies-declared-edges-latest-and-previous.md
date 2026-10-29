---
status: accepted
date: 2026-10-29
---

# Automap's CI verifies what the map declares, for latest and previous versions

## Context

Verifying every combination is infeasible: 25 libraries with two versions each is about 3.4×10⁷ builds per platform. The research's worked estimate for 25 libraries on 4 platforms put a sweep of declared pairings at about 4,400 runner-minutes and an all-pairs sweep at about 16,400 ([06](../research/06-verification.md)). Contracts make most compatibility a matter of bindings, N+M ([ADR-0011](0011-automap-owns-one-contract-per-seam.md)), and composition evidence covers interactions among three or more libraries ([ADR-0029](0029-evidence-subjects-and-platform.md)). Godot's extension tests use the latest and previous stable releases as reference points; vcpkg rebuilds only what an input change touches. Evidence goes stale after 30 days ([ADR-0031](0031-evidence-states-current-stale-broken.md)).

## Options

1. **All pairs** of libraries, versions and platforms.
2. **Declared edges only:** bindings, pairings with recipes, and reference compositions, for a window of versions.
3. **Sampling** of combinations.

## Decision

Option 2. On every supported platform, Automap's CI verifies:

- every binding,
- every pairing that has a recipe,
- a handful of reference compositions,

for each library's latest and previous version. It re-verifies whenever an input changes and at least every 30 days. Older versions retire by no longer being re-verified, so their evidence goes stale. There is no all-pairs testing.

## Consequences

- A pairing is declared by giving it a recipe; undeclared pairs carry no evidence.
- Retiring a version needs no separate deprecation process.
