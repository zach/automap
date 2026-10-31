// flecs's binding to the diagnostics contract: its log, abort and
// perf-trace hooks report to the facade.
// SPDX-License-Identifier: MIT-0
#include <cstdlib>
#include <memory>
#include <vector>

#include "automap/context.h"
#include "automap/diag.h"
#include "automap/lifecycle.h"
#include "flecs_setup.h"

namespace {

// The perf-trace zones open on this thread, innermost last.
thread_local std::vector<std::unique_ptr<am::diag::Zone>> zones;

} // namespace

// flecs's levels: -4 fatal, -3 error, -2 warning, -1 and 0 info, above 0
// debug and trace.
void am_flecs_log(int32_t level, const char*, int32_t, const char* message) {
    am::diag::Level facade = am::diag::Level::trace;
    if (level <= -3) facade = am::diag::Level::error;
    else if (level == -2) facade = am::diag::Level::warn;
    else if (level <= 0) facade = am::diag::Level::info;
    am::diag::log(facade, "flecs", message);
}

// flecs calls abort_ after logging the failed assertion.
void am_flecs_abort() {
    am::diag::assert_failed("flecs", "ecs_os_abort", "flecs aborted", nullptr, 0);
    std::abort();
}

void am_flecs_perf_trace_push(const char*, size_t, const char* name) {
    zones.push_back(std::make_unique<am::diag::Zone>(name));
}

void am_flecs_perf_trace_pop(const char*, size_t, const char*) {
    if (!zones.empty()) zones.pop_back();
}

// Routes flecs's log, abort and perf-trace hooks through the diagnostics
// facade. Runs in the diagnostics phase.
bool am_bind_flecs_diagnostics(am::Context&) {
    ecs_os_api_t& api = am_flecs_os_api();
    api.log_ = &am_flecs_log;
    api.abort_ = &am_flecs_abort;
    api.perf_trace_push_ = &am_flecs_perf_trace_push;
    api.perf_trace_pop_ = &am_flecs_perf_trace_pop;
    am::lifecycle::init_step("flecs:os_api.log+abort+perf_trace");
    return true;
}
