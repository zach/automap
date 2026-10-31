// Probes for flecs's binding to the jobs contract.
// SPDX-License-Identifier: MIT-0
#ifdef AUTOMAP_PROBE
#include "automap/jobs.h"
#include "automap/probe.h"
#include "flecs_setup.h"

namespace {

// jobs.routed-through-interface: the world runs its worker stages as tasks,
// and a frame's tasks are submitted through the jobs interface.
void am_probe_flecs_jobs(am::Context&) {
    ecs_world_t* world = am_flecs_world();
    auto before = am::jobs::submitted_count("flecs");
    ecs_progress(world, 0);
    bool routed = ecs_using_task_threads(world) &&
                  (ecs_get_stage_count(world) <= 1 || am::jobs::submitted_count("flecs") > before);
    am::probe::check("jobs.routed-through-interface", "flecs", routed, "JOBS_NOT_ROUTED(flecs)");
}
const am::probe::Registration registration("flecs--jobs", &am_probe_flecs_jobs);

} // namespace
#endif
