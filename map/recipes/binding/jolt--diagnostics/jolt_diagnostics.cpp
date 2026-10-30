// Jolt's binding to the diagnostics contract.
// SPDX-License-Identifier: MIT-0
#include <Jolt/Jolt.h>

#include <cstdarg>
#include <cstdio>

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

// Routes Jolt's trace output and, in builds with JPH_ENABLE_ASSERTS (Debug),
// its assertions through the diagnostics facade. Runs in the diagnostics
// phase.
bool am_bind_jolt_diagnostics(am::Context&) {
    JPH::Trace = &am_jolt_trace;
    JPH_IF_ENABLE_ASSERTS(JPH::AssertFailed = &am_jolt_assert_failed;)
    am::lifecycle::init_step("jolt:Trace+AssertFailed");
    return true;
}
