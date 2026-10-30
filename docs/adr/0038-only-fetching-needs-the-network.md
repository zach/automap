---
status: accepted
date: 2026-10-29
---

# Only fetching needs the network

## Context

Some functions may rely on a remote hosted registry or package repository ([ADR-0006](0006-product-shape-map-cli-router-skill.md)). The research lists reproducible offline builds and vendoring-friendly output as must-haves for the persona ([05](../research/05-developer-needs-2026.md)). vcpkg already builds offline once its asset and binary caches are warm, and `vcpkg export` bundles dependencies ([01](../research/01-package-managers.md)). Evidence for the developer's composition is produced by the CLI in their repo ([ADR-0032](0032-developer-composition-evidence-lives-in-their-repo.md)).

## Options

1. **Network for everything:** composing, building and verifying contact Automap's hosted services.
2. **Network for fetching only:** everything else works offline once fetched data is cached.
3. **Fully offline:** the map and all sources ship with the CLI.

## Decision

Option 2. Fetching is the only thing that needs the network: updating the map, and downloading library sources, Automap registry ports and published evidence. Composing, building and verifying the developer's composition all work offline once those are cached. The CLI can also vendor everything a composition needs into the developer's repo, for fully offline, reproducible builds.

## Consequences

- The map and the Automap registry can be hosted remotely without any build or verification depending on them being reachable.
- A vendored composition builds and verifies with no network at all.
