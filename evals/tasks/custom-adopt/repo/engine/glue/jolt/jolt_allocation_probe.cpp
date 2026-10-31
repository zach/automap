// Probes for Jolt's binding to the allocation contract.
// SPDX-License-Identifier: MIT-0
#ifdef AUTOMAP_PROBE
#include <Jolt/Jolt.h>

#include "automap/alloc.h"
#include "automap/probe.h"

void* am_jolt_allocate(size_t size);
void* am_jolt_reallocate(void* block, size_t, size_t size);
void am_jolt_free(void* block);
void* am_jolt_aligned_allocate(size_t size, size_t alignment);
void am_jolt_aligned_free(void* block);

namespace {

// allocation.routed-through-facade: Jolt's allocation functions are the
// facade's, and the facade served Jolt during the run.
void am_probe_jolt_allocation(am::Context&) {
    bool routed = JPH::Allocate == &am_jolt_allocate && JPH::Reallocate == &am_jolt_reallocate &&
                  JPH::Free == &am_jolt_free && JPH::AlignedAllocate == &am_jolt_aligned_allocate &&
                  JPH::AlignedFree == &am_jolt_aligned_free && am::alloc::routed_count("jolt") > 0;
    am::probe::check("allocation.routed-through-facade", "jolt", routed, "ALLOC_NOT_ROUTED(jolt)");
}
const am::probe::Registration routed("jolt--allocation", &am_probe_jolt_allocation);

// allocation.hooks-before-first-allocation: after Jolt shuts down, it has
// released nothing the facade didn't allocate.
void am_probe_jolt_hooks_early(am::Context&) {
    bool early = !am::alloc::released_foreign_block("jolt");
    am::probe::check("allocation.hooks-before-first-allocation", "jolt", early, "ALLOC_HOOK_LATE(jolt)");
}
const am::probe::Registration after_shutdown("jolt--allocation", &am_probe_jolt_hooks_early, am::probe::When::after_shutdown);

} // namespace
#endif
