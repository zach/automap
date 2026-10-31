# Handoff: build, ticket 15

Date: 2026-10-31. Work happens on one branch per ticket from `main`; each merges through a pull request with a merge commit once CI is green, with the ticket's status set to `done` in the same pull request.

## Where things stand

- **Done and merged:** tickets 01 to 15. Ticket 16 (map releases, fetch and cache, evidence states) is next.
- **Automap's CI publishes evidence.** The `Verify map` workflow (`.github/workflows/verify-map.yml`, driven by `ci/cmd/verifymap`) runs on every push to main, daily and on demand. It verifies every declared edge that needs it on Linux, Windows and macOS, runs the seeds, attests each evidence file with GitHub artifact attestations and commits evidence and `.sigstore.json` bundles into `map/evidence/`. Its first run on main published probed evidence for all 11 edges on all three platforms.
- **The public map runs on all three platforms** in CI: SDL3 offscreen on Linux (lavapipe) and with its native video driver on hosted Windows (WARP) and macOS (paravirtual Metal), through per-platform `headless` facts.
- **Evals** run on demand only, after asking the product owner: each spends the `ANTHROPIC_KEY` secret. `testbed-start` and `testbed-extend` pass.

## Things to know

- Pushes from the `Verify map` workflow use `GITHUB_TOKEN`, so they trigger no further workflows. Branch runs never publish; to try the workflow on a branch, add the branch to its push trigger and remove it before merging.
- On Windows, vcpkg is cloned in full: the git vcpkg runs there can't fetch a blobless clone's objects lazily. Elsewhere it is cloned without blobs, never shallow, since pinned port versions need its history.
- `gh attestation verify <file> --bundle <file>.sigstore.json --repo zach/automap` verifies a piece of evidence; this container can't reach Sigstore's trust root, so it runs only elsewhere.
- Long job logs exceed the GitHub tool's output limit; they are saved to a file, and `scratchpad/loggrep.py <file> <regex>` greps them.

## Next

Ticket 16, then 17 onward through the frontier of `docs/issues/automap-v1/`.
