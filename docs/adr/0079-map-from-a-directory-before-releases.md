---
status: accepted
date: 2026-10-30
---

# The CLI can read the map from a directory, recorded as a commit or as unreleased

## Context

Every lock records the map release it was resolved against ([ADR-0053](0053-map-releases-are-immutable-and-recorded-in-locks.md)), and the manifest, lock and evidence formats require one. The first map release is published only in slice 3 ([ADR-0061](0061-first-three-vertical-slices.md)), but slices 1 and 2, the CLI's tests against fixture maps ([ADR-0077](0077-tests-attach-at-the-cli-process.md)), and Automap's CI verifying the map's main branch before a release is cut all need the CLI to read a map that isn't a release. A fresh agent given the repo found no answer to this.

## Options

1. **Releases only:** publish a release before anything else can run.
2. **A map directory, recorded honestly:** the CLI can read the map from a directory, and records what it read.

## Decision

Option 2. The CLI reads the public map from a directory when given `--map <dir>` or the `AUTOMAP_MAP` environment variable, instead of from a release.

- From a clean git checkout, `map_release` records `git:<commit>`, which can be fetched again. Automap's CI uses this to verify the map's main branch before a release is cut.
- From any other directory, `map_release` records `unreleased`. The CLI never signs evidence produced against an unreleased map and never reports it as known-good.

Map releases stay the default once they exist.

## Consequences

- Slices 1 and 2, the CLI's tests and early evals run before the first release.
- Every lock still points at something that can be fetched again, unless it says `unreleased`.
