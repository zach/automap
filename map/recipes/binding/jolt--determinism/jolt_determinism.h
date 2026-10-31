// Jolt's binding to the determinism contract: a fixed simulation whose
// final state hash is the same on every platform when Jolt is a
// deterministic build (JPH_CROSS_PLATFORM_DETERMINISTIC).
// SPDX-License-Identifier: MIT-0
#pragma once
#include <cstdint>

#include "automap/context.h"

// Drops a fixed stack of boxes onto a floor in the composition's physics
// system, steps it two seconds at 60 Hz on the composition's job system,
// and returns the FNV-1a hash of the saved state, then removes the bodies.
// Run it before the game adds bodies of its own.
std::uint64_t am_jolt_replay_hash(am::Context& ctx);
