# 19: Watcher, curation agents and the cooldown gate

**What to build:** A watcher notices upstream releases, tags and commits for every catalog library, and curation agents open map changes with updated facts, pins, content hashes, recipes and probes. Automap's CI and the cooldown decide what merges (ADR-0037, ADR-0051).

**Blocked by:** [15: Automap's CI verifies declared edges and signs evidence into the map](15-automap-ci-signs-evidence-into-the-map.md)

**Status:** ready-for-agent

- [ ] A new upstream release of a catalog library produces a map change within a day of release.
- [ ] Facts are extracted from headers where possible, and every change passes schema and cross-file validation.
- [ ] A change merges only when the release is at least 7 days old and its evidence is known-good on all three platforms; a security release skips the wait, never the evidence.
- [ ] Changes that add slots or libraries, or a contract major version, wait for a human (ADR-0069).
