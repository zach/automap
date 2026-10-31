// Jolt's binding to the serialization contract.
// SPDX-License-Identifier: MIT-0
#include "jolt_serialization.h"

#include <cstdio>

#include <Jolt/Jolt.h>

std::string am_jolt_rebuild_key() {
    using JPH::uint64; // JPH_VERSION_ID names it unqualified
    char key[96];
    std::snprintf(key, sizeof key, "jolt-%d.%d.%d-%s-%08x", JPH_VERSION_MAJOR, JPH_VERSION_MINOR, JPH_VERSION_PATCH,
                  AUTOMAP_VARIANT_JOLT, static_cast<unsigned>(JPH_VERSION_ID));
    return key;
}
