// Jolt's binding to the determinism contract.
// SPDX-License-Identifier: MIT-0
#include "jolt_determinism.h"

#include <vector>

#include <Jolt/Jolt.h>

#include <Jolt/Physics/Body/BodyCreationSettings.h>
#include <Jolt/Physics/Collision/Shape/BoxShape.h>
#include <Jolt/Physics/PhysicsSystem.h>
#include <Jolt/Physics/StateRecorderImpl.h>

#include "jolt_setup.h"

std::uint64_t am_jolt_replay_hash(am::Context& ctx) {
    JPH::PhysicsSystem& system = *ctx.jolt.physics_system;
    JPH::BodyInterface& bodies = system.GetBodyInterface();
    std::vector<JPH::BodyID> created;

    JPH::BodyCreationSettings floor(new JPH::BoxShape(JPH::Vec3(20.0f, 1.0f, 20.0f)), JPH::RVec3(0.0f, -1.0f, 0.0f),
                                    JPH::Quat::sIdentity(), JPH::EMotionType::Static, am_jolt_layers::non_moving);
    created.push_back(bodies.CreateAndAddBody(floor, JPH::EActivation::DontActivate));
    for (int i = 0; i < 16; ++i) {
        // Each box slightly off the one below, so the stack topples.
        JPH::RVec3 position(0.1f * float(i % 3), 0.6f + 1.05f * float(i), 0.07f * float(i % 5));
        JPH::BodyCreationSettings box(new JPH::BoxShape(JPH::Vec3(0.5f, 0.5f, 0.5f)), position, JPH::Quat::sIdentity(),
                                      JPH::EMotionType::Dynamic, am_jolt_layers::moving);
        created.push_back(bodies.CreateAndAddBody(box, JPH::EActivation::Activate));
    }
    system.OptimizeBroadPhase();
    for (int step = 0; step < 120; ++step) {
        system.Update(1.0f / 60.0f, 1, ctx.jolt.temp_allocator, ctx.jolt.job_system);
    }

    JPH::StateRecorderImpl recorder;
    system.SaveState(recorder);
    std::string state = recorder.GetData();
    std::uint64_t hash = 14695981039346656037ull;
    for (unsigned char c : state) hash = (hash ^ c) * 1099511628211ull;

    bodies.RemoveBodies(created.data(), int(created.size()));
    bodies.DestroyBodies(created.data(), int(created.size()));
    return hash;
}
