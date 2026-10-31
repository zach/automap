// flecs's binding to the jobs contract: the pipeline's worker stages run as
// jobs on the jobs interface, through flecs's task API.
// SPDX-License-Identifier: MIT-0
#include <cstdint>

#include "automap/context.h"
#include "automap/jobs.h"
#include "automap/lifecycle.h"
#include "flecs_setup.h"

namespace {

// One worker stage's task, until flecs joins it.
struct Task {
    ecs_os_thread_callback_t callback = nullptr;
    void* arg = nullptr;
    void* result = nullptr;
    am::jobs::Counter done;
};

void run(void* raw) {
    Task* task = static_cast<Task*>(raw);
    task->result = task->callback(task->arg);
}

} // namespace

ecs_os_thread_t am_flecs_task_new(ecs_os_thread_callback_t callback, void* arg) {
    Task* task = new Task;
    task->callback = callback;
    task->arg = arg;
    am::jobs::submit("flecs", &run, task, &task->done);
    return static_cast<ecs_os_thread_t>(reinterpret_cast<std::uintptr_t>(task));
}

void* am_flecs_task_join(ecs_os_thread_t handle) {
    Task* task = reinterpret_cast<Task*>(static_cast<std::uintptr_t>(handle));
    am::jobs::wait(&task->done);
    void* result = task->result;
    delete task;
    return result;
}

// Gives flecs the jobs interface as its task API. Runs in the jobs phase,
// before setup glue makes the world and asks for task threads.
bool am_bind_flecs_jobs(am::Context&) {
    ecs_os_api_t& api = am_flecs_os_api();
    api.task_new_ = &am_flecs_task_new;
    api.task_join_ = &am_flecs_task_join;
    am::lifecycle::init_step("flecs:os_api.task");
    return true;
}
