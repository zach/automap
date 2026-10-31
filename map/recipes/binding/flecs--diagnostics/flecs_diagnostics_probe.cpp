// Probes for flecs's binding to the diagnostics contract.
// SPDX-License-Identifier: MIT-0
#ifdef AUTOMAP_PROBE
#include "automap/diag.h"
#include "automap/probe.h"
#include "flecs_setup.h"

void am_flecs_log(int32_t level, const char*, int32_t, const char* message);
void am_flecs_abort();
void am_flecs_perf_trace_push(const char*, size_t, const char* name);

namespace {

// diagnostics.routed-through-facade: flecs runs with the facade's log,
// abort and perf-trace hooks, and a warning flecs logs reaches the facade.
void am_probe_flecs_diagnostics(am::Context&) {
    auto before = am::diag::message_count("flecs");
    ecs_warn("automap probe: diagnostics routing");
    bool routed = ecs_os_api.log_ == &am_flecs_log && ecs_os_api.abort_ == &am_flecs_abort &&
                  ecs_os_api.perf_trace_push_ == &am_flecs_perf_trace_push && am::diag::message_count("flecs") > before;
    am::probe::check("diagnostics.routed-through-facade", "flecs", routed, "DIAG_NOT_ROUTED(flecs)");
}
const am::probe::Registration registration("flecs--diagnostics", &am_probe_flecs_diagnostics);

} // namespace
#endif
