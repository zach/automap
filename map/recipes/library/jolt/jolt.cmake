# Jolt's build fragment. Glue that derives from Jolt classes (the jobs
# adapter, the layer interfaces) compiles without RTTI, as the joltphysics
# port builds Jolt.
# SPDX-License-Identifier: MIT-0
find_package(Jolt CONFIG REQUIRED)

target_link_libraries(automap_glue PUBLIC Jolt::Jolt)
target_compile_options(automap_glue PRIVATE $<IF:$<CXX_COMPILER_ID:MSVC>,/GR-,-fno-rtti>)
