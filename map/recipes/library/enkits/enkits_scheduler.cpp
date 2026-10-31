// enkiTS's setup glue.
// SPDX-License-Identifier: MIT-0
#include "enkits_scheduler.h"

enki::TaskSchedulerConfig& am_enkits_config() {
    static enki::TaskSchedulerConfig config;
    return config;
}

enki::TaskScheduler& am_enkits_scheduler() {
    static enki::TaskScheduler scheduler;
    return scheduler;
}
