# Handoff: past the testbed

Date: 2026-11-10. The product owner built a complete game, Spintacular (`github.com/zach/spintacular`), on Automap 0.1.0 and logged 50 findings. This session checked each against the code, cut a feature from them, and built it. The evaluation is [`docs/research/08-spintacular-field-report.md`](../research/08-spintacular-field-report.md); the feature spec and tickets are in [`docs/issues/past-the-testbed/`](../issues/past-the-testbed/).

## Where things stand

- **Tickets 01 to 17 are done;** 18 (the eval run and the 0.2.0 release) waits for the product owner. The work is one stack of branches, each on the one before, with a pull request each (#68 to #82) and #84 at the top, which contains them all.
- **Bugs fixed:** two libraries on one vcpkg port (01); an in-repo library's pin on a Windows checkout (02); `upgrade` not carrying a private map's changed glue, and a lock problem reported once per platform (03); a private map's commit pin failing in CI's shallow clone (14, found on the way).
- **The CLI:** `--version`, `help`, a usage written from the command table (04); a failed build's first error in the result, and `FETCH_FAILED` with the URL when vcpkg can't download (05); `find` over slots and descriptions with short entries (06); `.gitignore`, `eject` removing the agent files, the executable found anywhere in the build tree (07); `automap remove` and `automap resolve` (13); private maps pinned by content, composable uncommitted (14); `SYSTEM_PACKAGE_MISSING` before building (12).
- **The router skill** routes to two guides the CLI writes beside it, `game.md` and `private-maps.md` (08).
- **The map:** SDL3's window settings in the context and the Dear ImGui frame in pieces (09); Dear ImGui from an Automap registry port, so SDL3 builds without IBus, D-Bus and libsystemd (11); the platform contract 1.1.0 with `system_packages` (12); SDL_shadercross without DXC and `automap_compile_shaders()` (15); Tracy's `on-demand` variant (16); `text` with stb_truetype and `data-files` with nlohmann/json (17, [ADR-0084](../adr/0084-the-map-covers-text-and-data-files.md)).
- **CI:** the reference workflow follows the composition (10); a public-map test compares every library's vcpkg install plan in each reference composition with its plan alone (11); a test applies every seeded violation in the public map to its reference glue without building (09).
- **Decisions:** [ADR-0084](../adr/0084-the-map-covers-text-and-data-files.md) (coverage) and [ADR-0085](../adr/0085-a-composition-has-one-set-of-slots-for-every-build-type.md) (one set of slots for every build type).

## Things to know

- **The fixture build tests run in a cloud session:** `VCPKG_ROOT=$HOME/.cache/automap/vcpkg/<release> go test ./...` builds the stub libraries, which need no downloads (about six minutes). The public map's libraries still build only in CI.
- **`vcpkg install --dry-run` works in a cloud session too:** it needs the registries and no source downloads. `AUTOMAP_TEST_PUBLIC_MAP=1 VCPKG_ROOT=… go test ./cmd/automap -run TestALibrarysPortBuildsTheSameFeaturesInEveryReferenceComposition` runs the install-plan check.
- **A session can't push to the `registry` branch.** A changed or new port lands on `main` first, on its own, with the map still pinning what it had; `Publish registry` publishes it and writes the registry commit into `map/map.json`; then the map change that uses it merges `main` in. #83 did this for the SDL_shadercross and Dear ImGui ports.
- **vcpkg builds a port with the union of the features every port in the plan asks of it,** and a top-level `default-features: false` doesn't override another port's request. A variant can only promise a feature is off when no port in the map asks for it; the install-plan check is what catches a new one.
- **Any map change makes every edge's evidence stale,** so Verify map on `main` re-verifies about 90 jobs after the stack merges, and commits fresh evidence. A map release is cut after that run is green.
- **Syntax-checking glue locally:** `git clone` works through the proxy, so a library's headers can be fetched and the glue checked with `clang++ -fsyntax-only` against them before CI builds it.
- **Runner queues are slow on this account.** Stacked branches each start a CI run on push; cancel the runs below the top of the stack and let the top one stand for them.
- The reference workflow's two calls in `reference-workflow.yml` upload artifacts with the same names; a developer's repo calls it once.
- `ci/reference-repo` is a composition at `map-2026.10.31`, written by a development CLI, so the reference workflow's fallback list of system packages is what it exercises; move it to the next map release once that is cut.

## Next

1. When #84's CI is green, merge it (a merge commit); the stacked pull requests below it close as merged. Watch Verify map and CI on `main`, and fix forward.
2. When Verify map on `main` is green, ask the product owner for the map release (`Release map` with `map-YYYY.MM.DD`), the `testbed-to-game` eval run (`Evals` with `tasks=testbed-to-game`) and `cli-v0.2.0` (`Release CLI` with `0.2.0`); then tick ticket 18.
3. Move `ci/reference-repo` to the new release with `automap upgrade`, and update the README's release line.
