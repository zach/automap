# 02c — Library catalog, part C: simulation, core runtime, math, scripting, networking, profiling

**Scope.** This document profiles the libraries for the simulation, runtime and glue slots that an Automap composition would wire together. It covers physics (Jolt, Box2D v3, PhysX 5, Bullet for contrast, plus Box3D, which is new), navigation (Recast/Detour), entity models (flecs, EnTT), core runtime (allocators, job systems, containers, logging), math libraries and the "math-types seam", profiling (Tracy, Optick, Superluminal), scripting and hot reload (Lua 5.4/5.5, LuaJIT, Luau, sol2, AngelScript, Wren, Live++, cr.h, jet-live) and networking (GameNetworkingSockets, ENet, yojimbo/netcode, Steamworks). Each profile records license, flavor, latest release, health, platform reach and the **integration seams** that glue code must handle. The document ends with a cross-library **seams catalog** that separates machine-checkable seams from agent guidance. Facts were taken from primary sources wherever possible: release feeds, headers, CMake files, changelogs and docs fetched from the repositories themselves.

Accessed: 2026-10-29

**Method and limits.** Release data comes from GitHub `releases.atom` feeds and commit pages. Seam details come from headers, CMake files and docs downloaded from `raw.githubusercontent.com` and read directly. The session's network policy blocked several vendor sites: lua.org, angelcode.com, liveplusplus.tech and superluminal.eu. The shared web-search budget ran out partway through the research. Claims that depend on those sources are marked **[Unverified]**, or are cited to a search-result summary and labeled as such.

---

## Key findings

1. **[Verified]** **Most seams are explicit hook points with documented signatures. Glue code is therefore regular enough to template.** The hooks found:
   - Jolt: `Allocate`/`Reallocate`/`Free`/`AlignedAllocate`/`AlignedFree` function-pointer globals, plus `Trace` and `AssertFailed` [4][7].
   - Box2D: `b2SetAllocator`, `b2SetAssertCallback` and `b2SetLogCallback` [19].
   - flecs: an `ecs_os_api_t` struct with malloc, thread, task, log, abort and perf-trace callbacks [34].
   - Recast/Detour: `rcAllocSetCustom`, `dtAllocSetCustom` and `rcAssertFailSetCustom` [32].
   - ENet: `enet_initialize_with_callbacks` [78].
   - Lua and Luau: a `lua_Alloc` passed to `lua_newstate` [67][70].
   - Dear ImGui: `SetAllocatorFunctions` [60].
   - PhysX: `PxAllocatorCallback` and `PxCpuDispatcher` [29].

   Godot 4.6's Jolt module is a production example of this glue: it routes allocator, trace and assert hooks into its own engine, then calls `Factory` and `RegisterTypes` [14].
2. **[Verified]** **ABI-affecting build defines are increasingly self-declaring and checkable.**
   - Jolt packs 12 feature bits into `JPH_VERSION_ID`, which `RegisterTypes()` and `VerifyJoltVersionID()` check at runtime [3][5]. The bits cover double precision, cross-platform determinism, FP exceptions, profiler, external profiler, debug renderer, temp allocator, custom allocator, object-layer bits, asserts and object stream.
   - Tracy 0.14 encodes its build options into a link symbol (`GetProfiler_CFG_…`), so a mismatch fails at link time [63][64].
   - Box2D 3.2's `config.h` labels `BOX2D_DOUBLE_PRECISION` "Affects ABI" and exposes `b2IsDoublePrecision()` [19].
   - `IMGUI_CHECKVERSION()` compares struct sizes between the compiled library and the caller [60].
   - unordered_dense puts everything in a versioned inline namespace [48].
3. **[Verified]** **Version churn breaks glue in 2025–2026, so recipes must be pinned to version ranges.**
   - Box2D v3.2.0 (2026-10-19) changed both the task-callback and allocator-callback signatures, renamed `b2SetAssertFcn`, and switched world positions to `b2Pos`/`b2WorldTransform` [18].
   - Lua 5.5 added a seed parameter to `lua_newstate` [67], which breaks sol2 3.5.0. The issue is still open [71].
   - EnTT v4.0.0 requires C++20 [36]. Taskflow v4 also requires C++20 [45].
   - Jolt `master` adds a pure-virtual `JobSystem::CreateJobs` [6][13]. The change is not listed in `APIChanges.md` because that file skips changes whose only effect is a compile error [11].
4. **[Verified]** **Determinism depends on build flags, and only some libraries guarantee it.**
   - Box2D v3 is deterministic across thread counts and 64-bit platforms by default. It relies on `/fp:precise`, `-ffp-contract=off` and its own trig functions, and different compiler flags break it [21].
   - Jolt is cross-platform deterministic only when built with `CROSS_PLATFORM_DETERMINISTIC` (about 8% slower), using the same source and the same defines on every platform [10].
   - PhysX exposes only `eENABLE_ENHANCED_DETERMINISM` [29].
   - netcode and yojimbo now build "strict FP on every target" [79][80].
5. **[Verified]** **The console/NDA tier exists across several libraries, each handled through a private channel.**
   - Jolt runs on "Platform Blue" only with a `PlatformBlue.h` obtained from the platform's developer forum [2].
   - GameNetworkingSockets console code is "not allowed to distribute here; contact us" [77].
   - Older PhysX changelogs say console code is subject to platform NDA [28].
   - LuaJIT documents `ps4build`, `ps5build`, `nxbuild` and `xb1build`, with the JIT and FFI disabled [69].
   - Tracy 0.14 added `TRACY_PLATFORM_HEADER` for porting to unsupported platforms [63][64].
6. **[Verified]** **The math-types seam is wide; no two of the profiled libraries share a vector type.** Jolt has `Vec3` and `RVec3`, where `RVec3` becomes `DVec3` under `JPH_DOUBLE_PRECISION` [8]. Box2D has `b2Vec2` plus `b2Pos`, which is double in large-world mode [18]. ImGui has `ImVec2` with `IM_VEC2_CLASS_EXTRA` conversion hooks [60]. ozz-animation has `ozz::math::Float3` [61], PhysX `PxVec3`, Bullet `btVector3` (switched by `BT_USE_DOUBLE_PRECISION`) [31], and Recast raw `float*` [32]. flecs and EnTT define no math types.
7. **[Verified]** **Maintenance health varies widely, and many libraries have a bus factor of about 1.**
   - Very active in Oct 2026: Jolt, Box2D, flecs, EnTT, Tracy, mimalloc, unordered_dense, quill, Luau (weekly releases), GNS, netcode/yojimbo and LuaJIT.
   - Stalled: sol2 (last release Feb 2025), Optick (2022), marl (CI-only commits), Bullet (2022), Wren (0.4.0 in 2021), HandmadeMath (2023), rtm (2024) and TLSF (2016).
   - Sources: [1][17][24][33][35][38][40][44][62][66][71][74].
8. **[Verified]** **Some libraries already ship machine-readable or agent-facing data that Automap could ingest.** Tracy 0.14 ships an MCP server so AI agents can analyze `.tracy` captures [63][64]. Box2D 3.2 ships JSON descriptions of its public API, one per precision [18]. Jolt's `APIChanges.md` marks serialization-breaking changes with *SBS* [11].
9. **[Verified]** **License metadata has to be tracked per version and per file.** The PhysX SDK moved from BSD-3 to Apache-2.0 in 5.11.0, and its headers now carry `SPDX: Apache-2.0`. The repository's root `LICENSE.md` still reads BSD-3 [27][28][29].
10. **[Inference]** **The job-system seam is the hardest to standardize because the libraries assume different scheduling semantics.** Jolt uses jobs with dependencies and barriers; Box2D 3.2 uses fire-and-finish tasks; flecs uses long-running workers or per-frame "task threads"; PhysX uses tasks with continuations. Bridging each job library to each consumer is an N×M glue matrix unless Automap defines a neutral scheduler adapter that reduces it to N+M.

---

## 1. Physics

### 1.1 Profiles

