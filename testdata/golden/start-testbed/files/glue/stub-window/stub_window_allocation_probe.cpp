#ifdef AUTOMAP_PROBE
#include <stub_window.h>

#include "automap/alloc.h"
#include "automap/probe.h"

// allocation.routed-through-facade: stub-window's hooks are the facade's,
// and the facade served stub-window during the smoke run.
namespace {
void am_probe_stub_window_allocation() {
    stub_window_alloc_fn alloc = nullptr;
    stub_window_free_fn release = nullptr;
    stub_window_get_allocator(&alloc, &release);
    bool routed = alloc != nullptr && release != nullptr && am::alloc::routed_count("stub-window") > 0;
    am::probe::check("allocation.routed-through-facade", "stub-window", routed, "ALLOC_NOT_ROUTED(stub-window)");
}
const am::probe::Registration registration("stub-window--allocation", &am_probe_stub_window_allocation);

// allocation.hooks-before-first-allocation: after shutdown, stub-window has
// released nothing the facade didn't allocate.
void am_probe_stub_window_hooks_early() {
    bool early = !am::alloc::released_foreign_block("stub-window");
    am::probe::check("allocation.hooks-before-first-allocation", "stub-window", early, "ALLOC_HOOK_LATE(stub-window)");
}
const am::probe::Registration after_shutdown("stub-window--allocation", &am_probe_stub_window_hooks_early, am::probe::When::after_shutdown);
} // namespace
#endif
