// Probes for Jolt's binding to the diagnostics contract.
// SPDX-License-Identifier: MIT-0
#ifdef AUTOMAP_PROBE
#include <Jolt/Jolt.h>

#include "automap/diag.h"
#include "automap/probe.h"

void am_jolt_trace(const char* format, ...);
#ifdef JPH_ENABLE_ASSERTS
bool am_jolt_assert_failed(const char* expression, const char* message, const char* file, JPH::uint line);
#endif

namespace {

// diagnostics.routed-through-facade: Jolt's trace reaches the facade and,
// where the build has Jolt's assertions (JPH_ENABLE_ASSERTS, Debug builds),
// its assert hook is the facade's. Without assertions only the trace can be
// checked, and the result says so.
void am_probe_jolt_diagnostics(am::Context&) {
    auto before = am::diag::message_count("jolt");
    JPH::Trace("automap probe: diagnostics routing");
    bool routed = JPH::Trace == &am_jolt_trace && am::diag::message_count("jolt") > before;
#ifdef JPH_ENABLE_ASSERTS
    routed = routed && JPH::AssertFailed == &am_jolt_assert_failed;
    const char* condition = "";
#else
    const char* condition =
        "Jolt's assert hook wasn't checked: this build compiles Jolt's assertions out (no JPH_ENABLE_ASSERTS, a non-Debug build); only its trace was.";
#endif
    am::probe::check("diagnostics.routed-through-facade", "jolt", routed, "DIAG_NOT_ROUTED(jolt)", condition);
}
const am::probe::Registration registration("jolt--diagnostics", &am_probe_jolt_diagnostics);

} // namespace
#endif
