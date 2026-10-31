// flecs's setup glue.
// SPDX-License-Identifier: MIT-0
#include "flecs_setup.h"

#include "automap/jobs.h"
#include "automap/lifecycle.h"

namespace {

ecs_world_t* world = nullptr;

} // namespace

ecs_os_api_t& am_flecs_os_api() {
    static ecs_os_api_t* api = [] {
        ecs_os_set_api_defaults();
        return new ecs_os_api_t(ecs_os_api);
    }();
    return *api;
}

ecs_world_t* am_flecs_world() { return world; }

bool am_flecs_setup_init(am::Context& ctx) {
    ecs_os_set_api(&am_flecs_os_api());
    world = ecs_init();
    if (world == nullptr) return false;
    // flecs's worker tasks block until the main thread releases them, so
    // every one must have a jobs worker to itself: the main thread runs the
    // first stage, and there are no more stages than workers.
    ecs_set_task_threads(world, am::jobs::workers());
    ctx.flecs.world = world;
    am::lifecycle::init_step("flecs:ecs_init+ecs_set_task_threads");
    return true;
}

void am_flecs_setup_shutdown(am::Context& ctx) {
    ecs_fini(world);
    world = nullptr;
    ctx.flecs.world = nullptr;
    am::lifecycle::shutdown_step("flecs:ecs_fini");
}
