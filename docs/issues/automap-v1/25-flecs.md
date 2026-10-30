# 25: flecs

**What to build:** flecs fills the ECS slot. Its OS API routes allocation, logging and abort to the facades, its task threads run on the jobs interface, its perf-trace hooks reach the profiler, and its handle and serialization facts are recorded.

**Blocked by:** [21: enkiTS and the jobs slot](21-enkits-and-the-jobs-slot.md), [23: spdlog and Tracy as diagnostics backends](23-spdlog-and-tracy.md)

**Status:** ready-for-agent

- [ ] `automap extend flecs` on a composition with enkiTS and Tracy ends probed on all three platforms.
- [ ] Every new probe fails its seeded violation.
