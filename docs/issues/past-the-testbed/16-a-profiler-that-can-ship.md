# 16: A profiler that can ship, and libraries for development builds only

**What to build:** Tracy gets a variant that can ship: on demand, so it records only while a profiler is connected. A composition can hold a library in development builds only. Field notes [43].

**Blocked by:** None

**Status:** ready-for-agent

- [ ] Tracy's `on-demand` variant builds the client with `TRACY_ON_DEMAND`; its description says it records only while a profiler is connected, and what it still does (listens for one).
- [ ] The diagnostics probe for Tracy passes for both variants: it checks what each variant promises.
- [ ] Verify map on the branch is green for Tracy's edges, both variants, on all three platforms.
- [ ] The product owner decides whether a composition can hold a library in development builds only (an ADR: it changes what a composition is); if yes, a follow-up ticket builds it.

## Comments

- A variant of a library the map already covers is curation, not coverage ([ADR-0051](../../adr/0051-agents-curate-the-public-map.md)). One set of slots and choices per composition is the model ([ADR-0017](../../adr/0017-composition-is-a-manifest-and-a-lock.md)), so choices that differ by build type are the product owner's decision.
