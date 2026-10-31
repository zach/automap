# flecs's build fragment. The port's CMake wrapper makes flecs::flecs name
# the library on static and dynamic triplets alike.
# SPDX-License-Identifier: MIT-0
find_package(flecs CONFIG REQUIRED)

target_link_libraries(automap_glue PUBLIC flecs::flecs)
