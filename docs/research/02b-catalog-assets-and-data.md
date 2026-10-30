# Library catalog, part B: assets, data definition and asset-pipeline patterns

This note covers three areas. It profiles the C++ asset libraries a custom engine is likely to compose: models, images and textures, geometry, animation, fonts and text, file watching and virtual file systems. It profiles the data-definition and serialization libraries: binary schemas, archive-style serializers, JSON, reflection including C++26 static reflection, SQLite, compression and hashing. It then extracts the recurring asset-pipeline patterns from reference engines and asks whether Automap should model data definition and the asset pipeline as first-class *slots* with contracts. For each library it records license, API flavor, latest release, maintenance health, platform reach, integration seams and bridges.

Accessed: 2026-10-29

**Method and limits.** Most facts come from the upstream repositories themselves. I read READMEs, changelogs, headers, tags and commit history from treeless git clones made on 2026-10-29. Health figures are my own analysis of `git log` on each default branch for 2025-10-29 to 2026-10-29: commits, distinct author names, and the top author's share of commits. They are crude. They count bot and merge commits and miss work on side branches. Three limits shaped the sourcing:

- The session's WebSearch budget was exhausted before this task could run a single search.
- `gcc.gnu.org` and `docs.godotengine.org` failed DNS resolution under the session's network policy. Following the tool guidance, I did not try other non-GitHub sites, so vendor pages for Unreal, Unity, RAD/Oodle and Our Machinery were not reached.
- The GitHub REST API was blocked, but `git` over HTTPS worked. Where a doc site has a GitHub source (godot-docs, o3de.org, cplusplus/draft, UnityCsReference, gcc-mirror, llvm-project), I used that instead.

Anything I could not check this way is marked **[Unverified]**.

---

## Key findings

1. **[Verified]** Several "default" asset libraries changed shape in 2026, so API flavor and version pins are map data, not trivia:
   - tinygltf v3 is a C11 POD API with arena allocation, and the C++ v1/v2 API now lives in `attic/` [3].
   - tinyexr v3 is a pure-C11 rewrite [7].
   - fastgltf after 0.9.x requires C++20 [1], and EnTT v4.0.0 raised its floor from C++17 to C++20 [33].
   - KTX-Software v5 (still a release candidate) breaks `ktxBasisParams` and removes the legacy tools [8].
2. **[Verified]** Bridges between libraries carry version skew and vendored copies of shared dependencies:
   - KTX v5.0.0-rc2 vendors Basis Universal **v2.1**, while Basis itself is at **v2.5** with the new XUASTC and XUBC7 codecs [8][9].
   - ozz-animation 0.17's `gltf2ozz` vendors the legacy tinygltf **v2.9** C++ header [16].
   - fastgltf embeds simdjson and simdutf [1], the Basis transcoder carries its own Zstd path [9], and tinyexr v3 vendors zstd and libdeflate [7]. A composed engine can therefore link the same library two or three times.
3. **[Verified]** Several commonly recommended libraries are dormant:
   - xatlas: no commits since 2022-07-26 [14].
   - refl-cpp: last release 2022-11 [35].
   - cereal: no release since 2022-02 [27].
   - bitsery: no commits in 12 months [28].
   - ACL: last release 2023-12, and no commits on `develop` since 2025-09 [17].
   - bc7enc_rdo: code unchanged since 2021; its README now points to "bc7f" in Basis Universal [10].

   Many healthy libraries are effectively single-maintainer: meshoptimizer (top author 99.7% of commits), EnTT (99%), tinyexr (91%), ufbx (83%) [git analysis].
4. **[Verified]** C++26 static reflection is in the standard text but not yet portable:
   - P2996 is in N5050 (2026-06-01), "the final draft for C++26" that "forms the basis of the Draft International Standard", with `__cplusplus` = `202603L` [36].
   - GCC 16.1 (2026-04-30) ships it behind `-freflection` ("Enable experimental C++26 Reflection") [37].
   - Upstream Clang's status page still says "No" [38], and the only Clang implementation is Bloomberg's experimental fork [39].
   - glaze, reflect-cpp and simdjson already ship opt-in P2996 modes that need GCC 16 or 16.2 [32][34][30].
   - MSVC status is **[Unverified]**.
5. **[Verified]** The text slot gained a third rendering approach. HarfBuzz 14.0.0 (2026-04-01) added `libharfbuzz-gpu`, which does GPU text rasterization based on the Slug algorithm with GLSL, WGSL, MSL and HLSL shader sources. 15.0.0 (2026-10-22) declared it stable under HarfBuzz's API/ABI guarantees [21]. It competes with msdfgen-style atlases [18][19] and FreeType bitmaps [20].
6. **[Verified]** The integration seams vary along axes Automap must model explicitly:
   - Allocator hooks may be per-call (ufbx, cgltf, zstd, ACL, FlatBuffers), process-global (`meshopt_setAllocator`, ozz `SetDefaulAllocator`, SQLite `SQLITE_CONFIG_MALLOC`) or compile-time macros (stb) [4][2][41][17][25][13][16][40][6].
   - Error models differ: Expected/result codes (fastgltf, cgltf, ufbx), exceptions plus RTTI (cereal), or throw-by-default (nlohmann) [1][2][4][27][31].
7. **[Verified]** The reference pipelines converge on the same contracts:
   - **Path-independent stable IDs:** Godot UIDs in the resource header, the `.import` file or a `.uid` sidecar; O3DE source UUIDs plus builder sub-IDs, with optional `.meta` sidecars; Unity `.meta` GUIDs; ezEngine asset GUIDs.
   - **Versioned importers:** Godot `get_format_version`, Unity `ScriptedImporter(version, …)`, O3DE builder fingerprints, AssetCooker `Version`.
   - **Declared dependencies:** Unity `DependsOn*`, O3DE source/job/product dependencies, Bevy `full_hash`.
   - **Reachability-driven packaging:** O3DE seed lists, Godot "export selected scenes (and dependencies)".

   Sources: [47][48][49][50][51][56][57][58].
8. **[Verified]** No mature, cross-platform, standalone C++ asset cooker exists. jlaumon's AssetCooker exists and is maintained, but it is Windows-only (it depends on NTFS USN journals), at v0.0.2 (2026-09-29), and MPL-2.0 [56]. All production-grade pipelines found are embedded in an engine.
9. **[Verified]** Upstream projects now publish agent-facing material, and some forbid agent contributions:
   - DirectXTex ships consumer-facing `SKILL.md` usage skills, added 2026-06-06 [11].
   - SQLite, HarfBuzz, xxHash, LZ4, assimp, tinygltf and tinyexr ship contributor-facing `AGENTS.md` files [40][21][43][42][5][3][7].
   - stb and PhysicsFS forbid AI-generated contributions [6][24].
10. **[Inference]** Treat serialization and reflection as ordinary slots with contracts. Model the asset pipeline as a cross-cutting **contract layer**: identity, determinism, versioned cache keys, dependency declaration, handles and reload, packaging manifest. Back it with tiered reference recipes rather than one swappable "pipeline slot". Leave the authoring data model (Truth- or DC-like) as custom code with guidance.

---

## Part 1: Asset libraries

### 1.1 Profiles (identity, release, health, reach)

Health = commits / distinct authors / top author's share on the default branch over the 12 months to 2026-10-29 (my git analysis). "Tool" means import or cook time; "RT" means shipped in the runtime.

