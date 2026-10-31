# Written by automap from automap.manifest.json. Edit the manifest, not this file.
# The developer's CMakeLists.txt includes this file once.

set(AUTOMAP_ROOT ${CMAKE_CURRENT_LIST_DIR}/../..)
set(AUTOMAP_CXX_STANDARD 20)

add_library(automap_glue OBJECT
    ${AUTOMAP_ROOT}/glue/automap/alloc.cpp
    ${AUTOMAP_ROOT}/glue/automap/diag.cpp
    ${AUTOMAP_ROOT}/glue/automap/lifecycle.cpp
    ${AUTOMAP_ROOT}/glue/automap/probe.cpp
    ${AUTOMAP_ROOT}/glue/automap/startup.cpp
    ${AUTOMAP_ROOT}/glue/dear-imgui--sdl3/imgui_sdl3_gpu.cpp
    ${AUTOMAP_ROOT}/glue/dear-imgui--sdl3/imgui_sdl3_gpu_probe.cpp
    ${AUTOMAP_ROOT}/glue/dear-imgui/imgui_allocation.cpp
    ${AUTOMAP_ROOT}/glue/dear-imgui/imgui_allocation_probe.cpp
    ${AUTOMAP_ROOT}/glue/dear-imgui/imgui_build_flags_probe.cpp
    ${AUTOMAP_ROOT}/glue/dear-imgui/imgui_setup.cpp
    ${AUTOMAP_ROOT}/glue/sdl3/sdl3_allocation.cpp
    ${AUTOMAP_ROOT}/glue/sdl3/sdl3_allocation_probe.cpp
    ${AUTOMAP_ROOT}/glue/sdl3/sdl3_diagnostics.cpp
    ${AUTOMAP_ROOT}/glue/sdl3/sdl3_diagnostics_probe.cpp
    ${AUTOMAP_ROOT}/glue/sdl3/sdl3_setup.cpp
)
target_include_directories(automap_glue PUBLIC ${AUTOMAP_ROOT}/glue)
target_compile_features(automap_glue PUBLIC cxx_std_${AUTOMAP_CXX_STANDARD})
if(AUTOMAP_PROBE)
    target_compile_definitions(automap_glue PUBLIC AUTOMAP_PROBE)
endif()
target_compile_definitions(automap_glue PUBLIC AUTOMAP_LIBRARY_DEAR_IMGUI=1 AUTOMAP_LIBRARY_SDL3=1)

include(${CMAKE_CURRENT_LIST_DIR}/dear-imgui.cmake)
include(${CMAKE_CURRENT_LIST_DIR}/sdl3.cmake)
