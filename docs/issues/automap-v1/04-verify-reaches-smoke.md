# 04: Verify reaches smoke

**What to build:** `automap verify` configures and builds the composition with CMake and vcpkg, runs its executable headless for a set number of frames, and writes evidence at the builds or smoke level, unsigned, beside the lock. The fixture's stub libraries are real C++ built through vcpkg, from a git registry the tests create from a directory in the fixture map, so this runs in seconds.

**Blocked by:** [03: Start writes a composition](03-start-writes-a-composition.md)

**Status:** done

- [x] On the fixture composition, `verify` reaches smoke on Linux in under a minute with warm caches.
- [x] A build failure reports `BUILD_FAILED` with the build log as its details file and level none; a crash or unclean shutdown reports `SMOKE_FAILED` and level builds.
- [x] Evidence validates against the evidence schema, names the lock's hash and the lock file's digest, records run conditions, and is marked unsigned (ADR-0032, ADR-0072).
- [x] `start` ends by running `verify` for the host platform and reports its level.

## Comments

- The fixture's stub libraries are `testdata/map/registry/ports/stub-window` and `stub-ui`, C libraries whose source ships inside their ports. The tests create a git registry from that directory, computing its versions database from the ports' git trees, and point the fixture map's `automap_registry` at it.
- Build tests need `VCPKG_ROOT`; they skip without it, except in CI, which bootstraps vcpkg on the runner and sets `AUTOMAP_TEST_BUILDS=1` so they can't skip there.
- `verify` uses the vcpkg at `VCPKG_ROOT`. Bootstrapping the vcpkg tool release `map.json` names when `VCPKG_ROOT` isn't set (SPEC §5.5) is an acceptance criterion of ticket 36, with the rest of the caches.
- Below the probed level, the evidence records each applicable probed rule as `unverifiable` because no probe ran, since an evidence record needs at least one result.
- The composition checks deferred from ticket 02 run whenever a command reads a composition: `MANIFEST_INVALID` and `LOCK_MISMATCH(<library>)`.