| Library | License | Flavor | Latest release | Health | Platforms (public) |
|---|---|---|---|---|---|
| **Jolt Physics** | MIT [2] | C++17. Compiled library; static or shared via `JPH_BUILD_SHARED_LIBS` [1][9]. No RTTI or exceptions [2]. | 5.6.0 (2026-07-11) [1]. `master` is at 5.6.1-dev [3]. | Commits through 2026-10-26, mostly by jrouwe (bus factor about 1) [13]. CLA required [2]. | Windows x86/x64/ARM64; Linux on x86/x64/ARM32/ARM64/RISC-V64/LoongArch64/PPC64LE; FreeBSD; Android; macOS; iOS; MinGW; WASM via JoltPhysics.js; "Platform Blue (a popular game console)" [2]. Shipped in Horizon Forbidden West and Death Stranding 2 [2]. Godot integration [2]. |
| **Box2D v3** | MIT [23] | Portable C17 [23]. Compiled library. | v3.2.0 (2026-10-19). Previous: 3.1.1 (2025-08-07) [17]. | Commits through 2026-10-20 by erincatto [24]. Funded by GitHub Sponsors [23]. README discloses LLM use for tests, samples, build files, code review and benchmarks [23]. | CI covers MinGW, clang-cl, Windows ARM64, Swift and Zig builds [18]. WASM via a third-party port [23]. Console use: **[Unverified]** (nothing in README or docs; a search was not possible). |
| **Box3D** (new) | MIT [25] | C17, from the same author and with the same design (data-oriented, SIMD, cross-platform determinism, recording and replay) [25]. | No releases; docs reference `v0.1.0` [25]. | Active: commits from 2026-07-09 through 2026-10-28 [25]. | Samples use D3D11, Metal and GL 4.5 via sokol [25]. Early-stage, so not yet profiled in depth. |
| **PhysX 5** | Apache-2.0 since 5.11.0, per the changelog and header SPDX lines. Root `LICENSE.md` still reads BSD-3 [27][28][29]. | C++. Compiled library, built with a packman/CMake generator script [28]. | Combined release "Omniverse PhysX 110.3 / SDK 5.9.1" (2026-10-27) [26]. `main` reports 5.12.0 [28]. | Active, NVIDIA-staffed, with heavy Omniverse/Isaac focus (ovphysx releases) [26][27]. | Runtime support listed as Linux and Windows 10+ x64 only; GPU path needs CUDA 12.8 and Volta or later [28]. GPU source was opened in 5.6 (2025) per secondary sources [30]. The README still says GPU binaries are fetched on demand [28], so this needs re-checking. |
| **Bullet** (contrast) | zlib [31] | C++. Compiled library. | 3.24 (2022-04-26); PyBullet 3.2.5 (2023-01-27) [31]. | Effectively dormant for games [31]. | Desktop and mobile (historical). |

### 1.2 Physics seams

| Seam | Jolt | Box2D 3.2 | PhysX 5 | Bullet |
|---|---|---|---|---|
| Allocator | Five function-pointer globals. `RegisterDefaultAllocator()` installs the defaults; `JPH_DISABLE_CUSTOM_ALLOCATOR` turns the pointers into link-time functions. `JPH_OVERRIDE_NEW_DELETE` routes class-level `new`/`delete` [4]. Godot assigns the five pointers to its own `Memory::*_static` [14]. A per-step `TempAllocator` is separate (`TempAllocatorImpl`) [12]. | `b2SetAllocator(b2AllocCallbackFcn(size, alignment), b2FreeCallbackFcn(mem, size))`. The signature changed in 3.2 [18][19]. | `PxAllocatorCallback`, passed to `PxCreateFoundation(version, allocator, errorCallback)` [29]. | `btAlignedAllocSetCustom` and `btAlignedAllocSetCustomAligned` [31]. |
| Jobs / threading | Abstract `JobSystem`. Implement `GetMaxConcurrency`, `CreateJob`, `CreateJobs` (pure virtual on `master`), `CreateBarrier`, `DestroyBarrier`, `WaitForJobs`, `QueueJob(s)` and `FreeJob`. Deriving from `JobSystemWithBarrier` supplies the barrier. `JobSystemThreadPool` is the reference implementation, with thread init/exit hooks [6]. Godot maps it onto `WorkerThreadPool::add_native_task` [15]. | `b2WorldDef.workerCount`, `enqueueTask(task, taskContext, userContext)`, `finishTask(userTask, userContext)` and `userTaskContext`. Returning NULL from `enqueueTask` means "ran inline". Without callbacks, Box2D now runs its own threads. 3.2 removed the item range and worker index, and `B2_MAX_WORKERS` dropped from 64 to 32 [18][20]. | `PxCpuDispatcher::submitTask` / `getWorkerCount` [29]. | Not profiled. |
| Math types | `Vec3`/`Mat44` (SIMD-aligned). `RVec3`/`RMat44` become `DVec3`/`DMat44` under `JPH_DOUBLE_PRECISION` [8]. | `b2Vec2` and `b2Rot`. In double mode `b2Pos` and `b2WorldTransform` hold double translation; in single precision they are aliases [18][20]. | `PxVec3` and `PxTransform`. | `btVector3`; `btScalar` is float or double [31]. |
| Handles | `BodyID`: 23-bit index, 8-bit sequence number, top bit reserved for the broadphase [8]. Bodies are accessed through `BodyInterface` and `BodyLockInterface` [10]. | Value IDs: `b2BodyId {int32 index1; uint16 world0; uint16 generation}`. Pack and unpack with `b2StoreBodyId`/`b2LoadBodyId` (uint64); `B2_IS_NULL` tests for null [20]. | Raw pointers to `PxBase`-derived objects, freed with `release()` [29]. | Raw pointers. |
| Init / shutdown | 1. `RegisterDefaultAllocator` (or custom hooks); 2. `Trace`/`AssertFailed`; 3. `Factory::sInstance = new Factory`; 4. `RegisterTypes()`; 5. custom shape `register_type` calls; 6. `TempAllocator`; 7. `JobSystem`; 8. `PhysicsSystem::Init`. Shutdown: `UnregisterTypes()`, then delete `Factory::sInstance` [12][14]. Only one RTTI factory exists [10]. | Set allocator and assert hooks before creating any world; `b2CreateWorld` / `b2DestroyWorld` [19]. | `PxCreateFoundation` → `PxCreatePhysics` → scene; release in reverse order [28][29]. | — |
| Errors / asserts | `JPH_ASSERT` is active only with `JPH_ENABLE_ASSERTS`, which `JPH_DEBUG` turns on. `AssertFailed` returns bool (true = break). `master` replaces the `NDEBUG` coupling with `JPH_DEBUG` [3][7][11]. | `B2_ASSERT` is active when `!NDEBUG` or `B2_ENABLE_ASSERT`. 3.2 rejects and logs invalid input in release builds [18][19]. | `PxErrorCallback`; recent fixes report errors as `eINTERNAL_ERROR` instead of crashing [28]. | — |
| Determinism | Same binary and same API call order is deterministic. `CROSS_PLATFORM_DETERMINISTIC` gives cross-compiler/OS/architecture determinism at about 8% cost. Same defines are required on every platform. Broadphase query order and listener order are not deterministic [10]. `master` removed the `mDeterministicSimulation` toggle [11]. | On by default across thread counts and 64-bit platforms. Requires precise FP on MSVC, `-ffp-contract=off` on clang/GCC and Box2D's own atan2/sin/cos [21]. Order follows creation order [21]. | `eENABLE_ENHANCED_DETERMINISM` flag [29]. No cross-platform claim was found. | — |
| ABI defines | `JPH_DOUBLE_PRECISION`, `JPH_CROSS_PLATFORM_DETERMINISTIC`, `JPH_FLOATING_POINT_EXCEPTIONS_ENABLED`, `JPH_PROFILE_ENABLED`, `JPH_EXTERNAL_PROFILE`, `JPH_DEBUG_RENDERER`, `JPH_DISABLE_TEMP_ALLOCATOR`, `JPH_DISABLE_CUSTOM_ALLOCATOR`, `JPH_OBJECT_LAYER_BITS` (16/32/64), `JPH_ENABLE_ASSERTS`, `JPH_OBJECT_STREAM`. All are encoded in `JPH_VERSION_ID` and checked [3][5]. Not encoded: SIMD flags (`USE_AVX2` etc.) and the MSVC static CRT (default ON) [9]. | `BOX2D_DOUBLE_PRECISION` ("Affects ABI"), `BOX2D_DISABLE_SIMD`, `BOX2D_VALIDATE`, `BOX2D_EXPORT`/`BOX2D_DLL`. Override them through `BOX2D_USER_CONFIG`. AVX2 is now selected at runtime [18][19]. | `PX_PHYSICS_VERSION` is passed to `PxCreateFoundation` [29]. | `BT_USE_DOUBLE_PRECISION`, queryable at runtime via `btIsDoublePrecision()`; version via `btGetVersion()` (326 on master) [31]. |
| Serialization | `ObjectStream` (`JPH_OBJECT_STREAM`) for "uncooked" settings. `SaveBinaryState` for cooked data, whose format changes between versions and depends on `JPH_DOUBLE_PRECISION` and `JPH_OBJECT_LAYER_BITS`. `PhysicsSystem::SaveState`/`RestoreState` for rollback. Only load trusted data [10][11]. | Recording and replay, plus experimental snapshots, added in 3.2 [18]. | `PxSerialization` binary collections [28]. | — |
| Existing bridges | `DebugRendererSimple` (DrawLine, DrawTriangle, DrawText3D) [Jolt `Renderer/DebugRendererSimple.h`]. External profiler hook via `JPH_EXTERNAL_PROFILE`, with user-supplied start/end functions; there is no built-in Tracy integration [7]. Godot module [14][15]. | `b2DebugDraw` callbacks [18]. `BOX2D_PROFILE` fetches Tracy **pinned at v0.14.1** with `TRACY_MANUAL_LIFETIME` [23]. | OmniPVD [28]. | — |

