# 16: A profiler that can ship, and libraries for development builds only

**What to build:** Tracy gets a variant that can ship: on demand, so it records only while a profiler is connected. A composition can hold a library in development builds only. Field notes [43].

**Blocked by:** None

**Status:** ready-for-agent

- [x] Tracy's `on-demand` variant builds the client with `TRACY_ON_DEMAND`; its description says it records only while a profiler is connected, and what it still does (listens for one).
- [x] The diagnostics probe for Tracy fits both variants: a facade zone opens a Tracy zone in both, and the client's queue gains the zone's events in the default variant and nothing in the on-demand one. The ABI seed defines `TRACY_FIBERS`, which neither variant is built with.
- [ ] Verify map on the branch is green for Tracy's edges, both variants, on all three platforms.
- [x] The product owner decides whether a composition can hold a library in development builds only: it can't; a composition has one set of slots for every build type ([ADR-0085](../../adr/0085-a-composition-has-one-set-of-slots-for-every-build-type.md)).

## Comments

- A variant of a library the map already covers is curation, not coverage ([ADR-0051](../../adr/0051-agents-curate-the-public-map.md)). One set of slots and choices per composition is the model ([ADR-0017](../../adr/0017-composition-is-a-manifest-and-a-lock.md)), so choices that differ by build type are the product owner's decision.
- `TRACY_ONLY_LOCALHOST` is a compile-time option of the client that the curated port doesn't expose, so the on-demand variant still listens on the network; its description says so. Listening on localhost only would take a port of Tracy in the Automap registry.
- `game.md` says to compose the on-demand variant for a build that ships, or to remove the profiler before the release build.
