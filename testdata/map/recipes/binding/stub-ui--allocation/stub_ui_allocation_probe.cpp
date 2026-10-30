#ifdef AUTOMAP_PROBE
#include "automap/alloc.h"
#include "automap/probe.h"

// allocation.routed-through-facade: the facade served stub-ui during the
// smoke run.
void am_probe_stub_ui_allocation() {
    bool routed = am::alloc::routed_count("stub-ui") > 0;
    am::probe::check("allocation.routed-through-facade", "stub-ui", routed, "ALLOC_NOT_ROUTED(stub-ui)");
}
#endif