---

## 2. Navigation: Recast/Detour

| Attribute | Value |
|---|---|
| License / flavor | zlib. C++98, no dependencies [32]. |
| Latest release | 1.6.0 (2023-05-21). Earlier: 1.5.1 (2016) [32]. |
| Health | Commits through 2026-02-27, maintained mainly by grahamboree (Catch2 3.12 update, rasterization tests) [32]. Release cadence is very slow. |
| Allocator | `rcAllocSetCustom(rcAllocFunc(size, rcAllocHint PERM/TEMP), rcFreeFunc)` and `dtAllocSetCustom`. 1.6.0 aligned `rcNew`/`rcDelete` with them [32]. |
| Asserts / logging | `rcAssertFailSetCustom`. Logging and timers go through `rcContext::doLog` and `doStartTimer` overrides [32]. |
| Threading | No scheduler hook. Building tiles in parallel and using one `dtNavMeshQuery` per thread are left to the caller **[Inference]**; no thread-safety statement was found in the headers. |
| Math types | Raw `float*` triplets. |
| Handles / ownership | `dtPolyRef` and `dtTileRef` are 32-bit, or 64-bit with `DT_POLYREF64`. `addTile(..., DT_TILE_FREE_DATA)` hands tile memory ownership to the navmesh, which frees it with `dtFree` [32]. |
| ABI defines | `DT_POLYREF64` (CMake `RECASTNAVIGATION_DT_POLYREF64`) and `DT_VIRTUAL_QUERYFILTER` [32]. |
| Errors | `dtStatus` bitfield return codes [32]. |
| Debug bridge | `duDebugDraw` virtual interface (begin, vertex, end) in DebugUtils [32]. |

---

## 3. Entity model

| | **flecs** | **EnTT** |
|---|---|---|
| License / flavor | MIT. C99 core plus a C++17 API that "doesn't use STL containers". Shipped as an amalgamated `flecs.c`/`flecs.h` [34]. | MIT. Header-only. **C++20 required since v4.0.0** [36][37]. |
| Latest release | v4.1.6 (2026-06-29). Breaking changes: a new entity-range API, query groups no longer ordered by default, and a changed amalgamation build [33]. | v4.0.0 (2026-07-23). Before that, v3.16.0 (2025-11-20) [36]. `master` headers report 4.1.0 [37]. |
| Health | Daily commits (2026-10-29), essentially a single maintainer [35]. | Daily commits on `wip` (2026-10-27) by skypjack alone [38]. Sponsors; used in Minecraft [37]. |
| Allocator | `ecs_os_api_t.malloc_`, `realloc_`, `calloc_` and `free_`. Configure with `ecs_os_set_api_defaults()`, then `ecs_os_set_api()` before `ecs_init` [34]. | Standard allocator template parameters (`basic_registry`, `basic_storage<Type, Entity, Allocator>`) [37]. |
| Threading | Choose one of two modes. `ecs_set_threads` creates long-running workers. `ecs_set_task_threads` drives an **external job system** through the `task_new_`/`task_join_` OS-API hooks. Switching mode ends the other [34]. | No scheduler; the user owns threading. `ENTT_USE_ATOMIC` makes internal counters atomic [37]. |
| Handles | `ecs_entity_t` is a uint64: low 32 bits index, bits 32–47 generation [34]. | `entt::entity`, an enum class over `ENTT_ID_TYPE` (default uint32: 20-bit index, 12-bit version) [37]. |
| Errors / asserts | `FLECS_DEBUG` vs `FLECS_NDEBUG` (derived from `NDEBUG`), `FLECS_SOFT_ASSERT`, `FLECS_KEEP_ASSERT`, `FLECS_SANITIZE`. The log and abort hooks live in the OS API [34]. | `ENTT_ASSERT` (overridable), `ENTT_DISABLE_ASSERT`, `ENTT_NO_EXCEPTION` (renamed in v4) [36][37]. |
| ABI / config defines | `FLECS_CUSTOM_BUILD` plus addon defines (`FLECS_CPP`, `FLECS_META`, `FLECS_REST`, `FLECS_PERF_TRACE`, ...). `FLECS_DEBUG_INFO` "adds additional debug information to internal data structures" [34], so a debug/release mismatch between flecs.c and its consumers is risky **[Inference]**. | `ENTT_ID_TYPE`, `ENTT_SPARSE_PAGE`, `ENTT_PACKED_PAGE`, `ENTT_USE_ATOMIC`, `ENTT_NO_ETO`, `ENTT_NO_MIXIN`, `ENTT_STANDARD_CPP`. An optional `<entt/ext/config.h>` is picked up automatically if present [37]. Because EnTT is header-only, every translation unit must agree on these **[Inference]**. |
| Math types | None. | None. |
| Serialization | JSON (`FLECS_JSON`) and reflection (`FLECS_META`) [34]. | Snapshot and meta facilities (meta API changed in v4) [36]. |
| Bridges | `perf_trace_push_`/`perf_trace_pop_` hooks (`FLECS_PERF_TRACE`) map onto Tracy-style zones [34]. A flecs+Jolt example was **[Unverified]**: none found, and search was unavailable. | None found. |

Other options, mentioned briefly: gaia-ecs reached v1.0.0 (2026-08-22) [39] but was not profiled.

---

## 4. Core runtime

### 4.1 Allocators

| Library | License | Latest | Seams |
|---|---|---|---|
| **mimalloc** | MIT | v3.6.0 (2026-10-23), which adds profiling hooks. v2.5.2 and v1.15.2 are legacy (2026-09-12) [40]. | Drop-in malloc override. v3 adds true first-class heaps usable from any thread. Ports exist for Windows, macOS, Linux, WASM, BSDs and others. Recent releases fix Xbox builds [40]. |
| **rpmalloc** | Unlicense OR MIT | 2.0.1 (2026-07-15); 2.0.0 (2026-07-10) after a two-year gap [41]. | `rpmalloc_initialize[_config]`, with optional custom map/commit/unmap. Per-thread init and finalize. `ENABLE_OVERRIDE`; first-class heaps need `RPMALLOC_FIRST_CLASS_HEAPS=1`. Windows needs `rpnew.h` in exactly one TU. Static-library linkers may drop the override unless `rpmalloc_linker_reference` is called [41]. Tracy bundles its own copy [63]. |
| **TLSF** (mattconte) | BSD | v3.1 (2016) per README [42]. | O(1) pool allocator that you add and remove pools to. **Not thread-safe; "the user must provide this"** [42]. yojimbo vendors it [80]. |

### 4.2 Job systems and library-defined job interfaces

