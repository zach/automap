---
status: accepted
date: 2026-10-30
---

# The exploration merges to main when the definition of done is signed off

## Context

The exploration has run on the `exploration` branch; `main` holds only the license. Automap is built in one repository with everything on the main branch ([ADR-0065](0065-one-repository-registry-on-its-own-branch.md)), and the instruction files switch to `AGENTS.md` when implementation starts ([ADR-0066](0066-build-repo-uses-agents-md.md)). A fresh agent given the repo couldn't tell whether implementation had been authorised or on which branch. Git history is the record of what changed in this repo's documents.

## Options

1. **Merge now,** before the definition-of-done review ends.
2. **Merge when the definition of done is signed off,** keeping the exploration's commits.
3. **Squash on merge.**

## Decision

Option 2. The `exploration` branch merges to `main`, with a merge commit, once the product owner signs off the definition-of-done review. Merging is the exploration's last step and the precondition for ticket 01. Ticket 01 starts on `main` and switches the instruction files to `AGENTS.md`, which ends the exploration's "no product code" rule. The `spike/testbed-agent` branch stays unmerged as the spike's archive.

## Consequences

- `main` carries the full history of how each decision was reached.
- Implementation never starts from an unreviewed spec.
