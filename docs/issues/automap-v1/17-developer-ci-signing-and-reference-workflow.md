# 17: Developer CI signing and the reference workflow

**What to build:** Composition evidence produced in the developer's CI is signed with the CI's identity, and the router skill carries a reference CI workflow that verifies a composition on all three platforms. Slices 1 and 2 are complete against a published map release.

**Blocked by:** [16: Map releases, fetch and cache, and evidence states](16-map-releases-and-evidence-states.md)

**Status:** done

- [x] `verify` in GitHub Actions with an identity token signs its evidence; run locally, evidence is marked unsigned (ADR-0032).
- [x] A repo using the reference workflow gets signed composition evidence for Windows, Linux and macOS.
- [x] The `testbed-start` and `testbed-extend` evals pass against the published map release, and both reference compositions are probed on all three platforms from Automap's CI.

## Comments

- `verify` decides its verifier from the environment (`verify.Verifier`): in GitHub Actions it is `developer-ci` with the repo's workflow ref as its id, signed when the job has an identity token (`ACTIONS_ID_TOKEN_REQUEST_*`, from `id-token: write`) and the map is a release or a git commit; anywhere else it is `developer-local` and unsigned. The workflow does the signing, with GitHub artifact attestations over the evidence file, as Automap's own CI does, so the CLI carries no Sigstore client.
- `.github/workflows/automap-verify.yml` is the reference workflow: a reusable workflow that installs the toolchain, vcpkg and the CLI on Windows, Linux and macOS runners, verifies, attests each platform's evidence and commits it with its `.sigstore.json` bundle beside the lock. The router skill carries the caller a developer adds. Automap's CI calls it on `ci/reference-repo/`, a testbed `start` wrote against `map-2026.10.31`: on the ticket's branch it reached probed with signed evidence on all three platforms.
- Evals read the newest published map release unless given a map directory. The Evals run 41154929174 (2026-10-31) passed `testbed-start` and `testbed-extend` against `map-2026.10.31`, both probed. Both reference compositions are probed on all three platforms in the map's published evidence (ticket 15).
- The first run on main hit GitHub's unauthenticated API limit on a shared macOS runner, so a release the lock records now downloads straight from its release files, and the CLI sends `GITHUB_TOKEN` to GitHub's API when it is set; the reference workflow and evals provide it.
- The CLI's tests never reach the network or a CI identity: the harness points release and evidence sources at nowhere and strips `GITHUB_` and `ACTIONS_` variables unless a test sets them.
