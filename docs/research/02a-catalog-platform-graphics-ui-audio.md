# Library catalog A: platform & GPU, shaders, tools/UI, audio

**Scope.** Compact, source-checked profiles of candidate libraries for four Automap *slots*: platform/window/input/GPU (SDL3 + SDL_GPU, raylib, sokol, bgfx, GLFW, Dawn, wgpu-native); shader toolchain (Slang, glslang, SPIRV-Cross, DXC, SDL_shadercross, sokol-shdc); tools/UI (Dear ImGui, ImGuizmo, ImPlot/ImPlot3D, RmlUi, Clay; Nuklear/microui as contrast); audio (miniaudio, SoLoud, SDL3 audio, OpenAL Soft, Steam Audio, FMOD, Wwise). For each library: license, form, latest release, maintenance health, platform reach, and the *integration seams* (allocators, threading, lifecycle, errors, math types, handles, build defines, shader formats) an agent must get right when wiring them together.

**Accessed: 2026-10-29.**

**Method and limits.** Most facts come from the libraries' own repositories: shallow clones of each default branch (HEAD as of 2026-10-27/29) for headers, READMEs and changelogs, and git history for release tags and commit statistics. GitHub release pages were read with WebFetch. "12-mo commits / humans / top author" covers 2025-10-29..2026-10-29, counts commit-author names with bots excluded, and is only a rough proxy for bus factor [62]. Several hosts could not be reached from this session: wiki.libsdl.org, fmod.com, audiokinetic.com, raysan5.itch.io and en.wikipedia.org. The web-search budget also ran out late in the work. So FMOD and Wwise terms rest on search-engine summaries of audiokinetic.com pages plus secondary articles, and are marked accordingly.

---

## Key findings

1. **[Verified]** **SDL_GPU has no web backend.** Its driver list is Metal, D3D12 and Vulkan, plus a `PrivateGPUDriver` slot for NDA platforms [14]. The header documents only those three public APIs [1], and SDL's full git history has no commit mentioning WebGPU [62]. An SDL3 + SDL_GPU testbed therefore **cannot target the browser**. Web builds need a second GPU path: SDL_Renderer/GLES, WebGPU through `emdawnwebgpu` plus `imgui_impl_wgpu` [43], or sokol, bgfx or raylib.
2. **[Verified]** **The SDL_GPU shader contract is strict and can be checked by a machine.** The format mask passed at device creation picks the backend. Each backend then accepts only its own formats: D3D12 always takes DXBC, and takes DXIL only if Shader Model 6.0 is available [1][14]. Resource bindings must follow fixed per-format layouts: SPIR-V descriptor sets 0–3, HLSL `space0..3`, MSL argument tables [1]. Dear ImGui's SDL_GPU backend embeds SPIR-V, DXBC, MSL and metallib blobs [39].
3. **[Verified]** **The ecosystem is converging on SDL_GPU.** Dear ImGui added `imgui_impl_sdlgpu3` in 1.91.7 (2025-01-14) [38]. RmlUi added an SDL GPU renderer in 6.2 and deprecated its SDL_Renderer backend in 6.3 in its favour [48]. SDL 3.4.0 (2026-01-01) added a 2D renderer that runs on a GPU device and lets the two share textures [2][3].
4. **[Verified]** **Many candidates have no semver releases, or churn fast.** Dear ImGui listed breaking changes in every release of the past 12 months (1.92.5 through 1.93.0) [38]. sokol has no versioned releases and logged 62 changelog entries in 12 months, 10 of them breaking [18]. bgfx is untagged and exposes only `BGFX_API_VERSION 167` [23]. SDL_shadercross has never been tagged [37]. Clay's last tag is v0.14 (2025-06-06) [49]. Pinning must therefore support commit SHAs, not only version ranges.
5. **[Verified]** **The shader toolchain is shifting.** glslang's HLSL front-end has been deprecated since April 2026 and will be removed at the next major version [33]. Slang's Metal and WGSL targets are "experimental", and WGSL is "work in-progress" [32]. SPIRV-Cross emits GLSL, MSL, HLSL and JSON reflection but not WGSL [34]. sokol-shdc uses Tint for WGSL [22].
6. **[Verified]** **Console reach exists but is private.** SDL's ports for Switch/Switch 2, PS4 and PS5 live in separate NDA-only repos and are free under zlib to licensed developers. The Xbox code (GDK) is in public SDL, but the GDKX libraries are restricted to licensed developers [5][6][7]. SDL_GPU defines `SDL_GPU_SHADERFORMAT_PRIVATE` "for NDA'd platforms" [1]. bgfx's GNM backend is "only for licensed PS4 developers" [23]. Wwise console access requires validating developer status [57].
7. **[Verified]** **Upstream projects restrict AI-generated contributions.** SDL bans LLM use in contributions outright [13], and so does SDL_shadercross, which ships the same policy file [37]. sokol forbids LLM-generated changes to its headers and docs [20]. RmlUi requires that all GitHub communication be written by humans [48]. Dear ImGui rejects AI-generated prose [44]. DXC requires a human in the loop [35]. glslang and SPIRV-Cross have AI-assisted-contribution clauses [33][34].
8. **[Verified]** **Bus-factor and dormancy risks.** Share of commits by the top author over the last 12 months [62]: OpenAL Soft 575/599, sokol about 1008/1074, ImPlot3D 132/148, GLFW 51/60, Dear ImGui about 713/875, miniaudio `dev` 91/116. SoLoud and microui have had no commits since August 2024, and SoLoud's last release is from 2020 [62].
9. **[Verified]** **License hazards.** OpenAL Soft is LGPL-2.0-or-later [54]. glslang's LICENSE combines BSD-3, Apache-2.0 and a "GPL 3 with special bison exception" section [33]. FMOD and Wwise are proprietary, with free tiers gated by budget or revenue [56–61].
10. **[Inference]** The strongest tier-1 desktop testbed is SDL 3.4.x + SDL_GPU, shaders through SDL_shadercross (SPIR-V or HLSL in) or Slang, Dear ImGui (docking) + ImGuizmo + ImPlot, and miniaudio. A web-first testbed is better served by sokol (WebGL2/WebGPU) or raylib (WebGL).

---

## 1. Platform / window / input / GPU slot

### 1.1 Profiles

