// enkiTS's setup glue: the one task scheduler the composition runs, and the
// configuration its bindings fill in before the jobs phase initializes it.
// SPDX-License-Identifier: MIT-0
#pragma once
#include <TaskScheduler.h>

// The scheduler's configuration. The allocation binding sets its custom
// allocator in the memory phase and the diagnostics binding its profiler
// callbacks in the diagnostics phase; the jobs binding initializes the
// scheduler with it in the jobs phase.
enki::TaskSchedulerConfig& am_enkits_config();

// The composition's task scheduler.
enki::TaskScheduler& am_enkits_scheduler();
