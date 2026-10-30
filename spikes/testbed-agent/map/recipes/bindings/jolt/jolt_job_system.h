// Reference glue: adapter from Jolt's JobSystem to the Automap jobs shared interface (DRAFT, spike).
#pragma once

#include <Jolt/Jolt.h>
#include <Jolt/Core/FixedSizeFreeList.h>
#include <Jolt/Core/JobSystemWithBarrier.h>

class AutomapJoltJobSystem final : public JPH::JobSystemWithBarrier {
public:
    AutomapJoltJobSystem(JPH::uint max_jobs, JPH::uint max_barriers);

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
};
