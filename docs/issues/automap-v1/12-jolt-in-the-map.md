# 12: Jolt in the map

**What to build:** Jolt 5.6.0 enters the public map: library entry with facts including math types, library recipe (build fragment with `-fno-rtti` for glue deriving from Jolt classes; setup glue for the physics system, layers and temp allocator), allocation, diagnostics and jobs bindings, and the math-types contract. Extending the testbed with Jolt reaches probed on Linux in Automap's CI.

**Blocked by:** [08: SDL3 and Dear ImGui in the map](08-sdl3-and-dear-imgui-in-the-map.md), [11: Jobs contract and ThreadSanitizer](11-jobs-contract-and-threadsanitizer.md)

**Status:** ready-for-agent

- [ ] `automap extend jolt` on the testbed ends probed on Linux, including `VerifyJoltVersionID()` and the ThreadSanitizer run.
- [ ] The Jolt jobs adapter routes every job through the jobs interface and waits for in-flight jobs on unbind.
- [ ] Every Jolt probe fails its seeded violation in CI.
- [ ] Jolt's assert hook result records the build condition when the run is not a Debug build.
