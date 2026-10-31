// Jolt's setup glue.
// SPDX-License-Identifier: MIT-0
#include "jolt_setup.h"

#include <Jolt/Core/Factory.h>
#include <Jolt/Core/TempAllocator.h>
#include <Jolt/Physics/Collision/BroadPhase/BroadPhaseLayer.h>
#include <Jolt/Physics/PhysicsSystem.h>
#include <Jolt/RegisterTypes.h>

#include "automap/lifecycle.h"
#include "automap/probe.h"

namespace {

namespace broad_phase {
constexpr JPH::BroadPhaseLayer non_moving(0);
constexpr JPH::BroadPhaseLayer moving(1);
constexpr JPH::uint count = 2;
} // namespace broad_phase

class BroadPhaseLayers final : public JPH::BroadPhaseLayerInterface {
public:
    JPH::uint GetNumBroadPhaseLayers() const override { return broad_phase::count; }

    JPH::BroadPhaseLayer GetBroadPhaseLayer(JPH::ObjectLayer layer) const override {
        return layer == am_jolt_layers::non_moving ? broad_phase::non_moving : broad_phase::moving;
    }

#if defined(JPH_EXTERNAL_PROFILE) || defined(JPH_PROFILE_ENABLED)
    const char* GetBroadPhaseLayerName(JPH::BroadPhaseLayer layer) const override {
        return layer == broad_phase::non_moving ? "non-moving" : "moving";
    }
#endif
};

class ObjectVsBroadPhase final : public JPH::ObjectVsBroadPhaseLayerFilter {
public:
    bool ShouldCollide(JPH::ObjectLayer layer, JPH::BroadPhaseLayer broad) const override {
        return layer == am_jolt_layers::moving || broad == broad_phase::moving;
    }
};

class ObjectPairs final : public JPH::ObjectLayerPairFilter {
public:
    bool ShouldCollide(JPH::ObjectLayer a, JPH::ObjectLayer b) const override {
        return a == am_jolt_layers::moving || b == am_jolt_layers::moving;
    }
};

BroadPhaseLayers broad_phase_layers;
ObjectVsBroadPhase object_vs_broad_phase;
ObjectPairs object_pairs;

// Placeholder capacities the game replaces with its own.
constexpr JPH::uint max_bodies = 1024;
constexpr JPH::uint max_body_pairs = 1024;
constexpr JPH::uint max_contact_constraints = 1024;
constexpr size_t temp_allocator_bytes = 10 * 1024 * 1024;

} // namespace

bool am_jolt_setup_init(am::Context& ctx) {
    // build-flags.abi-defines-match: Jolt aborts in RegisterTypes() when
    // the glue's version id (its version and ABI defines) differs from the
    // library's, so the check comes first and its failure is recorded
    // before it.
#ifdef AUTOMAP_PROBE
    am::probe::expect_failure("build-flags.abi-defines-match", "", "ABI_DEFINE_MISMATCH(jolt)");
#endif
    bool match = JPH::VerifyJoltVersionID();
#ifdef AUTOMAP_PROBE
    am::probe::check("build-flags.abi-defines-match", "", match, "ABI_DEFINE_MISMATCH(jolt)");
#endif
    if (!match) {
        JPH::Trace("Jolt was built with a different version or ABI defines than the glue.");
        return false;
    }

    JPH::Factory::sInstance = new JPH::Factory();
    JPH::RegisterTypes();
    am::lifecycle::init_step("jolt:Factory+RegisterTypes");

    ctx.jolt.temp_allocator = new JPH::TempAllocatorImpl(temp_allocator_bytes);
    ctx.jolt.physics_system = new JPH::PhysicsSystem();
    ctx.jolt.physics_system->Init(max_bodies, 0, max_body_pairs, max_contact_constraints, broad_phase_layers,
                                  object_vs_broad_phase, object_pairs);
    am::lifecycle::init_step("jolt:PhysicsSystem::Init");
    return true;
}

void am_jolt_setup_shutdown(am::Context& ctx) {
    delete ctx.jolt.physics_system;
    ctx.jolt.physics_system = nullptr;
    delete ctx.jolt.temp_allocator;
    ctx.jolt.temp_allocator = nullptr;
    am::lifecycle::shutdown_step("jolt:PhysicsSystem-destroyed");

    JPH::UnregisterTypes();
    delete JPH::Factory::sInstance;
    JPH::Factory::sInstance = nullptr;
    am::lifecycle::shutdown_step("jolt:UnregisterTypes+Factory");
}
