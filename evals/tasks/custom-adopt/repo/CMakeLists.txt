# Acme's engine: its in-house scheduler and renderer, SDL3 and Jolt from
# vcpkg, and glue of its own in engine/glue.
# SPDX-License-Identifier: MIT-0
cmake_minimum_required(VERSION 3.25)
project(acme_engine LANGUAGES C CXX)

add_subdirectory(engine/acme-jobs)
add_subdirectory(engine/acme-renderer)
find_package(SDL3 CONFIG REQUIRED)
find_package(Jolt CONFIG REQUIRED)
find_package(Threads REQUIRED)

# The engine's glue builds through Automap's glue target.
include(cmake/automap/automap.cmake)

add_executable(acme_game engine/core/engine_main.cpp)
target_link_libraries(acme_game PRIVATE
    SDL3::SDL3 Jolt::Jolt
    Acme::Jobs Acme::Renderer
    Threads::Threads automap_glue)
