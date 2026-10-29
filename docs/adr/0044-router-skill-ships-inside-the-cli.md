---
status: accepted
date: 2026-10-29
---

# The router skill ships inside the CLI and is written into the developer's repo

## Context

Automap has one router skill that teaches agents to drive the CLI ([ADR-0006](0006-product-shape-map-cli-router-skill.md)). The Agent Skills format has no version field, and guidance that didn't match the installed version made agents' results worse in SWE-Skills-Bench ([04](../research/04-agent-ecosystem.md)). A skill installed separately from the CLI, through a marketplace or `npx skills`, can drift out of step with it. Only fetching needs the network ([ADR-0038](0038-only-fetching-needs-the-network.md)).

## Options

1. **Separate skill distribution:** the router skill is installed from plugin marketplaces or `npx skills`, independently of the CLI.
2. **Bundled with the CLI:** the CLI writes the router skill into the developer's repo at its own version.

## Decision

Option 2. The router skill ships inside the CLI; there is no separate skill install. The CLI writes the skill into the developer's repo, at the CLI's own version, in each place the developer's agents read skills from (for example `.claude/skills/` for Claude Code). Upgrading the CLI rewrites the skill. Plugin marketplaces are at most a way to discover and install the CLI.

## Consequences

- The router skill and the CLI can never disagree about versions.
- Every agent working on the repo, on any machine or in CI, reads the same committed skill.
