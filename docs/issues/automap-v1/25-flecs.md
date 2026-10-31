# 25: flecs

**What to build:** flecs fills the ECS slot. Its OS API routes allocation, logging and abort to the facades, its task threads run on the jobs interface, its perf-trace hooks reach the profiler, and its handle and serialization facts are recorded.

**Blocked by:** [21: enkiTS and the jobs slot](21-enkits-and-the-jobs-slot.md), [23: spdlog and Tracy as diagnostics backends](23-spdlog-and-tracy.md)

**Status:** done

- [x] `automap extend flecs` on a composition with enkiTS and Tracy ends probed on all three platforms.
- [x] Every new probe fails its seeded violation.

## Comments

- flecs 4.1.6 is the curated port at the map's baseline. Setup glue copies flecs's default OS API for the bindings to fill (allocation in the memory phase; log, abort and perf trace in the diagnostics phase; the task API in the jobs phase), installs it, makes the world and calls `ecs_set_task_threads(world, am::jobs::workers())`.
- flecs's worker tasks block until the main thread releases them, so setup glue never asks for more stages than the jobs interface has workers; the jobs binding's notes say so, against the jobs contract's guidance on blocking workers.
- The diagnostics binding installs flecs's perf-trace hooks as facade zones, but the port builds flecs without `FLECS_PERF_TRACE`, so flecs calls none; a build with it would come from the Automap registry.
- flecs's handle and serialization facts wait for those contracts (ticket 29), which already plans flecs's handle and serialization probes.
- `TestTheTestbedExtendsWithEnkiTSTracyAndFlecsProbed` runs flecs's tasks on enkiTS with Tracy profiling, ThreadSanitizer included.
