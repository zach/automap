// Jolt's binding to the jobs contract.
// SPDX-License-Identifier: MIT-0
#pragma once
#include "automap/context.h"

// Creates the job system Jolt runs its jobs on, which submits every Jolt
// job to the jobs interface. Provides jolt.job_system. Runs in the jobs
// phase, after the scheduler starts.
bool am_bind_jolt_jobs(am::Context& ctx);

// Waits for every Jolt job in flight, then destroys the job system.
void am_unbind_jolt_jobs(am::Context& ctx);
