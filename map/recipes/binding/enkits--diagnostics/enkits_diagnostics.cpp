// enkiTS's binding to the diagnostics contract: its profiler callbacks
// report each task thread to the facade and open a profile zone while a
// thread waits.
// SPDX-License-Identifier: MIT-0
#include <cstdio>
#include <memory>
#include <vector>

#include "automap/context.h"
#include "automap/diag.h"
#include "automap/lifecycle.h"
#include "enkits_scheduler.h"

namespace {

// The zones open on this thread, innermost last.
thread_local std::vector<std::unique_ptr<am::diag::Zone>> zones;

void open(const char* name) { zones.push_back(std::make_unique<am::diag::Zone>(name)); }

void close() {
    if (!zones.empty()) zones.pop_back();
}

} // namespace

void am_enkits_thread_start(uint32_t threadnum) {
    char message[64];
    std::snprintf(message, sizeof message, "task thread %u started", threadnum);
    am::diag::log(am::diag::Level::trace, "enkits", message);
}

void am_enkits_thread_stop(uint32_t threadnum) {
    char message[64];
    std::snprintf(message, sizeof message, "task thread %u stopped", threadnum);
    am::diag::log(am::diag::Level::trace, "enkits", message);
}

void am_enkits_suspend_start(uint32_t) { open("enkiTS: waiting for new tasks"); }
void am_enkits_suspend_stop(uint32_t) { close(); }
void am_enkits_wait_start(uint32_t) { open("enkiTS: waiting for a task to complete"); }
void am_enkits_wait_stop(uint32_t) { close(); }

// Routes enkiTS's profiler callbacks through the diagnostics facade. Runs
// in the diagnostics phase, before the jobs phase initializes the
// scheduler.
bool am_bind_enkits_diagnostics(am::Context&) {
    enki::ProfilerCallbacks& callbacks = am_enkits_config().profilerCallbacks;
    callbacks.threadStart = &am_enkits_thread_start;
    callbacks.threadStop = &am_enkits_thread_stop;
    callbacks.waitForNewTaskSuspendStart = &am_enkits_suspend_start;
    callbacks.waitForNewTaskSuspendStop = &am_enkits_suspend_stop;
    callbacks.waitForTaskCompleteStart = &am_enkits_wait_start;
    callbacks.waitForTaskCompleteStop = &am_enkits_wait_stop;
    callbacks.waitForTaskCompleteSuspendStart = &am_enkits_suspend_start;
    callbacks.waitForTaskCompleteSuspendStop = &am_enkits_suspend_stop;
    am::lifecycle::init_step("enkits:profilerCallbacks");
    return true;
}
