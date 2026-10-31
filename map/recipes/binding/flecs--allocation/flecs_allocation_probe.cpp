// Probes for flecs's binding to the allocation contract.
// SPDX-License-Identifier: MIT-0
#ifdef AUTOMAP_PROBE
#include "automap/alloc.h"
#include "automap/probe.h"
#include "flecs_setup.h"

void* am_flecs_malloc(ecs_size_t size);
void am_flecs_free(void* block);

namespace {

// allocation.routed-through-facade: flecs runs with the facade's
// functions, and the facade served flecs during the run.
void am_probe_flecs_allocation(am::Context&) {
    bool routed = ecs_os_api.malloc_ == &am_flecs_malloc && ecs_os_api.free_ == &am_flecs_free &&
                  am::alloc::routed_count("flecs") > 0;
    am::probe::check("allocation.routed-through-facade", "flecs", routed, "ALLOC_NOT_ROUTED(flecs)");
}
const am::probe::Registration routed("flecs--allocation", &am_probe_flecs_allocation);

// allocation.hooks-before-first-allocation: after the world is gone, flecs
// has released nothing the facade didn't allocate.
void am_probe_flecs_hooks_early(am::Context&) {
    bool early = !am::alloc::released_foreign_block("flecs");
    am::probe::check("allocation.hooks-before-first-allocation", "flecs", early, "ALLOC_HOOK_LATE(flecs)");
}
const am::probe::Registration after_shutdown("flecs--allocation", &am_probe_flecs_hooks_early, am::probe::When::after_shutdown);

} // namespace
#endif
