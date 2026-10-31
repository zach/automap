// fastgltf's binding to the assets contract: glTF files and their external
// buffers loaded through the assets interface.
// SPDX-License-Identifier: MIT-0
#pragma once
#include <filesystem>

#include <fastgltf/core.hpp>

// Loads the glTF or GLB file at path, and every buffer it refers to by a
// relative path, through the assets interface. The loaded buffers are
// sources::Array, as Options::LoadExternalBuffers would leave them; images
// stay sources::URI for the image decoder to load.
fastgltf::Expected<fastgltf::Asset> am_fastgltf_load(const std::filesystem::path& path,
                                                     fastgltf::Extensions extensions = fastgltf::Extensions::None);
