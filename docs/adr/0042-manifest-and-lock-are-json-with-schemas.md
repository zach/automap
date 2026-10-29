---
status: accepted
date: 2026-10-29
---

# The manifest and lock are JSON with published schemas; the lock is canonical

## Context

The agent writes the manifest, the CLI writes the lock, and evidence names the lock's exact pins ([ADR-0017](0017-composition-is-a-manifest-and-a-lock.md), [ADR-0030](0030-evidence-is-signed-and-stored-in-the-map.md)). vcpkg's manifest is JSON with a published JSON Schema, which the research rated the easiest manifest for an agent to write and check ([01](../research/01-package-managers.md)). The agent is the main reader of these files ([ADR-0005](0005-agent-is-primary-steward.md)).

## Options

1. **JSON with JSON Schema.**
2. **TOML,** with comments.
3. **YAML.**

## Decision

Option 1. The manifest and the lock are both JSON, each validated against a published JSON Schema. Notes go in `$comment` fields, as `vcpkg.json` allows. The lock is written in canonical form using the JSON Canonicalization Scheme (RFC 8785), so the same composition always hashes the same.

## Consequences

- One format across Automap's files and vcpkg's avoids translation mistakes.
- Evidence has a stable hash to name for a composition.
