// Automap jobs interface, jobs contract 1.0.0: the reference thread pool.
// SPDX-License-Identifier: MIT-0
#include "automap/jobs.h"

#include <condition_variable>
#include <deque>
#include <map>
#include <mutex>
#include <string>
#include <thread>
#include <vector>

#include "automap/choices.h"
#include "automap/lifecycle.h"

#if defined(AUTOMAP_JOBS_SCHEDULER_JOBS_SLOT)
// Defined by the jobs slot's binding.
bool am_jobs_slot_start(int workers);
void am_jobs_slot_submit(am::jobs::Job job, void* data);
bool am_jobs_slot_run_one();
void am_jobs_slot_stop();
#endif

namespace am::jobs {
namespace {

struct Task {
    Job job;
    void* data;
    Counter* done;
};

struct Pool {
    std::mutex mutex;
    std::condition_variable wake;
    std::deque<Task> queue;
    std::vector<std::thread> threads;
    std::map<std::string, std::uint64_t> submitted;
    bool stopping = false;
};

Pool& pool() {
    static Pool* p = new Pool;
    return *p;
}

void run(const Task& t) {
    t.job(t.data);
    if (t.done != nullptr) t.done->pending.fetch_sub(1, std::memory_order_acq_rel);
}

#if !defined(AUTOMAP_JOBS_SCHEDULER_JOBS_SLOT)
// Pops one queued task, or returns false when there is none.
bool pop(Task* out) {
    Pool& p = pool();
    std::lock_guard<std::mutex> lock(p.mutex);
    if (p.queue.empty()) return false;
    *out = p.queue.front();
    p.queue.pop_front();
    return true;
}

void worker_loop() {
    Pool& p = pool();
    for (;;) {
        Task t;
        {
            std::unique_lock<std::mutex> lock(p.mutex);
            p.wake.wait(lock, [&] { return p.stopping || !p.queue.empty(); });
            if (p.queue.empty()) return; // stopping, and nothing left in flight
            t = p.queue.front();
            p.queue.pop_front();
        }
        run(t);
    }
}
#endif

} // namespace

void submit(const char* library, Job job, void* data, Counter* done) {
    Pool& p = pool();
    if (done != nullptr) done->pending.fetch_add(1, std::memory_order_acq_rel);
    {
        std::lock_guard<std::mutex> lock(p.mutex);
        ++p.submitted[library];
#if !defined(AUTOMAP_JOBS_SCHEDULER_JOBS_SLOT)
        p.queue.push_back({job, data, done});
#endif
    }
#if defined(AUTOMAP_JOBS_SCHEDULER_JOBS_SLOT)
    struct Slot {
        static void run_task(void* raw) {
            Task* t = static_cast<Task*>(raw);
            run(*t);
            delete t;
        }
    };
    am_jobs_slot_submit(&Slot::run_task, new Task{job, data, done});
#else
    p.wake.notify_one();
#endif
}

void wait(Counter* done) {
    while (done->pending.load(std::memory_order_acquire) > 0) {
#if defined(AUTOMAP_JOBS_SCHEDULER_JOBS_SLOT)
        if (!am_jobs_slot_run_one()) std::this_thread::yield();
#else
        Task t;
        if (pop(&t)) run(t);
        else std::this_thread::yield();
#endif
    }
}

int workers() { return AUTOMAP_JOBS_WORKERS; }

std::uint64_t submitted_count(const char* library) {
    Pool& p = pool();
    std::lock_guard<std::mutex> lock(p.mutex);
    auto it = p.submitted.find(library);
    return it == p.submitted.end() ? 0 : it->second;
}

} // namespace am::jobs

bool am_jobs_setup_init(am::Context&) {
#if defined(AUTOMAP_JOBS_SCHEDULER_JOBS_SLOT)
    if (!am_jobs_slot_start(AUTOMAP_JOBS_WORKERS)) return false;
#else
    am::jobs::Pool& p = am::jobs::pool();
    p.stopping = false;
    for (int i = 0; i < AUTOMAP_JOBS_WORKERS; ++i) p.threads.emplace_back(am::jobs::worker_loop);
#endif
    am::lifecycle::init_step("automap:jobs.start");
    return true;
}

void am_jobs_setup_shutdown(am::Context&) {
#if defined(AUTOMAP_JOBS_SCHEDULER_JOBS_SLOT)
    am_jobs_slot_stop();
#else
    // Workers drain the queue before they exit, and joining waits for the
    // jobs they are running: no job outlives the pool.
    am::jobs::Pool& p = am::jobs::pool();
    {
        std::lock_guard<std::mutex> lock(p.mutex);
        p.stopping = true;
    }
    p.wake.notify_all();
    for (std::thread& t : p.threads) t.join();
    p.threads.clear();
#endif
    am::lifecycle::shutdown_step("automap:jobs.stop");
}
