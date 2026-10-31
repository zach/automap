// flecs's binding to the serialization contract.
// SPDX-License-Identifier: MIT-0
#include "flecs_serialization.h"

#include <cstdio>

#include <flecs.h>

std::string am_flecs_rebuild_key() {
    char key[96];
    std::snprintf(key, sizeof key, "flecs-%d.%d.%d-%s", FLECS_VERSION_MAJOR, FLECS_VERSION_MINOR, FLECS_VERSION_PATCH,
                  AUTOMAP_VARIANT_FLECS);
    return key;
}
