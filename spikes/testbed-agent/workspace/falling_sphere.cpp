// Game code: a sphere falling onto a static floor, simulated by Jolt.
#include "falling_sphere.h"

#include <Jolt/Jolt.h>
#include <Jolt/Core/JobSystem.h>
#include <Jolt/Core/TempAllocator.h>
#include <Jolt/Physics/Body/BodyCreationSettings.h>
#include <Jolt/Physics/Collision/Shape/BoxShape.h>
#include <Jolt/Physics/Collision/Shape/SphereShape.h>
#include <Jolt/Physics/PhysicsSystem.h>

#include "automap/lifecycle.h"

using am::lifecycle::Phase;

namespace {

namespace Layers {
constexpr JPH::ObjectLayer non_moving = 0;
constexpr JPH::ObjectLayer moving = 1;
constexpr JPH::ObjectLayer count = 2;
} // namespace Layers

namespace BroadPhaseLayers {
constexpr JPH::BroadPhaseLayer non_moving(0);
constexpr JPH::BroadPhaseLayer moving(1);
constexpr JPH::uint count = 2;
} // namespace BroadPhaseLayers

class ObjectLayerPairFilter final : public JPH::ObjectLayerPairFilter {
public:
    bool ShouldCollide(JPH::ObjectLayer a, JPH::ObjectLayer b) const override {
        return a == Layers::moving || b == Layers::moving; // static bodies never collide with each other
    }
};

class BroadPhaseLayerInterface final : public JPH::BroadPhaseLayerInterface {
public:
    JPH::uint GetNumBroadPhaseLayers() const override { return BroadPhaseLayers::count; }
    JPH::BroadPhaseLayer GetBroadPhaseLayer(JPH::ObjectLayer layer) const override {
        JPH_ASSERT(layer < Layers::count);
        return layer == Layers::non_moving ? BroadPhaseLayers::non_moving : BroadPhaseLayers::moving;
    }
#if defined(JPH_EXTERNAL_PROFILE) || defined(JPH_PROFILE_ENABLED)
    const char* GetBroadPhaseLayerName(JPH::BroadPhaseLayer layer) const override {
        return layer == BroadPhaseLayers::non_moving ? "non_moving" : "moving";
    }
#endif
};

class ObjectVsBroadPhaseLayerFilter final : public JPH::ObjectVsBroadPhaseLayerFilter {
public:
    bool ShouldCollide(JPH::ObjectLayer layer, JPH::BroadPhaseLayer bp) const override {
        return layer == Layers::moving || bp == BroadPhaseLayers::moving;
    }
};

constexpr float sphere_radius = 0.5f;
constexpr float sphere_start_height = 10.0f;

BroadPhaseLayerInterface g_bp_layers;
ObjectVsBroadPhaseLayerFilter g_object_vs_bp;
ObjectLayerPairFilter g_object_pairs;

JPH::TempAllocatorImpl* g_temp_allocator = nullptr;
JPH::PhysicsSystem* g_physics = nullptr;
JPH::JobSystem* g_job_system = nullptr;
JPH::BodyID g_floor;
JPH::BodyID g_sphere;

} // namespace

bool falling_sphere_init_physics(JPH::JobSystem* job_system) {
    g_job_system = job_system;
    g_temp_allocator = new JPH::TempAllocatorImpl(4 * 1024 * 1024);
    g_physics = new JPH::PhysicsSystem();
    g_physics->Init(/*max_bodies*/ 64, /*num_body_mutexes*/ 0, /*max_body_pairs*/ 64, /*max_contact_constraints*/ 64,
                    g_bp_layers, g_object_vs_bp, g_object_pairs);
    am::lifecycle::init_step(Phase::systems, "jolt", "TempAllocator+PhysicsSystem::Init");
    return true;
}

void falling_sphere_create_bodies() {
    JPH::BodyInterface& bodies = g_physics->GetBodyInterface();

    // Floor: a 100 x 2 x 100 m static box whose top surface is at y = 0.
    JPH::BodyCreationSettings floor(new JPH::BoxShape(JPH::Vec3(50.0f, 1.0f, 50.0f)), JPH::RVec3(0, -1, 0),
                                    JPH::Quat::sIdentity(), JPH::EMotionType::Static, Layers::non_moving);
    g_floor = bodies.CreateAndAddBody(floor, JPH::EActivation::DontActivate);

    JPH::BodyCreationSettings sphere(new JPH::SphereShape(sphere_radius), JPH::RVec3(0, sphere_start_height, 0),
                                     JPH::Quat::sIdentity(), JPH::EMotionType::Dynamic, Layers::moving);
    sphere.mRestitution = 0.5f;
    g_sphere = bodies.CreateAndAddBody(sphere, JPH::EActivation::Activate);

    g_physics->OptimizeBroadPhase();
    am::lifecycle::init_step(Phase::game, "jolt", "create floor and sphere bodies");
}

void falling_sphere_step(float dt) { g_physics->Update(dt, /*collision_steps*/ 1, g_temp_allocator, g_job_system); }

void falling_sphere_reset() {
    JPH::BodyInterface& bodies = g_physics->GetBodyInterface();
    bodies.SetPositionAndRotation(g_sphere, JPH::RVec3(0, sphere_start_height, 0), JPH::Quat::sIdentity(),
                                  JPH::EActivation::Activate);
    bodies.SetLinearAndAngularVelocity(g_sphere, JPH::Vec3::sZero(), JPH::Vec3::sZero());
}

float falling_sphere_height() {
    return static_cast<float>(g_physics->GetBodyInterface().GetCenterOfMassPosition(g_sphere).GetY());
}

void falling_sphere_destroy_bodies() {
    JPH::BodyInterface& bodies = g_physics->GetBodyInterface();
    bodies.RemoveBody(g_sphere);
    bodies.DestroyBody(g_sphere);
    bodies.RemoveBody(g_floor);
    bodies.DestroyBody(g_floor);
    am::lifecycle::shutdown_step(Phase::game, "jolt", "destroy floor and sphere bodies");
}

void falling_sphere_shutdown_physics() {
    delete g_physics;
    g_physics = nullptr;
    delete g_temp_allocator;
    g_temp_allocator = nullptr;
    g_job_system = nullptr;
    am::lifecycle::shutdown_step(Phase::systems, "jolt", "destroy PhysicsSystem+TempAllocator");
}
