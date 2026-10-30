// Reference glue: adapter from Jolt's JobSystem to the Automap jobs shared interface (DRAFT, spike).
#pragma once

#include <atomic>

#include <Jolt/Jolt.h>
#include <Jolt/Core/FixedSizeFreeList.h>
#include <Jolt/Core/JobSystemWithBarrier.h>

class AutomapJoltJobSystem final : public JPH::JobSystemWithBarrier {
public:
    AutomapJoltJobSystem(JPH::uint max_jobs, JPH::uint max_barriers);
    ~AutomapJoltJobSystem() override; // waits until no submitted task still references a job

    int GetMaxConcurrency() const override;
    JPH::JobHandle CreateJob(const char* name, JPH::ColorArg color, const JobFunction& function,
                             JPH::uint32 num_dependencies = 0) override;

protected:
    void QueueJob(Job* job) override;
    void QueueJobs(Job** jobs, JPH::uint num_jobs) override;
    void FreeJob(Job* job) override;

private:
    static void run(void* context);

    using AvailableJobs = JPH::FixedSizeFreeList<Job>;
    AvailableJobs jobs_;
    std::atomic<JPH::uint32> in_flight_{0}; // tasks submitted to the scheduler that have not finished run()
};
