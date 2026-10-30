---
status: accepted
date: 2026-10-30
---

# The CLI is written in Go and ships as a single static binary

## Context

The CLI must run offline on Windows, macOS and Linux ([ADR-0038](0038-only-fetching-needs-the-network.md), [ADR-0056](0056-v1-platforms-desktop-only.md)); write the router skill into developers' repos ([ADR-0044](0044-router-skill-ships-inside-the-cli.md)); validate JSON Schemas and canonicalize locks ([ADR-0042](0042-manifest-and-lock-are-json-with-schemas.md)); sign and verify in-toto attestations through Sigstore ([ADR-0030](0030-evidence-is-signed-and-stored-in-the-map.md)); and drive vcpkg and CMake. Agents write and maintain it. Library tags checked on 2026-10-30: Sigstore's Go library `sigstore-go` v1.4.0; in-toto's `in-toto-golang` v0.11.0; Sigstore's Rust library `sigstore-rs` v0.14.0 (pre-1.0); `sigstore-python` v4.5.1 (needs a Python runtime).

## Options

1. **Go.**
2. **Rust.**
3. **Python.**
4. **C++.**

## Decision

Option 1. The CLI is written in Go and ships as a single static binary for each v1 platform.

## Consequences

- No runtime is needed on developers' machines, and cross-compiling for the three platforms is trivial.
- Sigstore signing and in-toto attestations use maintained Go implementations.
