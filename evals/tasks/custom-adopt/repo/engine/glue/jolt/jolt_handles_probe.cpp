// Probes for Jolt's binding to the handles contract.
// SPDX-License-Identifier: MIT-0
#ifdef AUTOMAP_PROBE
#include "automap/probe.h"
#include "jolt_handles.h"

namespace {

// handles.bits-preserved: body ids with each of their index and sequence
// bits set, the largest id, and the invalid id convert to engine handles and
// back unchanged. (The top bit is the broad phase's; no body id sets it.)
void am_probe_jolt_handles(am::Context&) {
    bool preserved = true;
    for (int bit = 0; bit < 31; ++bit) {
        JPH::BodyID id(std::uint32_t(1) << bit);
        preserved = preserved && am::handles::to_jolt(am::handles::from_jolt(id)) == id;
    }
    for (JPH::BodyID id : {JPH::BodyID(JPH::BodyID::cMaxBodyIndex, JPH::BodyID::cMaxSequenceNumber), JPH::BodyID()}) {
        preserved = preserved && am::handles::to_jolt(am::handles::from_jolt(id)) == id;
    }
    am::probe::check("handles.bits-preserved", "jolt", preserved, "HANDLE_TRUNCATED(jolt)");
}
const am::probe::Registration registration("jolt--handles", &am_probe_jolt_handles);

} // namespace
#endif
