// Automap shared interface for the jobs contract 0.1.0 (DRAFT, spike).
#include "jobs.h"

#include <atomic>
#include <condition_variable>
#include <deque>
#include <mutex>
#include <thread>
#include <utility>
#include <vector>

namespace am::jobs {
namespace {
Scheduler* g_scheduler = nullptr;
std::atomic<std::uint64_t> g_executed{0};
} // namespace

void install(Scheduler* s) { g_scheduler = s; }
Scheduler& scheduler() { return *g_scheduler; }
std::uint64_t executed_count() { return g_executed.load(std::memory_order_relaxed); }

struct ThreadPool::Impl {
    std::mutex mutex;
    std::condition_variable wake;
    std::deque<std::pair<TaskFn, void*>> queue;
    std::vector<std::thread> threads;
    bool stopping = false;

    void run() {
        for (;;) {
            std::pair<TaskFn, void*> task;
            {
                std::unique_lock<std::mutex> lock(mutex);
                wake.wait(lock, [this] { return stopping || !queue.empty(); });
                if (stopping && queue.empty()) return;
                task = queue.front();
                queue.pop_front();
            }
            task.first(task.second);
            g_executed.fetch_add(1, std::memory_order_relaxed);
        }
    }
};

ThreadPool::ThreadPool(unsigned workers) : impl_(new Impl) {
    for (unsigned i = 0; i < workers; ++i) impl_->threads.emplace_back([this] { impl_->run(); });
}

ThreadPool::~ThreadPool() {
    {
        std::lock_guard<std::mutex> lock(impl_->mutex);
        impl_->stopping = true;
    }
    impl_->wake.notify_all();
    for (auto& t : impl_->threads) t.join();
    delete impl_;
}

unsigned ThreadPool::worker_count() const { return static_cast<unsigned>(impl_->threads.size()); }

void ThreadPool::submit(TaskFn fn, void* context) {
    {
        std::lock_guard<std::mutex> lock(impl_->mutex);
        impl_->queue.emplace_back(fn, context);
    }
    impl_->wake.notify_one();
}

} // namespace am::jobs
