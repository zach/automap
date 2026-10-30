# 15: Automap's CI verifies declared edges and signs evidence into the map

**What to build:** Automap's CI derives its matrix from the map: every binding, every pairing with a recipe and each reference composition, for each library's latest and previous version and each variant, on each v1 platform. It runs every probe's seeded violation, signs each evidence statement with GitHub artifact attestations, and commits evidence into the map. It re-verifies on any input change and often enough that no declared edge's evidence passes 30 days (ADR-0030, ADR-0033).

**Blocked by:** [12: Jolt in the map](12-jolt-in-the-map.md), [14: Windows and macOS verification](14-windows-and-macos-verification.md)

**Status:** done

- [x] The testbed and testbed + Jolt have signed evidence for all three platforms in the map, verifiable with GitHub's attestation tooling.
- [x] Changing one recipe re-verifies exactly the edges that depend on it.
- [x] A probe that passes its seeded violation fails the job, and its rule doesn't count toward the published level until fixed (ADR-0071).
- [x] A scheduled run keeps every declared edge's evidence under 30 days old.

## Comments

- `ci/cmd/verifymap plan` derives the edges from the map: each binding and pairing recipe at the latest and previous versions it applies to and each variant, and each reference composition in `ci/compositions.json` (testbed, testbed + Jolt). An edge needs verifying on a platform its libraries target when it has no evidence there, its evidence is older than 23 days (a week before it goes stale), or one of its inputs changed since the commit its evidence's `map_release` names: its libraries' entries and library recipes, its own recipe, the contracts, harness and testbed, and the CLI. Edges with the same libraries share one job per platform.
- `verifymap run` starts the composition with `automap start`, runs every seed that applies (a library rule's in every job, a composition rule's in the reference compositions, where the libraries its mistake depends on are), and writes each edge's evidence at `evidence/<kind>/<id>/<pins>/<platform>.json`. An edge's level comes from its own rules; a rule whose probe missed its seed is recorded not-applicable as guidance and the job fails (ADR-0071).
- The `Verify map` workflow runs on every push to main, daily, and on demand. It signs every evidence file of a job with `actions/attest-build-provenance` and stores the Sigstore bundle beside each file as `<platform>.sigstore.json`, so `gh attestation verify <file> --repo zach/automap` checks it. A final job checks that the map still loads and commits the evidence to main.
- Getting there on Windows and macOS: SDL3's x11, wayland and vulkan features apply on Linux only (a vcpkg platform expression per feature in `vcpkg_features`); `headless` facts can differ per platform, since SDL's offscreen driver can't host a D3D12 or Metal swapchain while hosted Windows and macOS runners have a desktop session; vcpkg is cloned in full on Windows, whose vcpkg-run git couldn't fetch a blobless clone's objects. The public-map CI job runs on all three runners.
- On this branch's run of the workflow every one of the 15 jobs (5 compositions × 3 platforms) reached probed with every applicable seed caught, and its evidence was attested. Publishing happens on main.