| Library | License | Flavor | Latest release (date) | Health 12 mo | Public platform reach | Tool / RT |
|---|---|---|---|---|---|---|
| **fastgltf** [1] | MIT | C++20 on main (0.9.x is the last C++17 line); optional C++20 module | v0.9.1 (2026-09-27) | 230 / 19 / 77% | CI on x64, ARM, macOS; Android APK asset loading | RT loader and exporter |
| **cgltf** [2] | MIT | C99 single header, reader and writer | v1.15 (2025-02-09); master 2026-02-02 | 2 / 2 (recent PRs from the meshoptimizer author) | Portable C | RT |
| **tinygltf** (contrast) [3] | MIT | **v3: C11 POD API**, arena memory, no STL, RTTI or exceptions; C++ API moved to `attic/` | v3.0.1 (2026-07-31) | 120 / 4 / 65% | WASM demo; freestanding JSON tests | RT and tool |
| **ufbx** [4] | MIT or Unlicense | C99 single source (also builds as C++) | v0.23.1 (2026-09-25) | 153 / 4 / 83% | CI is bit-exact on Windows, macOS, Linux (x86, ARM, PPC) and WASI | RT and tool; aims at FBX SDK parity; fuzzed |
| **assimp** (contrast) [5] | BSD-3 | C++17 plus C API; 40+ formats | v6.0.5 (2026-04-30) | 275 / 92 / 19% | Desktop, Android, iOS | Tool |
| **stb_image** [6] | MIT or Unlicense | C single header | No tags; header v2.30; master 2026-10-22; `dev` branch 2026-10-29 | 21 / 4 on master, plus 59 security-fix commits on `dev` by F. Giesen in Oct 2026 | Portable | RT and tool; header warns about DoS on untrusted data |
| **tinyexr** [7] | BSD-3 | **v3: pure C11** (v1 C++ header deprecated) | v3.2.0 (2026-07-09) | 496 / 5 / 91%; PRs only from one-time sponsors of US$5,000 or more | WASM, freestanding core, SIMD dispatch | RT and tool |
| **KTX-Software / libktx** [8] | Apache-2.0 (plus listed third-party licenses) | C library plus unified `ktx` CLI | Stable v4.4.2 (2025-10-04); **v5.0.0-rc2 (2026-08-17)**, pending KTX spec Rev. 5 | 152 / 18 / 64% | Desktop; JS/Emscripten, Java and Python bindings | RT (load and transcode) and tool |
| **Basis Universal** [9] | Apache-2.0 | C++ encoder plus transcoder | **v2.5** (tag `v2_50`, 2026-07-22) | 577 / 8 / 61% | Native, WASM (web and WASI), early Python | Encoder is tool; transcoder is RT |
| **bc7enc_rdo** [10] | MIT or Unlicense (bc7e.ispc is Apache-2.0) | C++ / ISPC | No tags; code unchanged since 2021; README (2026-01) points to bc7f | 4 / 1 (README only) | Desktop | Tool, effectively superseded |
| **DirectXTex** [11] | MIT | C++, Windows-first | `may2026` (2026-05-08) | 91 / 13 / 45% | VS 2022/2026, clang-cl, MinGW; non-Win32 builds need DirectX-Headers; Xbox extension build options | Tool plus RT DDS loaders |
| **Compressonator** [12] | MIT (core/SDK license files) | C/C++ SDK (`CMP_Core`), CLI, GUI | V4.6.0 (2026-10-19) | 1 / 1 (squashed code drops) | Windows, Linux | Tool |
| astc-encoder (alt.) [63] | Apache-2.0 | C++ | 5.7.0 (2026-07-31) | n/a | n/a | Tool |
| **meshoptimizer** [13] | MIT | C API (C++ templates optional) | v1.3 (2026-09-25); v1.0 shipped 2025-12-06 | 672 / 3 / 99.7% | Portable | Optimize, simplify and meshlets at tool time; decode at RT |
| **xatlas** [14] | MIT | C++11 | No tags; **last commit 2022-07-26** | 0 | Portable | Tool, dormant |
| UVAtlas (alt.) [15] | MIT | C++ | `oct2025` (2025-11-01) | n/a | Windows-centric | Tool |
| **ozz-animation** [16] | MIT | C++ with its own SoA math and IO archives | 0.17.0 (2026-08-01) | 46 / 2 / 87% | Desktop plus web samples | Offline builders are tool; sampling and blending jobs are RT |
| **ACL** [17] | MIT | C++11 header-only; RTM math | v2.1.0 (2023-12-05); `develop` 2025-09-13 | **0** on `develop` | CI: Windows (x86, x64, ARM64, ARM64EC), Linux, macOS ARM64, Emscripten; manual iOS/Android; shipped inside Unreal | Compress is tool; decompress is RT |
| **msdfgen** [18] | MIT | C++ (`MSDFGEN_CORE_ONLY` has no dependencies) | v1.13 (2025-11-30); `+new-skia-api` 2026-03-05 | 6 / 2 | Portable | Mostly tool; core usable at RT |
| **msdf-atlas-gen** [19] | MIT | C++ | v1.4 (2026-03-05) | 5 / 1 | Portable | Tool |
| **FreeType** [20] | FTL or GPLv2 (choose one) | C | 2.15.0 (2026-10-12) | 247 / 30 / 39% | Universal | RT |
| **HarfBuzz** [21] | "Old MIT" | C API, C++ implementation | **15.0.0 (2026-10-22)** | 2488 / 61 / 82% | Universal, WASM | RT; adds `-raster`, `-vector` and **`-gpu`** libraries |
| **stb_truetype** [6] | MIT or Unlicense | C single header | v1.26 | (stb) | Portable | RT; "NO SECURITY GUARANTEE — DO NOT USE THIS ON UNTRUSTED FONT FILES" |
| **efsw** [22] | MIT | C++ | 1.7.3 (2026-10-28) | 125 / 10 / 81% | inotify, IOCP, FSEvents/kqueue, BSD kqueue, generic polling fallback | Dev-time hot reload |
| **dmon** [23] | BSD-2 | C single header | No tags; last commit 2026-02-23 (FreeBSD support added) | 5 / 4 | Windows, Linux, macOS, FreeBSD | Dev-time |
| **PhysicsFS** [24] | zlib | C | 3.2.0 (2022-09-30); `main` is 3.3.0-dev | 69 / 11 / 43% | Portable; SDL3 bridge in `extras/` | RT VFS (archives, mounts, write dir) |

### 1.2 Integration seams