| Scheduler / interface | License, version, health | How a consumer plugs in |
|---|---|---|
| **enkiTS** | zlib. C++11 and C APIs. v1.12 (2026-07-04) [43]. | `TaskSchedulerConfig { numTaskThreadsToCreate, numExternalTaskThreads, profilerCallbacks, customAllocator }`, `RegisterExternalTaskThread()`, and pinned tasks for main or IO threads [43]. Platforms: Windows, Linux, macOS, Android ("should work on iOS") [43]. |
| **marl** | Apache-2.0. Hybrid thread/fiber scheduler, C++11 [44]. | Fibers mean a blocking non-marl call stalls the worker [44]. Platforms: Windows, macOS, Linux, FreeBSD, Fuchsia, Emscripten, Android, iOS [44]. Only CI commits since 2025; "not an officially supported Google product" [44]. |
| **Taskflow** | MIT. v4.1.0 (2026-06-20). **v4 requires C++20** [45]. | Task-graph executor. Integration specifics for external pools were **[Unverified]** because the release notes are hosted off GitHub. |
| **oneTBB** | Apache-2.0. 2023.2.0 (2026-10-28), together with the first open-source release of the Thread Composability Manager [46]. | `task_arena` and `parallel_phase` leave policies [46]. |
| Jolt `JobSystem` | See §1.2 [6]. | Provide your own subclass, typically via `JobSystemWithBarrier`. Godot's subclass is a reference implementation [15]. |
| Box2D `enqueueTask` / `finishTask` | See §1.2 [18][20]. | Wrap each Box2D task as one job, and implement `finishTask` as wait-for-that-job. |
| flecs task threads | See §3 [34]. | `task_new_` and `task_join_` map onto job submit and wait. |
| PhysX `PxCpuDispatcher` | [29] | Implement `submitTask` so the task's `run()` and `release()` execute on your workers. |

### 4.3 Containers

| Library | License | Latest | Notes and seams |
|---|---|---|---|
| **EASTL** | BSD-3 [47] | 3.27.01 (2025-11-15), with C++14 compatibility fixes [47]. | Depends on EABase. **You must define two `operator new[]` overloads** (name, flags, debugFlags, file, line; aligned variant) [47]. |
| **ankerl::unordered_dense** | MIT | v5.3.3 (2026-10-22). Many releases in Sep–Oct 2026 [48]. | C++17. `std` and `pmr` allocators. v5 changed hash values (persisted data must be rebuilt) and renamed bucket types. Uses a **versioned inline namespace** [48]. |
| **parallel-hashmap / gtl** | Apache-2.0 [49] | phmap 2.0.0 (2025-01-21); gtl 1.2.0 (2024-09-21) [49]. | Slow cadence. |
| **plf::colony** | zlib [50] | v7.7.54 (commit 2026-10-23); no GitHub releases [50]. | C++98 through C++23 compatible [50]. Status of std::hive standardization: **[Unverified]** (search budget exhausted). |

### 4.4 Logging and formatting

| Library | License | Latest | Seams |
|---|---|---|---|
| **spdlog** | MIT | 1.17.0 (2026-01-04), bundling fmt 12.1.0 [51]. | Header-only (C++11) or compiled. `SPDLOG_USE_STD_FORMAT`, `SPDLOG_FMT_EXTERNAL` and `SPDLOG_FMT_EXTERNAL_HO` are mutually exclusive (CMake enforces this). `SPDLOG_NO_EXCEPTIONS` [51]. The compiled library and its consumers must agree on the fmt backend **[Inference]**. |
| **{fmt}** | MIT-style [52] | 12.2.0 (2026-06-16): adds a C11 API (`fmt-c`) and a C++20 modules target [52]. | Version skew between a bundled fmt (inside spdlog) and the engine's own fmt is a classic ODR seam **[Inference]**. |
| **quill** | MIT | v13.1.0 (2026-10-17); v13.0.0 (2026-08-30) [53]. | Asynchronous, C++17. `Backend::start()` launches a backend thread, which is a lifecycle seam [53]. |

---

## 5. Math libraries and the math-types seam

### 5.1 Libraries

| Library | License | Flavor | Latest | Layout-affecting config |
|---|---|---|---|---|
| **GLM** | Happy Bunny (modified MIT) / MIT [54] | C++, header-only | 1.0.3 (feed 2026-01-12; notes dated 2025-12-31) [54] | `GLM_FORCE_DEFAULT_ALIGNED_GENTYPES` (changes padding), `GLM_FORCE_DEPTH_ZERO_TO_ONE`, `GLM_FORCE_LEFT_HANDED`, `GLM_FORCE_INTRINSICS`/`PURE`, `GLM_FORCE_PRECISION_*`, `GLM_FORCE_SIZE_T_LENGTH`. `GLM_FORCE_MESSAGES` prints the active configuration [54]. |
| **HandmadeMath** | Public domain [55] | Single header for C and C++ | 2.0.0 (2023-02-20) [55] | Configurable angle units and handedness [55]. |
| **DirectXMath** | MIT [56] | All-inline SIMD C++ | "June 2026" hotfix of 3.21 (2026-07-15) [56] | Non-Windows builds need a `sal.h`. `_XM_NO_XMVECTOR_OVERLOADS_` on clang/GCC. SSE/AVX/NEON paths [56]. |
| **cglm** | MIT [57] | C99, inline and call APIs, plus a struct API | 0.9.7 (2026-10-25) [57] | `CGLM_ALL_UNALIGNED`. Since 0.9.7 the call-API headers no longer include the inline API [57]. |
| **Realtime Math (rtm)** | MIT [58] | C++11, header-only (from ACL) | 2.3.1 (2024-03-16) [58] | — |
| **linalg.h** | Unlicense [59] | C++11, single header | 2.2 (no date shown) [59] | — |

### 5.2 Which math types each library owns

| Library | Own types | Conversion hook | Notes |
|---|---|---|---|
| Jolt | `Vec3`, `Vec4`, `Mat44`, `Quat`; `RVec3`/`RMat44` (= `DVec3`/`DMat44` in double mode) [8] | None built in. Write free-function adapters. | 16-byte SIMD alignment (`JPH_VECTOR_ALIGNMENT`) [8]. |
| Box2D | `b2Vec2`, `b2Rot`, `b2Transform`; `b2Pos`, `b2WorldTransform` [18][20] | None (C structs) | Single precision: `b2Pos` aliases `b2Vec2`. Double precision: double translation [18]. |
| ImGui | `ImVec2`, `ImVec4` [60] | `IM_VEC2_CLASS_EXTRA` / `IM_VEC4_CLASS_EXTRA` in `imconfig.h` or `IMGUI_USER_CONFIG` [60] | `IMGUI_CHECKVERSION()` checks `sizeof(ImVec2)` and related sizes [60]. |
| ozz-animation | `ozz::math::Float2/3/4`, `Float4x4`, SIMD types [61] | None | C++17; custom allocator via `memory::SetDefaulAllocator` (spelling as in the source) [61]. 0.17.0 (2026-08-01) [61]. |
| Recast/Detour | `float*` (x,y,z), y-up [32] | Not needed (POD arrays) | — |
| PhysX | `PxVec3`, `PxQuat`, `PxTransform` | None | — |
| Bullet | `btVector3`, `btScalar` [31] | None | `BT_USE_DOUBLE_PRECISION` switches `btScalar` to double [31]. |
| flecs / EnTT | None [34][37] | — | The engine chooses its own math components. |

**Seam summary [Inference]:** a composition picks one "engine math" library (GLM, DirectXMath, HandmadeMath or rtm) and needs one adapter per library that owns its own types. In practice that means Jolt, Box2D, ImGui, ozz and PhysX. Precision must also agree across the composition: a large-world engine needs both `JPH_DOUBLE_PRECISION` and `BOX2D_DOUBLE_PRECISION` on, plus a double-capable engine-math type.

---

## 6. Profiling

