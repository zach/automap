#include <stub_physics.h>

#include "automap/context.h"
#include "automap/jobs.h"
#include "automap/lifecycle.h"

namespace {
am::jobs::Counter step_tasks;

void submit(stub_physics_task_fn task, void* data, void*) { am::jobs::submit("stub-physics", task, data, &step_tasks); }
void wait(void*) { am::jobs::wait(&step_tasks); }
} // namespace

// Routes stub-physics's tasks through the jobs interface. Runs in the jobs
// phase, after the scheduler starts.
bool am_bind_stub_physics_jobs(am::Context&) {
    stub_physics_set_job_hooks(&submit, &wait, nullptr);
    am::lifecycle::init_step("stub-physics:set_job_hooks");
    return true;
}
