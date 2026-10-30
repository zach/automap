# 37: The custom-eject eval and the eject measure

**What to build:** Journey 3's eject step as an eval task, and the eject measure computed across every journey composition.

**Blocked by:** [34: Custom-integration evals](34-custom-integration-evals.md), [35: automap eject](35-automap-eject.md)

**Status:** ready-for-agent

- [ ] `custom-eject`: the session ejects Acme's engine, which then builds and passes its smoke test without Automap.
- [ ] The eval report computes the eject measure: the share of ejected journey compositions that build and pass smoke without Automap (target 100%, ADR-0063).
