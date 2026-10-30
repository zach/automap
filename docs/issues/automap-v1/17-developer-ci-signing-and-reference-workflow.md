# 17: Developer CI signing and the reference workflow

**What to build:** Composition evidence produced in the developer's CI is signed with the CI's identity, and the router skill carries a reference CI workflow that verifies a composition on all three platforms. Slices 1 and 2 are complete against a published map release.

**Blocked by:** [16: Map releases, fetch and cache, and evidence states](16-map-releases-and-evidence-states.md)

**Status:** ready-for-agent

- [ ] `verify` in GitHub Actions with an identity token signs its evidence; run locally, evidence is marked unsigned (ADR-0032).
- [ ] A repo using the reference workflow gets signed composition evidence for Windows, Linux and macOS.
- [ ] The `testbed-start` and `testbed-extend` evals pass against the published map release, and both reference compositions are probed on all three platforms from Automap's CI.
