#ifdef AUTOMAP_PROBE
#include <stub_physics.h>

#include "automap/jobs.h"
#include "automap/probe.h"

// jobs.routed-through-interface: a step submits its tasks through the jobs
// interface, and they run.
namespace {
void am_probe_stub_physics_jobs(am::Context&) {
    auto submitted = am::jobs::submitted_count("stub-physics");
    long long before = stub_physics_checksum();
    stub_physics_step();
    bool routed = am::jobs::submitted_count("stub-physics") > submitted && stub_physics_checksum() > before;
    am::probe::check("jobs.routed-through-interface", "stub-physics", routed, "JOBS_NOT_ROUTED(stub-physics)");
}
const am::probe::Registration registration("stub-physics--jobs", &am_probe_stub_physics_jobs);
} // namespace
#endif
