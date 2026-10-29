---
status: accepted
date: 2026-10-29
---

# New releases enter the map after a 7-day cooldown and known-good evidence

## Context

The brief names release-age cooldowns as an example of supply-chain hygiene. No C++ package manager has one ([01](../research/01-package-managers.md)); elsewhere they have become default practice (pnpm 11 holds back releases for a day by default; Dependabot waits 3 days). Dependabot's cooldown for vcpkg measures the age of the registry commit, not of the upstream release. So a cooldown has to be Automap's own curation rule. The persona wants no forced upgrades ([05](../research/05-developer-needs-2026.md)), so a short delay costs them little.

## Options

1. **No cooldown:** a release enters the map as soon as it has evidence.
2. **Cooldown only:** a release enters the map after a waiting period.
3. **Cooldown and evidence:** a release enters the map only once it is old enough and has known-good evidence.

## Decision

Option 3. A new upstream release enters the map only once it is at least **7 days old** and has **known-good evidence**. Releases flagged as security fixes may skip the waiting period, never the evidence.

## Consequences

- "Latest" in the map always means "latest known-good", so an agent upgrading a composition never lands on an untested release by default.
- Compromised or hastily pulled releases are kept out for at least a week.
