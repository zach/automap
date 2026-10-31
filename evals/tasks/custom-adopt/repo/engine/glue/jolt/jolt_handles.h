// Jolt's binding to the handles contract: engine handles that carry Jolt's
// body ids. A BodyID is 32 bits (23-bit index, 8-bit sequence number, and a
// bit reserved for the broad phase); the engine handle keeps all of them.
// SPDX-License-Identifier: MIT-0
#pragma once
#include <cstdint>

#include <Jolt/Jolt.h>

#include <Jolt/Physics/Body/BodyID.h>

namespace am::handles {

// The engine's handle for a Jolt body.
struct JoltBody {
    std::uint32_t bits = JPH::BodyID::cInvalidBodyID;
};

inline JoltBody from_jolt(JPH::BodyID id) { return JoltBody{id.GetIndexAndSequenceNumber()}; }
inline JPH::BodyID to_jolt(JoltBody handle) { return JPH::BodyID(handle.bits); }

} // namespace am::handles
