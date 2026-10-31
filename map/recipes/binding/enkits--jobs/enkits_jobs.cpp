// enkiTS's binding to the jobs contract: the jobs slot's implementation of
// the jobs interface, which runs every job the composition submits as an
// enkiTS task.
// SPDX-License-Identifier: MIT-0
#include <algorithm>
#include <memory>
#include <mutex>
#include <vector>

#include "automap/jobs.h"
#include "automap/lifecycle.h"
#include "enkits_scheduler.h"

namespace {

// Threads outside enkiTS that may submit jobs, such as a library's own
// thread; each registers with the scheduler the first time it submits.
constexpr uint32_t external_threads = 8;

// One submitted job, as an enkiTS task.
struct JobTask final : enki::ITaskSet {
    am::jobs::Job job;
    void* data;
    void ExecuteRange(enki::TaskSetPartition, uint32_t) override { job(data); }
};

std::mutex tasks_mutex;
std::vector<std::unique_ptr<JobTask>> tasks;

// Frees the tasks that have finished. enkiTS doesn't touch a task without
// dependents once it is complete.
void reap() {
    tasks.erase(std::remove_if(tasks.begin(), tasks.end(), [](const auto& t) { return t->GetIsComplete(); }), tasks.end());
}

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
    auto task = std::make_unique<JobTask>();
    task->job = job;
    task->data = data;
    JobTask* queued = task.get();
    {
        std::lock_guard<std::mutex> lock(tasks_mutex);
        reap();
        tasks.push_back(std::move(task));
    }
    scheduler.AddTaskSetToPipe(queued);
}

// Runs queued tasks on the calling thread while it waits; enkiTS returns
// when none is ready.
bool am_jobs_slot_run_one() {
    am_enkits_scheduler().WaitforTask(nullptr);
    return false;
}

void am_jobs_slot_stop() {
    am_enkits_scheduler().WaitforAllAndShutdown();
    {
        std::lock_guard<std::mutex> lock(tasks_mutex);
        tasks.clear();
    }
    am::lifecycle::shutdown_step("enkits:WaitforAllAndShutdown");
}
