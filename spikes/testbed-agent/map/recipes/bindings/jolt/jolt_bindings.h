// Reference glue: Jolt bindings to the allocation, diagnostics, jobs, lifecycle and build-flags contracts (DRAFT, spike).
#pragma once

namespace JPH { class JobSystem; }

void jolt_bind_allocation();             // memory phase; before any other Jolt call
void jolt_bind_diagnostics();            // diagnostics phase
JPH::JobSystem* jolt_create_job_system(); // jobs phase; adapter over am::jobs::scheduler()
void jolt_register_types();              // systems phase
void jolt_unregister_types();            // systems phase, shutdown
void jolt_destroy_job_system();          // jobs phase, shutdown
void jolt_register_probes();             // call after a smoke run, before shutdown
