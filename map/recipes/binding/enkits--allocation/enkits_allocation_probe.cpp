// Probes for enkiTS's binding to the allocation contract.
// SPDX-License-Identifier: MIT-0
#ifdef AUTOMAP_PROBE
#include "automap/alloc.h"
#include "automap/probe.h"
#include "enkits_scheduler.h"

void* am_enkits_alloc(size_t align, size_t size, void*, const char*, int);
void am_enkits_free(void* block, size_t, void*, const char*, int);

namespace {

// allocation.routed-through-facade: the scheduler runs with the facade's
// allocator, and the facade served enkiTS during the run.
void am_probe_enkits_allocation(am::Context&) {
    enki::TaskSchedulerConfig config = am_enkits_scheduler().GetConfig();
    bool routed = config.customAllocator.alloc == &am_enkits_alloc && config.customAllocator.free == &am_enkits_free &&
                  am::alloc::routed_count("enkits") > 0;
    am::probe::check("allocation.routed-through-facade", "enkits", routed, "ALLOC_NOT_ROUTED(enkits)");
}
const am::probe::Registration routed("enkits--allocation", &am_probe_enkits_allocation);

// allocation.hooks-before-first-allocation: after the scheduler shuts down,
// enkiTS has released nothing the facade didn't allocate.
void am_probe_enkits_hooks_early(am::Context&) {
    bool early = !am::alloc::released_foreign_block("enkits");
    am::probe::check("allocation.hooks-before-first-allocation", "enkits", early, "ALLOC_HOOK_LATE(enkits)");
}
const am::probe::Registration after_shutdown("enkits--allocation", &am_probe_enkits_hooks_early, am::probe::When::after_shutdown);

} // namespace
#endif
