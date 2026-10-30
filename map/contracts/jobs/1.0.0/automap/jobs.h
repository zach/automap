// Automap jobs interface, jobs contract 1.0.0.
// SPDX-License-Identifier: MIT-0
//
// Every library that runs work on worker threads submits it through this
// interface, tagged with the library's id, so the composition has one
// scheduler. The jobs choices in the choices header pick what backs it:
// the reference thread pool, sized by the workers choice, or the library in
// the jobs slot (whose binding defines the am_jobs_slot_* functions).
#pragma once
#include <atomic>
#include <cstdint>

#include "automap/context.h"

namespace am::jobs {

using Job = void (*)(void* data);

// Counts a group of submitted jobs still to finish.
struct Counter {
    std::atomic<int> pending{0};
};

// Submits job for a library. When done isn't null, it is counted until the
// job finishes.
void submit(const char* library, Job job, void* data, Counter* done);

// Waits until every job counted by done has finished, running queued jobs
// on the calling thread meanwhile, so waiting never starves the workers.
void wait(Counter* done);

// The number of worker threads.
int workers();

// Jobs a library has submitted.
std::uint64_t submitted_count(const char* library);

} // namespace am::jobs

// Starts the scheduler. Runs in the jobs phase, as step automap:jobs.start.
bool am_jobs_setup_init(am::Context& ctx);

// Finishes every job in flight, then stops the scheduler. Runs in the jobs
// phase at shutdown, as step automap:jobs.stop.
void am_jobs_setup_shutdown(am::Context& ctx);