| | **Tracy** | **Optick** | **Superluminal** |
|---|---|---|---|
| License | BSD-3 (repository `LICENSE`) [65] | MIT [66] | Proprietary **[Unverified]** |
| Latest | 0.14.1 (2026-08-22); 0.14.0 (2026-08-09) [62]. `master` is 0.14.2-dev with `ProtocolVersion = 83` [65]. | v1.4.0 (2022-05-15); last commit 2022-07-07 [66] | **[Unverified]**: superluminal.eu was blocked and search was exhausted. |
| Version coupling | The manual says to use the same profiler version on client and server, because "the network protocol may change in-between releases". The UI greys out incompatible clients [64]. NEWS: "no guarantee that version mismatched client and server will be able to talk" [63]. | — | — |
| ABI defines | `TRACY_ENABLE`, `TRACY_ON_DEMAND`, `TRACY_NO_CALLSTACK`, `TRACY_MANUAL_LIFETIME`, `TRACY_FIBERS` and about 30 other `TRACY_*` switches [65]. **Since 0.14.0, mismatches fail at link time through mangled symbol names.** One gap: the client omits `TRACY_ENABLE` while the library was built with it [63][64]. Since 0.14.0, `TRACY_ENABLE` is OFF by default in CMake/Meson [63]. | `optick.config.h` [66] | — |
| Multi-DLL | One profiler DLL containing `TracyClient.cpp`, plus `TRACY_IMPORTS` on Windows. `TRACY_MANUAL_LIFETIME` provides `StartupProfiler`/`ShutdownProfiler` [64]. | — | — |
| Platforms | CPU instrumentation for C, C++, Lua, Python and Fortran. GPU zones for OpenGL, Vulkan, D3D11/12, Metal, OpenCL, CUDA and WebGPU [65]. Reduced features on macOS/iOS [64]. `TRACY_PLATFORM_HEADER` hooks (thread id, user info, safe copy, allocator) for unsupported platforms, added in 0.14.0 [63][64]. | Windows, Linux, macOS; Xbox and PS4 columns in the README [66]. | **[Unverified]** |
| Bridges | Box2D `BOX2D_PROFILE` (pins v0.14.1) [23]. Jolt via `JPH_EXTERNAL_PROFILE` user hooks [7]. flecs `perf_trace_*` hooks [34]. enkiTS `ProfilerCallbacks` [43]. Lua bindings [65]. **MCP server for AI agents** (0.14.0) [63][64]. | UE4 plugin [66]. | — |

---

## 7. Scripting and hot reload

| Library | License | Latest / health | Seams |
|---|---|---|---|
| **Lua 5.5** | MIT | 5.5.0 released 2025-12-22 per a search summary of lua.org [68]. Mirror tags: v5.5.0 (2025-12-13), v5.5.1 (2026-07-23), v5.4.9 (2026-08-07) [67]. | **`lua_newstate(lua_Alloc, ud, unsigned seed)` gained a third parameter.** `lua_resetthread` and `lua_setcstacklimit` are deprecated. GC parameters moved to `LUA_GCPARAM`. `global` is now reserved and `for` control variables are read-only. "Never assume binaries are compatible between versions" [67]. Errors use `throw` when compiled as C++ and `longjmp` when compiled as C (`LUAI_THROW`), so the build mode must match the bindings [67]. |
| **Lua 5.4** | MIT | 5.4.9 tag (2026-08-07) [67] | Same `lua_Alloc` seam with the two-argument `lua_newstate`. |
| **LuaJIT** | MIT [69] | No new release tags were checked. The `v2.1` branch has commits through 2026-10-21, all by Mike Pall [69]. One commit mentions a "v3.0 backport"; a public v3.0 branch was not visible **[Unverified]**. | Lua 5.1 API. On iOS and consoles (PS4, PS5, Vita, Xbox 360, Xbox One, Switch NX32/64) the JIT and FFI are disabled; only the interpreter runs [69]. |
| **Luau** | MIT [70] | 0.744 (2026-10-23); weekly cadence [70] | Runtime C++11; compiler and analysis C++17. Same `lua_newstate(lua_Alloc, ud)` signature. No `__gc` (use `lua_newuserdatadtor`). Sandboxing via `luaL_sandbox`. Used in Roblox, Alan Wake 2 and Warframe [70]. |
| **sol2** | MIT [71] | v3.5.0 (2025-02-25); last commit 2025-03-07 [71]. **No Lua 5.5 support:** issue #1747 is open and PR #1723 unmerged. Issue #1754 promotes LuaBridge3 as an "active alternative"; the maintainer has not replied [71]. | Header-only C++17. Supports Lua 5.1–5.4 and LuaJIT [71]. |
| LuaBridge3 (alternative) | — | 3.0-rc13 (2026-10-18). The feed shows Luau 0.743 and a Lua 5.5 addition in rc5 [72]. | Not profiled further. Its 5.5/Luau claims come from its own author [71][72]. |
| **AngelScript** | zlib-style text in `angelscript.h` [73] | 2.38.0. The header confirms the version [73]; the 2025-08-08 date comes from a search summary of angelcode.com [73]. | Native calling-convention support per platform **[Unverified]** (angelcode.com was blocked). |
| **Wren** | MIT | 0.4.0 (2021-04-09); last commit 2025-11-19 [74] | C99 / C++98 embedding API [74]. Low activity. |
| **cr.h** | MIT [75] | No releases; commits 2026-06-23 [75] | Host/guest DLL reload. `CR_STATE` globals are preserved; crashes roll back to the previous build. Host and DLL must share the same dynamic CRT (`/MD`). Multiple plugins are flaky on Linux/macOS [75]. |
| **jet-live** | MIT [76] | Last commit 2026-03-02 [76] | Linux x86_64 and macOS arm only (Windows: "no plans"). Needs `compile_commands.json`, `-falign-functions=16`, and debug, non-stripped builds. Not thread-safe [76]. |
| **Live++** | Commercial | **[Unverified]**: liveplusplus.tech was blocked and the search returned nothing relevant. | — |

**DLL-reload gameplay pattern [Inference, supported by cr.h docs [75]].** Gameplay code lives in a shared library that exposes a C ABI entry point. State lives in host-owned memory or tagged globals. Everything that crosses the boundary must share the allocator and CRT. Every library with global hook state needs the hooks re-installed or shared across the reload boundary: Jolt's `Allocate`, `Trace`, `Factory::sInstance`; Box2D's allocator; Tracy's profiler instance (`TRACY_IMPORTS`) [4][19][64].

---

## 8. Networking

| Library | License | Latest / health | Seams |
|---|---|---|---|
| **GameNetworkingSockets** | BSD-3 (Valve copyright header) [77] | v1.6.1 (2026-10-03), which adds iOS, tvOS and visionOS with "no public interfaces changed". Commits through 2026-10-02 [77]. | `GameNetworkingSockets_Init` / `_Kill`. `SteamNetworkingSockets_SetCustomMemoryAllocator(malloc, free, realloc)` works **only if built with `STEAMNETWORKINGSOCKETS_ENABLE_MEM_OVERRIDE`**, must be called exactly once before anything else, and "a few allocations" still go to the CRT [77]. Handles are uint32 (`HSteamNetConnection`, `HSteamListenSocket`, `HSteamNetPollGroup`) [77]. Dependencies: protobuf (protoc and runtime must be the same version; protobuf 22+ forces C++17 and pulls in Abseil) and one of OpenSSL, libsodium or bcrypt (bcrypt "to satisfy an Xbox requirement") [77]. Console code is distributed privately [77]. |
| **Steamworks networking** | Proprietary SDK | — | Same `ISteamNetworkingSockets` API as GNS; adds Steam Datagram Relay [77]. |
| **ENet** | MIT [78] | 1.3.18 (header); latest commit 2026-06-23 [78] | C. `enet_initialize_with_callbacks(version, {malloc, free, no_memory})`. Peers are pointers (`ENetPeer*`) [78]. |
| **netcode** | BSD-3 [79] | 1.4.8 (2026-09-13). Several security releases in 2026 (1.4.0 "security release", single-use connect tokens) [79]. | C. `netcode_init`/`netcode_term` manage global state. Per-client/server `allocate_function`, `free_function` and `allocator_context`. `netcode_set_assert_function` [79]. |
| **yojimbo** | BSD-3 [80] | 1.14.0 (2026-10-16). Vendored libraries build with `-ffp-contract=off`; a system TLSF is supported [80]. | C++. `InitializeYojimbo`/`ShutdownYojimbo`. Built on netcode, reliable and serialize, and vendors libsodium and TLSF [80]. Patreon-funded [80]. |

---

## 9. Per-slot comparison

