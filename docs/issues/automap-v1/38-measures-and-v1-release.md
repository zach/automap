# 38: The four measures and the v1 release checklist

**What to build:** The eval suite and the map compute the four v1 measures on every map release, and the v1 release, CLI 0.1.0, goes to the product owner as a human checkpoint (ADR-0063, ADR-0065, ADR-0069).

**Blocked by:** [30: Starter-engine reference composition and journey 2 evals](30-starter-engine-composition-and-evals.md), [34: Custom-integration evals](34-custom-integration-evals.md), [37: The custom-eject eval and the eject measure](37-custom-eject-eval-and-measure.md)

**Status:** ready-for-agent

- [ ] The eval report shows: request-to-running-testbed time (warm cache, with cold reported alongside); first-attempt success across all eval tasks; freshness (share of catalog libraries at their latest release within 14 days, share of evidence current); and eject.
- [ ] All four measures meet their targets on one map release.
- [ ] The release checklist is presented to the product owner for the v1 decision, releasing the CLI as 0.1.0 (`cli-v0.1.0`).
