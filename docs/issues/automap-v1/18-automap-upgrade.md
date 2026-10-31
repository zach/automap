# 18: automap upgrade

**What to build:** `automap upgrade [<library>...] [--map-release <release>]` moves a composition to a newer map release, re-resolves, applies changed reference glue where the composition's glue still matches the old reference, reports where it doesn't, and ends with `verify`.

**Blocked by:** [16: Map releases, fetch and cache, and evidence states](16-map-releases-and-evidence-states.md)

**Status:** done

- [x] On a fixture map with two releases, `upgrade` moves `map_release`, re-resolves and updates glue that matches the old reference; glue that doesn't is left alone and reported as `GLUE_DIVERGED(<file>)` with the new reference beside it.
- [x] `upgrade <library>` moves only that library and anything its constraints require.
- [x] The result names failing rules with failure codes, so the agent can adapt glue using the recipe's notes.
- [x] Golden tests cover a clean upgrade, a diverged-glue upgrade and a failing-probe upgrade.

## Comments

- `upgrade` reads the lock's release as the old map and the newest release (or `--map-release`, or a `--map` directory) as the new one. Libraries named move to their newest version there; the others are held at their locked versions (`mapmodel.Map.Hold`), and a library whose locked version isn't in the new release moves as its constraints require. `composition.Upgrade` records the new release and adds any new contract's choice defaults and any new recipe's glue entries.
- It computes the composition's files under both maps. Files the CLI generates (manifest, lock, vcpkg files, `automap.cmake`, the choices header and the startup unless developer-owned) are rewritten; every file copied from the map is updated when missing or still the old release's version, left alone when the developer changed it and the reference didn't, and otherwise left alone with the new reference written beside it as `<file>.reference` and a `GLUE_DIVERGED(<file>)` warning. It ends with `verify`, whose failures carry their rules' failure codes.
- Every command with a lock now reads each library at the version the lock pins, so facts and recipes match the pins after a partial upgrade.
- Golden tests cover a clean upgrade (`upgrade-clean`), diverged glue (`upgrade-diverged`) and a failing probe (`upgrade-failing-probe`, compared by its result alone, since its lock names a registry the test builds per machine). Golden comparisons leave out `build/` and `evidence/`, which change every run.
