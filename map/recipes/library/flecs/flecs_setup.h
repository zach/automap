// flecs's setup glue: the composition's one world, and the OS API flecs
// runs on, which the bindings fill in before the world is made.
// SPDX-License-Identifier: MIT-0
#pragma once
#include <flecs.h>

#include "automap/context.h"

// The OS API table flecs is given: flecs's defaults, with the allocation
// binding's hooks set in the memory phase, the diagnostics binding's in the
// diagnostics phase and the jobs binding's in the jobs phase.
ecs_os_api_t& am_flecs_os_api();

// The composition's world, once the systems phase has made it.
ecs_world_t* am_flecs_world();

// Installs the OS API, makes the world, and runs its pipeline's worker
// stages as tasks: one stage per jobs worker, the main thread's included.
bool am_flecs_setup_init(am::Context& ctx);

// Destroys the world.
void am_flecs_setup_shutdown(am::Context& ctx);
