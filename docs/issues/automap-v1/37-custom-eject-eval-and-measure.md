# 37: The custom-eject eval and the eject measure

**What to build:** Journey 3's eject step as an eval task, and the eject measure computed across every journey composition.

**Blocked by:** [34: Custom-integration evals](34-custom-integration-evals.md), [35: automap eject](35-automap-eject.md)

**Status:** done

- [x] `custom-eject`: the session ejects Acme's engine, which then builds and passes its smoke test without Automap.
- [x] The eval report computes the eject measure: the share of ejected journey compositions that build and pass smoke without Automap (target 100%, ADR-0063).

## Comments

`custom-eject` passed against map release `map-2026.10.31.1` in the Evals workflow's run [41238211570](https://github.com/zach/automap/actions/runs/41238211570): the session ejected Acme's engine in 141 s, and the ejected repo built with plain CMake and vcpkg and passed its 120-frame smoke run without the CLI. The report's eject measure was 1 of 1 (100%).