| Library | Allocator hook | Threading | Math types | Ownership of results | Errors | Notable build options / RTTI / exceptions |
|---|---|---|---|---|---|---|
| fastgltf [1] | Only buffer and image allocation (map/unmap callbacks); does not hook all heap allocations | `Parser` "is not thread-safe" (one per thread); base64 decode callback lets you run decoding on your own scheduler | Own `fastgltf::math`; `glm_element_traits.hpp` and `dxmath_element_traits.hpp` adapters for accessor copies | `Asset` owns std-container data; GPU-mapped writes possible | `Expected<T>` with an `Error` enum; noexcept factories | `FASTGLTF_USE_64BIT_FLOAT`, `FASTGLTF_ENABLE_CPP_MODULES`; embeds simdjson and simdutf |
| cgltf [2] | `cgltf_memory_options.alloc_func/free_func`, plus file `read/release` callbacks | n/a | Plain floats | `cgltf_data` freed by `cgltf_free` | `cgltf_result` enum | Parses EXT and KHR meshopt compression descriptors; decoding is left to meshoptimizer |
| tinygltf v3 [3] | Arena; one `tg3_model_free()` frees everything | n/a | POD | Arena-owned | `tg3_error_stack` with severity and source locations | `TINYGLTF3_ENABLE_FS` and `…_STB_IMAGE` off by default; fuzzed; no RTTI or exceptions |
| ufbx [4] | `ufbx_allocator_opts`: alloc/realloc/free plus **memory and allocation limits**; the allocator is stored in the scene | `ufbx_thread_pool` callbacks (init, run, wait, free) plug into an engine job system | `ufbx_vec3`, `ufbx_quat` etc., with `UFBX_*_CONVERSION` macros to user types; `UFBX_REAL_IS_FLOAT` | Plain-data scene; `ufbx_free_scene` | `ufbx_error`; progress and cancel callbacks | Single `.c` file; sanitizes scenes by default |
| assimp [5] | Custom `IOSystem`/`IOStream` via `SetIOHandler()` before `ReadFile()` (I/O, not memory) | n/a | Own `ai*` types | n/a | n/a | Many `ASSIMP_BUILD_*_IMPORTER` options, `ASSIMP_DOUBLE_PRECISION`, `ASSIMP_NO_EXPORT` |
| stb_image [6] | `STBI_MALLOC/REALLOC/FREE` macros (compile-time, global) | Thread-local failure reason since 2.24 | n/a | Caller frees with `stbi_image_free` | NULL plus `stbi_failure_reason` | `STBI_NO_STDIO`, `STBI_ASSERT`, `STBI_NO_FAILURE_STRINGS` |
| tinyexr v3 [7] | Allocator hook | C11 threads, opt-in | n/a | n/a | n/a | Vendored zstd and libdeflate; libdeflate dropped in freestanding and WASM builds |
| libktx [8] | No allocator hook found in `ktx.h`; custom I/O via `ktxStream` | Basis encode takes a thread count | n/a | `ktxTexture` objects | `KTX_error_code` | v5: `ktxBasisParams.uastc` became `codec` (ETC1S, UASTC LDR 4x4, UASTC HDR 4x4, UASTC HDR 6x6 Intermediate); vendors basis_universal v2.1 and astc-encoder |
| Basis [9] | n/a | WASM multithreading supported | n/a | n/a | n/a | `basisu_transcoder_init()` **must** be called first (global tables); `BASISD_SUPPORT_KTX2_ZSTD` controls the built-in Zstd |
| DirectXTex [11] | n/a | OpenMP (`BC_USE_OPENMP`) | DXGI formats | n/a | n/a | `BUILD_DX11`, `BUILD_DX12`, `ENABLE_OPENEXR_SUPPORT`, `BUILD_XBOX_EXTS_*`; NuGet packages deprecated in favor of vcpkg |
| Compressonator [12] | n/a | n/a | n/a | n/a | n/a | V4.6 hardened DDS, TGA and BC1–BC7 parsing; release CLI and GUI refuse to run unless DLLs are AMD-signed |
| meshoptimizer [13] | `meshopt_setAllocator` is **process-global**, with stack-order allocate/free; experimental `meshopt_setAllocatorCache` | Encoder version setters "not thread safe" | `float*` plus stride | Caller-provided buffers | Return sizes or 0 | `meshopt_encodeIndexVersion` / `…VertexBufferLevel` select bitstream versions; v1.x adds `buildMeshletsFlex/Spatial`, `partitionClusters`, opacity-map functions |
| xatlas [14] | `SetAlloc(realloc, free)`, global | n/a | n/a | `Atlas*` | `AddMeshError` | Progress callback |
| ozz [16] | `ozz::memory::Allocator` with **global** `SetDefaulAllocator` | n/a | Own SoA `ozz::math` | Builders produce runtime objects | n/a | `ozz_build_fbx` requires the Autodesk FBX SDK; `ozz_build_gltf` uses vendored tinygltf v2.9 |
| ACL [17] | `acl::iallocator` passed per call | n/a | **RTM** (`nfrechette/rtm`) | Caller-allocated compressed tracks | Game-provided assert macro | Header-only C++11 |
| msdfgen [18] | n/a | `MSDFGEN_USE_OPENMP` | n/a | n/a | n/a | `MSDFGEN_CORE_ONLY`, `MSDFGEN_USE_SKIA`, `MSDFGEN_USE_VCPKG` (default ON), `MSDFGEN_DISABLE_SVG/PNG` |
| HarfBuzz [21] | Build-time configuration (`CONFIG.md`) | `HB_NO_MT` / `HB_TINY` disable thread safety | n/a | Ref-counted objects | n/a | `HB_MINI`, `HB_LEAN`, `HB_TINY` profiles; amalgamated `harfbuzz-world.cc` (14.0) |
| PhysicsFS [24] | `PHYSFS_setAllocator` | "Mostly thread safe"; state setters are mutexed, but per-file reads, writes and seeks are **not** locked | n/a | n/a | Per-thread error codes | `PHYSFS_Io` lets you wrap your own streams |
| efsw [22] | n/a | Watches asynchronously; the README does not name the callback thread, so assume it is not the main thread **[Inference]** | n/a | n/a | n/a | Falls back to polling if a native backend fails |

### 1.3 Bridges and known-good pairings (verified links only)

