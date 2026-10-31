// Probes for Jolt's binding to the diagnostics contract.
// SPDX-License-Identifier: MIT-0
#ifdef AUTOMAP_PROBE
#include <Jolt/Jolt.h>

#include <Jolt/Core/Profiler.h>

#include "automap/diag.h"
#include "automap/probe.h"

void am_jolt_trace(const char* format, ...);
#ifdef JPH_ENABLE_ASSERTS
bool am_jolt_assert_failed(const char* expression, const char* message, const char* file, JPH::uint line);
#endif
#ifdef JPH_EXTERNAL_PROFILE
#include <atomic>
#include <cstdint>
extern std::atomic<std::uint64_t> am_jolt_profile_zones;
#endif

namespace {

// diagnostics.routed-through-facade: Jolt's trace reaches the facade; where
// the build has Jolt's assertions (JPH_ENABLE_ASSERTS, Debug builds), its
// assert hook is the facade's; and where it has Jolt's external profile
// hooks (JPH_EXTERNAL_PROFILE), its profile scopes open facade zones. Without assertions only the trace can be
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
#ifdef JPH_EXTERNAL_PROFILE
    // Jolt's external profile hooks open facade zones: a profile scope
    // reaches the glue.
    auto zones = am_jolt_profile_zones.load();
    {
        using namespace JPH; // JPH_PROFILE names Jolt's types unqualified
        JPH_PROFILE("automap probe: profile routing");
    }
    routed = routed && am_jolt_profile_zones.load() > zones;
#endif
    am::probe::check("diagnostics.routed-through-facade", "jolt", routed, "DIAG_NOT_ROUTED(jolt)", condition);
}
const am::probe::Registration registration("jolt--diagnostics", &am_probe_jolt_diagnostics);

} // namespace
#endif
