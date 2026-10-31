// Probes for flecs's binding to the handles contract.
// SPDX-License-Identifier: MIT-0
#ifdef AUTOMAP_PROBE
#include "automap/probe.h"
#include "flecs_handles.h"

namespace {

// handles.bits-preserved: entity ids with each of their 64 bits set convert
// to engine handles and back unchanged.
void am_probe_flecs_handles(am::Context&) {
    bool preserved = true;
    for (int bit = 0; bit < 64; ++bit) {
        ecs_entity_t id = std::uint64_t(1) << bit;
        preserved = preserved && am::handles::to_flecs(am::handles::from_flecs(id)) == id;
    }
    am::probe::check("handles.bits-preserved", "flecs", preserved, "HANDLE_TRUNCATED(flecs)");
}
const am::probe::Registration registration("flecs--handles", &am_probe_flecs_handles);

} // namespace
#endif
