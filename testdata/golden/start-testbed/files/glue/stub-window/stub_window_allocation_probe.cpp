#ifdef AUTOMAP_PROBE
#include <stub_window.h>

#include "automap/alloc.h"
#include "automap/probe.h"

// allocation.routed-through-facade: stub-window's hooks are the facade's,
// and the facade served stub-window during the smoke run.
void am_probe_stub_window_allocation() {
    stub_window_alloc_fn alloc = nullptr;
    stub_window_free_fn release = nullptr;
    stub_window_get_allocator(&alloc, &release);
    bool routed = alloc != nullptr && release != nullptr && am::alloc::routed_count("stub-window") > 0;
    am::probe::check("allocation.routed-through-facade", "stub-window", routed, "ALLOC_NOT_ROUTED(stub-window)");
}
#endif
