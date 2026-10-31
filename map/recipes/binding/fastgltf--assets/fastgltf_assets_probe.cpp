// Probes for fastgltf's binding to the assets contract.
// SPDX-License-Identifier: MIT-0
#ifdef AUTOMAP_PROBE
#include <chrono>
#include <cstdio>
#include <filesystem>
#include <string>
#include <variant>

#include "automap/assets.h"
#include "automap/probe.h"
#include "fastgltf_assets.h"

namespace {

// A one-triangle glTF whose positions are in an external buffer.
const char sample_gltf[] = R"({"asset":{"version":"2.0"},
"buffers":[{"uri":"triangle.bin","byteLength":36}],
"bufferViews":[{"buffer":0,"byteLength":36}],
"accessors":[{"bufferView":0,"componentType":5126,"count":3,"type":"VEC3","min":[0,0,0],"max":[1,1,0]}],
"meshes":[{"primitives":[{"attributes":{"POSITION":0}}]}],
"nodes":[{"mesh":0,"name":"triangle"}],"scenes":[{"nodes":[0]}],"scene":0})";
const float sample_positions[] = {0, 0, 0, 1, 0, 0, 0, 1, 0};

void write(const std::filesystem::path& path, const void* data, std::size_t size) {
    if (std::FILE* f = std::fopen(path.string().c_str(), "wb")) {
        std::fwrite(data, 1, size, f);
        std::fclose(f);
    }
}

// The import's output, as text: what game code reads from the asset.
std::string summary(const fastgltf::Asset& asset) {
    std::string out;
    for (const auto& mesh : asset.meshes) out += "mesh " + std::string(std::string_view(mesh.name)) + " " + std::to_string(mesh.primitives.size()) + "\n";
    for (const auto& node : asset.nodes) out += "node " + std::string(std::string_view(node.name)) + "\n";
    for (const auto& accessor : asset.accessors) out += "accessor " + std::to_string(accessor.count) + "\n";
    for (const auto& buffer : asset.buffers) {
        if (const auto* array = std::get_if<fastgltf::sources::Array>(&buffer.data)) {
            out.append(reinterpret_cast<const char*>(array->bytes.data()), array->bytes.size());
        }
    }
    return out;
}

// assets.reads-through-interface: loading the sample reads the glTF file
// and its buffer through the assets interface. assets.output-identical:
// loading it twice gives the same asset.
void am_probe_fastgltf_assets(am::Context&) {
    auto stamp = std::chrono::steady_clock::now().time_since_epoch().count();
    std::filesystem::path dir = std::filesystem::temp_directory_path() / ("automap-probe-" + std::to_string(stamp));
    std::filesystem::create_directories(dir);
    write(dir / "triangle.gltf", sample_gltf, sizeof sample_gltf - 1);
    write(dir / "triangle.bin", sample_positions, sizeof sample_positions);

    auto before = am::assets::read_count("fastgltf");
    auto first = am_fastgltf_load(dir / "triangle.gltf");
    bool through = first.error() == fastgltf::Error::None && am::assets::read_count("fastgltf") == before + 2;
    auto second = am_fastgltf_load(dir / "triangle.gltf");
    bool identical = first.error() == fastgltf::Error::None && second.error() == fastgltf::Error::None &&
                     summary(first.get()) == summary(second.get()) && first->buffers.size() == 1 &&
                     std::holds_alternative<fastgltf::sources::Array>(first->buffers[0].data);

    std::error_code ignored;
    std::filesystem::remove_all(dir, ignored);
    am::probe::check("assets.reads-through-interface", "fastgltf", through, "ASSET_READ_BYPASSED(fastgltf)");
    am::probe::check("assets.output-identical", "fastgltf", identical, "ASSET_OUTPUT_DIFFERS(fastgltf)");
}
const am::probe::Registration registration("fastgltf--assets", &am_probe_fastgltf_assets);

} // namespace
#endif
