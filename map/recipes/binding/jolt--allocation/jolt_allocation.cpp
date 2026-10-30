// Jolt's binding to the allocation contract.
// SPDX-License-Identifier: MIT-0
#include <Jolt/Jolt.h>

#include "automap/alloc.h"
#include "automap/context.h"
#include "automap/lifecycle.h"

void* am_jolt_allocate(size_t size) { return am::alloc::allocate("jolt", size); }
void* am_jolt_reallocate(void* block, size_t, size_t size) { return am::alloc::reallocate("jolt", block, size); }
void am_jolt_free(void* block) { am::alloc::release("jolt", block); }
void* am_jolt_aligned_allocate(size_t size, size_t alignment) { return am::alloc::allocate_aligned("jolt", size, alignment); }
void am_jolt_aligned_free(void* block) { am::alloc::release_aligned("jolt", block); }

// Routes Jolt's allocations through the allocation facade. Runs in the
// memory phase, before JPH::Factory and JPH::RegisterTypes().
bool am_bind_jolt_allocation(am::Context&) {
    JPH::Allocate = &am_jolt_allocate;
    JPH::Reallocate = &am_jolt_reallocate;
    JPH::Free = &am_jolt_free;
    JPH::AlignedAllocate = &am_jolt_aligned_allocate;
    JPH::AlignedFree = &am_jolt_aligned_free;
    am::lifecycle::init_step("jolt:allocation-hooks");
    return true;
}
