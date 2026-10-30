---
status: accepted
date: 2026-10-29
---

# Each NDA platform's content lives in its own private map

## Context

Private content lives in private maps ([ADR-0047](0047-private-maps.md)). For console platforms, the platform holders set the rules ([05](../research/05-developer-needs-2026.md), [02a](../research/02a-catalog-platform-graphics-ui-audio.md)): the platform holder confirms developer status, then the vendor releases privately. SDL keeps one private repo per console, so a developer with one platform's NDA sees nothing of another's, while SDL's public docs still say which consoles have ports and who to contact. FMOD and Wwise gate console builds the same way.

## Options

1. **One private map for all consoles.**
2. **One private map per NDA platform,** with the public map naming only that a console variant exists and who distributes it.
3. **No mention of consoles in the public map at all.**

## Decision

Option 2. Each NDA platform's libraries, variants, recipes and evidence live in a private map of their own, one per platform. The public map says only that a console variant exists and who distributes it, never anything about the platform's APIs.

## Consequences

- Holding one platform's NDA never exposes another platform's content.
- An agent can tell the developer where to get console support without leaking anything.
