# Reference glue: the testbed composition's top-level build file (DRAFT, spike).
# Sandbox substitution: libraries come from FetchContent at their pinned commits instead of vcpkg.
cmake_minimum_required(VERSION 3.25)
project(automap_testbed LANGUAGES C CXX)

set(CMAKE_CXX_STANDARD 20)          # build-flags choice: cxx_standard
set(CMAKE_CXX_STANDARD_REQUIRED ON)

include(FetchContent)

# sdl3 3.4.18 (pin from map/libraries/sdl3.json)
set(SDL_SHARED OFF CACHE BOOL "" FORCE)
set(SDL_STATIC ON CACHE BOOL "" FORCE)
set(SDL_TEST_LIBRARY OFF CACHE BOOL "" FORCE)
set(SDL_TESTS OFF CACHE BOOL "" FORCE)
set(SDL_EXAMPLES OFF CACHE BOOL "" FORCE)
if(CMAKE_SYSTEM_NAME STREQUAL "Linux" AND AUTOMAP_HEADLESS)
  set(SDL_X11 OFF CACHE BOOL "" FORCE)
  set(SDL_WAYLAND OFF CACHE BOOL "" FORCE)
  set(SDL_UNIX_CONSOLE_BUILD ON CACHE BOOL "" FORCE) # SDL 3.4 refuses to configure without X11/Wayland unless this is set
endif()
FetchContent_Declare(SDL3 GIT_REPOSITORY https://github.com/libsdl-org/SDL GIT_TAG 829a65d769d935c4852f8159e964312c0957260a)

# dear-imgui 1.93.0, docking variant
FetchContent_Declare(imgui GIT_REPOSITORY https://github.com/ocornut/imgui GIT_TAG b916186e66aa3c02b72c6ca17455516512ed8049)

# jolt 5.6.0, default variant (pin and options from map/libraries/jolt.json)
foreach(opt TARGET_UNIT_TESTS TARGET_HELLO_WORLD TARGET_PERFORMANCE_TEST TARGET_SAMPLES TARGET_VIEWER
            ENABLE_ALL_WARNINGS OVERRIDE_CXX_FLAGS INTERPROCEDURAL_OPTIMIZATION ENABLE_INSTALL
            JPH_USE_DX12 JPH_USE_VK JPH_USE_MTL)
  set(${opt} OFF CACHE BOOL "" FORCE)
endforeach()
FetchContent_Declare(JoltPhysics GIT_REPOSITORY https://github.com/jrouwe/JoltPhysics GIT_TAG e77f175595e64cb44218cc9d9d56fc365ad0e36a
  SOURCE_SUBDIR Build)

FetchContent_MakeAvailable(SDL3 imgui JoltPhysics)

add_library(imgui STATIC
  ${imgui_SOURCE_DIR}/imgui.cpp ${imgui_SOURCE_DIR}/imgui_draw.cpp ${imgui_SOURCE_DIR}/imgui_tables.cpp
  ${imgui_SOURCE_DIR}/imgui_widgets.cpp ${imgui_SOURCE_DIR}/imgui_demo.cpp
  ${imgui_SOURCE_DIR}/backends/imgui_impl_sdl3.cpp ${imgui_SOURCE_DIR}/backends/imgui_impl_sdlgpu3.cpp)
target_include_directories(imgui PUBLIC ${imgui_SOURCE_DIR} ${imgui_SOURCE_DIR}/backends)
target_link_libraries(imgui PUBLIC SDL3::SDL3)

add_executable(testbed
  main.cpp falling_sphere.cpp
  glue/automap/alloc.cpp glue/automap/diag.cpp glue/automap/jobs.cpp glue/automap/lifecycle.cpp glue/automap/probe.cpp
  glue/sdl3_bindings.cpp glue/imgui_bindings.cpp glue/imgui_sdl3_gpu.cpp
  glue/jolt_bindings.cpp glue/jolt_job_system.cpp)
target_include_directories(testbed PRIVATE glue)
target_link_libraries(testbed PRIVATE SDL3::SDL3 imgui Jolt::Jolt) # Jolt::Jolt carries Jolt's PUBLIC JPH_* defines
# Jolt is built with -fno-rtti (CPP_RTTI_ENABLED defaults OFF), so its out-of-line classes have no typeinfo.
# Translation units that derive from Jolt classes must also be compiled without RTTI, or they fail to link.
set_source_files_properties(glue/jolt_bindings.cpp glue/jolt_job_system.cpp falling_sphere.cpp
  PROPERTIES COMPILE_OPTIONS "$<$<CXX_COMPILER_ID:GNU,Clang>:-fno-rtti>")