| Slot | Default pick (inference) | Alternatives | Deciding seam |
|---|---|---|---|
| 3D physics | Jolt 5.6 | PhysX 5 (Win/Linux, GPU), Box3D (early), Bullet (avoid) | Job-system bridge; `JPH_*` defines; double precision |
| 2D physics | Box2D 3.2 | — | 3.2 API break; double precision; native vs external tasks |
| Navigation | Recast/Detour 1.6 (pin a commit) | — | `DT_POLYREF64`; tile ownership |
| ECS | flecs 4.1 (C/C++17) | EnTT 4 (C++20 only) | C++ standard; external task threads (flecs) |
| Allocator | mimalloc v3 | rpmalloc 2, TLSF (pools) | Override vs explicit heaps; CRT and DLL boundary |
| Jobs | enkiTS | Taskflow 4 (C++20), oneTBB, marl (stalled) | Adapter semantics per consumer |
| Containers | std + unordered_dense | EASTL (needs `operator new[]`), plf::colony | Persisted hashes (unordered_dense v5) |
| Logging | spdlog 1.17 or quill 13 | — | fmt vs `std::format` configuration |
| Math | Project choice | GLM, DirectXMath, HandmadeMath, rtm, cglm | Adapters to Jolt, Box2D, ImGui, ozz |
| Profiling | Tracy 0.14.x (pinned across all consumers) | Optick (dormant), Superluminal (commercial) | Version and protocol lock; link-time macro check |
| Scripting | Lua 5.4 + sol2, or Luau | Lua 5.5 + LuaBridge3, AngelScript, Wren | sol2 lacks 5.5; `lua_newstate` signature |
| Hot reload | cr.h (DLL pattern) | Live++ (commercial), jet-live (Linux/macOS) | Shared CRT and allocator; hook re-install |
| Networking | GNS 1.6 | ENet (light), yojimbo/netcode | protobuf/crypto dependencies; allocator override needs a build flag |

---

## 10. Seams catalog

Legend for "Checkable": **M** = machine-checkable (defines, versions, signatures, runtime self-checks); **P** = partly; **G** = guidance for agents.

| Seam | How the libraries differ (concrete) | Glue an integration needs | Checkable |
|---|---|---|---|
| **Allocation** | Several hook styles:<br>• Function-pointer globals: Jolt [4], Recast/Detour [32], Box2D [19], Bullet [31], ImGui [60].<br>• Struct of callbacks: flecs OS API [34], ENet [78].<br>• Per-instance: netcode config [79], Lua/Luau `lua_Alloc` [67][70].<br>• Interface object: PhysX [29], ozz [61], enkiTS `CustomAllocator` [43].<br>• Template allocator: EnTT, unordered_dense [37][48].<br>• Build-flag-gated: GNS `ENABLE_MEM_OVERRIDE` [77].<br>• User must define symbols: EASTL `operator new[]` [47].<br>• Global override: mimalloc, rpmalloc [40][41].<br>Signatures differ in size+alignment (Box2D 3.2), free-with-size, and realloc with old size (Jolt). | One engine allocator facade with adapters per signature. Install every hook before the library's first allocation. Keep CRT and DLL boundaries consistent (cr.h `/MD`, ImGui per-DLL `SetAllocatorFunctions`). | **P**: hook presence and signatures can be extracted from headers. Ordering and "exactly once" rules (GNS) are **G**. |
| **Jobs / threading** | • Jolt: dependency jobs plus barriers; `WaitForJobs` blocks [6].<br>• Box2D 3.2: enqueue/finish per task; internal threads if no callbacks; max 32 workers [18].<br>• flecs: workers or external task threads [34].<br>• PhysX: `submitTask` [29].<br>• Tracy: thread naming and fibers.<br>• quill: owns a backend thread [53].<br>• GNS: internal service thread **[Inference]**.<br>• marl: fibers, so blocking calls stall [44]. | One adapter per (scheduler × consumer) pair, or one neutral Automap scheduler interface. Avoid deadlock when a library waits from inside a worker (Jolt `WaitForJobs`, Box2D `finishTask`). Count workers consistently (Box2D's 32 cap; Jolt's `GetMaxConcurrency`). | **G** (semantics), **M** (interface signature and version) |
| **Math types** | See §5.2. Precision modes: Jolt `RVec3`, Box2D `b2Pos`, Bullet `btScalar`. Alignment: Jolt SIMD, GLM `FORCE_DEFAULT_ALIGNED`, cglm `ALL_UNALIGNED`. Handedness and clip-space flags in GLM. | Adapter headers between engine math and each library's types. Precision consistency across physics libraries. Up-axis conversion (Recast is y-up). | **P**: precision and alignment defines are **M**; handedness and axis conventions are **G**. |
| **Handles / ownership** | • Generational IDs: Jolt `BodyID` (23+8 bits) [8], Box2D `b2*Id` [20], flecs uint64 (32+16) [34], EnTT (20+12) [37].<br>• Opaque uint32 handles: GNS [77].<br>• Salted refs: Detour `dtPolyRef` (32/64) [32].<br>• Raw pointers with `release()`: PhysX [29], ENet [78].<br>• Ownership transfer: Detour `DT_TILE_FREE_DATA` [32]. | Mapping tables from ECS entity to physics ID and back (store Box2D IDs as uint64 via `b2StoreBodyId`). Lifetime hooks on entity destruction. Never cache raw PhysX or ENet pointers across destroy. | **G** (ID bit widths are **M**) |
| **Lifecycle / init order** | • Jolt: allocator → Trace/Assert → Factory → `RegisterTypes` → custom types → TempAllocator/JobSystem → PhysicsSystem; reverse to shut down [12][14].<br>• GNS: allocator once, then `GameNetworkingSockets_Init` [77].<br>• netcode: `netcode_init`/`term`; yojimbo: `Initialize`/`ShutdownYojimbo` [79][80].<br>• Tracy: `StartupProfiler`/`ShutdownProfiler` under `MANUAL_LIFETIME`, and shutdown is mandatory [63][64].<br>• flecs: OS API before `ecs_init` [34]. rpmalloc: process and thread init [41]. quill: `Backend::start` [53]. | A generated init/shutdown graph: allocator → log → profiler → jobs → subsystems, then reverse. Re-run hook installation on DLL reload. | **P**: ordering can be expressed as a DAG and validated. |
| **Error handling / asserts / logging** | • Bool-returning assert hooks: Jolt (true = break) [7]; Box2D int-returning [19].<br>• Void hooks: Recast [32], netcode [79].<br>• Error-callback objects: PhysX [29].<br>• Status codes: Detour `dtStatus` [32].<br>• Exceptions vs longjmp: Lua [67]; EnTT `ENTT_NO_EXCEPTION` [37]; spdlog `NO_EXCEPTIONS` [51]; Jolt has none [2].<br>• Box2D 3.2 logs and rejects bad input in release builds [18]. | Route every library into the engine's log and assert. Agree on exceptions on or off across the composition. | **P**: exception and RTTI flags are **M**; severity mapping is **G**. |
| **Serialization** | Jolt `SaveBinaryState` is version- and define-dependent, with *SBS* changes listed in APIChanges [10][11]. Box2D recording and snapshots (3.2) [18]. unordered_dense v5 changed hashes [48]. Lua bytecode is incompatible across versions [67]. PhysX binary collections [28]. | Treat engine-level data as the source of truth; re-cook library blobs when the version or defines change (Jolt advises the same) [10]. | **M**: version and define fingerprints can gate caches. |
| **Build flags / ABI** | Self-checking: Jolt `JPH_VERSION_ID` [3][5], Tracy link mangling [63], ImGui `IMGUI_CHECKVERSION` [60], PhysX version argument [29], Box2D `b2IsDoublePrecision` [19], unordered_dense inline namespace [48]. Unchecked: EnTT config, flecs addons and debug, GLM force flags, `DT_POLYREF64`, spdlog fmt mode, Lua C vs C++, MSVC CRT (Jolt defaults to the static CRT) [9][32][34][37][51][54][67]. | Single source of truth for defines, propagated to every target as CMake INTERFACE definitions. Emit the self-check calls in generated init code. | **M** |
| **Determinism** | Box2D on by default, flag-sensitive [21]. Jolt opt-in define, same defines everywhere [10]. PhysX enhanced flag only [29]. netcode and yojimbo strict FP [79][80]. Box2D: "determinism does not mean your application will be deterministic" [21]. | A composition-level "determinism profile": compiler FP flags, required library defines, banned APIs (broadphase-order queries, listener ordering), and replay tests. | **P**: flags are **M**; app-level discipline is **G**. |
| **Version coupling** (extra) | Tracy client/server protocol [64]. GNS protoc/runtime protobuf [77]. Box2D pins Tracy 0.14.1 [23]. sol2 vs Lua 5.5 [71]. Jolt `CreateJobs` on `master` [6]. | Lockfile with cross-library constraints. CI per pinned tuple. | **M** |
| **Platform / NDA overlays** (extra) | Jolt `PlatformBlue.h` [2]. GNS private console code [77]. LuaJIT console build scripts [69]. Tracy `TRACY_PLATFORM_HEADER` [64]. | A private-registry overlay that injects headers or patches without publishing them. | **P** |

