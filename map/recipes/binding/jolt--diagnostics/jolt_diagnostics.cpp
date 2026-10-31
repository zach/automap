// Jolt's binding to the diagnostics contract.
// SPDX-License-Identifier: MIT-0
#include <Jolt/Jolt.h>

#include <Jolt/Core/Profiler.h>

#include <atomic>
#include <cstdarg>
#include <cstdint>
#include <cstdio>
#include <new>

#include "automap/context.h"
#include "automap/diag.h"
#include "automap/lifecycle.h"

void am_jolt_trace(const char* format, ...) {
    char message[1024];
    va_list args;
    va_start(args, format);
    std::vsnprintf(message, sizeof message, format, args);
    va_end(args);
    am::diag::log(am::diag::Level::info, "jolt", message);
}

#ifdef JPH_ENABLE_ASSERTS
bool am_jolt_assert_failed(const char* expression, const char* message, const char* file, JPH::uint line) {
    return am::diag::assert_failed("jolt", expression, message, file, static_cast<int>(line));
}
#endif

#ifdef JPH_EXTERNAL_PROFILE
// Jolt's profile scopes, in builds with its external profile hooks (the
// Automap registry's), each open a facade zone in the scope's own storage.
static_assert(sizeof(am::diag::Zone) <= 64, "a facade zone fits in Jolt's measurement storage");

// Profile scopes Jolt has opened.
std::atomic<std::uint64_t> am_jolt_profile_zones{0};

#ifdef JPH_SHARED_LIBRARY
// A shared Jolt calls these through JPH::ProfileStartMeasurement and
// JPH::ProfileEndMeasurement, which the binding sets.
void am_jolt_profile_start(const char* name, JPH::uint32, JPH::uint8* storage) {
    new (storage) am::diag::Zone(name);
    am_jolt_profile_zones.fetch_add(1, std::memory_order_relaxed);
}

void am_jolt_profile_end(JPH::uint8* storage) { reinterpret_cast<am::diag::Zone*>(storage)->~Zone(); }
#else
// A static Jolt leaves the measurement's constructor and destructor for the
// program to define.
JPH::ExternalProfileMeasurement::ExternalProfileMeasurement(const char* inName, JPH::uint32) {
    new (mUserData) am::diag::Zone(inName);
    am_jolt_profile_zones.fetch_add(1, std::memory_order_relaxed);
}

JPH::ExternalProfileMeasurement::~ExternalProfileMeasurement() { reinterpret_cast<am::diag::Zone*>(mUserData)->~Zone(); }
#endif
#endif

// Routes Jolt's trace output, in builds with JPH_ENABLE_ASSERTS (Debug) its
// assertions, and in builds with JPH_EXTERNAL_PROFILE its profile scopes
// through the diagnostics facade. Runs in the diagnostics phase.
bool am_bind_jolt_diagnostics(am::Context&) {
    JPH::Trace = &am_jolt_trace;
    JPH_IF_ENABLE_ASSERTS(JPH::AssertFailed = &am_jolt_assert_failed;)
#if defined(JPH_EXTERNAL_PROFILE) && defined(JPH_SHARED_LIBRARY)
    JPH::ProfileStartMeasurement = &am_jolt_profile_start;
    JPH::ProfileEndMeasurement = &am_jolt_profile_end;
#endif
    am::lifecycle::init_step("jolt:Trace+AssertFailed");
    return true;
}
