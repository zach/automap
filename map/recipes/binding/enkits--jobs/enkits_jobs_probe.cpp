// Probes for enkiTS's binding to the jobs contract.
// SPDX-License-Identifier: MIT-0
#ifdef AUTOMAP_PROBE
#include <atomic>
#include <chrono>

#include "automap/jobs.h"
#include "automap/probe.h"
#include "enkits_scheduler.h"

namespace {

std::atomic<int> on_task_threads{0};

void note_thread(void*) {
    uint32_t n = am_enkits_scheduler().GetThreadNum();
    if (n != enki::NO_THREAD_NUM && n != 0) on_task_threads.fetch_add(1);
}

// jobs.routed-through-interface: jobs submitted through the jobs interface
// run on enkiTS's task threads. Waiting runs queued jobs on the calling
// thread too, and enkiTS's task threads may still be starting (on Windows
// they take longer), so batches are submitted until one runs on a task
// thread, for up to two seconds.
void am_probe_enkits_jobs(am::Context&) {
    const auto deadline = std::chrono::steady_clock::now() + std::chrono::seconds(2);
    do {
        am::jobs::Counter done;
        for (int i = 0; i < 64; ++i) am::jobs::submit("enkits-probe", &note_thread, nullptr, &done);
        am::jobs::wait(&done);
    } while (on_task_threads.load() == 0 && std::chrono::steady_clock::now() < deadline);
    am::probe::check("jobs.routed-through-interface", "enkits", on_task_threads.load() > 0, "JOBS_NOT_ROUTED(enkits)");
}
const am::probe::Registration registration("enkits--jobs", &am_probe_enkits_jobs);

} // namespace
#endif