---

## Leads checked

| Lead | Verdict | Source |
|---|---|---|
| Simulation candidates: Jolt, Box2D v3, Recast/Detour, flecs/EnTT | **Confirmed**, with caveats. All are maintained in 2026. Box2D 3.2 breaks the API; EnTT 4 requires C++20; Recast's last release was 2023 but commits continue. Box3D is a new 3D option. | [1][17][18][25][32][33][36] |
| Glue candidates: Tracy, Lua + sol2, GNS/ENet | **Corrected in part.** Tracy is confirmed (active, but its version is locked across client and server). **sol2 is stalled and does not support Lua 5.5.** Lua 5.4 + sol2 or Luau are safer pins. GNS is confirmed active but dependency-heavy; ENet is confirmed stable with low activity. | [62][64][71][77][78] |
| "Godot 4.6 made Jolt the default physics engine" | **Confirmed, with a nuance**: Jolt is the default **for new 3D projects**; existing projects keep their setting. This rests on search summaries of the official 4.6 release page and secondary coverage. The Godot 4.6 source contains the Jolt module [14]. | [16][14] |
| "The hard part is the seams: allocator ownership, job systems and threading, math types, handles, asset lifecycle, serialization, build flags" | **Confirmed**. Each seam appears concretely in headers and in Godot's production glue. Determinism and version coupling should be added. Asset lifecycle appears here only as Detour tile ownership and Jolt cooked data. | [10][14][15][18][32] |

---

## Implications for Automap (inference)

All points in this section are inference.

1. **The manifest should model hooks, not just dependencies.** Per library, record:
   - `hooks.allocator`: style and signature;
   - `hooks.assert` and `hooks.log`;
   - `jobs.interface`: kind such as `virtual-jobsystem`, `enqueue-finish`, `task-threads` or `dispatcher`;
   - `handles`: type and bit layout;
   - `lifecycle`: init and shutdown steps plus constraints such as "exactly once" or "before first call";
   - `abi_defines`: each entry with a must-match flag and a verification method: `runtime-check` (Jolt, ImGui, PhysX, Box2D), `link-mangle` (Tracy), `inline-namespace` (unordered_dense) or `none`.
2. **Much seam data can be extracted from headers semi-automatically** (hook typedefs, `option()` lists, config headers), as this research did with grep. A "seam extractor" plus agent and human curation looks cheaper than hand-authoring.
3. **Known-good pairings must be version-pinned tuples backed by CI evidence.** Box2D 3.1→3.2, Lua 5.4→5.5 and Jolt 5.6→master each break glue code that would otherwise look correct.
4. **A `doctor` command (CLI or MCP) is feasible:**
   - scan `compile_commands.json` for define consistency;
   - emit and run self-checks (`VerifyJoltVersionID`, `IMGUI_CHECKVERSION`, `b2IsDoublePrecision`);
   - check that Tracy versions are locked;
   - check FP flags for a determinism profile.
5. **Job systems justify a neutral adapter layer** (an Automap "scheduler contract") so recipes scale as N+M rather than N×M. Godot's `JoltJobSystem` is the template [15].
6. **The private-registry tier has precedents:** Jolt's NDA header, GNS's private console code, and Tracy's platform header. These argue for an overlay mechanism: a public recipe plus private patch or header packs.
7. **Ingest existing machine-readable artifacts** rather than re-describing them: Box2D's API JSON, Jolt's *SBS* changelog markers, release Atom feeds, and Tracy's MCP server for performance verification inside agent loops.
8. **Health metadata must be refreshed continuously.** sol2's stall and Tracy's coupling show how quickly a "known-good" choice goes stale.

## Open questions

1. Live++ and Superluminal: current platform lists (PS5, Xbox, Linux) and licensing terms. Not verified because the hosts were blocked.
2. Will Jolt 5.7 ship the pure-virtual `CreateJobs`, and will Godot's bridge adopt it?
3. Will sol2 merge PR #1723 (Lua 5.5)? Is LuaBridge3 3.0 final close?
4. PhysX: the root license (BSD-3) and SDK license (Apache-2.0) disagree. Are GPU binaries still downloaded, or must they be built? Is any console path public?
5. Is Box2D 3.2's thread-count independence preserved under every external scheduler? This needs a cross-scheduler replay test.
6. Does a maintained flecs + Jolt or EnTT + Jolt reference integration exist that could serve as a recipe seed? None was found.
7. How should NDA overlays be versioned and verified without exposing headers (Jolt `PlatformBlue.h`)?
8. Status of std::hive standardization (plf::colony), and Taskflow 4 integration hooks for external pools. Both unverified because the search budget was exhausted.

---

## Sources

All accessed 2026-10-29. "Feed" means the GitHub `releases.atom` feed of the repository. "master" or "main" means the default branch as of the access date.

