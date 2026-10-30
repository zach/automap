// Probes for Jolt's binding to the jobs contract.
// SPDX-License-Identifier: MIT-0
#ifdef AUTOMAP_PROBE
#include <Jolt/Jolt.h>

#include <Jolt/Physics/Body/BodyCreationSettings.h>
#include <Jolt/Physics/Collision/Shape/BoxShape.h>
#include <Jolt/Physics/Collision/Shape/SphereShape.h>
#include <Jolt/Physics/PhysicsSystem.h>

#include "automap/jobs.h"
#include "automap/probe.h"
#include "jolt_setup.h"

namespace {

// jobs.routed-through-interface: one PhysicsSystem::Update with a falling
// sphere above a floor submits Jolt's jobs through the jobs interface.
void am_probe_jolt_jobs(am::Context& ctx) {
    JPH::BodyInterface& bodies = ctx.jolt.physics_system->GetBodyInterface();
    JPH::BodyID floor = bodies.CreateAndAddBody(
        JPH::BodyCreationSettings(new JPH::BoxShape(JPH::Vec3(10.0f, 1.0f, 10.0f)), JPH::RVec3(0.0f, -1.0f, 0.0f),
                                  JPH::Quat::sIdentity(), JPH::EMotionType::Static, am_jolt_layers::non_moving),
        JPH::EActivation::DontActivate);
    JPH::BodyID sphere = bodies.CreateAndAddBody(
        JPH::BodyCreationSettings(new JPH::SphereShape(0.5f), JPH::RVec3(0.0f, 2.0f, 0.0f), JPH::Quat::sIdentity(),
                                  JPH::EMotionType::Dynamic, am_jolt_layers::moving),
        JPH::EActivation::Activate);
    auto before = am::jobs::submitted_count("jolt");
    ctx.jolt.physics_system->Update(1.0f / 60.0f, 1, ctx.jolt.temp_allocator, ctx.jolt.job_system);
    bool routed = am::jobs::submitted_count("jolt") > before;
    for (JPH::BodyID id : {sphere, floor}) {
        bodies.RemoveBody(id);
        bodies.DestroyBody(id);
    }
    am::probe::check("jobs.routed-through-interface", "jolt", routed, "JOBS_NOT_ROUTED(jolt)");
}
const am::probe::Registration registration("jolt--jobs", &am_probe_jolt_jobs);

} // namespace
#endif
