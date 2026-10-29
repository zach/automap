---
status: accepted
date: 2026-10-29
---

# Anyone can propose map changes; evidence always comes from Automap's CI

## Context

Agents curate the public map under evidence gates, and humans decide coverage ([ADR-0051](0051-agents-curate-the-public-map.md)). Whether community-submitted verification runs could ever enter the map was left open ([ADR-0032](0032-developer-composition-evidence-lives-in-their-repo.md)). Accepting someone else's runs safely would require trusted, isolated build workflows (SLSA Build Level 3) ([06](../research/06-verification.md)). Re-running a contribution on Automap's own CI is cheap: a full sweep costs about $85 at private-runner prices. Some upstream maintainers already publish agent-facing material for their libraries, such as DirectXTex's usage skills ([02b](../research/02b-catalog-assets-and-data.md)).

## Options

1. **Closed curation:** only Automap's curation agents and maintainers change the map.
2. **Open contributions, Automap's evidence only:** anyone proposes changes; Automap's CI produces all evidence.
3. **Open contributions with contributor-supplied evidence** from trusted workflows.

## Decision

Option 2. Anyone (people, agents or upstream maintainers) can propose changes to the public map as pull requests: facts, recipes, probes or new library versions. Contributions pass the same evidence gates as curation agents. Evidence is always produced by Automap's CI, never accepted from a contributor's own runs. Adding new slots or libraries still needs the human coverage decision.

## Consequences

- "Known-good" means exactly one thing: Automap's CI ran the probes.
- There is no contributor trust tier to manage.
