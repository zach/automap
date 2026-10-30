# 16: Map releases, fetch and cache, and evidence states

**What to build:** The public map is published as immutable, dated releases, and the CLI resolves against them. The CLI caches what it fetches and works offline afterwards. It fetches published evidence for a lock's pins and reports each piece's state.

**Blocked by:** [15: Automap's CI verifies declared edges and signs evidence into the map](15-automap-ci-signs-evidence-into-the-map.md)

**Status:** ready-for-agent

- [ ] Tagging `map-YYYY.MM.DD` publishes an archive of the map with its SHA-256; locks record the release they were resolved against (ADR-0053).
- [ ] `start` and `extend` resolve against the newest release by default, unless `--map` or `AUTOMAP_MAP` names a directory (ADR-0079); once cached, every command except fetching works with the network off (ADR-0038).
- [ ] `find`, `show` and `verify` report each relevant piece of evidence as current, stale or broken, and warn on stale or broken (ADR-0031).
- [ ] Evidence is fetched for the lock's exact pins from the map's main branch.
