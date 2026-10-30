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
    ${AUTOMAP_ROOT}/glue/stub-ui--stub-window/stub_ui_stub_window.cpp
    ${AUTOMAP_ROOT}/glue/stub-ui--stub-window/stub_ui_stub_window_probe.cpp
    ${AUTOMAP_ROOT}/glue/stub-ui/stub_ui_allocation.cpp
    ${AUTOMAP_ROOT}/glue/stub-ui/stub_ui_allocation_probe.cpp
    ${AUTOMAP_ROOT}/glue/stub-ui/stub_ui_diagnostics.cpp
    ${AUTOMAP_ROOT}/glue/stub-ui/stub_ui_diagnostics_probe.cpp
    ${AUTOMAP_ROOT}/glue/stub-ui/stub_ui_setup.cpp
    ${AUTOMAP_ROOT}/glue/stub-window/stub_window_allocation.cpp
    ${AUTOMAP_ROOT}/glue/stub-window/stub_window_allocation_probe.cpp
    ${AUTOMAP_ROOT}/glue/stub-window/stub_window_diagnostics.cpp
    ${AUTOMAP_ROOT}/glue/stub-window/stub_window_diagnostics_probe.cpp
    ${AUTOMAP_ROOT}/glue/stub-window/stub_window_setup.cpp
)
target_include_directories(automap_glue PUBLIC ${AUTOMAP_ROOT}/glue)
target_compile_features(automap_glue PUBLIC cxx_std_${AUTOMAP_CXX_STANDARD})
if(AUTOMAP_PROBE)
    target_compile_definitions(automap_glue PUBLIC AUTOMAP_PROBE)
endif()

include(${CMAKE_CURRENT_LIST_DIR}/stub-ui.cmake)
include(${CMAKE_CURRENT_LIST_DIR}/stub-window.cmake)
