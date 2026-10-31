// Jolt's binding to the serialization contract: the rebuild key for data
// Jolt produces, such as saved physics state and binary shapes. It names
// Jolt's version and the variant the composition builds, and Jolt's own
// JPH_VERSION_ID, which packs the version with its ABI-affecting defines.
// SPDX-License-Identifier: MIT-0
#pragma once
#include <string>

// The key data Jolt produced is stored under; data under another key is
// rebuilt rather than read.
std::string am_jolt_rebuild_key();
