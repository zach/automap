---
status: accepted
date: 2026-10-29
---

# Evidence is current, stale or broken

## Context

Evidence names exact pins ([ADR-0030](0030-evidence-is-signed-and-stored-in-the-map.md)), so a new library version doesn't invalidate old evidence; it simply has none yet. What still decays is time: CI runner images roll (at most two generally available images per OS), toolchains update, and hidden inputs such as system SDKs change underneath unchanged pins ([06](../research/06-verification.md)). vcpkg keeps a list of expected build failures that is re-checked on schedule, and an unexpected pass counts as a CI failure, which forces negative claims to retire once they stop being true. The research costed a monthly full re-verification sweep at about $85 at private-runner prices, and nothing on free public runners.

## Options

1. **No expiry:** evidence stays valid until its inputs change.
2. **Expiry with current and stale states only.**
3. **Current, stale and broken states,** with failed runs kept as negative evidence.

## Decision

Option 3. Evidence is always in exactly one of three states:

- **current:** produced within the last 30 days.
- **stale:** older than 30 days; shown with a warning and queued for a re-run.
- **broken:** the latest run failed; kept and re-checked on schedule, so a pairing that gets fixed retires its own negative claim.

Known-good requires current evidence.

## Consequences

- Keeping evidence current requires re-verifying at least every 30 days; the cadence is decided in branch 11.
- Agents can distinguish "known-good", "probably fine, re-verify before relying on it" and "known to fail".