| Library | License | Language / form | Latest release (date) | 12-mo commits / humans / top author [62] | Platform reach (public statements) |
|---|---|---|---|---|---|
| **SDL3** (+SDL_GPU) | zlib [1] | C API, compiled lib (CMake), static or shared | **3.4.18** (tag 2026-10-02); 3.4.0 published 2026-01-01 [3][4]; `main` is 3.5.0 (dev, odd minor = prerelease) [12] | 2160 / 271 / Lantinga 636 | Windows, macOS 10.14+, Linux, iOS/tvOS/visionOS, Android, Emscripten, SteamOS, Xbox (GDK) and more [5]; Switch/Switch 2, PS4, PS5 through NDA-only forks [5][6]; Xbox needs licensed GDKX [7] |
| **raylib** | zlib/libpng [17] | C99, compiled lib or unity build; vendors GLFW, miniaudio, stb_* | **6.2** (2026-10-15); 6.0 (2026-04-23) [16]; `master` 6.5-dev [17] | 1331 / 261 / raysan5 719 | Windows, Linux, macOS, RPi/DRM, Android, HTML5 [17]; platform backends GLFW, SDL, RGFW, Win32 (exp.), Emscripten (exp.), Memory (headless) [16] |
| **sokol** (app, gfx, glue, audio…) | zlib [18] | C, STB-style single-file headers (`SOKOL_IMPL`) | **No versioned releases**; dated CHANGELOG (latest 27-Oct-2026) [18] | 1074 / 33 / floooh ~1008 | sokol_app: Windows, macOS, Linux, iOS, Android, HTML5 [19]. No gamepad API in sokol_app [19]. No public console statement found |
| **bgfx** | BSD-2-Clause [23] | C++ with C99 bindings; compiled; GENie build [24] | **Untagged**; `BGFX_API_VERSION 167` [23] | 608 / 51 / Karadžić ~517 | Android, iOS/iPadOS/tvOS 16+, Linux, macOS 13+, PS4, RPi, UWP/Xbox One, Wasm, Windows 7+ [23] |
| **GLFW** (contrast) | zlib [25] | C99, compiled | **3.5.1** (2026-07-31); 3.5.0 skipped because of a bad tag [26]; `master` 3.6.0 [25] | 60 / 9 / Löwy 51 | Windows 7+, macOS 10.11+, Linux X11/Wayland [25]. No official web or mobile |
| **Dawn** | BSD-3-Clause [27] | C/C++ (`webgpu.h` + C++ wrapper); compiled, CMake with `DAWN_FETCH_DEPENDENCIES` [27] | Daily tags `vYYYYMMDD.HHMMSS`; latest v20261027.184652 (2026-10-28); GitHub binaries are a "best-effort service" [28] | 4745 / 128 / (Google team) | Backends: D3D12 (preferred), Metal 2.3 (macOS 13+, iOS 14+ best effort), Vulkan 1.1 (Linux, ChromeOS; Android WIP), GLES 3.1 in progress, D3D11 WIP; **Xbox not supported** [27]. Emscripten through `emdawnwebgpu` [28] |
| **wgpu-native** | MIT or Apache-2.0 [29] | Rust library with C API (`webgpu.h` + `wgpu.h`) [29] | **v29.0.1.1** (2026-06-23) [30] | 64 / 21 / (no dominant author) | Prebuilt Windows, macOS, Linux, Android, iOS [30]. MSRV 1.87 [29] |

### 1.2 Integration seams

