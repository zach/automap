---
status: accepted
date: 2026-10-30
---

# Automap is built in one repository; the registry is published to its own branch

## Context

Automap has a CLI ([ADR-0064](0064-cli-is-written-in-go.md)), a public map with dated releases ([ADR-0053](0053-map-releases-are-immutable-and-recorded-in-locks.md)), an eval suite ([ADR-0062](0062-no-telemetry-eval-suite-measures-success.md)), docs, and the Automap registry of vcpkg overlay ports ([ADR-0034](0034-vcpkg-is-the-v1-build-backend.md)). vcpkg's docs (checked 2026-10-30) require a registry to have `ports/` and `versions/` at its root; a git registry entry may name a branch through its `reference` field, and vcpkg's own docs repo serves its sample registry from a `vcpkg-registry` branch. Some changes cross the CLI, the map and the evals together, such as a new contract version.

## Options

1. **One repository,** with the registry published to its own branch.
2. **Separate repositories** for the CLI, the map and the registry.

## Decision

Option 1. Automap is built in this one repository. The CLI, the public map, the eval suite and the docs live on the main branch. The Automap registry is published to its own branch. CLI releases and map releases get separate tags, such as `cli-v0.1.0` and `map-2026.11.01`. CLI versions follow semver, and the first release is 0.1.0.

## Consequences

- Building agents work in one context, and changes that cross the CLI, map and evals land together.
- The registry branch is generated from the main branch, so there is no second repository to keep in sync.
