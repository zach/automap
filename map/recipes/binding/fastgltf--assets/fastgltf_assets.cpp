// fastgltf's binding to the assets contract.
// SPDX-License-Identifier: MIT-0
#include "fastgltf_assets.h"

#include <cstring>
#include <variant>
#include <vector>

#include "automap/assets.h"

namespace {

// fastgltf 1.0 boxes a buffer's URI and replaces StaticVector with
// static_vector; 0.9 has neither container header.
#if __has_include(<fastgltf/containers/static_vector.hpp>)
const fastgltf::URI& uri_of(const fastgltf::sources::URI& source) { return *source.uri; }
using Bytes = fastgltf::static_vector<std::byte>;
constexpr fastgltf::Error missing_buffer = fastgltf::Error::MissingExternalFile;
#else
const fastgltf::URI& uri_of(const fastgltf::sources::URI& source) { return source.uri; }
using Bytes = fastgltf::StaticVector<std::byte>;
constexpr fastgltf::Error missing_buffer = fastgltf::Error::MissingExternalBuffer;
#endif

} // namespace

fastgltf::Expected<fastgltf::Asset> am_fastgltf_load(const std::filesystem::path& path, fastgltf::Extensions extensions) {
    std::vector<std::byte> bytes;
    if (!am::assets::read("fastgltf", path.string().c_str(), bytes)) return fastgltf::Error::InvalidPath;
    auto data = fastgltf::GltfDataBuffer::FromBytes(bytes.data(), bytes.size());
    if (data.error() != fastgltf::Error::None) return data.error();
    fastgltf::Parser parser(extensions);
    // Without Options::LoadExternalBuffers fastgltf reads no file itself:
    // each external buffer is loaded below, through the interface.
    auto asset = parser.loadGltf(data.get(), path.parent_path(), fastgltf::Options::LoadGLBBuffers);
    if (asset.error() != fastgltf::Error::None) return asset;
    for (fastgltf::Buffer& buffer : asset->buffers) {
        auto* uri = std::get_if<fastgltf::sources::URI>(&buffer.data);
        if (uri == nullptr || !uri_of(*uri).isLocalPath()) continue;
        std::vector<std::byte> contents;
        if (!am::assets::read("fastgltf", (path.parent_path() / uri_of(*uri).fspath()).string().c_str(), contents) ||
            contents.size() < uri->fileByteOffset + buffer.byteLength) {
            return missing_buffer;
        }
        fastgltf::sources::Array loaded{Bytes(buffer.byteLength), uri->mimeType};
        std::memcpy(loaded.bytes.data(), contents.data() + uri->fileByteOffset, buffer.byteLength);
        buffer.data = std::move(loaded);
    }
    return asset;
}
