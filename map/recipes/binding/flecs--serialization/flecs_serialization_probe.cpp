// Probes for flecs's binding to the serialization contract.
// SPDX-License-Identifier: MIT-0
#ifdef AUTOMAP_PROBE
#include <cstdio>

#include <flecs.h>

#include "automap/probe.h"
#include "flecs_serialization.h"

namespace {

// serialization.rebuild-key-complete: the rebuild key names flecs's version
// and the composition's variant of it.
void am_probe_flecs_serialization(am::Context&) {
    std::string key = am_flecs_rebuild_key();
    char version[32];
    std::snprintf(version, sizeof version, "%d.%d.%d", FLECS_VERSION_MAJOR, FLECS_VERSION_MINOR, FLECS_VERSION_PATCH);
    bool complete = key.find(version) != std::string::npos && key.find(AUTOMAP_VARIANT_FLECS) != std::string::npos;
    am::probe::check("serialization.rebuild-key-complete", "flecs", complete, "REBUILD_KEY_INCOMPLETE(flecs)");
}
const am::probe::Registration registration("flecs--serialization", &am_probe_flecs_serialization);

} // namespace
#endif
