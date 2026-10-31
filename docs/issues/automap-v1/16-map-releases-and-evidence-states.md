# 16: Map releases, fetch and cache, and evidence states

**What to build:** The public map is published as immutable, dated releases, and the CLI resolves against them. The CLI caches what it fetches and works offline afterwards. It fetches published evidence for a lock's pins and reports each piece's state.

**Blocked by:** [15: Automap's CI verifies declared edges and signs evidence into the map](15-automap-ci-signs-evidence-into-the-map.md)

**Status:** done

- [x] Tagging `map-YYYY.MM.DD` publishes an archive of the map with its SHA-256; locks record the release they were resolved against (ADR-0053).
- [x] `start` and `extend` resolve against the newest release by default, unless `--map` or `AUTOMAP_MAP` names a directory (ADR-0079); once cached, every command except fetching works with the network off (ADR-0038).
- [x] `find`, `show` and `verify` report each relevant piece of evidence as current, stale or broken, and warn on stale or broken (ADR-0031).
- [x] Evidence is fetched for the lock's exact pins from the map's main branch.

## Comments

- The `Release map` workflow publishes a tag `map-YYYY.MM.DD`: it checks the map loads, archives `map/` reproducibly (sorted, fixed times and owners, `gzip -n`) and attaches the archive and its `.sha256` to a GitHub release. Run on demand with a release name, it creates the tag itself; the first release, `map-2026.10.31`, was published that way from main after this ticket merged.
- `internal/release` lists releases through GitHub's releases API, downloads and checks one against its SHA-256 (`MAP_RELEASE_INVALID(<release>)` otherwise) and unpacks it into the cache. Without `--map` or `AUTOMAP_MAP`, `start` reads the newest release (the newest cached one when offline) and every other command the release the lock records; `extend` keeps the lock's release, since moving to a newer one is `upgrade`'s job (ADR-0053). A lock resolved against a directory names no release, so commands on it still need the directory.
- Published evidence is read for the composition's bindings and pairings at the lock's exact pins (and, for `find` and `show`, each library's at its newest version and default variant) from the map on main, cached, and read from the cache offline; with a map directory it is read from the directory. It is broken when its latest run failed (a rule failed, or the run didn't reach smoke), stale after 30 days, and current otherwise; missing evidence isn't listed. `EVIDENCE_STALE` and `EVIDENCE_BROKEN` warnings name the subject and platform.
- `AUTOMAP_RELEASES`, `AUTOMAP_EVIDENCE_URL` and `AUTOMAP_CACHE` point the CLI elsewhere; the process-level tests serve releases and evidence from a local HTTP server.
