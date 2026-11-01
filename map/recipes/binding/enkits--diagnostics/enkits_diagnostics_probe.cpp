// Probes for enkiTS's binding to the diagnostics contract.
// SPDX-License-Identifier: MIT-0
#ifdef AUTOMAP_PROBE
#include <chrono>
#include <thread>

#include "automap/diag.h"
#include "automap/probe.h"
#include "enkits_scheduler.h"

void am_enkits_thread_start(uint32_t threadnum);

namespace {

// diagnostics.routed-through-facade: the scheduler runs with the facade's
// profiler callbacks, and its task threads reported to the facade. A task
// thread reports as it starts, and enkiTS's task threads may still be
// starting (on Windows they take longer), so the report is awaited for up
// to two seconds.
void am_probe_enkits_diagnostics(am::Context&) {
    enki::TaskSchedulerConfig config = am_enkits_scheduler().GetConfig();
    bool hooked = config.profilerCallbacks.threadStart == &am_enkits_thread_start;
    const auto deadline = std::chrono::steady_clock::now() + std::chrono::seconds(2);
    while (hooked && am::diag::message_count("enkits") == 0 && std::chrono::steady_clock::now() < deadline) {
        std::this_thread::sleep_for(std::chrono::milliseconds(5));
    }
    bool routed = hooked && am::diag::message_count("enkits") > 0;
    am::probe::check("diagnostics.routed-through-facade", "enkits", routed, "DIAG_NOT_ROUTED(enkits)");
}
const am::probe::Registration registration("enkits--diagnostics", &am_probe_enkits_diagnostics);

} // namespace
#endif
