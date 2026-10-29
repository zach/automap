---
status: accepted
date: 2026-10-29
---

# The asset pipeline is a seam, not a slot

## Context

The research split data and assets three ways ([02b](../research/02b-catalog-assets-and-data.md)):

- **Serialization, JSON and reflection** have interchangeable, maintained libraries with clear seams, so they are ordinary slots under the existing model ([ADR-0012](0012-slots-have-no-automap-api.md)).
- **The asset pipeline** has no reusable library to fill a slot: the one standalone cooker found (AssetCooker) is Windows-only and at v0.0.2, and every production pipeline is embedded in its engine. Yet Godot, O3DE, Unity, Bevy and ezEngine converge on the same contracts: stable IDs, importers whose output depends only on their input and version, declared file dependencies, content-hash cache keys, and reload.
- **The asset data model and editor** are the most opinionated part of any engine and stay the developer's own code.

cgltf, fastgltf, ufbx, libktx and PhysicsFS already expose injectable file-I/O hooks. Service seams with many hook styles get a shared interface ([ADR-0025](0025-shared-interfaces-for-service-seams.md)).

## Options

1. **An asset-pipeline slot** filled by a swappable pipeline library.
2. **An assets seam** with a contract every asset library is bound to.
3. **Recipes and guidance only,** with no contract.

## Decision

Option 2. The asset pipeline is an eleventh seam, **assets**, not a slot. Its contract covers:

- whether each library is used at build time or at runtime;
- output determined by input, library version and settings;
- every file a library reads reported through an injectable file-I/O hook;
- a reload hook.

It has a shared file-I/O interface, like the allocator facade.

## Consequences

- Serialization, JSON and reflection libraries fill ordinary slots.
- The asset data model and editor stay the developer's own code.
- Like every contract, assets has at least one probed rule in v1 ([ADR-0027](0027-v1-probe-floor-and-classification.md)): output identical across runs with the same input, version and settings, and every file read going through the file-I/O interface.
