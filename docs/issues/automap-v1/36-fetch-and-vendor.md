# 36: fetch and vendor for offline work

**What to build:** `automap fetch` fills every cache a composition needs; `automap vendor` copies sources, ports and evidence into the repo for fully offline, reproducible builds (ADR-0038).

**Blocked by:** [16: Map releases, fetch and cache, and evidence states](16-map-releases-and-evidence-states.md)

**Status:** done

- [x] After `fetch`, `verify` succeeds with the network disabled.
- [x] After `vendor`, a fresh clone builds and verifies with the network disabled and no caches.
- [x] With `VCPKG_ROOT` unset, the CLI bootstraps the vcpkg tool release the public map's `map.json` names into its cache, and `fetch` fills it (SPEC §5.5).