| Bridge | Evidence | Contract implied |
|---|---|---|
| fastgltf / cgltf + meshoptimizer | fastgltf exposes `EXT_meshopt_compression` and `KHR_meshopt_compression` data [1]; cgltf parses both (a 2026 PR by meshoptimizer's author) [2]; **KHR_meshopt_compression** is a Khronos extension with status "Release Candidate", merged 2026-01-24 [46] | The loader hands compressed bufferViews to `meshopt_decode*`; decoder and bitstream versions must match the encoder |
| bgfx `geometryc` = cgltf + meshoptimizer | `tools/geometryc/geometryc.cpp` includes both [59] | A known-good cook-time pairing used in a shipping open-source renderer |
| KTX2 + Basis Universal | libktx vendors Basis (v2.1 in v5.0.0-rc2) and transcodes UASTC HDR 4x4 and HDR 6x6 to ASTC HDR or BC6H; new targets `KTX_TTF_RGB_HALF` and `KTX_TTF_RGB9E5` [8] | **Skew:** Basis v2.5 adds XUASTC LDR 4x4–12x12, XUBC7 and ASTC LDR 4x4–12x12 [9], none of which appear in the KTX rc2 notes [8]. Pin the pair, not each library |
| fastgltf `KHR_texture_basisu`, `MSFT_texture_dds`, `EXT_texture_webp/avif` | Indices exposed on textures [1] | The loader does not decode images ("Image decoding ❌") [1]; an image slot must be wired in |
| ozz + glTF | `gltf2ozz` vendors tinygltf v2.9 C++ [16] | Conflicts with tinygltf v3 (now C) if both are linked; the vendored copy is header-only |
| ozz + FBX | `fbx2ozz` needs the Autodesk FBX SDK [16] | **[Inference]** ufbx [4] is the obvious license-friendly replacement recipe |
| ACL ⇄ Unreal | README: shipped "as part of each engine release since UE 5.13" [17] (sic; probably 5.3) | ACL is production-proven but upstream activity is low |
| msdf-atlas-gen ← msdfgen ← FreeType | msdf-atlas-gen v1.4 updates to msdfgen 1.13; `MSDF_ATLAS_MSDFGEN_EXTERNAL` [19][18] | Version lockstep between atlas tool and generator |
| HarfBuzz + FreeType | HarfBuzz docs direct hinted rasterization to FreeType [21]; FreeType 2.15 OFF5 support was derived from HarfBuzz [20] | Both now support ISO OFF 5th edition (>64k glyphs, VARC) [20][21] |
| PhysicsFS + SDL3 | `extras/physfssdl3.{c,h}` [24] | VFS ↔ platform I/O adapter |
| Vendored copies of shared dependencies | fastgltf ⊃ simdjson and simdutf [1]; Basis ⊃ Zstd path [9]; tinyexr v3 ⊃ zstd and libdeflate [7]; KTX ⊃ basisu and astc-encoder [8] | **Contract:** the manifest must record vendored dependencies (name, version, symbol-prefixed or not) to detect duplicate symbols and ODR risk |

**Registry coverage (vcpkg at 2026-10-28, Conan Center and xmake-repo at 2026-10-29)** [60][61][62]:

- **Present in vcpkg (version):** fastgltf 0.9.1, cgltf 1.15, tinygltf 3.0.0, assimp 6.0.5, stb 2026-08-02, tinyexr 3.2.0, ktx **4.4.2**, basisu 2.50, directxtex 2026-05-07, meshoptimizer 1.3, msdfgen 1.13+new-skia-api, freetype **2.14.3**, harfbuzz **14.6.0**, efsw 1.7.2, physfs 3.2.0, uvatlas 2025-10-28; dmon is listed as `septag-dmon`.
- **Lag:** vcpkg trails upstream for FreeType, HarfBuzz and KTX.
- **Absent from vcpkg:** ufbx, xatlas, ozz-animation, nfrechette ACL, msdf-atlas-gen, Compressonator, astc-encoder.
- **Name collision:** vcpkg's `acl` port is *POSIX Access Control Lists*.
- **Conan:** has `nfrechette-acl`, `ozz-animation`, `msdf-atlas-gen`, `fastgltf`, `cgltf`, `ktx`, `libbasisu`; lacks ufbx and xatlas.
- **xmake-repo:** has ufbx, ozz-animation, astc-encoder; lacks xatlas.

---

## Part 2: Data definition and serialization

### 2.1 Profiles

| Library | License | Flavor | Latest release (date) | Health 12 mo | Key seams |
|---|---|---|---|---|---|
| **FlatBuffers** [25] | Apache-2.0 | Schema IDL plus `flatc` codegen; zero-copy reads | v25.12.19 (2025-12-19; re-tag 2026-02-06); no newer tag | 164 / 75 / 13% | `flatbuffers::Allocator` (allocate, deallocate, reallocate_downward) per builder; verifier for untrusted buffers; flexbuffers is the schemaless variant |
| **Cap'n Proto** [26] | MIT | Schema plus codegen; zero-copy; RPC; KJ base library | v1.5.0 (2026-07-09); `master` is **2.0-dev** | 431 / 39 / 38% | `CAPNP_LITE` drops reflection APIs; KJ uses exceptions, and under `-fno-exceptions` it "logs an error and returns" |
| **cereal** [27] | BSD-3 | Header-only C++11 archives (binary, JSON, XML) | v1.3.2 (2022-02-27) | 3 / 3 | **Throws** `cereal::Exception`; uses `typeid` for versioning and polymorphism, so RTTI is required; `CEREAL_CLASS_VERSION` |
| **bitsery** [28] | MIT | Header-only C++11, no IDL, no macros | v5.2.5 (2025-10-08) | **0** | Configurable error checking, endianness, extensions for forward/backward compatibility, allocator support for pointers |
| **yyjson** [29] | MIT | C | 0.13.0 (2026-09-09) | 126 / 28 / 35% | `yyjson_alc` allocator passed per call; `YYJSON_DISABLE_{READER,WRITER,INCR_READER,FILE,UTILS}` |
| **simdjson** [30] | Apache-2.0 or MIT | C++17 On-Demand parser plus builder | v5.0.8 (2026-10-29); v5.0.0 2026-09-28 | 343 / 60 / 64% | Usable with or without exceptions; input needs SIMD padding (handled automatically for string inputs); **C++26 reflection** (de)serialization benchmarked on GCC 16.2 |
| **nlohmann/json** [31] | MIT | C++11 header-only DOM | v3.12.0 (2025-04-11) | 610 / 57 / 66% | Exceptions by default (`JSON_NOEXCEPTION` opt-out); allocator template parameter |
| **glaze** [32] | MIT | **C++23**, compile-time reflection | v9.1.0 (2026-10-28) | 916 / 45 / 58% | **No exceptions** (helpers optional); JSON, BEVE, CBOR, MsgPack, TOML 1.1, YAML, CSV; MSVC needs `/Zc:preprocessor`; **P2996 mode** on GCC 16+ or clang-p2996 |
| **EnTT meta** [33] | MIT | Runtime reflection (`entt::meta_factory`); C++20 since v4 | v4.0.0 (2026-07-22) | 810 / 8 / 99% | Swappable or multiple `meta_context`s; built-in RTTI system (`entt::type_info`) |
| **reflect-cpp** [34] | MIT | C++20 aggregate reflection (default) or opt-in C++26 mode | v0.26.2 (2026-10-19) | 128 / 40 / 41% | 15+ formats through adapters (Avro, BSON, Cap'n Proto, CBOR, cereal, flexbuffers, msgpack, TOML, XML, YAML, …); C++26 mode needs `REFLECTCPP_USE_CPP26_REFLECTION` and **GCC 16.2+** (`-freflection`) or clang-p2996 (`-freflection-latest`) |
| **refl-cpp** [35] | MIT | C++17 macros | v0.12.4 (2022-11-05) | **0** | Dormant |
| **SQLite** [40] | Public domain | C | 3.54.1 (2026-10-16) | 2227 / 4 / 64% | `SQLITE_CONFIG_MALLOC`; single-thread, multi-thread or serialized modes; VFS registration for platform I/O; `ext/wasm`; no PRs without prior agreement |
| **zstd** [41] | BSD-3 or GPLv2 | C | v1.5.7 (2025-02-19); `dev` reports 1.6.0 | 147 / 19 / 60% | `ZSTD_customMem` per context (static-linking-only API); threading needs `ZSTD_MULTITHREAD`; dictionaries |
| **LZ4** [42] | BSD-2 (lib), GPLv2+ (CLI) | C | v1.10.0 (2024-07-21) | 528 / 23 / 86% | A user-memory-functions branch merged 2026-10-28 (unreleased) |
| **Oodle** [45] | Proprietary | C SDK | **[Unverified]** | n/a | See below |
| **xxHash** [43] | BSD-2 | C, header-only option | v0.8.4 (2026-09-18) | 231 / 16 / 80% | XXH3 is the recommended default; outputs "remain stable across xxHash releases"; `xxhash.h` is generated |
| **BLAKE3** [44] | CC0, Apache-2.0 or Apache-2.0 with LLVM exception | C and Rust | 1.8.7 (2026-08-20) | 34 / 11 / 71% | C hashing is single-threaded unless built with `BLAKE3_USE_TBB` (oneTBB); SIMD up to AVX-512 and NEON |

**Oodle [Unverified].** I could not reach radgametools.com or Epic's documentation (see Method). From prior knowledge only: RAD Game Tools joined Epic Games in 2021. Oodle Data, Texture and Network are proprietary SDKs, licensed commercially and bundled for Unreal Engine licensees. That makes Oodle the archetype of a "private registry" library: binary or source-under-license, with redistribution constraints. Automap's manifest must be able to express it without publishing it.

### 2.2 C++26 static reflection (P2996): state on 2026-10-29

| Item | Status | Source |
|---|---|---|
| Standard text | P2996R13 was adopted at the 2025 Sofia meeting. N5050 (2026-06-01) is "the final draft for C++26 … basis of the Draft International Standard", with `__cplusplus` = `202603L`. N5054 (2026-07-16) is the first C++29 working draft | [36][38][39] |
| ISO publication / DIS ballot result | **[Unverified]**. The editors' reports confirm only that the DIS is based on N5050; I found no ballot outcome | [36] |
| GCC | 16.1.0 tagged 2026-04-30, 16.2.0 tagged 2026-08-07; `-freflection` = "Enable experimental C++26 Reflection." (off by default); ships `<meta>` and `gcc/cp/reflect.cc` | [37] |
| Clang (upstream) | `cxx_status.html` on main lists P2996R13 and P3394R4 as **"No"**; latest release llvmorg-23.1.4 | [38] |
| Clang fork | Bloomberg clang-p2996: "highly experimental … DO NOT use this project to build any artifacts destined for production"; `p2996` branch active (2026-09-24) | [39] |
| MSVC, EDG | **[Unverified]**; vendor blogs could not be reached | n/a |
| Library adoption | glaze (GCC 16+ / clang-p2996) [32]; reflect-cpp (GCC 16.2+) [34]; simdjson (C++26 "compile-time parsing" and reflection examples) [30] | |

**[Inference]** In late 2026, P2996 can serve as an *optional, GCC-only accelerator* behind a C++20 fallback, which is exactly how glaze and reflect-cpp ship it. It cannot yet be a baseline contract for an engine that must build with MSVC or console toolchains.

---

## Part 3: Asset-pipeline patterns and reference systems

### 3.1 Reference systems

| System | Identity | Import / cook split | Cache and invalidation | Dependencies | Runtime loading / reload | Packaging | Verification |
|---|---|---|---|---|---|---|---|
| **Godot 4.x** (master is 4.8-beta) | `uid://` (int64 `ResourceUID::ID`) stored in the `.tscn`/`.tres` header, in the `<asset>.import` file, or in a `<file>.uid` sidecar for loaders with custom UID support [48][49] | Source plus `.import` (committed) → hidden `.godot/imported/*-<md5>.ctex` (not committed); "If any of the files present in this folder is erased … reimported automatically" [47] | Reimports when the source's MD5 changes [47]; importer `get_format_version()` [49] | Scene `ext_resource` entries carry uid and path [48] | `ResourceLoader.load_threaded_request` / `…_get_status`; cache modes [48] | Export modes include "Export selected scenes (and dependencies)" [48]; `.import` lists per-feature variants (`path.<feature>`), e.g. per texture-compression format [49] | Verified from godot-docs and source |
| **O3DE** | Source UUID generated from the scan-folder-relative path; product ID = source UUID + builder-defined **sub-ID** ("must be deterministic and unique"); optional experimental `.meta` sidecar holding a UUID plus `legacyUuids` for relocation [51] | Asset Builders run CreateJobs (single-threaded) and ProcessJob (multi-threaded) and write products to the Asset Cache [51] | Job fingerprint built from source and dependency timestamps plus builder versions; changing a builder's fingerprint or UUID reprocesses its assets; Asset Cache Server shares product ZIPs across a team [51] | Three kinds: **source** (process-time), **job** (needs another job's products; can name specific sub-products) and **product** (runtime and bundling) [51] | `AZ::Data::Asset<T>` smart-pointer handle; `AssetManager::GetAsset`; `OnAssetReady`; dependency load modes Preload, QueueLoad, NoLoad; **hot reload**, with the warning that copied-out data is not refreshed [51] | Asset Bundler with **seed lists**; product dependencies drive inclusion [51] | Verified from o3de.org source |
| **Unity** | `.meta` sidecar holds the GUID and importer settings (source comments: "its .meta GUID rewritten on disk", "stored in the .meta file") [50] | Importers produce artifacts; `ScriptedImporter(int version, …)` [50] | `ArtifactKey`; importer version is declared in `[ScriptedImporter(version, ext)]` [50]; reimport-on-version-bump semantics and the Accelerator/cache server are **[Unverified]** | `AssetImportContext.DependsOnSourceAsset(path or GUID)`, `DependsOnArtifact`, `DependsOnCustomDependency` plus `AssetDatabase.RegisterCustomDependency(name, Hash128)` [50] | **[Unverified]** (Addressables docs not reached) | **[Unverified]** | Partly verified (UnityCsReference 7000.0.0a12) |
| **Unreal** (DDC, Zen, cook) | **[Unverified]** | **[Unverified]**; from prior knowledge, "cooking" converts editor packages into per-platform runtime data | **[Unverified]**; from prior knowledge, the Derived Data Cache is a key/value store keyed by input hashes plus per-type version GUIDs, with local, shared and cloud tiers. Zen Storage Server is a local service for DDC and cooked output | **[Unverified]** | **[Unverified]** | **[Unverified]**; from prior knowledge, Asset Manager rules and primary-asset labels drive chunking | Tried: `EpicGames/UnrealEngine` and `EpicGames/zen` need authenticated GitHub access; Epic docs were not reached |
| **Naughty Dog DC** | n/a | "DC, a custom data scripting system developed in Racket"; a "multitude of domain specific languages" for gameplay scripting, animation, effects and sound on PS3 (Uncharted, The Last of Us) [54] | **[Unverified]** | **[Unverified]** | **[Unverified]** | **[Unverified]** | Verified from the RacketCon 2013 keynote abstract; GDC talk details not reached |
| **Our Machinery "The Truth"** | **[Unverified]** | **[Unverified]** | **[Unverified]** | **[Unverified]** | **[Unverified]** | **[Unverified]** | ourmachinery.com not reached; `OurMachinery/*` GitHub repos are private or nonexistent. From prior knowledge: an in-memory, typed object/property store with 64-bit object IDs, immutable read snapshots and atomic write commits, which yields undo, change notification and collaboration |
| **ezEngine** | Asset GUID assigned at creation and never changed; duplicating files by copy-paste is warned against [58] | Asset documents are "transformed" into runtime data [58] | n/a | n/a | n/a | Per-platform **asset profiles** [58] | Verified from docs-src |
| **Bevy** (Rust; pattern reference) | Path-based; `.meta` (RON) holds loader or processor plus settings, not a GUID [57] | `AssetProcessor`, a "build system for assets": `LoadTransformAndSave` or `Process` [57] | `ProcessedInfo.hash` (asset bytes + `.meta`) and `full_hash` (plus every process dependency's `full_hash`); stated values: automatic, configurable, lossless, **deterministic** [57] | `process_dependencies` [57] | Optional `file_watcher` reprocesses on change [57] | n/a | Verified from source (0.20.0-dev) |
| **AssetCooker** (jlaumon) | Path rules over "Repos" | Rules in TOML or Lua run command lines (e.g. texconv), with declared `OutputPaths` [56] | NTFS **USN journal** change tracking; a rule `Version` "to force all commands to run again" [56] | DepFiles in its own format or Make `.d` format [56] | n/a | `NoOrphanFiles` and "Find Orphan Files" [56] | Verified; MPL-2.0, Windows-only ("USN Journals are a Windows' only feature"), v0.0.2 |
| **bgfx tools** | n/a | Standalone `geometryc` and `shaderc` cookers, plus `texturev`/`geometryv` viewers [59]; texture cooking (`texturec`) is not in `bgfx/tools` (believed to live in the bimg repo; **[Unverified]**) | None (plain CLIs) | n/a | n/a | n/a | Verified tool list |

### 3.2 Patterns, and the contract each imposes on libraries and slots

| Pattern | Seen in | What a library or slot must honor | Library facts that matter |
|---|---|---|---|
| **Stable identity independent of path** | Godot UID [48][49]; O3DE UUID + sub-ID and `.meta` [51]; Unity `.meta` GUID [50]; ezEngine GUID [58] | Importers must emit **deterministic sub-IDs** for multi-product sources (a mesh file → meshes, materials, skeleton, clips); runtime handles key on ID, not path | ufbx and fastgltf expose stable element indices, not GUIDs; ID derivation is glue code. **[Inference]** |
| **Import/cook split** (source → intermediate → platform product) | Godot `.import` → `.godot/imported` [47]; O3DE builders [51]; ezEngine transform [58]; Bevy processor [57] | Libraries tagged **tool-time vs runtime**; runtime builds must not need tool-only dependencies | Split libraries: Basis encoder/transcoder [9], ACL compress/decompress [17], meshoptimizer encode/decode [13], ozz offline/runtime [16]; tool-only: DirectXTex, Compressonator, xatlas, msdf-atlas-gen [11][12][14][19] |
| **Hash-keyed derived-data cache** | Godot md5 [47]; Bevy `hash`/`full_hash` [57]; O3DE fingerprints and Asset Cache Server [51]; Unreal DDC **[Unverified]** | **Deterministic output** for (input bytes, settings, *library version*, *bitstream/format version*); thread count must not change output; libraries must expose a version string usable in cache keys | meshoptimizer exposes `MESHOPTIMIZER_VERSION` and explicit encoder versions [13]; xxHash guarantees stable outputs across releases [43]; whether zstd and Basis output is bit-stable across thread counts and versions was **not verified** |
| **Versioned importers** | Godot `get_format_version` [49]; Unity `ScriptedImporter` version [50]; O3DE builder fingerprint [51]; AssetCooker `Version` [56] | Each recipe's importer declares a version that changes whenever the output changes, including on library upgrades | Library version bumps (KTX 4→5, tinygltf 2→3) should invalidate caches automatically. **[Inference]** |
| **Explicit dependency graph** | O3DE source/job/product deps [51]; Unity `DependsOn*` [50]; Bevy process deps [57]; AssetCooker DepFiles [56] | Importers must *report* every file they read (glTF external buffers and images, MTL files, fonts); I/O must go through injectable callbacks so reads can be recorded | Hooks exist: cgltf `file.read` [2], fastgltf `LoadExternal*` options [1], ufbx `ufbx_open_file_cb` [4], PhysicsFS `PHYSFS_Io` [24], libktx `ktxStream` [8] |
| **Handle-based async loading with hot reload** | O3DE `Asset<T>` + `OnAssetReady` + hot reload [51]; Godot threaded load [48] | A handle type with load states; a reload callback; a rule on whether consumers may cache copies (O3DE warns they go stale) [51]; a file watcher | efsw / dmon [22][23]; watcher callbacks run on background threads, so a reload queue is needed. **[Inference]** |
| **Reachability-driven packaging** | O3DE seed lists + product deps [51]; Godot "selected scenes (and dependencies)" [48]; AssetCooker orphan detection [56] | Products must declare runtime dependencies; hard-coded loads must be declared as seeds [51] | Archive or VFS slot (PhysicsFS) mounts the packaged output [24] |
| **Platform variants** | Godot `path.<feature>` [49]; O3DE per-platform jobs [51]; ezEngine profiles [58] | Texture and mesh recipes are parameterized by target (BCn/ASTC/ETC; meshlets vs not) | KTX2 + Basis transcode at runtime reduces the variant count [8][9]; DirectXTex Xbox options are build-time [11] |
| **Authoring data model** (Truth, DC) | DC [54]; Truth **[Unverified]** | Schema evolution, reflection, undo and diff; this is where serialization and reflection choices bite | P2996 not yet portable [36][38]; EnTT meta / reflect-cpp / glaze are C++20/23 fallbacks [33][34][32] |

### 3.3 Should data definition and the asset pipeline be first-class slots? (inference)

**Strongest argument for first-class slots with contracts.** Every pipeline examined, in four engines and three standalone tools, independently converges on the same small set of contracts: stable IDs, versioned deterministic importers, declared dependencies, content-hash cache keys, handles with reload, and reachability manifests [47]–[58]. Those contracts are exactly what makes the Part 1 libraries composable. A texture recipe is cacheable only if Basis/KTX output is deterministic and versioned. A glTF importer can participate in dependency tracking only if its I/O is injectable. Hot reload works only if the runtime handle contract is honored. If Automap does not name these contracts, an agent cannot check them, and agents need them most where code is custom: a contract is the test oracle for generated glue. Serialization is the same: exception and RTTI policy, schema-evolution rules and codegen steps cascade into every other slot, much as math types do.

**Strongest argument for custom code with recipes.** No reusable library actually *fills* an "asset pipeline" slot:
- AssetCooker is Windows-only at v0.0.2 [56].
- Every production pipeline is embedded in its engine's object model (Godot resources, O3DE `AZ::Data`, Unity artifacts) [47]–[51].
- The authoring data model is the most opinionated part of an engine (DC is a Racket DSL platform [54]).

A slot implies interchangeable implementations. Here there would be one, the user's, so a slot would be fiction plus a large standards effort with high risk of premature lock-in. Recipes and guidance carry the patterns without pretending they are swappable.

**Synthesis.** Split the question three ways:
1. **Serialization format, JSON and reflection are real slots.** There are interchangeable, maintained libraries with clear seams (Part 2).
2. **The asset pipeline is a contract layer, not a slot.** It is cross-cutting like allocators or threading. Every asset library gets conformance metadata (below), and each tier gets a reference recipe that composes them.
3. **The authoring data model and editor stay custom**, with pattern guidance (Truth, DC, Godot resources, O3DE).

---

## Leads checked

| Lead | Verdict | Notes and sources |
|---|---|---|
| Candidate asset libraries: fastgltf/cgltf, ufbx, stb_image, meshoptimizer, KTX/Basis Universal, xatlas, ozz-animation, msdfgen | **Corrected** | All exist and most are healthy [1][2][4][6][13][8][9][16][18], with four corrections: **xatlas is dormant** (last commit 2022-07-26) [14]; cgltf has had no release since 2025-02 and is low-activity [2]; KTX v5 is still RC and vendors Basis v2.1, behind Basis v2.5 [8][9]; HarfBuzz-gpu is a new text option [21] |
| Glue: zstd/LZ4, xxHash/BLAKE3, SQLite, FlatBuffers | **Confirmed**, with caveats | All maintained [41][42][43][44][40][25]; zstd has had no release since 2025-02 [41], LZ4 none since 2024-07 [42], FlatBuffers no tag since 2025-12 [25] |
| Proven patterns: GUID sidecars (Unity .meta), import/cook split, hash-keyed DDC (Unreal DDC/Zen), dependency graphs, handle-based async loading with hot reload, reachability-driven packaging | **Confirmed except Unreal** | Godot, O3DE, Unity (partly), Bevy, ezEngine and AssetCooker verified [47]–[51][56]–[58]; **Unreal DDC/Zen could not be verified** (Epic docs and repos not reachable). Note that Godot does **not** use a GUID sidecar for most assets; the UID lives in the resource header or `.import`, with `.uid` sidecars only for some types [48][49] |
| Reference systems: Godot importer, O3DE Asset Processor, Naughty Dog DC, Our Machinery "The Truth" | **Partly confirmed** | Godot and O3DE verified in depth [47]–[51]; DC's existence and purpose verified from the RacketCon 2013 abstract [54]; **The Truth could not be verified** (site and repos unreachable) |
| jlaumon's AssetCooker exists and is maintained | **Confirmed**, with limits | Exists; last commit 2026-10-04; v0.0.2 (2026-09-29); Windows-only; MPL-2.0 [56] |
| "Custom code concentrates in gameplay, tools, the renderer frontend, the asset pipeline, and data definition" | **Partly supported (inference)** | Consistent for the asset pipeline and data model, given the absence of reusable standalone pipelines [56] and engine-embedded pipelines [47]–[51]. Gameplay and renderer frontend were not studied here |

---

## Implications for Automap (inference)

1. **The manifest needs per-library fields beyond name@version:**
   - `api_flavor`: C, C++17, C++20, C++23, plus an optional C++26 mode.
   - `role`: tool, runtime or both.
   - `alloc_hook`: per-call, global or compile-time.
   - `thread_safety`: reentrant, per-instance or global state.
   - `error_model`: result codes, Expected, or exceptions.
   - `rtti_required`, `exceptions_required`, `untrusted_input_safe` (contrast ufbx, tinygltf v3 and tinyexr v3, which are fuzzed, with stb_truetype's warning) [4][3][7][6].
   - `vendored_deps` [1][7][8][9].
   - `determinism` and `format_version` for cacheability [13][43].
2. **Pin bridges, not just libraries.** "KTX 5.0-rc2 + Basis 2.1" and "Basis 2.5 standalone" are different pairings with different codec sets [8][9]. Known-good pairings should be first-class records (e.g. cgltf + meshoptimizer as in bgfx [59]).
3. **Compute health signals from git and record them.** Dormant entries (xatlas, refl-cpp, cereal, bitsery, ACL) need explicit replacement recipes (UVAtlas, reflect-cpp/glaze, ufbx for FBX) and a "maintenance risk" flag. Single-maintainer concentration is common and worth surfacing [git analysis].
4. **Registry naming is ambiguous.** vcpkg's `acl` is POSIX ACL. Map entries need canonical IDs plus per-registry aliases (`nfrechette-acl` on Conan) and an availability matrix, since ufbx, ozz and xatlas are missing from vcpkg [60][61][62].
5. **Ingest upstream agent material and respect contribution policy.** Prefer upstream skills where they exist (DirectXTex `skills/`) [11]. Carry a `contribution_policy` field so an agent never proposes upstream PRs to stb or PhysicsFS (AI-generated contributions banned) [6][24], and knows SQLite takes no PRs without agreement [40] and tinyexr accepts PRs only from sponsors [7].
6. **Asset pipeline as a contract layer.** Define contracts (`AssetId`, `ImporterVersion`, `DeclaredDeps`, `CacheKey`, `Handle<T>` + `on_reload`, `ProductDeps`) and per-tier reference recipes:
   - *testbed*: runtime loading straight from source, no cook;
   - *starter*: import/cook split with a hash-keyed local cache and file watching;
   - *custom*: GUID sidecars or headers, a shared cache, reachability packaging, platform variants, and private or console registries.
7. **C++ standard floor is a tier decision.** C++17 rules out fastgltf ≥1.0 and EnTT v4. C++23 is required for glaze. C++26 reflection is GCC-only and experimental [1][33][32][37][38].
8. **The text slot should list three approaches:** FreeType + HarfBuzz raster atlases, msdfgen/msdf-atlas-gen distance fields, and HarfBuzz-gpu (Slug) outlines [20][18][19][21].

## Open questions

- MSVC and EDG P2996 timelines, and the C++26 DIS ballot outcome and ISO publication date (not reachable this session).
- The patent and licensing position of the Slug algorithm as used by `libharfbuzz-gpu`. HarfBuzz states the basis [21]; I found no licensing note.
- Are Basis Universal and zstd outputs bit-identical across thread counts and minor versions? Their cacheability depends on it.
- When will KTX v5 go final with Basis v2.5's XUASTC and XUBC7, and will KTX spec Rev. 5 cover them [8]?
- Unreal DDC and Zen key composition, Unity Accelerator protocol, and Our Machinery Truth details: primary sources need network access to Epic, Unity and archived Our Machinery content.
- Oodle licensing terms for non-Unreal engines [45].
- Console constraints (threads, file I/O, SIMD, allocators) for these libraries. Only DirectXTex publicly exposes Xbox build options [11]; the rest needs NDA-tier data.
- Could a small cross-platform cooker (an AssetCooker-like rule engine with portable change detection) be an Automap recipe deliverable rather than a slot?

## Sources

All accessed 2026-10-29. Repository facts were read from git clones of the named repositories at the stated tag or branch tip. "Health" figures are the author's own `git log` analysis (2025-10-29 to 2026-10-29).

1. fastgltf — https://github.com/spnda/fastgltf — README.md, docs/options.rst, docs/overview.rst, docs/changelog.rst, include/fastgltf/core.hpp (tag v0.9.1, 2026-09-27; main 2026-10-29)
2. cgltf — https://github.com/jkuhlmann/cgltf — cgltf.h, LICENSE (master 2026-02-02; tag v1.15, 2025-02-09)
3. tinygltf — https://github.com/syoyo/tinygltf — README.md, AGENTS.md (tag v3.0.1, 2026-07-31)
4. ufbx — https://github.com/ufbx/ufbx — README.md, ufbx.h, LICENSE (tag v0.23.1, 2026-09-25)
5. assimp — https://github.com/assimp/assimp — Readme.md, CMakeLists.txt, AGENTS.md, LICENSE (tag v6.0.5, 2026-04-30)
6. stb — https://github.com/nothings/stb — stb_image.h (v2.30), stb_truetype.h (v1.26), README.md, CONTRIBUTING.md, LICENSE (master 2026-10-22; dev 2026-10-29)
7. tinyexr — https://github.com/syoyo/tinyexr — README.md, AGENTS.md, LICENSE (tag v3.2.0, 2026-07-09)
8. KTX-Software — https://github.com/KhronosGroup/KTX-Software — RELEASE_NOTES.md, lib/include/ktx.h, CMakeLists.txt, external/basis_universal/README.md (tag v5.0.0-rc2, 2026-08-17; v4.4.2, 2025-10-04)
9. Basis Universal — https://github.com/BinomialLLC/basis_universal — README.md (v2.5), transcoder/basisu_transcoder.h (tag v2_50, 2026-07-22)
10. bc7enc_rdo — https://github.com/richgel999/bc7enc_rdo — README.md, LICENSE, commit log (master 2026-07-30)
11. DirectXTex — https://github.com/microsoft/DirectXTex — README.md, CHANGELOG.md, CMakeLists.txt, skills/directxtex-usage/SKILL.md (tag may2026, 2026-05-08; skills added 2026-06-06)
12. Compressonator — https://github.com/GPUOpen-Tools/compressonator — readme.md, docs/source/revisions.rst, license/ (tag V4.6.0, 2026-10-19)
13. meshoptimizer — https://github.com/zeux/meshoptimizer — src/meshoptimizer.h, LICENSE.md (tag v1.3, 2026-09-25)
14. xatlas — https://github.com/jpcy/xatlas — README.md, source/xatlas/xatlas.h (master 2022-07-26)
15. UVAtlas — https://github.com/microsoft/UVAtlas — LICENSE, tags (oct2025, 2025-11-01)
16. ozz-animation — https://github.com/guillaumeblanc/ozz-animation — CHANGES.md, CMakeLists.txt, include/ozz/base/memory/allocator.h, include/ozz/base/io/archive.h, src/animation/offline/gltf/extern/tiny_gltf.h (tag 0.17.0, 2026-08-01)
17. Animation Compression Library — https://github.com/nfrechette/acl — README.md, includes/acl/core/iallocator.h, .gitmodules (develop 2025-09-13; tag v2.1.0, 2023-12-05)
18. msdfgen — https://github.com/Chlumsky/msdfgen — CHANGELOG.md, CMakeLists.txt (tag v1.13, 2025-11-30)
19. msdf-atlas-gen — https://github.com/Chlumsky/msdf-atlas-gen — CHANGELOG.md, CMakeLists.txt (tag v1.4, 2026-03-05)
20. FreeType (GitHub mirror) — https://github.com/freetype/freetype — docs/CHANGES, LICENSE.TXT (tag VER-2-15-0; release dated 2026-10-12)
21. HarfBuzz — https://github.com/harfbuzz/harfbuzz — NEWS (15.0.0, 2026-10-22; 14.2.0, 2026-04-20; 14.0.0, 2026-04-01), README.md, CONFIG.md, AGENTS.md, COPYING
22. efsw — https://github.com/SpartanJ/efsw — README.md, LICENSE (tag 1.7.3, 2026-10-28)
23. dmon — https://github.com/septag/dmon — LICENSE, commit log (master 2026-02-23)
24. PhysicsFS — https://github.com/icculus/physfs — src/physfs.h, CMakeLists.txt, AGENTS.md, extras/ (main 2026-09-30; tag release-3.2.0, 2022-09-30)
25. FlatBuffers — https://github.com/google/flatbuffers — CHANGELOG.md, include/flatbuffers/allocator.h, include/flatbuffers/base.h (tag v25.12.19)
26. Cap'n Proto — https://github.com/capnproto/capnproto — c++/configure.ac, c++/CMakeLists.txt, c++/src/kj/exception.h, LICENSE (tag v1.5.0, 2026-07-09)
27. cereal — https://github.com/USCiLab/cereal — README.md, include/cereal/cereal.hpp, LICENSE (tag v1.3.2, 2022-02-27; master 2026-03-10)
28. bitsery — https://github.com/fraillt/bitsery — README.md, LICENSE (tag v5.2.5, 2025-10-08)
29. yyjson — https://github.com/ibireme/yyjson — src/yyjson.h, LICENSE (tag 0.13.0, 2026-09-09)
30. simdjson — https://github.com/simdjson/simdjson — README.md, LICENSE, LICENSE-MIT (tag v5.0.8, 2026-10-29)
31. nlohmann/json — https://github.com/nlohmann/json — README.md (tag v3.12.0, 2025-04-11)
32. glaze — https://github.com/stephenberry/glaze — README.md (tag v9.1.0, 2026-10-28)
33. EnTT — https://github.com/skypjack/entt — README.md, docs/md/meta.md, docs/md/core.md (tag v4.0.0, 2026-07-22; v3.16.0 README for comparison)
34. reflect-cpp — https://github.com/getml/reflect-cpp (docs reference munich-data-engineering) — README.md, docs/content/cpp26_reflection.md (tag v0.26.2, 2026-10-19)
35. refl-cpp — https://github.com/veselink1/refl-cpp — tags (v0.12.4, 2022-11-05)
36. ISO C++ draft repository — https://github.com/cplusplus/draft — papers/n5047.md (2026-05-12), papers/n5051.md (2026-06-01), papers/n5055.md (2026-07-16), papers/n5033.md, source/meta.tex at tag n5050
37. GCC (GitHub mirror) — https://github.com/gcc-mirror/gcc — tags releases/gcc-16.1.0 (2026-04-30) and releases/gcc-16.2.0 (2026-08-07); gcc/c-family/c.opt; libstdc++-v3/include/std/meta; gcc/cp/reflect.cc
38. Clang C++ status — https://github.com/llvm/llvm-project/blob/main/clang/www/cxx_status.html (main, 2026-10-29); tag llvmorg-23.1.4
39. Bloomberg clang-p2996 — https://github.com/bloomberg/clang-p2996 — README.md (default branch 2025-06-29); p2996 branch tip 2026-09-24
40. SQLite (GitHub mirror) — https://github.com/sqlite/sqlite — src/sqlite.h.in, LICENSE.md, AGENTS.md, ext/ (tag version-3.54.1, 2026-10-16)
41. Zstandard — https://github.com/facebook/zstd — lib/zstd.h (dev, 2026-10-05), LICENSE, COPYING (tag v1.5.7, 2025-02-19)
42. LZ4 — https://github.com/lz4/lz4 — LICENSE, lib/lz4.h, AGENTS.md, commit log (tag v1.10.0, 2024-07-21)
43. xxHash — https://github.com/Cyan4973/xxHash — README.md, AGENTS.md, LICENSE (tag v0.8.4, 2026-09-18)
44. BLAKE3 — https://github.com/BLAKE3-team/BLAKE3 — c/README.md, LICENSE_* (tag 1.8.7, 2026-08-20)
45. RAD Game Tools, Oodle — https://www.radgametools.com/oodle.htm — **not accessed** (network policy); cited only as the intended primary source
46. Khronos glTF, KHR_meshopt_compression — https://github.com/KhronosGroup/glTF/tree/main/extensions/2.0/Khronos/KHR_meshopt_compression — README.md (merged 2026-01-24; status Release Candidate)
47. Godot docs, Import process — https://github.com/godotengine/godot-docs/blob/master/tutorials/assets_pipeline/import_process.rst (master 2026-10-28)
48. Godot docs — engine_details/file_formats/tscn.rst, tutorials/migrating/upgrading_to_godot_4.4.rst, tutorials/export/exporting_projects.rst, classes/class_resourceloader.rst — https://github.com/godotengine/godot-docs (master 2026-10-28)
49. Godot source — https://github.com/godotengine/godot — core/io/resource_importer.{h,cpp}, core/io/resource_loader.cpp, core/io/resource_uid.h, version.py (master 2026-10-27, 4.8-beta)
50. Unity C# reference source — https://github.com/Unity-Technologies/UnityCsReference — Editor/Mono/AssetPipeline/AssetImportContext.bindings.cs, Modules/AssetPipelineEditor/Public/ScriptedImporter.cs, Modules/AssetPipelineEditor/ImportSettings/AssetImporterEditor.cs, Editor/Mono/AssetDatabase/AssetsModifiedProcessor.cs, ImportBlocksEditor/… (commit "Unity 7000.0.0a12", 2026-10-22)
51. O3DE documentation source — https://github.com/o3de/o3de.org — content/docs/user-guide/assets/pipeline/{asset-dependencies-and-identifiers,metadata,asset-builders}.md, assets/custom-assets/builder/create-jobs.md, assets/asset-processor/{faster-scanning,asset-cache-server}.md, assets/runtime.md, packaging/asset-bundler/{_index,overview,assets-resolving,bundle-assets-for-release}.md (main 2026-10-23)
52. Unreal Engine, Derived Data Cache docs — https://dev.epicgames.com/documentation/en-us/unreal-engine/derived-data-cache-in-unreal-engine — **not accessed**
53. Unreal Engine, Zen Storage Server docs — https://dev.epicgames.com/documentation/en-us/unreal-engine/zen-storage-server-for-unreal-engine — **not accessed**; https://github.com/EpicGames/zen requires authentication
54. RacketCon 2013 program (Dan Liebgold, Naughty Dog keynote abstract) — https://github.com/racket/racket-lang-org/blob/master/rcon/2013/all.rkt
55. Our Machinery, "The Story behind The Truth: Designing a Data Model" — https://ourmachinery.com/post/the-story-behind-the-truth-designing-a-data-model/ — **not accessed**
56. AssetCooker — https://github.com/jlaumon/AssetCooker — Readme.md, LICENSE (MPL-2.0) (tag v0.0.2, 2026-09-29; master 2026-10-04)
57. Bevy — https://github.com/bevyengine/bevy — crates/bevy_asset/src/processor/mod.rs, crates/bevy_asset/src/meta.rs, Cargo.toml (0.20.0-dev; main 2026-10-29)
58. ezEngine docs source — https://github.com/ezEngine/docs-src — pages/docs/assets/assets-overview.md, pages/docs/assets/asset-profiles.md (main 2026-10-29)
59. bgfx — https://github.com/bkaradzic/bgfx — tools/, tools/geometryc/geometryc.cpp (master 2026-10-29)
60. vcpkg — https://github.com/microsoft/vcpkg — ports/*/vcpkg.json (master 2026-10-28)
61. Conan Center Index — https://github.com/conan-io/conan-center-index — recipes/ (master 2026-10-29)
62. xmake-repo — https://github.com/xmake-io/xmake-repo — packages/ (master 2026-10-29)
63. astc-encoder — https://github.com/ARM-software/astc-encoder — LICENSE.txt (tag 5.7.0, 2026-07-31)
