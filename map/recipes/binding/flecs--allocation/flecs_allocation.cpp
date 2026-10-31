// flecs's binding to the allocation contract.
// SPDX-License-Identifier: MIT-0
#include "automap/alloc.h"
#include "automap/context.h"
#include "automap/lifecycle.h"
#include "flecs_setup.h"

void* am_flecs_malloc(ecs_size_t size) { return am::alloc::allocate("flecs", static_cast<std::size_t>(size)); }

void* am_flecs_calloc(ecs_size_t size) { return am::alloc::allocate_zeroed("flecs", 1, static_cast<std::size_t>(size)); }

void* am_flecs_realloc(void* block, ecs_size_t size) {
    return am::alloc::reallocate("flecs", block, static_cast<std::size_t>(size));
}

void am_flecs_free(void* block) { am::alloc::release("flecs", block); }

// Routes flecs's allocations through the allocation facade. Runs in the
// memory phase, before setup glue installs the OS API and makes the world.
bool am_bind_flecs_allocation(am::Context&) {
    ecs_os_api_t& api = am_flecs_os_api();
    api.malloc_ = &am_flecs_malloc;
    api.calloc_ = &am_flecs_calloc;
    api.realloc_ = &am_flecs_realloc;
    api.free_ = &am_flecs_free;
    am::lifecycle::init_step("flecs:os_api.memory");
    return true;
}
