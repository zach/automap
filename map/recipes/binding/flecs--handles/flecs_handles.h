// flecs's binding to the handles contract: engine handles that carry
// flecs's entity ids. An ecs_entity_t is 64 bits: a 32-bit index, a 16-bit
// generation that tells a recycled index from the entity that had it, and
// id flags in the top bits; the engine handle keeps all of them.
// SPDX-License-Identifier: MIT-0
#pragma once
#include <cstdint>

#include <flecs.h>

namespace am::handles {

// The engine's handle for a flecs entity.
struct FlecsEntity {
    std::uint64_t bits = 0;
};

inline FlecsEntity from_flecs(ecs_entity_t entity) { return FlecsEntity{entity}; }
inline ecs_entity_t to_flecs(FlecsEntity handle) { return handle.bits; }

} // namespace am::handles
