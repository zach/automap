# 04: Verify reaches smoke

**What to build:** `automap verify` configures and builds the composition with CMake and vcpkg, runs its executable headless for a set number of frames, and writes evidence at the builds or smoke level, unsigned, beside the lock. The fixture's stub libraries are real C++ built through vcpkg, from a git registry the tests create from a directory in the fixture map, so this runs in seconds.

**Blocked by:** [03: Start writes a composition](03-start-writes-a-composition.md)

**Status:** ready-for-agent

- [ ] On the fixture composition, `verify` reaches smoke on Linux in under a minute with warm caches.
- [ ] A build failure reports `BUILD_FAILED` with the build log as its details file and level none; a crash or unclean shutdown reports `SMOKE_FAILED` and level builds.
- [ ] Evidence validates against the evidence schema, names the lock's hash and the lock file's digest, records run conditions, and is marked unsigned (ADR-0032, ADR-0072).
- [ ] `start` ends by running `verify` for the host platform and reports its level.
