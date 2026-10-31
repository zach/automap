// enkiTS's binding to the jobs contract: the jobs slot's implementation of
// the jobs interface, which runs every job the composition submits as an
// enkiTS task.
// SPDX-License-Identifier: MIT-0
#include "automap/jobs.h"
#include "automap/lifecycle.h"
#include "enkits_scheduler.h"

namespace {

// Threads outside enkiTS that may submit jobs, such as a library's own
// thread; each registers with the scheduler the first time it submits.
constexpr uint32_t external_threads = 8;

// Deletes its job's task once the task completes: enkiTS doesn't touch a
// task again by the time its dependents' OnDependenciesComplete runs.
struct DeleteOnCompletion final : enki::ICompletable {
    enki::Dependency dependency;
    void OnDependenciesComplete(enki::TaskScheduler* scheduler, uint32_t thread) override {
        ICompletable::OnDependenciesComplete(scheduler, thread);
        delete dependency.GetDependencyTask(); // deletes this too, as a member
    }
};

// One submitted job, as an enkiTS task that deletes itself when done.
struct JobTask final : enki::ITaskSet {
    am::jobs::Job job;
    void* data;
    DeleteOnCompletion deleter;
    JobTask(am::jobs::Job j, void* d) : job(j), data(d) { deleter.SetDependency(deleter.dependency, this); }
    void ExecuteRange(enki::TaskSetPartition, uint32_t) override { job(data); }
};

// Registers a thread enkiTS doesn't know, for as long as the thread lives.
struct ExternalThread {
    bool registered = false;
    ~ExternalThread() {
        if (registered) am_enkits_scheduler().DeRegisterExternalTaskThread();
    }
};
thread_local ExternalThread external;

} // namespace

bool am_jobs_slot_start(int workers) {
    enki::TaskSchedulerConfig& config = am_enkits_config();
    config.numTaskThreadsToCreate = static_cast<uint32_t>(workers);
    config.numExternalTaskThreads = external_threads;
    am_enkits_scheduler().Initialize(config);
    am::lifecycle::init_step("enkits:Initialize");
    return true;
}

void am_jobs_slot_submit(am::jobs::Job job, void* data) {
    enki::TaskScheduler& scheduler = am_enkits_scheduler();
    if (scheduler.GetThreadNum() == enki::NO_THREAD_NUM && !external.registered) {
        external.registered = scheduler.RegisterExternalTaskThread();
    }
    scheduler.AddTaskSetToPipe(new JobTask(job, data));
}

// Runs queued tasks on the calling thread while it waits; enkiTS returns
// when none is ready.
bool am_jobs_slot_run_one() {
    am_enkits_scheduler().WaitforTask(nullptr);
    return false;
}

void am_jobs_slot_stop() {
    am_enkits_scheduler().WaitforAllAndShutdown();
    am::lifecycle::shutdown_step("enkits:WaitforAllAndShutdown");
}
