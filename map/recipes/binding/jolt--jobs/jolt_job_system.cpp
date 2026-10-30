// Jolt's binding to the jobs contract: a JPH::JobSystem that runs every
// Jolt job through the composition's jobs interface.
// SPDX-License-Identifier: MIT-0
#include "jolt_job_system.h"

#include <Jolt/Jolt.h>

#include <Jolt/Core/FixedSizeFreeList.h>
#include <Jolt/Core/JobSystemWithBarrier.h>
#include <Jolt/Physics/PhysicsSettings.h>

#include <atomic>
#include <chrono>
#include <thread>

#include "automap/jobs.h"
#include "automap/lifecycle.h"

namespace {

class AutomapJobSystem final : public JPH::JobSystemWithBarrier {
public:
    AutomapJobSystem() {
        JobSystemWithBarrier::Init(JPH::cMaxPhysicsBarriers);
        jobs_.Init(JPH::cMaxPhysicsJobs, JPH::cMaxPhysicsJobs);
    }

    // A worker can still be between a job's Execute() and Release(), which
    // may free the job into this system: wait for every job in flight.
    ~AutomapJobSystem() override {
        while (in_flight_.load(std::memory_order_acquire) > 0) std::this_thread::yield();
    }

    int GetMaxConcurrency() const override { return am::jobs::workers() + 1; }

    JobHandle CreateJob(const char* name, JPH::ColorArg color, const JobFunction& function, JPH::uint32 dependencies) override {
        JPH::uint32 index;
        for (;;) {
            index = jobs_.ConstructObject(name, color, this, function, dependencies);
            if (index != Jobs::cInvalidObjectIndex) break;
            std::this_thread::sleep_for(std::chrono::microseconds(100));
        }
        Job* job = &jobs_.Get(index);
        sync_.fetch_add(1, std::memory_order_release); // publishes the job to whichever thread queues it
        JobHandle handle(job); // keeps a reference: the job may finish as soon as it is queued
        if (dependencies == 0) QueueJob(job);
        return handle;
    }

protected:
    void QueueJob(Job* job) override {
        sync_.fetch_add(0, std::memory_order_acq_rel);
        job->AddRef();
        in_flight_.fetch_add(1, std::memory_order_acq_rel);
        am::jobs::submit("jolt", &Run, job, nullptr);
    }

    void QueueJobs(Job** jobs, JPH::uint count) override {
        for (JPH::uint i = 0; i < count; ++i) QueueJob(jobs[i]);
    }

    void FreeJob(Job* job) override {
        sync_.fetch_add(0, std::memory_order_acq_rel);
        jobs_.DestructObject(job);
    }

private:
    using Jobs = JPH::FixedSizeFreeList<Job>;

    static void Run(void* data) {
        Job* job = static_cast<Job*>(data);
        AutomapJobSystem* self = static_cast<AutomapJobSystem*>(job->GetJobSystem());
        job->Execute();
        self->sync_.fetch_add(1, std::memory_order_release);
        job->Release();
        self->in_flight_.fetch_sub(1, std::memory_order_acq_rel);
    }

    Jobs jobs_;
    std::atomic<int> in_flight_{0};
    // Jolt hands a job between threads inside the library: a dependency
    // that completes there queues the job from another thread, and the
    // last release, there too, frees it after an atomic fence. The library
    // is built without ThreadSanitizer, which sees neither, so sync_ makes
    // the same ordering explicit: CreateJob and each worker release to it,
    // and QueueJob and FreeJob acquire from it.
    std::atomic<unsigned> sync_{0};
};

} // namespace

bool am_bind_jolt_jobs(am::Context& ctx) {
    ctx.jolt.job_system = new AutomapJobSystem();
    am::lifecycle::init_step("jolt:job-system-adapter");
    return true;
}

void am_unbind_jolt_jobs(am::Context& ctx) {
    delete ctx.jolt.job_system;
    ctx.jolt.job_system = nullptr;
    am::lifecycle::shutdown_step("jolt:job-system-adapter-destroyed");
}
