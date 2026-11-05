# SDL_shadercross's build fragment.
# SPDX-License-Identifier: MIT-0
# SDL_shadercross's targets link SDL3's headers, which its package doesn't find itself.
find_package(SDL3 CONFIG REQUIRED)
find_package(SDL3_shadercross CONFIG REQUIRED)

target_link_libraries(automap_glue PUBLIC SDL3_shadercross::SDL3_shadercross)
