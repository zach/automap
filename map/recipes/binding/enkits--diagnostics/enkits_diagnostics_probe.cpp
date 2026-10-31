// Probes for enkiTS's binding to the diagnostics contract.
// SPDX-License-Identifier: MIT-0
#ifdef AUTOMAP_PROBE
#include "automap/diag.h"
#include "automap/probe.h"
#include "enkits_scheduler.h"

void am_enkits_thread_start(uint32_t threadnum);

namespace {

// diagnostics.routed-through-facade: the scheduler runs with the facade's
// profiler callbacks, and its task threads reported to the facade.
void am_probe_enkits_diagnostics(am::Context&) {
    enki::TaskSchedulerConfig config = am_enkits_scheduler().GetConfig();
    bool routed = config.profilerCallbacks.threadStart == &am_enkits_thread_start && am::diag::message_count("enkits") > 0;
    am::probe::check("diagnostics.routed-through-facade", "enkits", routed, "DIAG_NOT_ROUTED(enkits)");
}
const am::probe::Registration registration("enkits--diagnostics", &am_probe_enkits_diagnostics);

} // namespace
#endif
