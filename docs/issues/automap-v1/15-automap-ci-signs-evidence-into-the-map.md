# 15: Automap's CI verifies declared edges and signs evidence into the map

**What to build:** Automap's CI derives its matrix from the map: every binding, every pairing with a recipe and each reference composition, for each library's latest and previous version and each variant, on each v1 platform. It runs every probe's seeded violation, signs each evidence statement with GitHub artifact attestations, and commits evidence into the map. It re-verifies on any input change and often enough that no declared edge's evidence passes 30 days (ADR-0030, ADR-0033).

**Blocked by:** [12: Jolt in the map](12-jolt-in-the-map.md), [14: Windows and macOS verification](14-windows-and-macos-verification.md)

**Status:** ready-for-agent

- [ ] The testbed and testbed + Jolt have signed evidence for all three platforms in the map, verifiable with GitHub's attestation tooling.
- [ ] Changing one recipe re-verifies exactly the edges that depend on it.
- [ ] A probe that passes its seeded violation fails the job, and its rule doesn't count toward the published level until fixed (ADR-0071).
- [ ] A scheduled run keeps every declared edge's evidence under 30 days old.
