# 12: Jolt in the map

**What to build:** Jolt 5.6.0 enters the public map: library entry with facts including math types, library recipe (build fragment with `-fno-rtti` for glue deriving from Jolt classes; setup glue for the physics system, layers and temp allocator), allocation, diagnostics and jobs bindings, and the math-types contract. Extending the testbed with Jolt reaches probed on Linux in Automap's CI.

**Blocked by:** [08: SDL3 and Dear ImGui in the map](08-sdl3-and-dear-imgui-in-the-map.md), [11: Jobs contract and ThreadSanitizer](11-jobs-contract-and-threadsanitizer.md)

**Status:** done

- [x] `automap extend jolt` on the testbed ends probed on Linux, including `VerifyJoltVersionID()` and the ThreadSanitizer run.
- [x] The Jolt jobs adapter routes every job through the jobs interface and waits for in-flight jobs on unbind.
- [x] Every Jolt probe fails its seeded violation in CI.
- [x] Jolt's assert hook result records the build condition when the run is not a Debug build.

## Comments

- Jolt 5.6.0 enters the public map in its default variant (single precision), from the curated `joltphysics` port at port version 1. Pinned port versions are read from vcpkg's history, so CI clones vcpkg without blobs instead of shallow.
- Setup glue calls `JPH::VerifyJoltVersionID()` before `RegisterTypes()`, which aborts on a mismatch, and records the failure first; the ABI seed defines `JPH_CROSS_PLATFORM_DETERMINISTIC` for the glue, a define that changes the version id without changing any linked symbol. `verify` now runs the probes whenever the build succeeded, so a startup that stops on such a check still reports why.
- The jobs adapter derives from `JPH::JobSystemWithBarrier`, submits every job through `am::jobs`, and its destructor waits for jobs in flight. Jolt hands jobs between threads inside its library, which isn't built with ThreadSanitizer (a dependency completing there queues a job from another thread; the last release there frees it after an atomic fence), so the adapter makes that ordering explicit with one atomic that `CreateJob` and the workers release to and `QueueJob` and `FreeJob` acquire from. A `jobs.tsan-clean` failure names ThreadSanitizer's log as its details file.
- The jobs contract's race seed increments its counter in each worker before the worker's first lock, so the race is reported whatever the composition's jobs do; in the Jolt composition, job dependencies otherwise ordered most accesses.
- `math-types.precision-agrees` is judged by `verify` from the libraries' `position_precision` facts against the composition's choice. A probe result can carry a reason; Jolt's diagnostics result says the assert hook wasn't checked when Jolt's assertions are compiled out, and verify's Debug builds check it.
- A composition rule's seed is named by the library or pairing whose recipe carries it (`lifecycle.init-order:jolt`). A seed that isn't caught keeps the failing run's log, or the probe or ThreadSanitizer log, as its details file.
