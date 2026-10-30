# 18: automap upgrade

**What to build:** `automap upgrade [<library>...] [--map-release <release>]` moves a composition to a newer map release, re-resolves, applies changed reference glue where the composition's glue still matches the old reference, reports where it doesn't, and ends with `verify`.

**Blocked by:** [16: Map releases, fetch and cache, and evidence states](16-map-releases-and-evidence-states.md)

**Status:** ready-for-agent

- [ ] On a fixture map with two releases, `upgrade` moves `map_release`, re-resolves and updates glue that matches the old reference; glue that doesn't is left alone and reported as `GLUE_DIVERGED(<file>)` with the new reference beside it.
- [ ] `upgrade <library>` moves only that library and anything its constraints require.
- [ ] The result names failing rules with failure codes, so the agent can adapt glue using the recipe's notes.
- [ ] Golden tests cover a clean upgrade, a diverged-glue upgrade and a failing-probe upgrade.
