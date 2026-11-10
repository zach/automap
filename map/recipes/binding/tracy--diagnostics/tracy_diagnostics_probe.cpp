// Probes for Tracy's binding to the diagnostics contract.
// SPDX-License-Identifier: MIT-0
#ifdef AUTOMAP_PROBE
#include <cstddef>

#include <tracy/Tracy.hpp>

#include "automap/diag.h"
#include "automap/probe.h"

std::size_t am_tracy_open_zones();

namespace {

// diagnostics.routed-through-facade: a zone opened through the facade opens
// a Tracy zone on this thread. Without on-demand, the client holds what it
// records while no profiler is connected, so the zone's begin and end
// events wait in this thread's queue. On demand, the client records only
// while a profiler is connected: with none, the queue gains nothing.
void am_probe_tracy_diagnostics(am::Context&) {
    std::size_t queued = tracy::GetToken()->size_approx();
    std::size_t open = am_tracy_open_zones();
    bool opened = false;
    {
        am::diag::Zone zone("automap probe: profiling routing");
        opened = am_tracy_open_zones() == open + 1;
    }
    std::size_t now = tracy::GetToken()->size_approx();
#ifdef TRACY_ON_DEMAND
    bool recorded = now <= queued;
#else
    bool recorded = now >= queued + 2;
#endif
    bool routed = TracyIsStarted && opened && recorded;
    am::probe::check("diagnostics.routed-through-facade", "tracy", routed, "DIAG_NOT_ROUTED(tracy)");
}
const am::probe::Registration registration("tracy--diagnostics", &am_probe_tracy_diagnostics);

} // namespace
#endif
