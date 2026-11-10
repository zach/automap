# SDL_shadercross's build fragment.
# SPDX-License-Identifier: MIT-0
# SDL_shadercross's static library calls SDL3 and its package finds neither
# SDL3's headers nor its library itself.
find_package(SDL3 CONFIG REQUIRED)
find_package(SDL3_shadercross CONFIG REQUIRED)

target_link_libraries(automap_glue PUBLIC SDL3_shadercross::SDL3_shadercross SDL3::SDL3)

# automap_compile_shaders(<target> <stage>:<file>...)
#
# Compiles HLSL shader sources to SPIR-V when <target> builds, into
# shaders/ beside its executable, each as <file's name>.spv: what the
# sdl-shadercross--sdl3 pairing's am_shadercross_create_shader() takes.
# <stage> is vertex, fragment or compute; the entry point is main; <file> is
# relative to the directory of the calling CMakeLists.txt.
#
#     automap_compile_shaders(mygame vertex:shaders/sprite.vert.hlsl fragment:shaders/sprite.frag.hlsl)
#
# The compiler is dxc: vcpkg installs it with SDL_shadercross on Windows and
# x64 Linux, and one on the PATH is used otherwise. Where there is none, as
# on macOS, <file>.spv beside the source is copied instead: commit the
# SPIR-V a Windows or Linux build wrote.
function(automap_compile_shaders target)
    if(TARGET ${target}_shaders)
        message(FATAL_ERROR "automap_compile_shaders: call it once for ${target}, with every shader.")
    endif()
    find_program(AUTOMAP_DXC NAMES dxc
        HINTS "${VCPKG_INSTALLED_DIR}/${VCPKG_TARGET_TRIPLET}/tools/directx-dxc" "${VCPKG_INSTALLED_DIR}/${VCPKG_HOST_TRIPLET}/tools/directx-dxc")
    set(built "${CMAKE_CURRENT_BINARY_DIR}/automap-shaders")
    set(outputs "")
    foreach(shader IN LISTS ARGN)
        if(NOT shader MATCHES "^(vertex|fragment|compute):(.+)$")
            message(FATAL_ERROR "automap_compile_shaders: ${shader} isn't <stage>:<file>, with stage vertex, fragment or compute.")
        endif()
        set(stage "${CMAKE_MATCH_1}")
        get_filename_component(source "${CMAKE_MATCH_2}" ABSOLUTE BASE_DIR "${CMAKE_CURRENT_SOURCE_DIR}")
        get_filename_component(name "${source}" NAME)
        set(output "${built}/${name}.spv")
        if(stage STREQUAL "vertex")
            set(profile vs_6_0)
        elseif(stage STREQUAL "fragment")
            set(profile ps_6_0)
        else()
            set(profile cs_6_0)
        endif()
        if(AUTOMAP_DXC)
            # On Linux dxc loads libdxcompiler.so from vcpkg's lib directory.
            get_filename_component(dxc_dir "${AUTOMAP_DXC}" DIRECTORY)
            add_custom_command(OUTPUT "${output}"
                COMMAND ${CMAKE_COMMAND} -E make_directory "${built}"
                COMMAND ${CMAKE_COMMAND} -E env "LD_LIBRARY_PATH=${dxc_dir}/../../lib"
                        "${AUTOMAP_DXC}" -spirv -fspv-flatten-resource-arrays -T ${profile} -E main "${source}" -Fo "${output}"
                DEPENDS "${source}"
                COMMENT "Compiling ${name} to SPIR-V"
                VERBATIM)
        elseif(EXISTS "${source}.spv")
            add_custom_command(OUTPUT "${output}"
                COMMAND ${CMAKE_COMMAND} -E make_directory "${built}"
                COMMAND ${CMAKE_COMMAND} -E copy_if_different "${source}.spv" "${output}"
                DEPENDS "${source}.spv"
                COMMENT "Copying ${name}.spv, compiled elsewhere"
                VERBATIM)
        else()
            message(FATAL_ERROR "automap_compile_shaders: this machine has no dxc (vcpkg has none for this platform, and none is on the PATH), and ${source}.spv doesn't exist. Compile ${name} on Windows or x64 Linux and commit the .spv beside it, or install dxc: the Vulkan SDK ships one.")
        endif()
        list(APPEND outputs "${output}")
    endforeach()
    add_custom_target(${target}_shaders
        COMMAND ${CMAKE_COMMAND} -E copy_directory "${built}" "$<TARGET_FILE_DIR:${target}>/shaders"
        DEPENDS ${outputs}
        VERBATIM)
    add_dependencies(${target} ${target}_shaders)
endfunction()
