---
status: accepted
date: 2026-10-29
---

# Evidence is a signed attestation stored in the map

## Context

Evidence must outlive the CI run that produced it. Since 2026-10-01, GitHub deletes public repositories' workflow runs, logs and artifacts after at most 90 days ([06](../research/06-verification.md)), so a claim that links to a CI run would break within a quarter. in-toto test-result attestations and SLSA verification summaries exist for exactly this kind of claim, and GitHub Actions can sign them (via `actions/attest`) on every plan for public repositories. The map already includes the evidence records behind its claims ([ADR-0006](0006-product-shape-map-cli-router-skill.md)).

## Options

1. **CI links:** evidence is a link to the run that produced it.
2. **Unsigned records in the map.**
3. **Signed attestations in the map.**

## Decision

Option 3. Every piece of evidence Automap publishes is a signed in-toto attestation stored in the map. Each record names:

- the subject's exact pins,
- the platform,
- the evidence level,
- the result of each rule,
- the date,
- the CI identity that signed it.

A CI link may appear in a record but is never the record.

## Consequences

- Evidence is versioned with the claims it backs.
- Anyone can verify who produced a piece of evidence and what it covered, without access to the original CI run.
