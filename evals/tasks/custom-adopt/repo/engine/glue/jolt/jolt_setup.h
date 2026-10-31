// Jolt's setup glue.
// SPDX-License-Identifier: MIT-0
#pragma once
#include <Jolt/Jolt.h>

#include <Jolt/Physics/Collision/ObjectLayer.h>

#include "automap/context.h"

// The object layers setup glue maps to broad-phase layers. They are
// placeholders the game replaces with its own.
namespace am_jolt_layers {
constexpr JPH::ObjectLayer non_moving = 0;
constexpr JPH::ObjectLayer moving = 1;
constexpr JPH::ObjectLayer count = 2;
} // namespace am_jolt_layers

// Registers Jolt's types and creates the physics system, with its layers,
// a temp allocator, and the job system from the jolt--jobs binding. Checks
// first that the glue was compiled with the library's version and ABI
// defines (JPH::VerifyJoltVersionID). Provides jolt.physics_system and
// jolt.temp_allocator.
bool am_jolt_setup_init(am::Context& ctx);

// Destroys the physics system, then unregisters Jolt's types and deletes
// the factory.
void am_jolt_setup_shutdown(am::Context& ctx);