1. Jolt Physics releases feed — https://github.com/jrouwe/JoltPhysics/releases.atom — 5.6.0 (2026-07-11)
2. Jolt README — https://github.com/jrouwe/JoltPhysics/blob/master/README.md — master
3. Jolt `Core.h` — https://github.com/jrouwe/JoltPhysics/blob/master/Jolt/Core/Core.h — 5.6.1-dev
4. Jolt `Memory.h` — https://github.com/jrouwe/JoltPhysics/blob/master/Jolt/Core/Memory.h
5. Jolt `RegisterTypes.h` — https://github.com/jrouwe/JoltPhysics/blob/master/Jolt/RegisterTypes.h
6. Jolt `JobSystem.h`, `JobSystemThreadPool.h`, `JobSystemWithBarrier.h` — https://github.com/jrouwe/JoltPhysics/tree/master/Jolt/Core
7. Jolt `IssueReporting.h`, `Profiler.h` — https://github.com/jrouwe/JoltPhysics/tree/master/Jolt/Core
8. Jolt `BodyID.h`, `Math/Real.h` — https://github.com/jrouwe/JoltPhysics/tree/master/Jolt
9. Jolt `Build/CMakeLists.txt` — https://github.com/jrouwe/JoltPhysics/blob/master/Build/CMakeLists.txt
10. Jolt `Docs/Architecture.md` (Deterministic Simulation; Saving Shapes) — https://github.com/jrouwe/JoltPhysics/blob/master/Docs/Architecture.md
11. Jolt `Docs/APIChanges.md` — https://github.com/jrouwe/JoltPhysics/blob/master/Docs/APIChanges.md — entries to 2026-10-22
12. Jolt `HelloWorld/HelloWorld.cpp` — https://github.com/jrouwe/JoltPhysics/blob/master/HelloWorld/HelloWorld.cpp
13. Jolt commit history — https://github.com/jrouwe/JoltPhysics/commits/master — to 2026-10-26
14. Godot `modules/jolt_physics/jolt_globals.cpp` — https://github.com/godotengine/godot/blob/4.6-stable/modules/jolt_physics/jolt_globals.cpp — 4.6-stable
15. Godot `modules/jolt_physics/spaces/jolt_job_system.{h,cpp}` — https://github.com/godotengine/godot/tree/4.6-stable/modules/jolt_physics/spaces — 4.6-stable
16. Godot 4.6 release page — https://godotengine.org/releases/4.6/ — via web-search summary; corroborated by https://gamefromscratch.com/godot-4-6-released/ and https://www.gdquest.com/library/godot_4_6_workflow_changes
17. Box2D releases feed — https://github.com/erincatto/box2d/releases.atom — v3.2.0 (2026-10-19)
18. Box2D v3.2 release notes — https://github.com/erincatto/box2d/blob/main/docs/release_notes_v320.md
19. Box2D `base.h`, `config.h` — https://github.com/erincatto/box2d/tree/main/include/box2d — 3.2.0
20. Box2D `types.h`, `id.h`, `math_functions.h` — https://github.com/erincatto/box2d/tree/main/include/box2d
21. Box2D `docs/simulation.md` (Determinism) — https://github.com/erincatto/box2d/blob/main/docs/simulation.md
22. Box2D v3.1 release notes — https://github.com/erincatto/box2d/blob/main/docs/release_notes_v310.md
23. Box2D `README.md`, `CMakeLists.txt`, `src/CMakeLists.txt` — https://github.com/erincatto/box2d
24. Box2D commit history — https://github.com/erincatto/box2d/commits/main — to 2026-10-20
25. Box3D README and commits — https://github.com/erincatto/box3d — to 2026-10-28
26. PhysX releases feed — https://github.com/NVIDIA-Omniverse/PhysX/releases.atom — 110.3 / SDK 5.9.1 (2026-10-27)
27. PhysX root `README.md`, `LICENSE.md` — https://github.com/NVIDIA-Omniverse/PhysX
28. PhysX SDK `physx/README.md`, `physx/version.txt` (5.12.0), `physx/CHANGELOG.md` — https://github.com/NVIDIA-Omniverse/PhysX/tree/main/physx
29. PhysX headers `PxFoundation.h`, `PxCpuDispatcher.h`, `PxPhysicsVersion.h`, `PxSceneDesc.h`, `PxBase.h` — https://github.com/NVIDIA-Omniverse/PhysX/tree/main/physx/include
30. "NVIDIA open-sources PhysX's GPU simulation code" — https://www.cgchannel.com/2025/04/nvidia-open-sources-physxs-gpu-simulation-code (2025-04-22); also https://heise.de/-10341975 — both via web-search summary
31. Bullet3 releases feed, `LICENSE.txt`, `btScalar.h`, `btAlignedAllocator.h` — https://github.com/bulletphysics/bullet3 — 3.24 (2022-04-26)
32. recastnavigation feed, README, CMakeLists, `RecastAlloc.h`, `RecastAssert.h`, `Recast.h`, `DetourAlloc.h`, `DetourNavMesh.h`, `DebugDraw.h`, commits — https://github.com/recastnavigation/recastnavigation — 1.6.0 (2023-05-21); commits to 2026-02-27
33. flecs releases feed — https://github.com/SanderMertens/flecs/releases.atom — v4.1.6 (2026-06-29)
34. flecs `flecs.h`, `os_api.h`, `addons/pipeline.h`, `private/api_defines.h`, README — https://github.com/SanderMertens/flecs — 4.1.6
35. flecs commit history — https://github.com/SanderMertens/flecs/commits/master — 2026-10-29
36. EnTT releases feed — https://github.com/skypjack/entt/releases.atom — v4.0.0 (2026-07-23)
37. EnTT `config/config.h`, `config/version.h`, `entity/fwd.hpp`, `entity/entity.hpp`, README — https://github.com/skypjack/entt — master
38. EnTT `wip` commit history — https://github.com/skypjack/entt/commits/wip — 2026-10-27
39. gaia-ecs releases feed — https://github.com/richardbiely/gaia-ecs/releases.atom — v1.0.0 (2026-08-22)
40. mimalloc releases feed and `dev3` readme — https://github.com/microsoft/mimalloc — v3.6.0 (2026-10-23)
41. rpmalloc releases feed and README — https://github.com/mjansson/rpmalloc — 2.0.1 (2026-07-15)
42. TLSF README — https://github.com/mattconte/tlsf — v3.1 (2016)
43. enkiTS releases feed, README, `TaskScheduler.h` — https://github.com/dougbinks/enkiTS — v1.12 (2026-07-04)
44. marl commits, README, LICENSE — https://github.com/google/marl — last commit 2026-04-27
45. Taskflow releases feed, README, LICENSE — https://github.com/taskflow/taskflow — v4.1.0 (2026-06-20)
46. oneTBB releases feed, README — https://github.com/uxlfoundation/oneTBB — 2023.2.0 (2026-10-28)
47. EASTL releases feed, README, `doc/FAQ.md` — https://github.com/electronicarts/EASTL — 3.27.01 (2025-11-15)
48. unordered_dense releases feed, v5.0.0 notes, README, `unordered_dense.h` — https://github.com/martinus/unordered_dense — v5.3.3 (2026-10-22)
49. parallel-hashmap and gtl releases feeds and LICENSE — https://github.com/greg7mdp/parallel-hashmap (2.0.0, 2025-01-21); https://github.com/greg7mdp/gtl (1.2.0, 2024-09-21)
50. plf_colony commits, README, LICENSE.md — https://github.com/mattreecebentley/plf_colony — v7.7.54 (2026-10-23)
51. spdlog releases feed, CMakeLists.txt, README — https://github.com/gabime/spdlog — 1.17.0 (2026-01-04)
52. {fmt} releases feed, LICENSE — https://github.com/fmtlib/fmt — 12.2.0 (2026-06-16)
53. quill releases feed, README, `Backend.h` — https://github.com/odygrd/quill — v13.1.0 (2026-10-17)
54. GLM releases feed, `manual.md` — https://github.com/g-truc/glm — 1.0.3 (2026-01-12)
55. HandmadeMath releases feed, README — https://github.com/HandmadeMath/HandmadeMath — v2.0.0 (2023-02-20)
56. DirectXMath releases feed, README, LICENSE — https://github.com/microsoft/DirectXMath — June 2026 (2026-07-15)
57. cglm releases feed, README, LICENSE — https://github.com/recp/cglm — v0.9.7 (2026-10-25)
58. rtm releases feed, README — https://github.com/nfrechette/rtm — 2.3.1 (2024-03-16)
59. linalg README — https://github.com/sgorsten/linalg — 2.2
60. Dear ImGui `imgui.h`, `imconfig.h` — https://github.com/ocornut/imgui — master, 1.93.1 WIP
61. ozz-animation releases feed, README, `allocator.h`, `vec_float.h` — https://github.com/guillaumeblanc/ozz-animation — 0.17.0 (2026-08-01)
62. Tracy releases feed — https://github.com/wolfpld/tracy/releases.atom — 0.14.1 (2026-08-22)
63. Tracy NEWS — https://github.com/wolfpld/tracy/blob/master/NEWS — through 0.14.1 plus unreleased section
64. Tracy manual — https://github.com/wolfpld/tracy/blob/master/manual/tracy.tex — master
65. Tracy `TracyProtocol.hpp` (ProtocolVersion 83), `TracyVersion.hpp` (0.14.2), `CMakeLists.txt`, README, LICENSE — https://github.com/wolfpld/tracy
66. Optick releases feed, commits, README, LICENSE — https://github.com/bombomby/optick — v1.4.0 (2022-05-15)
67. Lua official GitHub mirror: tags, `lua.h`, `manual/manual.of`, `ldo.c` — https://github.com/lua/lua — v5.5.1 tag (2026-07-23)
68. Lua version history — https://lua.org/versions.html — via web-search summary (host blocked); 5.5.0 released 2025-12-22
69. LuaJIT `v2.1` branch commits and branches, README, `doc/install.html` — https://github.com/LuaJIT/LuaJIT — commits to 2026-10-21
70. Luau releases feed, README, `VM/include/lua.h`, LICENSE.txt — https://github.com/luau-lang/luau — 0.744 (2026-10-23)
71. sol2 releases feed, commits, README, issues #1747 and #1754 — https://github.com/ThePhD/sol2 — v3.5.0 (2025-02-25)
72. LuaBridge3 releases feed — https://github.com/kunitoki/LuaBridge3/releases.atom — 3.0-rc13 (2026-10-18)
73. AngelScript `angelscript.h` (mirror) — https://github.com/codecat/angelscript-mirror — 2.38.0; release date via web-search summary of https://angelcode.com/angelscript/ (host blocked)
74. Wren releases feed, commits, README — https://github.com/wren-lang/wren — 0.4.0 (2021-04-09)
75. cr.h README and commits — https://github.com/fungos/cr — commits to 2026-06-23
76. jet-live README and commits — https://github.com/ddovod/jet-live — commits to 2026-03-02
77. GameNetworkingSockets releases feed, commits, README, BUILDING.md, CMakeLists.txt, `steamnetworkingsockets.h`, `steamnetworkingtypes.h`, LICENSE — https://github.com/ValveSoftware/GameNetworkingSockets — v1.6.1 (2026-10-03)
78. ENet commits, `enet.h`, `callbacks.h`, LICENSE — https://github.com/lsalzman/enet — 1.3.18; commit 2026-06-23
79. netcode releases feed, `netcode.h`, README — https://github.com/mas-bandwidth/netcode — 1.4.8 (2026-09-13)
80. yojimbo releases feed, README, `yojimbo.h` — https://github.com/mas-bandwidth/yojimbo — 1.14.0 (2026-10-16)
