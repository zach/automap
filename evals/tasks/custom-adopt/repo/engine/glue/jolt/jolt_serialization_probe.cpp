// Probes for Jolt's binding to the serialization contract.
// SPDX-License-Identifier: MIT-0
#ifdef AUTOMAP_PROBE
#include <cstdio>

#include <Jolt/Jolt.h>

#include "automap/probe.h"
#include "jolt_serialization.h"

namespace {

// serialization.rebuild-key-complete: the rebuild key names Jolt's version,
// the composition's variant of it, and Jolt's version id with its
// ABI-affecting defines.
void am_probe_jolt_serialization(am::Context&) {
    using JPH::uint64; // JPH_VERSION_ID names it unqualified
    std::string key = am_jolt_rebuild_key();
    char version[32], id[16];
    std::snprintf(version, sizeof version, "%d.%d.%d", JPH_VERSION_MAJOR, JPH_VERSION_MINOR, JPH_VERSION_PATCH);
    std::snprintf(id, sizeof id, "%08x", static_cast<unsigned>(JPH_VERSION_ID));
    bool complete = key.find(version) != std::string::npos && key.find(AUTOMAP_VARIANT_JOLT) != std::string::npos &&
                    key.find(id) != std::string::npos;
    am::probe::check("serialization.rebuild-key-complete", "jolt", complete, "REBUILD_KEY_INCOMPLETE(jolt)");
}
const am::probe::Registration registration("jolt--serialization", &am_probe_jolt_serialization);

} // namespace
#endif
