---
status: accepted
date: 2026-10-29
---

# Licenses are facts; the CLI warns before adding copyleft or proprietary libraries

## Context

Indie stacks are overwhelmingly permissive (zlib, MIT), and LGPL is effectively unusable on consoles: SDL's maintainer wrote a zlib replacement for OpenAL Soft (LGPL) to ship on Switch ([05](../research/05-developer-needs-2026.md)). vcpkg already writes an SPDX license record for every installed package and can print a license report ([01](../research/01-package-managers.md)). Full license and provenance reporting is the audit job, which comes later ([ADR-0002](0002-core-job-is-extend.md)).

## Options

1. **Nothing until the audit job.**
2. **License as a fact with a warning at extend time,** carrying vcpkg's license data into the lock.
3. **Full audit reporting in v1.**

## Decision

Option 2. The map records each library's license as a fact. The CLI warns before adding an LGPL, GPL or proprietary library to a composition, saying why (for example, "can't ship on consoles without relinking"). It carries vcpkg's license data into the lock, so a report is always possible. Full audit reporting waits for the audit job.

## Consequences

- A copyleft library is caught at extend time, not at a console submission.
- License data costs little to build because it reuses vcpkg's own.
