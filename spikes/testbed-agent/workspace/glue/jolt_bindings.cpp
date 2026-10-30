// Reference glue: Jolt bindings (DRAFT, spike).
#include "jolt_bindings.h"

#include <cstdarg>
#include <cstdio>

#include <Jolt/Jolt.h>
#include <Jolt/Core/Factory.h>
#include <Jolt/Core/Memory.h>
#include <Jolt/Physics/PhysicsSettings.h>
#include <Jolt/RegisterTypes.h>

#include "automap/alloc.h"
#include "automap/diag.h"
#include "automap/jobs.h"
#include "automap/lifecycle.h"
#include "automap/probe.h"
#include "jolt_job_system.h"

using am::alloc::Client;
using am::lifecycle::Phase;

namespace {
bool g_factory_existed_at_install = true;
AutomapJoltJobSystem* g_job_system = nullptr;

void* am_allocate(size_t size) { return am::alloc::allocate(Client::jolt, size); }
void* am_reallocate(void* block, size_t, size_t new_size) { return am::alloc::reallocate(Client::jolt, block, new_size); }
void am_free(void* block) { am::alloc::release(Client::jolt, block); }
void* am_aligned_allocate(size_t size, size_t alignment) { return am::alloc::allocate_aligned(Client::jolt, size, alignment); }
void am_aligned_free(void* block) { am::alloc::release_aligned(Client::jolt, block); }

void am_trace(const char* format, ...) {
    char buffer[1024];
    va_list args;
    va_start(args, format);
    std::vsnprintf(buffer, sizeof(buffer), format, args);
    va_end(args);
    am::diag::log(am::diag::Level::info, "jolt", buffer);
}

#ifdef JPH_ENABLE_ASSERTS
bool am_assert_failed(const char* expression, const char* message, const char* file, JPH::uint line) {
    return am::diag::assert_failed("jolt", expression, message, file, static_cast<int>(line));
}
#endif
} // namespace

void jolt_bind_allocation() {
    g_factory_existed_at_install = JPH::Factory::sInstance != nullptr;
    JPH::Allocate = am_allocate;
    JPH::Reallocate = am_reallocate;
    JPH::Free = am_free;
    JPH::AlignedAllocate = am_aligned_allocate;
    JPH::AlignedFree = am_aligned_free;
    am::lifecycle::init_step(Phase::memory, "jolt", "JPH allocation hooks");
}

void jolt_bind_diagnostics() {
    JPH::Trace = am_trace;
    JPH_IF_ENABLE_ASSERTS(JPH::AssertFailed = am_assert_failed;)
    am::lifecycle::init_step(Phase::diagnostics, "jolt", "JPH::Trace+AssertFailed");
}

JPH::JobSystem* jolt_create_job_system() {
    g_job_system = new AutomapJoltJobSystem(JPH::cMaxPhysicsJobs, JPH::cMaxPhysicsBarriers);
    am::lifecycle::init_step(Phase::jobs, "jolt", "AutomapJoltJobSystem");
    return g_job_system;
}

void jolt_register_types() {
    JPH::Factory::sInstance = new JPH::Factory();
    JPH::RegisterTypes();
    am::lifecycle::init_step(Phase::systems, "jolt", "Factory+RegisterTypes");
}

void jolt_unregister_types() {
    JPH::UnregisterTypes();
    delete JPH::Factory::sInstance;
    JPH::Factory::sInstance = nullptr;
    am::lifecycle::shutdown_step(Phase::systems, "jolt", "UnregisterTypes+Factory");
}

void jolt_destroy_job_system() {
    delete g_job_system;
    g_job_system = nullptr;
    am::lifecycle::shutdown_step(Phase::jobs, "jolt", "AutomapJoltJobSystem");
}

void jolt_register_probes() {
    am::probe::check("allocation.hooks-before-first-allocation", "jolt", !g_factory_existed_at_install, "ALLOC_HOOK_LATE(jolt)");
    bool routed = JPH::Allocate == am_allocate && JPH::AlignedAllocate == am_aligned_allocate && am::alloc::routed_count(Client::jolt) > 0;
    am::probe::check("allocation.routed-through-facade", "jolt", routed, "ALLOC_NOT_ROUTED(jolt)");
    am::probe::check("diagnostics.log-routed", "jolt", JPH::Trace == am_trace, "DIAG_NOT_ROUTED(jolt)");
#ifdef JPH_ENABLE_ASSERTS
    am::probe::check("diagnostics.assert-routed", "jolt", JPH::AssertFailed == am_assert_failed, "ASSERT_NOT_ROUTED(jolt)");
#endif
    am::probe::check("build-flags.abi-match", "jolt", JPH::VerifyJoltVersionID(), "DEFINE_MISMATCH(jolt)");
    am::probe::check("jobs.routed-through-interface", "jolt", am::jobs::executed_count() > 0, "JOBS_NOT_ROUTED(jolt)");
}
