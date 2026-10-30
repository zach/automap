// Automap shared interface for the jobs contract 0.1.0 (DRAFT, spike).
#pragma once
#include <cstdint>

namespace am::jobs {

using TaskFn = void (*)(void* context);

// The neutral scheduler every library's job hooks submit to.
class Scheduler {
public:
    virtual ~Scheduler() = default;
    virtual unsigned worker_count() const = 0;
    virtual void submit(TaskFn fn, void* context) = 0;
};

void install(Scheduler* scheduler);
Scheduler& scheduler();

// Number of tasks executed through the interface.
std::uint64_t executed_count();

// Reference implementation for the "automap-thread-pool" choice.
class ThreadPool final : public Scheduler {
public:
    explicit ThreadPool(unsigned workers);
    ~ThreadPool() override;
    unsigned worker_count() const override;
    void submit(TaskFn fn, void* context) override;

private:
    struct Impl;
    Impl* impl_;
};

} // namespace am::jobs