| Library | Allocator hook | Threading / main thread | Lifecycle | Errors | Math / handles | Key build defines |
|---|---|---|---|---|---|---|
| SDL3 | `SDL_SetMemoryFunctions`: must be "the first call made into the SDL library"; unsafe once anything is allocated [10] | 115 functions in `SDL_video.h` are documented as main-thread only [1]. On the web, "rendering *must* be done on the main thread" [9]. `SDL_AppEvent` may run concurrently with `SDL_AppIterate` for events pushed off the main thread [8] | `SDL_main` or `SDL_MAIN_USE_CALLBACKS` (`SDL_AppInit/Iterate/Event/Quit`). Include `SDL_main.h` in exactly one TU. SDL calls `SDL_Quit` after `SDL_AppQuit` [8] | `bool` return + per-thread `SDL_GetError()` [11] | Opaque handle pointers (`SDL_GPUDevice*` etc.); no vector math types | `SDL_MAIN_HANDLED`, `SDL_MAIN_USE_CALLBACKS` [8]; CMake `SDL_PTHREADS` for Emscripten [9] |
| SDL_GPU | (uses SDL's allocator) | Command buffers can be built on several threads "as long as they are submitted in the correct order" [1] | Create device → `SDL_ClaimWindowForGPUDevice` → per frame: acquire command buffer, then swapchain texture, then passes, then submit [1] | NULL/false + `SDL_GetError()` [1] | Left-handed; NDC z∈[0,1]; +Y down in viewport and texture coordinates; SDL corrects Vulkan automatically [1] | Formats: `SPIRV`, `DXBC` (SM5.1), `DXIL` (SM6.0), `MSL`, `METALLIB`, `PRIVATE` [1] |
| raylib | `RL_MALLOC/RL_CALLOC/RL_REALLOC/RL_FREE` macros, forwarded into vendored miniaudio (`MA_MALLOC`) and dr_* [17] | Single global context (inference from the API shape) | `InitWindow`/`CloseWindow`; web builds use ASYNCIFY or `emscripten_set_main_loop` [17] | `TraceLog`; return values were made consistent in 6.2 [16] | Own C structs (`Vector2/3`, `Matrix`) | `PLATFORM_*`, `GRAPHICS_API_OPENGL_{11,21,33,43,ES2,ES3,SOFTWARE}`, `SUPPORT_MODULE_RAUDIO` [17] |
| sokol | `desc.allocator {alloc_fn, free_fn, user_data}` in each header, both or neither [19] | sokol_app owns the frame loop (callbacks) unless `SOKOL_NO_ENTRY` [19] | `sapp_run` → `init_cb` → `sg_setup(&(sg_desc){.environment = sglue_environment()})` [18] | `desc.logger` (sokol_log.h) + validation layer in debug builds [19] | Handles are small structs with IDs (`sg_buffer` etc.); no math lib | One backend define per implementation TU, identical for app and gfx: `SOKOL_GLCORE/GLES3/D3D11/METAL/WGPU/VULKAN`. Objective-C TU on Apple [18] |
| bgfx | `bx::AllocatorI*` in `bgfx::Init`; `CallbackI` for fatal, trace and cache [23] | API thread + render thread by default. Calling `renderFrame()` before `init()` gives single-threaded mode, meant for "platforms that don't allow creating a separate rendering thread" [23] | `init` → `frame` loop → `shutdown` | `CallbackI::fatal` | Handles are 16-bit index structs (`bgfx::TextureHandle`) | `BGFX_CONFIG_MULTITHREADED`; GENie project generation [24] |
| GLFW | `glfwInitAllocator` (called before `glfwInit`) [25] | 93 functions are main-thread only [25] | `glfwInit` → windows → `glfwTerminate` | Error callback + codes | Opaque pointers | Window and input only; no GPU API |
| Dawn / wgpu-native | n/a (C API) | Per `webgpu.h`; async callbacks | Instance → adapter → device; native requires choosing an implementation | Error callbacks | `WGPU*` opaque handles | `webgpu.h` has **no tagged release** (no git tags in webgpu-native/webgpu-headers) [31]. wgpu-native still ships its own `wgpu.h` extensions [29] |

### 1.3 SDL_GPU backend × shader-format matrix

| SDL_GPU driver | Public platforms | Shader formats accepted | Notes |
|---|---|---|---|
| `vulkan` | Windows, Linux, **Nintendo Switch**, "certain Android devices" [1] | SPIRV | Vulkan 1.0 + listed features; 3.4 added properties that relax them for Android [1][2] |
| `direct3d12` | Windows 10+, **Xbox One / Series (GDK)** [1] | DXBC always; DXIL if SM 6.0 [14] | `SDL_PROP_GPU_DEVICE_CREATE_D3D12_ALLOW_FEWER_RESOURCE_SLOTS_BOOLEAN` for Haswell/Broadwell [1] |
| `metal` | macOS 10.14+, iOS/tvOS 13+; simulators unsupported [1] | MSL, METALLIB | ImGui uses MSL for macOS 10.14+ and metallib for 14+ [39] |
| private driver | NDA platforms [14] | PRIVATE [1] | Only in private forks (inference) |
| *(WebGPU/WebGL)* | **none** [14] | — | No backend; searched SDL history for "webgpu": 0 commits [62] |

### 1.4 Known seams: platform/GPU

- **Web target vs SDL_GPU.** The ImGui `example_sdl3_sdlgpu3` has no Emscripten makefile, while `example_sdl3_opengl3` and `example_sdl3_wgpu` do [41][43]. A composition that selects "SDL3 + SDL_GPU" and "target: web" should fail validation, or switch the web GPU path to SDL_Renderer/GL or to WebGPU.
- **Emscripten main loop.** SDL documents that a blocking `while` loop "will not work on Emscripten". Use `emscripten_set_main_loop` or the SDL main callbacks, which default to a zero-fps (rAF) main loop [9]. raylib's examples default to `-sASYNCIFY` [17]. ImGui wraps its examples in an `EMSCRIPTEN_MAINLOOP_BEGIN` stub [43]. sokol_app owns the loop [19].
- **Threads on the web.** Turning on pthreads requires a server that sends cross-origin-isolation headers. SDL's guidance is "easier to not enable them at all" [9]. miniaudio's AudioWorklets need `-sAUDIO_WORKLET=1 -sWASM_WORKERS=1 -sASYNCIFY` [52].
- **Format mask ≠ accepted formats.** In `SDL_CreateGPUDevice`, the mask only *selects* a backend. `SDL_CreateGPUShader` validates against the backend's own supported set [14]. Example: ImGui's example requests `SPIRV|DXIL|MSL|METALLIB`, yet its backend submits **DXBC** on D3D12, which works because D3D12 always supports DXBC [39][41][14]. Agents should query `SDL_GetGPUShaderFormats()` rather than assume.
- **Binding layout is enforced by convention, not validated.** SDL warns it is "difficult for the API to automatically validate shaders" and recommends reflection [1]. SPIR-V/GLSL layout: vertex stage set 0 for samplers, storage textures and storage buffers, set 1 for uniforms; fragment stage sets 2 and 3. Bindings must start at 0 with no gaps, samplers first [1]. This can be checked with SPIRV-Cross/SDL_shadercross reflection.
- **Version policy.** SDL uses odd/even versions: an odd minor or odd patch means a prerelease. Newer stable versions are backward- but not forward-compatible [12]. A rule such as "SDL ≥3.4.0, even minor" can be enforced mechanically.
- **sokol backend consistency.** The backend define must match between sokol_gfx and sokol_app and the generated shader headers. Shader headers must come from a compatible sokol-shdc version [18]. sokol's Vulkan backend is "highly experimental" with frame-pacing issues on Windows+NVIDIA; WebGPU on native desktop is "experimental" [19].
- **bgfx WebGPU documentation drift.** The README says "WebGPU (Dawn Native only)" [23]. Yet commits from June–Oct 2026 add and fix Emscripten WebGPU, e.g. "WebGPU: Fixed rendering into expired canvas texture on Emscripten" (2026-10-27) [62]. Treat web WebGPU support in bgfx as **[Unverified]** until the docs catch up.
- **Native WebGPU naming.** Code against `webgpu.h` must also define which implementation is in use. ImGui requires `IMGUI_IMPL_WEBGPU_BACKEND_{DAWN,WGPU,WGVK}` natively and defaults to Dawn with Emscripten ≥4.0.10 (`--use-port=emdawnwebgpu`) [43]. sokol requires `--use-port=emdawnwebgpu` for WebGPU builds [19].
- **SDL 2D/GPU interop (3.4).** `SDL_CreateGPURenderer`/`SDL_GetGPURendererDevice` let a 2D renderer and SDL_GPU code share one device. `SDL_PROP_TEXTURE_CREATE_GPU_TEXTURE_POINTER` and `SDL_PROP_TEXTURE_GPU_TEXTURE_POINTER` wrap textures in either direction [2]. This is useful for debug overlays without a second device.
- **OpenXR.** SDL `main` (3.5 dev) gained "GPU: OpenXR integration" on 2026-01-30 [62]. It is not in any stable release yet.

---

## 2. Shader toolchain slot

### 2.1 Profiles

| Tool | License | Form | Latest release (date) | 12-mo commits / humans / top [62] | Inputs → outputs |
|---|---|---|---|---|---|
| **Slang** | Apache-2.0 WITH LLVM-exception [32] | C++ compiler lib + `slangc`; hosted under Khronos (SPDX "The Khronos Group") [32] | **v2026.20** (2026-10-16); releases roughly weekly to biweekly [62] | 2006 / 91 / 285 (no single owner) | Slang/HLSL → HLSL (D3D11/12), SPIR-V and GLSL (Vulkan), MSL (**experimental**), WGSL (**experimental, WIP**), CUDA, CPU [32]. Built on Windows, Linux, macOS; Wasm experimental [32] |
| **glslang** | BSD-3 core + others; LICENSE also contains "GPL 3 with special bison exception" text [33] | C++ lib + `glslang` CLI | **16.6.0** (2026-09-11) [62] | 318 / 79 / 26 | GLSL → SPIR-V; HLSL front-end **deprecated April 2026**, to be removed at the next major version, with ≥18 months' notice [33] |
| **SPIRV-Cross** | Apache-2.0 [34] | C++11 + stable C API (`spirv_cross_c.h`) [34] | Tags follow the Vulkan SDK: `vulkan-sdk-1.4.363.0` (2026-09-16) [62] | 248 / 45 / Arntzen 150 | SPIR-V → GLSL, MSL, HLSL, JSON reflection; no WGSL [34] |
| **DXC** | NCSA/LLVM release license [35] | `dxc`, `dxcompiler.[dll\|so]`, `dxil.[dll\|so]` [35] | **v1.9.2609** (Sep 2026, "Latest"); v1.10.x = SM 6.10 previews [36] | 497 / 48 / 110 | HLSL → DXIL, SPIR-V; Metal libraries through Apple's Metal Shader Converter when built from source [35] |
| **SDL_shadercross** | zlib [37] | C lib + CLI; vendors SPIRV-Cross, DXC, SPIRV-Tools [37] | **Never tagged**; header says 3.0.0 [37] | 21 / 6 / 16 | SPIR-V or HLSL → DXBC, DXIL, SPIR-V, MSL, HLSL; can return ready `SDL_GPUShader*` at runtime; reflection metadata API [37] |
| **sokol-shdc** | MIT [22] | C++ CLI; binaries through the `sokol-tools-bin` repo [22] | No releases [62] | 63 / 6 / 55 | Annotated Vulkan-style GLSL 450 → GLSL 300es/310es/410/430, HLSL4/5 (optionally bytecode), Metal (optionally bytecode), WGSL, SPIR-V; emits C/Zig/Rust/Odin/Nim/D/Jai/C3 headers or YAML reflection; uses glslang, SPIRV-Tools, SPIRV-Cross and Tint [22] |
| *bgfx shaderc* | BSD-2 | C++ CLI | (with bgfx) | — | bgfx GLSL dialect (`$input/$output`, `varying.def.sc`, `mul()`, `SAMPLER2D`) → GLSL/ESSL, HLSL, Metal, PSSL, SPIR-V profiles [24] |

### 2.2 What each GPU API consumes and how shaders reach it

| Consumer | Format it consumes | Typical path |
|---|---|---|
| SDL_GPU Vulkan | SPIR-V with SDL set layout [1] | GLSL→glslang, HLSL→DXC (`-spirv`), or Slang→SPIR-V |
| SDL_GPU D3D12 | DXBC SM5.1 or DXIL SM6.0 [1] | SDL_shadercross (DXBC needs d3dcompiler_47 or vkd3d-utils; DXIL needs dxcompiler + dxil) [37] |
| SDL_GPU Metal | MSL source or metallib [1] | SDL_shadercross/SPIRV-Cross → MSL (default MSL 1.2.0) [37]; `xcrun metal/metallib` for metallib [39] |
| sokol_gfx | Per-backend source or bytecode in a generated header [22] | sokol-shdc only (practical) |
| bgfx | shaderc binaries per renderer [24] | shaderc only |
| raylib | GLSL (GL 3.3 / ES2 / ES3 variants) | Hand-written GLSL per GL version |
| WebGPU (Dawn/wgpu/browser) | WGSL (SPIR-V is accepted natively only by some implementations) | Tint (Dawn), sokol-shdc, or Slang (experimental) |

### 2.3 Known seams: shaders

- **Toolchain pinning.** sokol-shdc output must match the sokol headers [18]. SDL_shadercross is untagged and vendors its own DXC and SPIRV-Cross [37]. DXC ships a stable line (v1.9) alongside a prerelease line (v1.10 = SM 6.10 preview, which needs a preview Agility SDK) [36]. A manifest should record tool SHA or version together with the shader artifacts.
- **Runtime DLLs.** SDL_shadercross says "strongly recommended that you ship SPIRV-Cross and DXIL dependencies along with your application" when you translate at runtime [37]. This becomes a packaging rule in the manifest.
- **HLSL source path.** With glslang's HLSL deprecated [33], HLSL-authored projects should route through DXC or Slang. A "glslang for HLSL" recipe should be flagged as legacy.
- **Reflection-driven create-info.** SDL_GPU needs exact sampler, storage and uniform counts in `SDL_GPUShaderCreateInfo` [1]. SDL_shadercross provides `SDL_ShaderCross_ReflectGraphicsSPIRV` [37], and SPIRV-Cross exposes JSON reflection [34]. A glue recipe can generate the counts instead of hand-writing them.
- **Coordinate conventions.** SDL_GPU uses a [0,1] depth range and corrects Vulkan's Y flip [1]. raylib/rlgl is OpenGL-convention (inference from its GL backends). Projection-matrix helpers must match the slot.

---

## 3. Tools / UI slot

### 3.1 Profiles

| Library | License | Language / form | Latest release (date) | 12-mo commits / humans / top [62] | Notes |
|---|---|---|---|---|---|
| **Dear ImGui** | MIT [42] | C++ (small C++11 subset, no STL headers) [42]; compile `.cpp` + chosen backends | **v1.93.0** (2026-10-14), tagged twice: `v1.93.0` and `v1.93.0-docking` [38][62] | 875 / 71 / ocornut ~713 | Funding through the wiki "Funding" page [44]. Releases about every 1–3 months, each with a "Breaking Changes" section [38] |
| **ImGuizmo** | MIT [45] | C++ sources (`src/`) | **1.10** (2026-05-11); before that 1.9 (2026-05-02), 1.83 (2021) [62] | 46 / 10 / Guillemet 30 | Header comment reads "v1.92.5 WIP", which mirrors ImGui numbering and not the tags [45] |
| **ImPlot** | MIT [46] | C++ sources | **v1.1** (2026-10-25); v1.0 (2026-04-05) [62] | 154 / 21 / Queiroz 113 | Most commits now come from the ImPlot3D author (Breno Cunha Queiroz); the original author has 8 [62] |
| **ImPlot3D** | MIT [47] | C++ sources | **v0.5** (2026-10-25) [62] | 148 / 7 / Queiroz 132 | Depends only on Dear ImGui, not on ImPlot [47] |
| **RmlUi** | MIT [48] | **C++17** (since 6.3); compiled lib (CMake); FreeType font engine by default [48] | **6.3** (2026-08-22); 6.2 (2026-01-11) [48][62] | 141 / 37 / Ragazzon ~91 | HTML/CSS-like (RML/RCSS); 6.3 adds CSS custom properties [48]. Readme lists Windows, Linux, macOS, Android, iOS, Switch [48] |
| **Clay** | zlib [49] | C single header (`CLAY_IMPLEMENTATION`); **requires C99, C++20, or MSVC** (`#error` otherwise) [49] | **v0.14** (2025-06-06); `main` active until 2026-05-20 with unreleased features [62] | 55 / 11 / Barker 17 | Arena memory, no hidden allocations; emits render commands [49] |
| Nuklear *(contrast)* | MIT or public domain [50] | ANSI C single header | v4.13.3 (2026-05-05) [62] | 123 / 17 / Loach 56 | Active community maintenance |
| microui *(contrast)* | MIT [51] | C, about 1k LOC | No tags; last commit 2024-08-13 [62] | 0 | Dormant |

### 3.2 Bridges and backends

| UI lib | Platform backends | Renderer backends relevant to the candidates |
|---|---|---|
| Dear ImGui (master) | `sdl3`, `sdl2`, `glfw`, `win32`, `osx`, `android`, `glut`, `allegro5`, `qnx`, `null` [38] | `sdlgpu3`, `sdlrenderer3`, `vulkan`, `dx9–12`, `metal`, `metal4`, `opengl2/3`, `wgpu` (Dawn, wgpu-native or WGVK) [38][43]. **sokol:** `util/sokol_imgui.h` (C++ or C via cimgui; uses `RendererHasTextures`, ImGui ≥1.92) [21]. **bgfx:** example-level integration only (inference) |
| Dear ImGui (docking) | same | Same backends plus multi-viewport support, e.g. sdlgpu3 [38]. Docking and multi-viewports exist only on this branch, which "is kept in sync with master regularly" [42] |
| RmlUi | GLFW, SDL, SFML, Win32, X11, Wayland [48] | GL2, GL3, DX11, DX12, Vulkan, **SDL_GPU** (picks SPIRV, then DXIL, then MSL); SDLrenderer deprecated [48] |
| Clay | (input supplied by the user) | raylib, SDL2, SDL3 (`SDL_Renderer`, not SDL_GPU), sokol, GLES3, web, cairo, win32_gdi, termbox2, playdate [49] |
| ImGuizmo / ImPlot / ImPlot3D | Inherit ImGui's | Draw through `ImDrawList`, so they work with any ImGui backend [45][46][47] |

### 3.3 Dear ImGui integration seams

| Seam | Detail |
|---|---|
| Allocator | `ImGui::SetAllocatorFunctions(alloc, free, user)`; `IMGUI_DISABLE_DEFAULT_ALLOCATORS`. DLL users must call `SetCurrentContext()` + `SetAllocatorFunctions()` per module [42] |
| Math interop | `IM_VEC2_CLASS_EXTRA` / `IM_VEC4_CLASS_EXTRA` in `imconfig.h` (or `IMGUI_USER_CONFIG`) for implicit conversion; or `IMGUI_DEFINE_MATH_OPERATORS` [42]. ImGuizmo takes column-major `float[16]` [45]. RmlUi has its own `RMLUI_VECTOR2_USER_EXTRA` etc. through `RMLUI_CUSTOM_CONFIGURATION`, plus `RMLUI_MATRIX_ROW_MAJOR` [48] |
| Threading | A context "may be not used from multiple threads in parallel". Render-thread staging goes through `imgui_threaded_rendering` (ImDrawDataSnapshot / ImTextureQueue) [42] |
| Init order (SDL3+SDL_GPU example) | `SDL_Init` → create window → `SDL_CreateGPUDevice(SPIRV\|DXIL\|MSL\|METALLIB)` → `SDL_ClaimWindowForGPUDevice` → `SDL_SetGPUSwapchainParameters` → `ImGui::CreateContext` → `ImGui_ImplSDL3_InitForSDLGPU` → `ImGui_ImplSDLGPU3_Init(&info)`. Shutdown: `SDL_WaitForGPUIdle` → `ImGui_ImplSDL3_Shutdown` → `ImGui_ImplSDLGPU3_Shutdown` → `ImGui::DestroyContext` [41] |
| Per-frame rule (sdlgpu3) | `ImGui_ImplSDLGPU3_PrepareDrawData()` **must** be called *before* the render pass that calls `ImGui_ImplSDLGPU3_RenderDrawData` ("MANDATORY") [39] |
| Texture IDs | `ImTextureID` is backend-specific. For sdlgpu3 it changed on 2025-08-08 (**breaking**) from `SDL_GPUTextureSamplerBinding*` to `SDL_GPUTexture*` [39]. Since 1.92, backends must handle `ImGuiBackendFlags_RendererHasTextures` (dynamic font atlas) [39][21] |
| Errors | `IM_ASSERT` (overridable); backends assert on SDL failures and point to `SDL_GetError()` [39] |
| Version coupling | `IMGUI_CHECKVERSION()` in backend init [40]. Backends live in the same repo, so take them from the same commit as core. Add-ons check `IMGUI_VERSION_NUM` (ImPlot gates on 1.88/1.89 renames; ImPlot3D has a pre-1.92 `ImTextureRef` fallback) [46][47]. The SDL3 backend logs compiled and linked SDL versions but has no minimum-version `#if` (inference from [40]) |
| Exceptions/RTTI | Not required by ImGui (inference: no use in the core per FAQ [42]). RmlUi auto-detects RTTI since 6.3 [48] |

### 3.4 Known seams: tools/UI

- **C++ standard floor.** Clay in a C++ TU needs C++20, except on MSVC [49]. RmlUi needs C++17 [48]. ImGui needs only a C++11 subset [42]. A composition's language standard must cover the maximum over all libraries compiled as C++. Clay can be isolated in a `.c` TU.
- **ImGuizmo frame order.** Call `ImGuizmo::BeginFrame()` right after the backend's `NewFrame`; call `SetImGuiContext` across DLLs [45]. Matrices are column-major `float[16]` [45]. Feeding them from SDL_GPU-convention ([0,1] depth) projections is **[Unverified]** for correctness; an agent should test with a known transform.
- **sokol_imgui.h** compiles as C++ against `imgui.h` or as C against cimgui. "Original" cimgui and "dear bindings" cimgui are both handled [21]. It requires ImGui ≥1.92 semantics [21].
- **RmlUi + SDL3** platform interfaces now require the window to be set before use (6.3) [48].

---

## 4. Audio slot

### 4.1 Profiles

| Library | License | Form | Latest release (date) | 12-mo commits / humans / top [62] | Platforms |
|---|---|---|---|---|---|
| **miniaudio** | Public domain or MIT-0 [52] | C single file today; from 0.12 a `miniaudio.h` + `miniaudio.c` split (`#include "miniaudio.c"` stays supported) [52] | **0.11.25** (2026-03-04); `dev` has 0.11.26 work and a `dev-0.12` branch [52][62] | dev: 116 / 21 / Reid 91 | Windows, macOS, iOS, Linux, BSDs, Android, RPi, Emscripten (Web Audio; opt-in AudioWorklets) [52]. Other platforms through **custom backends without modifying the source** [52] |
| **SDL3 audio** | zlib | Part of SDL | (SDL 3.4.18) | (SDL) | Same platforms as SDL. 3.4 added `SDL_PutAudioStreamDataNoCopy` and planar input [2] |
| **SoLoud** | zlib/libpng [53] | C++ with C API | RELEASE_20200207; last commit 2024-08-13 [62] | 0 | Backends include SDL, miniaudio, OpenAL, XAudio2, WASAPI [53]. **Dormant** |
| **OpenAL Soft** | **LGPL-2.0-or-later**; default HRTF data Apache-2.0 [54] | C API, C++20 implementation since 1.25.0 [54] | **1.25.2** (2026-05-11) [62] | 599 / 21 / Robinson 575 | Desktop backends (PipeWire, PulseAudio, ALSA, WASAPI…) [54]. Console and web **[Unverified]** |
| **Steam Audio** | Apache-2.0 [55] | C API (`phonon.h`), compiled | **4.8.1** (2026-02-11) [62] | 18 / 9 / 5 | Windows, Linux, macOS, Android, iOS [55]. CMake also has an Emscripten/WASM path [55]. Integrations: Unity, Unreal, **FMOD Studio 2.0+**, **Wwise 2023.1+** [55] |
| **FMOD** (proprietary) | Commercial EULA; free "Indie" tier | Binary SDK (Core + Studio APIs) | **[Unverified]**: fmod.com unreachable | n/a | Consoles supported (well known; not verified here) |
| **Wwise** (proprietary) | Commercial; free Indie tier [56] | Binary SDK through the Audiokinetic Launcher | **[Unverified]**: audiokinetic.com unreachable | n/a | "Full platform access", but console use needs validated developer status [57] |

### 4.2 Integration seams

| Library | Allocator | Threading | Lifecycle / errors | Build defines |
|---|---|---|---|---|
| miniaudio | `ma_allocation_callbacks` per object or config [52] | Data callback runs on the audio thread; end-of-sound callbacks also fire there [52] | `ma_engine_init/uninit`, `ma_result` codes | `MINIAUDIO_IMPLEMENTATION` (moving to compiling `miniaudio.c`), `MA_NO_*`, `MA_ENABLE_AUDIO_WORKLETS` [52] |
| SDL3 audio | SDL's | Stream callbacks "may run from any thread"; SDL holds the stream lock during them [15] | `SDL_OpenAudioDeviceStream`; bool + `SDL_GetError` | — |
| OpenAL Soft | none public (inference) | Internal mixer thread | Context and device; `alGetError` | CMake backend detection: "Double check that the appropriate backends were detected" [54] |
| Steam Audio | `IPLContextSettings.allocateCallback(size, alignment)` / `freeCallback` [55] | Simulation and audio threads (per docs; not checked in detail) | `IPLerror` = SUCCESS/FAILURE/OUTOFMEMORY/INITIALIZATION [55] | Optional Intel IPP, Embree, Radeon Rays, TrueAudio Next [55] |

### 4.3 Proprietary middleware: what a private registry must handle

| Aspect | FMOD | Wwise |
|---|---|---|
| Free tier | Free "Indie" license. Sources agree on **< $200K annual revenue** but disagree on the budget cap (**< $500K** in older coverage, **< $600K** in newer) [59][60][61]. **Conflicting / secondary only** | Indie: free for production budgets **up to $250K** [56][58]. The current Indie license has **no asset limit** [57] |
| Paid tiers | About $2K per title (Indie), about $6K (Basic) per a 2026 guide [61]: **[Unverified]** | Pro: $250K–$2M budget, "starting at $8,000" first platform; Premium above $2M [56][58]: figures from a search-engine summary of the pricing page |
| Ongoing checks | **[Unverified]** | Indie keys are checked against budget at milestones and every 6–12 months after launch [57] (search summary) |
| Consoles | **[Unverified]** (platform-holder verification expected) | Included at Indie, but "developers will have to validate their developer status" [57] |
| Distribution | Binary SDK downloaded from `fmod.com/download#fmodengine` [55]. Login requirement **[Unverified]** | Audiokinetic Launcher. SDK is found through the **`WWISEROOT`** env var (`$WWISEROOT/SDK/include/AK/AkPlatforms.h`); plugins are built with Wwise's `wp.py` [55] |
| Automap implications | Not vendorable. The manifest must reference an external SDK path, a license tier, and per-platform SDK packages; console SDK variants sit behind NDA | Same, plus an env-var-based locator and per-platform SDK installs managed by the Launcher |

### 4.4 Known seams: audio

- **Duplicate vendoring.** raylib vendors miniaudio 0.11.25 (plus dr_wav/dr_mp3/dr_flac, stb_*) and compiles it into `raudio` when `SUPPORT_MODULE_RAUDIO=1` [17]. Adding a separate miniaudio implementation TU to the same link risks duplicate symbols. Mitigations: set `SUPPORT_MODULE_RAUDIO=0`, or share one copy. This can be checked by a machine (symbol and implementation-define scan).
- **Audio-thread rules.** Callbacks from miniaudio and SDL may run off the main thread [15][52]. Game-side state shared with them needs locks or lock-free queues. This belongs in agent guidance.
- **LGPL linking.** Shipping OpenAL Soft statically triggers LGPL relinking obligations (inference from its LGPL-2+ license [54]). A license-policy check can flag it.
- **Web audio.** miniaudio's AudioWorklet mode needs ASYNCIFY and WASM workers [52]. That interacts with the GPU and main-loop choices in §1.4.

---

## 5. Per-slot comparison (summary)

| Slot | Candidate | Web | Mobile | Console (public) | Release discipline | Seam surface |
|---|---|---|---|---|---|---|
| Platform+GPU | SDL3 + SDL_GPU | SDL yes; **SDL_GPU no** | iOS/Android | NDA forks + GDK + PRIVATE GPU slot | Monthly stable patches (3.4.0→3.4.18) [4] | Strict shader layout; main-thread video |
| | raylib 6.2 | WebGL | Android | none stated | 2 releases/yr (6.0, 6.2) [16] | GL-only (+software); vendored deps |
| | sokol | WebGL2 + WebGPU | iOS/Android | none stated | Rolling, unversioned [18] | Backend-define and shdc coupling |
| | bgfx | WebGL2 (+WebGPU in progress) | iOS/Android | PS4 (GNM, licensed) | Rolling, API version only [23] | Own shader dialect; render thread |
| | Dawn / wgpu-native | (browser native) | Android/iOS (Dawn best effort) | Xbox unsupported (Dawn) | Daily (Dawn) / ~quarterly (wgpu-native) | `webgpu.h` not formally versioned [31] |
| Shaders | Slang | WGSL exp. | — | — | ~weekly | Experimental MSL/WGSL |
| | glslang + SPIRV-Cross | (via Tint for WGSL) | — | — | Vulkan SDK cadence | HLSL front-end deprecated |
| | DXC | — | — | Xbox implied (DXIL) | ~bimonthly + previews | Ship dxcompiler/dxil |
| | SDL_shadercross | — | — | — | Untagged | Vendors DXC/SPIRV-Cross |
| Tools/UI | Dear ImGui | yes (GL/WebGPU) | yes | (any backend you write) | 1–3 months, breaking every release | Backend pairing, texture IDs |
| | RmlUi | **[Unverified]** | Android/iOS | Switch (readme) | ~2/yr | C++17, FreeType |
| | Clay | yes (web renderer) | via renderer | via renderer | Last tag 2025 | C99/C++20 |
| Audio | miniaudio | yes | yes | custom backend | ~3 releases/yr | Audio thread |
| | Steam Audio | WASM path in CMake | yes | **[Unverified]** | ~2–3/yr (4.7.0 Aug 2025 → 4.8.1 Feb 2026) [62] | Optional vendor libs |
| | FMOD / Wwise | **[Unverified]** | yes | yes (gated) | Vendor-controlled | Licensing, external SDK |

---

## Leads checked

| Lead | Verdict | Evidence |
|---|---|---|
| "SDL 3.4 (January 2026) improved interop between its GPU API and its 2D renderer." | **Confirmed** | 3.4.0 release published 1 Jan 2026 (tag commit 2025-12-31). Adds `SDL_CreateGPURenderer`, `SDL_GetGPURendererDevice`, GPU render states for fragment shaders, and GPU-texture wrapping props [2][3][4] |
| "raylib 6.0 (April 2026) added a software renderer." | **Confirmed** (note: superseded) | "raylib 6.0 (23 April 2026) … New Software Renderer backend [rlsw]" [16]; `rlsw.h` is "An OpenGL 1.1-style software renderer" [17]. Current release is **6.2 (15 Oct 2026)** [16] |
| "Dear ImGui has an SDL_GPU backend." | **Confirmed** | `imgui_impl_sdlgpu3` added in 1.91.7 (2025-01-14) [38]; current changelog entries through 2026-09-07 [39] |
| Typical testbed stack: SDL3 + SDL_GPU or raylib; Slang or glslang + SPIRV-Cross; Dear ImGui (ImGuizmo, ImPlot), RmlUi, Clay; miniaudio; sokol_gfx and bgfx as alternatives | **Corrected (partially)** | All projects exist and are active, with these caveats: SDL_GPU has no web backend [14]; glslang's HLSL path is deprecated [33]; Slang's MSL/WGSL are experimental [32]; Clay's last tag is from 2025 and it needs C++20 in C++ TUs [49]; sokol's Vulkan backend is experimental [19]; RmlUi already has an SDL_GPU renderer [48], which strengthens the pairing |

---

## Implications for Automap *(inference)*

**Strongest v1 candidates for a tier-1 testbed (desktop):**

- **Platform/GPU:** SDL 3.4.x (stable, even-minor) with SDL_GPU. It has the widest public reach including documented console paths, monthly patch releases, a 271-person contributor base, and two UI libraries with first-party SDL_GPU backends.
- **Shaders:** SDL_shadercross as the "last mile" to DXBC/DXIL/MSL, fed SPIR-V from glslang (GLSL) or DXC (HLSL). Slang is an alternative authoring language where WGSL/MSL being experimental is acceptable.
- **Tools/UI:** Dear ImGui **docking** branch (pin to a `vX.Y.Z-docking` tag) + `imgui_impl_sdl3` + `imgui_impl_sdlgpu3`, ImGuizmo, ImPlot. RmlUi for in-game UI.
- **Audio:** miniaudio.
- **Web-first alternative recipe:** sokol (app+gfx+audio, `SOKOL_WGPU` or `SOKOL_GLES3`) with sokol-shdc and `sokol_imgui.h`, or raylib for simplicity. This should be modelled as a *separate composition*, not a flag on the SDL_GPU one.

**Seams that look checkable by a machine (manifest validators or CI):**

| Check | Basis |
|---|---|
| Version-range and stability policy (SDL odd/even; ImGui core + backends from the same commit; `IMGUI_CHECKVERSION`) | [12][40] |
| Target × backend feasibility (e.g. web + SDL_GPU → error; Xbox + Dawn → error) | [14][27] |
| Shader artifact formats per enabled GPU driver (SPIRV/DXBC/DXIL/MSL/METALLIB) | [1] |
| SDL_GPU binding-layout lint from SPIR-V reflection (set numbers, contiguous bindings, ordering) | [1][34][37] |
| Language-standard floor (Clay C++20 / C99; RmlUi C++17; OpenAL Soft C++20 to build) | [48][49][54] |
| Single-implementation-TU defines (`SOKOL_IMPL`, `MINIAUDIO_IMPLEMENTATION`, `CLAY_IMPLEMENTATION`) and backend-define consistency across sokol headers | [18][49][52] |
| Duplicate vendored libraries (raylib's miniaudio/stb vs standalone) | [17] |
| License policy (LGPL OpenAL Soft; glslang mixed; proprietary FMOD/Wwise need a declared license tier) | [33][54][56] |
| Pin by SHA for untagged libraries (sokol, bgfx, SDL_shadercross, sokol-shdc, Clay `main`) | [18][23][37][49] |

**Seams that need agent guidance (judgment or runtime behaviour):**

- Init and shutdown order across SDL, the GPU device, ImGui context and backends; `PrepareDrawData` before the render pass.
- Main-thread rules (SDL video, GLFW, the web) and choosing between main callbacks and a classic loop.
- Emscripten main-loop, ASYNCIFY and pthreads trade-offs.
- Audio-thread data sharing.
- Coordinate and projection conventions (depth range, handedness) when bridging ImGuizmo, raylib and SDL_GPU.
- Diagnosing "shader not working" with RenderDoc or Xcode, as SDL advises [1].

**Process implication.** Because SDL, sokol, RmlUi and ImGui restrict AI-generated contributions [13][20][44][48], the agent should keep fixes as local patches or glue recipes in the user's repo and draft issue text for a human to review. It should never file upstream issues or PRs on its own. Private registries are required for console SDL forks, PRIVATE GPU drivers, GDKX, the GNM backend and middleware console SDKs.

---

## Open questions

1. Does SDL plan an SDL_GPU WebGPU backend? None exists in code or history [14][62], and the SDL forum thread found was inconclusive [64]. wiki.libsdl.org and Discourse could not be read directly.
2. Is bgfx's Emscripten WebGPU path production-ready? The README ("Dawn Native only") conflicts with 2026 commits [23][62].
3. Current FMOD license thresholds (budget $500K vs $600K), tier prices, console download gating, and the current FMOD/Wwise SDK versions. Blocked: fmod.com and audiokinetic.com were unreachable, and the search budget was exhausted.
4. Do the private SDL console forks implement SDL_GPU through the PRIVATE driver, and which shader toolchain feeds it? (Inference from [1][14] only.)
5. Will `webgpu.h` reach a tagged 1.0? Today there are no tags [31], and wgpu-native still carries `wgpu.h` extensions [29].
6. Will the ImPlot maintainership shift (most 2026 commits by the ImPlot3D author) be formalised? [62]
7. When will Clay cut a release after v0.14, given the large unreleased changes on `main`? [49][62]

---

## Sources

All accessed 2026-10-29. "@sha" is the repository HEAD inspected.

1. SDL — `include/SDL3/SDL_gpu.h` (and `SDL_video.h`), libsdl-org/SDL @a45cd58 (2026-10-28). https://github.com/libsdl-org/SDL/blob/main/include/SDL3/SDL_gpu.h
2. SDL — `WhatsNew.txt`, 3.4.0 section. https://github.com/libsdl-org/SDL/blob/main/WhatsNew.txt
3. SDL — GitHub release "release-3.4.0" (published 1 Jan 2026). https://github.com/libsdl-org/SDL/releases/tag/release-3.4.0
4. SDL — tags release-3.4.0 … release-3.4.18 (3.4.18 tagged 2026-10-02). https://github.com/libsdl-org/SDL/tags
5. SDL — `docs/README-platforms.md`. https://github.com/libsdl-org/SDL/blob/main/docs/README-platforms.md
6. SDL — `docs/README-switch.md`, `docs/README-ps5.md`. https://github.com/libsdl-org/SDL/tree/main/docs
7. SDL — `docs/README-gdk.md`. https://github.com/libsdl-org/SDL/blob/main/docs/README-gdk.md
8. SDL — `docs/README-main-functions.md` and `include/SDL3/SDL_main.h`. https://github.com/libsdl-org/SDL/blob/main/docs/README-main-functions.md
9. SDL — `docs/README-emscripten.md`. https://github.com/libsdl-org/SDL/blob/main/docs/README-emscripten.md
10. SDL — `include/SDL3/SDL_stdinc.h` (`SDL_SetMemoryFunctions`). https://github.com/libsdl-org/SDL/blob/main/include/SDL3/SDL_stdinc.h
11. SDL — `include/SDL3/SDL_error.h`. https://github.com/libsdl-org/SDL/blob/main/include/SDL3/SDL_error.h
12. SDL — `docs/README-versions.md`. https://github.com/libsdl-org/SDL/blob/main/docs/README-versions.md
13. SDL — `AGENTS.md` / `CLAUDE.md` (LLM policy). https://github.com/libsdl-org/SDL/blob/main/AGENTS.md
14. SDL — `src/gpu/SDL_gpu.c` (backend list, format validation), `src/gpu/d3d12/SDL_gpu_d3d12.c`. https://github.com/libsdl-org/SDL/blob/main/src/gpu/SDL_gpu.c
15. SDL — `include/SDL3/SDL_audio.h`. https://github.com/libsdl-org/SDL/blob/main/include/SDL3/SDL_audio.h
16. raylib — `CHANGELOG` (6.2: 15 Oct 2026; 6.0: 23 Apr 2026). https://github.com/raysan5/raylib/blob/master/CHANGELOG
17. raylib — `src/raylib.h`, `src/rlgl.h`, `src/Makefile`, `src/platforms/`, `src/external/` (rlsw.h v1.6, miniaudio 0.11.25), `src/raudio.c`, `examples/Makefile`, `README.md`, `LICENSE` @89fa90f. https://github.com/raysan5/raylib
18. sokol — `README.md`, `CHANGELOG.md` (latest 27-Oct-2026) @bbac2ef. https://github.com/floooh/sokol
19. sokol — `sokol_gfx.h`, `sokol_app.h` header docs (feature matrix, allocator, backends). https://github.com/floooh/sokol/blob/master/sokol_app.h
20. sokol — `AGENTS.md` (LLM usage rules). https://github.com/floooh/sokol/blob/master/AGENTS.md
21. sokol — `util/sokol_imgui.h`. https://github.com/floooh/sokol/blob/master/util/sokol_imgui.h
22. sokol-tools — `docs/sokol-shdc.md`, `LICENSE`, `README.md` @61679bd. https://github.com/floooh/sokol-tools/blob/master/docs/sokol-shdc.md
23. bgfx — `README.md`, `include/bgfx/bgfx.h`, `include/bgfx/defines.h` (API 167), `LICENSE` @0d1856c. https://github.com/bkaradzic/bgfx
24. bgfx — `docs/tools.rst` (shaderc), `docs/build.rst` (GENie). https://github.com/bkaradzic/bgfx/tree/master/docs
25. GLFW — `README.md`, `include/GLFW/glfw3.h` (3.6.0 dev) @92dcf4c. https://github.com/glfw/glfw
26. GLFW — Releases page (3.5.1 latest; 3.5 skipped). https://github.com/glfw/glfw/releases
27. Dawn — `README.md`, `docs/support.md`, `docs/quickstart-cmake.md` @40d98d8. https://github.com/google/dawn
28. Dawn — GitHub Releases (v20261027.184652; nightly binaries; emdawnwebgpu). https://github.com/google/dawn/releases
29. wgpu-native — `README.md`, `CHANGELOG.md`, `.gitmodules` @589e6ca. https://github.com/gfx-rs/wgpu-native
30. wgpu-native — Releases (v29.0.1.1, 2026-06-23). https://github.com/gfx-rs/wgpu-native/releases
31. webgpu-headers — repository refs (no tags; checked with `git ls-remote`). https://github.com/webgpu-native/webgpu-headers
32. Slang — `README.md`, `LICENSE`, `AGENTS.md` @01dcd35; tag v2026.20 (2026-10-16). https://github.com/shader-slang/slang
33. glslang — `README.md` (HLSL deprecation), `LICENSE.txt`, `CONTRIBUTING.md` @64086a2; tag 16.6.0. https://github.com/KhronosGroup/glslang
34. SPIRV-Cross — `README.md`, `LICENSE`, `CONTRIBUTING.md` @57e4efc. https://github.com/KhronosGroup/SPIRV-Cross
35. DirectXShaderCompiler — `README.md`, `LICENSE.TXT`, `AIToolPolicy.md` @0d6c746. https://github.com/microsoft/DirectXShaderCompiler
36. DirectXShaderCompiler — Releases (v1.9.2609 latest; v1.10.x SM 6.10 previews). https://github.com/microsoft/DirectXShaderCompiler/releases
37. SDL_shadercross — `README.txt`, `include/SDL3_shadercross/SDL_shadercross.h`, `AGENTS.md`, `LICENSE.txt` @1ff05be. https://github.com/libsdl-org/SDL_shadercross
38. Dear ImGui — `docs/CHANGELOG.txt` (1.93.0, 2026-10-14; 1.91.7 SDLGPU3; 1.92.7 WGVK) @2fb554b; docking @83ccc61. https://github.com/ocornut/imgui/blob/master/docs/CHANGELOG.txt
39. Dear ImGui — `backends/imgui_impl_sdlgpu3.cpp`, `backends/sdlgpu3/build_instructions.txt`. https://github.com/ocornut/imgui/blob/master/backends/imgui_impl_sdlgpu3.cpp
40. Dear ImGui — `backends/imgui_impl_sdl3.cpp`. https://github.com/ocornut/imgui/blob/master/backends/imgui_impl_sdl3.cpp
41. Dear ImGui — `examples/example_sdl3_sdlgpu3/main.cpp`. https://github.com/ocornut/imgui/blob/master/examples/example_sdl3_sdlgpu3/main.cpp
42. Dear ImGui — `imconfig.h`, `imgui.h`, `docs/FAQ.md`, `LICENSE.txt`. https://github.com/ocornut/imgui/blob/master/docs/FAQ.md
43. Dear ImGui — `backends/imgui_impl_wgpu.h`, `examples/example_sdl3_wgpu/CMakeLists.txt`, `examples/libs/emscripten/emscripten_mainloop_stub.h`. https://github.com/ocornut/imgui/blob/master/backends/imgui_impl_wgpu.h
44. Dear ImGui — `.github/pull_request_template.md`, `.github/FUNDING.yml`. https://github.com/ocornut/imgui/tree/master/.github
45. ImGuizmo — `src/ImGuizmo.h`, `LICENSE`, tags 1.9/1.10 @105e4fd. https://github.com/CedricGuillemet/ImGuizmo
46. ImPlot — `implot.h`, `README.md`, `LICENSE`, tags v1.0/v1.1 @72de278. https://github.com/epezent/implot
47. ImPlot3D — `implot3d.h`, `README.md`, `LICENSE`, tag v0.5 @c54525f. https://github.com/brenocq/implot3d
48. RmlUi — `readme.md`, `changelog.md` (6.3/6.2), `Backends/`, `CMakeLists.txt`, `Include/RmlUi/Config/Config.h`, `AI_POLICY.md` @ae4e0dc. https://github.com/mikke89/RmlUi
49. Clay — `clay.h`, `README.md`, `renderers/`, `LICENSE.md` @e6cc369; tag v0.14. https://github.com/nicbarker/clay
50. Nuklear — repository, tag v4.13.3. https://github.com/Immediate-Mode-UI/Nuklear
51. microui — repository @0850aba. https://github.com/rxi/microui
52. miniaudio — `README.md`, `CHANGES.md`, `miniaudio.h` (master 0.11.25; `dev` @b7ba259). https://github.com/mackron/miniaudio
53. SoLoud — `README.md`, `LICENSE`, `src/backend/` @e82fd32. https://github.com/jarikomppa/soloud
54. OpenAL Soft — `README.md`, `COPYING`, `CMakeLists.txt` (CXX_STANDARD 20), `ChangeLog` (1.25.0) @b746364; tag 1.25.2. https://github.com/kcat/openal-soft
55. Steam Audio — `README.md` (4.8.1), `core/src/core/phonon.h`, `core/CMakeLists.txt`, `fmod/doc/build-instructions.rst`, `wwise/build/FindWwise.cmake` @480dd64. https://github.com/ValveSoftware/steam-audio
56. Audiokinetic — "Wwise Pricing" page (read only through a search-engine summary). https://www.audiokinetic.com/en/wwise/pricing/
57. Audiokinetic — "Free Wwise for Indie Developers – Full Platform Access with Unlimited Sounds" (search-engine summary). https://www.audiokinetic.com/en/blog/free-wwise-for-indie-developers/
58. Audiokinetic — "Wwise Licensing and Pricing Philosophy" (search-engine summary; dated about 3.5 years before access). https://www.audiokinetic.com/en/blog/wwise-licensing-and-pricing-philosophy/
59. Game Developer — "Small developers and creators can now use FMOD Studio for free" (secondary). https://www.gamedeveloper.com/audio/small-developers-and-creators-can-now-use-fmod-studio-for-free
60. GameFromScratch — "FMOD Studio Now Free For Indie Game Developers" (secondary). https://gamefromscratch.com/fmod-studio-now-free-for-indie-game-developers/
61. StraySpark — "Wwise vs FMOD vs MetaSounds: Choosing Audio Middleware for Your UE5 Game in 2026" (secondary; figures flagged by its author as needing verification). https://www.strayspark.studio/blog/wwise-fmod-metasounds-audio-middleware-comparison
62. Commit and tag statistics computed from treeless git clones of each repo above (window 2025-10-29..2026-10-29; bots excluded; author-name counts). Also covers tag dates and history searches ("webgpu" in SDL; "OpenXR" in SDL; WebGPU/Emscripten in bgfx). Data: GitHub repositories as listed.
63. Emscripten — "Using WebGPU in Emscripten" (emdawnwebgpu port; seen in search results only). https://emscripten.org/docs/porting/multimedia_and_graphics/WebGPU-support.html
64. SDL Discourse — "The new SDL3 GPU functions are awesome but can they also be used for the browser solution?" (search-engine summary only). https://discourse.libsdl.org/t/can-sdl3-gpu-functions-shader-be-compiled-to-wasm/59691
