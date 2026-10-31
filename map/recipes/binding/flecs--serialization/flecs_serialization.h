// flecs's binding to the serialization contract: the rebuild key for data
// flecs produces, such as worlds and entities serialized to JSON. It names
// flecs's version and the variant the composition builds.
// SPDX-License-Identifier: MIT-0
#pragma once
#include <string>

// The key data flecs produced is stored under; data under another key is
// rebuilt rather than read.
std::string am_flecs_rebuild_key();
