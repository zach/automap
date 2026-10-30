#ifdef AUTOMAP_PROBE
#include "automap/alloc.h"
#include "automap/probe.h"

// allocation.routed-through-facade: the facade served stub-ui during the
// smoke run.
namespace {
void am_probe_stub_ui_allocation(am::Context&) {
    bool routed = am::alloc::routed_count("stub-ui") > 0;
    am::probe::check("allocation.routed-through-facade", "stub-ui", routed, "ALLOC_NOT_ROUTED(stub-ui)");
}
const am::probe::Registration registration("stub-ui--allocation", &am_probe_stub_ui_allocation);

// allocation.hooks-before-first-allocation: after shutdown, stub-ui has
// released nothing the facade didn't allocate.
void am_probe_stub_ui_hooks_early(am::Context&) {
    bool early = !am::alloc::released_foreign_block("stub-ui");
    am::probe::check("allocation.hooks-before-first-allocation", "stub-ui", early, "ALLOC_HOOK_LATE(stub-ui)");
}
const am::probe::Registration after_shutdown("stub-ui--allocation", &am_probe_stub_ui_hooks_early, am::probe::When::after_shutdown);
} // namespace
#endif
