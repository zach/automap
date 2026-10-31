// enkiTS's binding to the allocation contract.
// SPDX-License-Identifier: MIT-0
#include "automap/alloc.h"
#include "automap/context.h"
#include "automap/lifecycle.h"
#include "enkits_scheduler.h"

void* am_enkits_alloc(size_t align, size_t size, void*, const char*, int) {
    return am::alloc::allocate_aligned("enkits", size, align);
}

void am_enkits_free(void* block, size_t, void*, const char*, int) { am::alloc::release_aligned("enkits", block); }

// Routes enkiTS's allocations through the allocation facade. Runs in the
// memory phase, before the jobs phase initializes the scheduler.
bool am_bind_enkits_allocation(am::Context&) {
    am_enkits_config().customAllocator.alloc = &am_enkits_alloc;
    am_enkits_config().customAllocator.free = &am_enkits_free;
    am::lifecycle::init_step("enkits:customAllocator");
    return true;
}
