---
status: accepted
date: 2026-10-29
---

# Evidence for an NDA platform comes only from whoever holds that NDA

## Context

Only NDA holders can legally build or run console code. vcpkg's Xbox triplets aren't tested even in vcpkg's own CI, and Defold runs console builds on self-hosted build servers ([01](../research/01-package-managers.md), [03](../research/03-prior-art.md)). GitHub signs public-repo attestations into a public transparency log, which would expose platform names ([06](../research/06-verification.md)). Each NDA platform's content lives in its own private map ([ADR-0048](0048-one-private-map-per-nda-platform.md)), and developers already produce evidence for their own compositions ([ADR-0032](0032-developer-composition-evidence-lives-in-their-repo.md)).

## Options

1. **Automap verifies consoles** under its own NDAs on everyone's behalf.
2. **Partner-run verification published publicly.**
3. **Evidence only from NDA holders,** signed with their own key and kept in the platform's private map.
4. **Unverified console claims.**

## Decision

Option 3. Evidence for an NDA platform comes only from whoever holds that NDA: the studio's own CI on its devkits, or a porting partner. It is signed with their own key, stored in that platform's private map, and never written to any public transparency log. Automap's public CI never builds for consoles.

## Consequences

- Console evidence needs no new mechanism, only a different key and a private home.
- The public map never carries console evidence.
