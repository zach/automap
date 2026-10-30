// Reference glue: adapter from Jolt's JobSystem to the Automap jobs shared interface (DRAFT, spike).
// Follows the pattern of Jolt's own JobSystemThreadPool.
#include "jolt_job_system.h"

#include <thread>

#include "automap/jobs.h"

AutomapJoltJobSystem::AutomapJoltJobSystem(JPH::uint max_jobs, JPH::uint max_barriers) {
    JobSystemWithBarrier::Init(max_barriers);
    jobs_.Init(max_jobs, max_jobs);
}

int AutomapJoltJobSystem::GetMaxConcurrency() const {
    return static_cast<int>(am::jobs::scheduler().worker_count()) + 1;
}

JPH::JobHandle AutomapJoltJobSystem::CreateJob(const char* name, JPH::ColorArg color, const JobFunction& function,
                                               JPH::uint32 num_dependencies) {
    JPH::uint32 index;
    for (;;) {
        index = jobs_.ConstructObject(name, color, this, function, num_dependencies);
        if (index != AvailableJobs::cInvalidObjectIndex) break;
        std::this_thread::yield();
    }
    Job* job = &jobs_.Get(index);
    JPH::JobHandle handle(job); // keeps a reference; the job may complete as soon as it is queued
    if (num_dependencies == 0) QueueJob(job);
    return handle;
}

void AutomapJoltJobSystem::QueueJob(Job* job) {
    job->AddRef(); // released by run() on the worker
    am::jobs::scheduler().submit(&AutomapJoltJobSystem::run, job);
}

void AutomapJoltJobSystem::QueueJobs(Job** jobs, JPH::uint num_jobs) {
    for (JPH::uint i = 0; i < num_jobs; ++i) QueueJob(jobs[i]);
}

void AutomapJoltJobSystem::FreeJob(Job* job) { jobs_.DestructObject(job); }

void AutomapJoltJobSystem::run(void* context) {
    Job* job = static_cast<Job*>(context);
    job->Execute();
    job->Release();
}
