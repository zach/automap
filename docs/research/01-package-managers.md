# 01: C++ package managers, what they can express, and which to build on

**Scope.** This file compares vcpkg, Conan 2, xmake/xrepo, and CPM.cmake with CMake FetchContent as possible build backends for Automap. It covers manifests and versioning, features and options, how each expresses compatibility, private registries and console platforms, platform reach, supply-chain hygiene (including release-age cooldowns), CMake integration (dependency providers, CPS, modules), and catalog coverage for 23 game-engine libraries. It ends with a scored recommendation. Build2, Bazel, Meson and Nix get a short note each, only where they bear on the backend choice.

**Accessed: 2026-10-29.**

**Method and limits.** Version numbers come from shallow clones of the three catalogs taken on the day: `microsoft/vcpkg` @ `d4c0a46` (2026-10-28), `conan-io/conan-center-index` @ `7fd5753` (2026-10-29) and `xmake-io/xmake-repo` @ `cf09fff` (2026-10-29). Upstream versions come from `git ls-remote`, and tag dates from the tagged commits. This session could not reach learn.microsoft.com, docs.conan.io, cmake.org, xmake.io or github.blog, so I read the **source repositories** of those docs on GitHub instead (`microsoft/vcpkg-docs` @ `eda2ef2`, `conan-io/docs` @ `25c21b6`, `xmake-io/xmake-docs` @ `9840261`, `Kitware/CMake` Help/ @ `9325f20`). The shared web-search budget ran out partway through. The claims affected are marked **[Unverified]**.

---

## Key findings

1. **[Verified]** **vcpkg has the most current catalog for Automap's test list.** It carries 21 of the 23 libraries, and 18 of those are at the latest upstream release. That includes Dear ImGui 1.93.0, which upstream tagged on 2026-10-14, and Box2D 3.2.0, tagged on 2026-10-18. Two are stale: sokol is still the 2023-10-07 snapshot, and bgfx is 1.129 against 1.167 upstream. ufbx and ozz-animation are missing. xmake-repo carries 22 libraries, 10 of them current. Conan Center carries 20, 9 of them current, and lags most on game libraries: Jolt is at 5.2.0 against 5.6.0, and GameNetworkingSockets at 1.4.1 against 1.6.1 [16][32][38][54].
2. **[Verified]** **All three catalog managers can express "library X built with feature Y", but not equally well.**
   - **Dear ImGui with its SDL3 and SDL_GPU backends.** vcpkg (`imgui[sdl3-binding,sdlgpu3-binding]`) and xmake (`configs = {sdl3 = true, sdl3_gpu = true}`) both let you select this. Conan Center's imgui recipe only compiles the SDL3 platform backend and leaves the other backends as loose source files [16][32][38].
   - **Jolt with double precision.** Only xmake exposes it. vcpkg policy forbids features that act as mutually exclusive alternatives, so getting double precision from vcpkg needs an overlay port or per-port triplet flags [6][7].
3. **[Verified]** **The managers describe compatibility in different kinds of language.**
   - vcpkg uses a closed platform-expression vocabulary plus triplets. The vocabulary has an identifier for Xbox but none for other consoles [4].
   - Conan uses code: `validate()` over the resolved graph, `compatibility()`, package_id modes, and `package_id_abi_options` (new in 2.26) [22][23][19].
   - xmake uses imperative Lua (`on_check` with `raise`).
   - CPM and FetchContent express nothing.

   None of them can express semantic seams such as allocator ownership, threading, math types or init order. **[Inference]**
4. **[Verified]** **No C++ package manager has a native release-age cooldown.** I found none in the vcpkg-tool, Conan, xmake or CPM docs or release notes. The only route today is outside the manager:
   - Dependabot's `cooldown` supports the `vcpkg` ecosystem, but only through `default-days`, and it applies to the `builtin-baseline` bumps Dependabot opens. Dependabot's cooldown now defaults to 3 days [49][14].
   - Renovate has a Conan manager, but its Conan datasource returns no release timestamps [48].
   - By contrast, npm (11.10, 2026-02), pnpm (11.0, on by default) and uv ship cooldowns natively [45][46][47].
5. **[Verified]** **Supply-chain features (SBOMs, signing, vulnerability scanning) are strongest in Conan, while vcpkg has the strongest provenance basics.**
   - vcpkg writes an SPDX 2.3 SBOM for every port and pins every download by SHA512. Since 2026-07 its SBOMs also carry PURLs [13][2].
   - Conan has experimental CycloneDX SBOMs, `conan audit` for CVE scanning, and `conan cache sign/verify` [19][27][28].
   - xmake pins a sha256 hash per version and nothing more [34].
   - CMake 4.5, at release candidate 2 on 2026-10-28, adds `install(SBOM)` [40].
6. **[Verified]** **CMake itself now reads and writes the Common Package Specification (CPS).** Since CMake 4.3 (2026-03-17), `find_package` can import CPS files, and `install(PACKAGE_INFO)` and `export(PACKAGE_INFO)` can write them [40]. The spec is still pre-1.0 (0.15.0) [44]. Conan 2.25+ supports a full round trip [19], while the vcpkg and xmake docs never mention CPS.
   - Only Conan's integration (cmake-conan) uses CMake's dependency-provider mechanism [31]. vcpkg uses a toolchain file [15], and CPM overrides `find_package` [39].
7. **[Verified]** **vcpkg is the only manager with a public console precedent: Xbox.** It ships community triplets for Xbox and a stub port (`ms-gdkx`) that locates the NDA GDK through environment variables [8]. I found no public PlayStation or Switch material in any of the four tools' docs. Conan has the strongest private infrastructure: Artifactory or `conan_server` remotes, a local recipes index, source backups, cache save/restore, and `settings_user.yml` for custom platforms [24][25][26].
8. **[Inference]** **The prior-conversation lead is right that managers don't solve semantic integration, but it undersells what they already encode.** Feature-to-dependency edges, propagation of ABI-affecting defines, and file-conflict rules are partial compatibility knowledge that Automap should reuse rather than duplicate.
9. **[Inference]** **Recommended backends:**
   - **vcpkg** as the primary backend: JSON manifests with a published schema, the most current catalog, a native fit with CMake, the Xbox precedent, and overlays for NDA platforms and for build variants vcpkg can't express as features.
   - **Conan** as a second backend for the custom-integration tier.
   - **CPM or FetchContent** as an escape hatch.

   Supporting more than one backend is feasible if Automap's own manifest is backend-neutral and its glue targets CMake imported targets.

---

## 1. Snapshot

| Tool | Current release (date) | Manifest language | Curated catalog |
|---|---|---|---|
| vcpkg | tool `2026-09-26`; registry `2026.10.16` (published 2026-10-20) [1][2] | JSON (`vcpkg.json`, `vcpkg-configuration.json`) with published JSON Schemas [17] | 2,877 ports per the release notes; 2,893 port directories at HEAD [1][16] |
| Conan 2 | 2.34.0 (2026-10-28) [19] | `conanfile.txt` (declarative, INI-like) or `conanfile.py` (Python) | ConanCenter: 1,974 recipe directories [32] |
| xmake / xrepo | v3.1.2 (2026-10-16) [33] | `xmake.lua` (Lua) | xmake-repo: 2,032 packages [38] |
| CPM.cmake | v0.43.2 (2026-09-24) [39] | CMake code | none (fetches from any URL) |
| FetchContent | ships with CMake: 4.4.5 (2026-10-27); 4.5.0-rc2 (2026-10-28) [40] | CMake code | none |

GitHub stars on 2026-10-29 were vcpkg 27.6k, Conan 9.5k, xmake 12.3k and CPM 4.2k [56]. **[Unverified]** I could not reach survey data on package-manager use among game developers (such as the JetBrains C++ ecosystem survey) after the search budget ran out, so stars are only a weak proxy.

## 2. Manifests, versioning, reproducibility

| | vcpkg | Conan 2 | xmake | CPM / FetchContent |
|---|---|---|---|---|
| **Version schemes** | `version`, `version-semver`, `version-date`, `version-string`, plus `#port-version`. Versions in different schemes are never compared [3] | Semver-like ordering with pre-releases and builds [20] | Semver strings; branches and commits also accepted [34] | Git tag, commit or URL |
| **Constraints** | `builtin-baseline` (a registry commit that sets a version floor for everything), `version>=` (lower bound only), and `overrides` (an exact pin that wins over everything else). vcpkg picks the **lowest** version that satisfies every constraint, so ranges and upper bounds aren't needed [3]. Baselines in transitive dependencies are ignored [3] | Ranges such as `[>=1.0 <2]`, `[~2.5]`, `[^1.2]` and `\|\|`; the latest version in range wins; pre-releases are opt-in [20]. `force` and `override` traits resolve conflicts [21] | `">=1.1.0 <1.2"`, `"~1.6"`, `"1.2.x"`. Only versions that exist in the repo, each with a sha256, are accepted unless `{verify = false}` [34] | CPM: `VERSION` plus `GIT_TAG` (defaults to `v<VERSION>`). In a diamond, the first version added wins and CPM warns [39] |
| **Lock and reproducibility** | No lockfile. The baseline commit plus overrides act as the lock, and every download is checked against a SHA512 [3][16] | `conan.lock` holds versions and **revisions**, which are content hashes of the recipe and binaries [19][20] | `set_policy("package.requires_lock", true)` writes `xmake-requires.lock`, recording the version and xmake-repo commit for each platform and architecture [34] | `CPMUsePackageLock(package-lock.cmake)` plus the `cpm-update-package-lock` target. The README recommends pinning commit hashes [39] |
| **Machine editing** | `vcpkg add port`, `x-update-baseline`, `x-add-version` [2] | `conan require add` (experimental); `graph info --format=json` | `xrepo` CLI | none |

## 3. Features, options, and the two probes

| | Variant mechanism | Constraints on variants | How a consumer selects |
|---|---|---|---|
| vcpkg | `features` and `default-features`. A feature can declare its own dependencies and its own `supports` [4] | **Must be additive**: "If `port[featureA]` installs and `port[featureB]` installs, then `port[featureA,featureB]` must install." Alternatives belong in overlay ports [6] | `{"name":"imgui","features":[…],"default-features":false}`. Anything else goes through triplet variables, set per port via `PORT` (for example `VCPKG_CMAKE_CONFIGURE_OPTIONS`) [7] |
| Conan | `options` (any values), `settings` (os, arch, compiler, build_type, cppstd…), `conf`, and profiles | Values can be non-additive, and `validate()` can reject a combination [22] | `pkg/*:opt=val` in a profile or conanfile, or as a trait in `requires(…, options={})`. Conan's docs warn that options set in recipes have weak guarantees and recommend profiles [21] |
| xmake | Per-package `configs` (bool, string or enum) plus built-ins (shared, debug, pic, runtimes) | Non-additive allowed | `add_requires("p", {configs = {…}})`, and `add_requireconfs("a.b", …)` for transitive packages [34] |
| CPM / FetchContent | `OPTIONS "X ON"` becomes cache variables; needs policies CMP0077 and CMP0126 [39] | No model at all: whatever the library's CMake accepts | Raw CMake |

**Probe A: Dear ImGui with its SDL3 platform backend and SDL_GPU renderer backend.**
- **vcpkg** (imgui 1.93.0) has features `sdl3-binding`, `sdl3-renderer-binding` and `sdlgpu3-binding`. `sdlgpu3-binding` depends on `imgui[sdl3-binding]` and `sdl3`, and there are also `docking-experimental`, `vulkan-binding` and `webgpu-binding` [16].
- **xmake** (1.92.9 and 1.92.9-docking) has configs `sdl3`, `sdl3_renderer`, `sdl3_gpu`, `vulkan`, `wgpu` and others [38].
- **Conan Center** (1.92.9b and 1.92.9b-docking) has only the option `with_sdl3_binding`. Every `imgui_impl_*` backend is copied into `res/bindings`, and the consumer compiles the ones it needs [32].
- **CPM:** Dear ImGui has no root `CMakeLists.txt` at v1.93.0, so the consumer has to write the build itself [58].

**Probe B: Jolt with double precision.**
- **vcpkg** (5.6.0) offers only the features `debugrenderer`, `profiler` and `rtti` [16]. Because features must be additive, double precision needs one of two workarounds:
  - an overlay port, or
  - a per-port triplet line such as `if(PORT STREQUAL "joltphysics") list(APPEND VCPKG_CMAKE_CONFIGURE_OPTIONS -DDOUBLE_PRECISION=ON) endif()`, which works because the portfile uses `vcpkg_cmake_configure` [7]. **[Inference, untested]**
- **xmake** (5.6.0) has a `double_precision` config that adds the `JPH_DOUBLE_PRECISION` define, alongside `object_layer_bits`, `cross_platform_deterministic` and the instruction-set configs [38].
- **Conan Center** (5.2.0) has only `shared` and `fPIC`, so the recipe would need a fork [32].
- **CPM:** `OPTIONS "DOUBLE_PRECISION ON"`. Jolt's CMake makes the define `PUBLIC` [55].

## 4. Compatibility expression

| Concern | vcpkg | Conan 2 | xmake | CPM / FetchContent |
|---|---|---|---|---|
| **Platform support** | `supports`, a platform expression over a fixed set of identifiers: `x64 arm64 wasm32 windows uwp xbox linux android ohos apple osx ios tvos visionos emscripten qnx vxworks static staticcrt native …`. An install plan that breaks it is refused unless `--allow-unsupported` [4][18] | `validate()` and `validate_build()` raise `ConanInvalidConfiguration` (exit code 6) and can inspect `self.dependencies[...].options` once the graph is resolved [22] | `on_check` with `raise(...)`, e.g. box2d "unsupported 32-bit" and ktx's arm64 limits [38] | none |
| **CRT and linkage** | Triplet variables `VCPKG_CRT_LINKAGE` and `VCPKG_LIBRARY_LINKAGE` [7] | `compiler.runtime` setting and `shared` option | Built-in `shared` and `runtimes` configs | The consumer's toolchain |
| **Binary identity and ABI** | Per-triplet builds; per-port customization inside triplets [7] | package_id from settings, options and dependencies, with embed and non-embed modes [23]. A `compatibility()` method and plugin; the default plugin treats different `cppstd` values as compatible [23]. Since 2.26, **`package_id_abi_options`** feeds a dependency option such as `shared` into consumers' package_id [19][23] | Hash of the config set | none |
| **Conflicts and mutual exclusion** | No field for it. Curation rule: "no two ports a user is expected to use concurrently may provide the same file", enforced by `FILE_CONFLICTS` in CI [6] | Version conflicts are errors; anything else has to be coded in `validate()` | none declarative | First declaration wins |
| **ABI-affecting defines** (Jolt `JPH_*`) | Comes from Jolt's own CMake config. For DLL builds the port also patches `Core.h` to hard-code `JPH_SHARED_LIBRARY` and others [16] | Hand-written `cpp_info.defines` (`JPH_OBJECT_STREAM`, AVX/SSE flags, `JPH_OBJECT_LAYER_BITS=16`), a copy of Jolt's defines that can drift from upstream [32] | `package:add("defines", …)` set per config [38] | The library's CMake `PUBLIC` definitions |

Jolt shows the stakes. Its CMake emits instruction-set defines "to ensure that child projects use the same settings… (a mismatch causes link errors)". It also offers a runtime check, `VerifyJoltVersionID()`, "to verify the library ABI is compatible with your" build [55]. Allocator policy reaches the build only as a flag (`JPH_DISABLE_CUSTOM_ALLOCATOR`) [55].

**What none of them can express (all [Inference] from the manifest schemas and recipe APIs reviewed above):**
- who owns allocation and which allocator hooks are installed when;
- which job system or threads a library may use, and whether it shares the engine's;
- math-type interop, for example Jolt `Vec3` against the engine's math library;
- handle and ownership conventions;
- lifecycle and init order, e.g. `SDL_Init` → GPU device → `ImGui_ImplSDLGPU3_Init`;
- error-handling policy beyond an exceptions on/off option;
- serialization formats;
- evidence-backed **pairings** across libraries, such as "these versions were tested together on Android".

The closest any of them come is Conan's `validate()`, which can assert facts about dependency options but not about runtime behavior.

## 5. Registries, private use, NDA platforms, offline, caching

| | vcpkg | Conan 2 | xmake | CPM / FetchContent |
|---|---|---|---|---|
| **Private catalogs** | Git and filesystem registries (a ports tree plus a versions database), with patterns routing packages to registries; `default-registry` can point at an internal mirror [5][9] | Remotes: Artifactory or `conan_server`. A **local-recipes-index** remote (experimental) serves a git repo in conan-center-index layout [25]. ConanCenter itself recommends hosting your own fork for production [29] | `add_repositories("name url-or-dir [branch]")` [34]; private source or binary packages [36] | Any git URL; no registry |
| **Overlays and customization** | `overlay-ports`, `overlay-triplets` [5]; triplet chainloading (`VCPKG_CHAINLOAD_TOOLCHAIN_FILE`) [7] | Profiles; `settings_user.yml` adds operating systems, architectures and compilers [24]; `conan config install` | Custom platforms and toolchains [37] | `CPM_<name>_SOURCE` points a package at a local checkout [39] |
| **Offline / air-gapped** | `vcpkg export` bundles, or git `insteadOf` plus asset caching [10]. Asset cache keyed by SHA512, with `x-block-origin` and `x-script` [12] | **Backup sources** keyed by sha256 [26]; `conan cache save/restore` [26]; a private remote | `pkg_searchdirs`, PAC mirrors and proxies, `xrepo export` [36] | `CPM_SOURCE_CACHE` allows offline configure [39] |
| **Binary cache** | Providers: `files`, `nuget`, `http`, and experimental `x-azblob`, `x-gcs`, `x-aws`, `x-cos`, `x-az-universal`. `x-gha` has been **removed** [11] | Native: binaries are first-class artifacts on remotes | Install "schemes" fall back from binary to source (since v3.0.7) [37] | None ("No pre-built binaries") [39] |

**NDA consoles.**
- **Xbox (vcpkg).** Community triplets `x64-xbox-xboxone[-static]` and `x64-xbox-scarlett[-static]` exist; they build with `WINAPI_FAMILY_GAMES` [8].
  - Most libraries build against the public Windows SDK.
  - Code that needs Direct3D 12.X declares a dependency on the `ms-gdkx` stub port, which finds the NDA *GDK with Xbox Extensions* through `GameDKCoreLatest` and `GameDKXboxLatest` (GDK October 2025 or later) [8].
  - These triplets are **not tested in vcpkg CI**, which builds 13 jobs across Windows, Linux, macOS arm64 and Android [8][16].
- **PlayStation and Switch: no public material.** Searches of the vcpkg, Conan and xmake doc repositories for PlayStation, Nintendo and console platforms found nothing. **[Unverified]** what studios actually do.
- **[Inference]** The natural mechanisms are a private overlay triplet with a chainloaded NDA toolchain (vcpkg), `settings_user.yml` plus a profile (Conan), or a custom xmake platform. One catch for vcpkg: platform expressions have no identifier for those systems, so a port's `supports` cannot name them, and negative expressions like `!windows` will silently evaluate true.

## 6. Platform reach and cross-compilation

| | Windows / macOS / Linux | Android | iOS | wasm (Emscripten) | Consoles | Cross model |
|---|---|---|---|---|---|---|
| vcpkg | Built-in triplets, CI-tested [16] | Built-in arm64, x64 and arm-neon triplets, CI-tested [16] | Community triplets for ios, ios-simulator, tvos and visionos [16] | Community triplet `wasm32-emscripten` [16] | Xbox community triplets; `xbox` platform identifier [4][8] | Separate host and target triplets; `"host": true` dependencies [4] |
| Conan | Yes; ConanCenter publishes binaries for these [29] | Yes; NDK example [30] | Yes (`iOS`, `tvOS`, `watchOS`, `visionOS` settings) [24] | `os=Emscripten`, `arch=wasm/wasm64`, `compiler=emcc`; example [24][30] | None in default settings; extend with `settings_user.yml` [24] | Two profiles, build and host |
| xmake | Yes | `-p android` | `-p iphoneos`, plus `appletvos`, `watchos`, `applexros` | `-p wasm` | none public; `cross` and custom platforms [37] | `-p`/`-a` with an SDK path; xmake-repo CI covers android, iphoneos, wasm, mingw, freebsd and cross [38] |
| CPM / FetchContent | Whatever each library's CMake supports, built in the consumer's single toolchain | ← | ← | ← | ← | No host/target split for tools **[Inference]** |

## 7. Supply-chain hygiene

| | Hash pinning | SBOM | Signing / provenance | Vulnerability scan | Release-age cooldown |
|---|---|---|---|---|---|
| vcpkg | SHA512 on every download [16]; registry versions database | SPDX 2.3 written to `share/<port>/vcpkg.spdx.json` [13]. PURLs since tool 2026-07-13; git tree gitoids in 2026-07-24, a pre-release that was not promoted [2]. CycloneDX only by converting [13] | None for packages. GitHub dependency-graph submission works [2]. The tool's own CI actions are pinned to commit SHAs since 2026-09-26 [2] | Through GitHub's dependency graph | **None native.** Dependabot `cooldown` supports the `vcpkg` ecosystem with `default-days` only (no semver-level settings) and applies to baseline bumps; Dependabot's default cooldown is 3 days [49][14] |
| Conan 2 | sha256 in `conandata.yml`; revisions; lockfiles [20][26] | CycloneDX 1.4 and 1.6 via `conan.tools.sbom`, **experimental** [27]; SBOM deployers (2.26); SPDX license expressions (2.30) [19] | Signing plugin (cosign, GPG, OpenSSL…; experimental) [28]; `conan cache sign/verify` (2.26, 2026-02-25) [19] | `conan audit` (2.14, 2025-03-12) [19] | **None.** Dependabot has no Conan ecosystem [49]. Renovate has a `conan` manager [48], but the datasource returns no timestamps; with Renovate's default `minimumReleaseAgeBehaviour: timestamp-required`, a configured age gate would never pass **[Inference from source]** [48] |
| xmake | sha256 per version; `verify = false` bypasses it [34] | None documented | None documented | None | None. xmake-repo's daily auto-update workflow uses tag-pinned actions and the xmake `master` branch with a write token [38] (an observation about the repo, not about the client) |
| CPM / FetchContent | `URL_HASH`; commit-hash pins [39] | None for dependencies. CMake 4.5 adds `install(SBOM)` and `export(SBOM)` for your own exports (release candidate) [40] | None | None | None |

**State of the art elsewhere, for comparison:**

| Tool | Setting | Since | Default |
|---|---|---|---|
| npm | `min-release-age` (days), plus `min-release-age-exclude` | 11.10.0, 2026-02-11 [45] | off |
| pnpm | `minimumReleaseAge` (minutes) | 10.x; **on by default** in 11.0, 2026-04-28 [46] | 1440 (one day) |
| uv | `exclude-newer` accepts relative durations ("7 days") | 0.9.17, 2025-12-09 [47] (secondary source) | off |
| Yarn | `npmMinimalAgeGate` | 4.10 [50] (secondary source) | reported on |
| Renovate | `minimumReleaseAge`, which needs release timestamps from the datasource | — [48] | off |
| Dependabot | `cooldown` | 2025; default 3 days by 2026 [49] | 3 days |

**Conclusion [Verified absence, Inference on implication]:** no C++ manager has an equivalent. The closest thing is a cooldown on vcpkg baseline commits through Dependabot. That window is measured from the registry commit, not from when the upstream library was released, so it is coarse.

## 8. CMake integration, dependency providers, CPS, modules

- **How each connects to CMake:**
  - **vcpkg:** a toolchain file (`CMAKE_TOOLCHAIN_FILE`, best set in presets), with a chainload hook for your own toolchain. It does *not* use dependency providers [15].
  - **Conan:** the `CMakeToolchain` and `CMakeDeps`/`CMakeConfigDeps` generators. The `cmake-conan` develop2 branch *is* a dependency provider, injected with `CMAKE_PROJECT_TOP_LEVEL_INCLUDES` (needs CMake 3.24+). It runs `conan install` at the first `find_package`, and only mainstream compilers are detected automatically [31].
  - **xmake:** the `xrepo.cmake` function `xrepo_package()` adds install prefixes to `CMAKE_PREFIX_PATH` [36].
  - **CPM:** overrides `find_package`; `CPM_USE_LOCAL_PACKAGES` tries `find_package` first [39].
- **Dependency providers** arrived in CMake 3.24 (`cmake_language(SET_DEPENDENCY_PROVIDER … SUPPORTED_METHODS FIND_PACKAGE FETCHCONTENT_MAKEAVAILABLE_SERIAL)`). FetchContent's `FIND_PACKAGE_ARGS` and `FETCHCONTENT_TRY_FIND_PACKAGE_MODE` let a `FetchContent_Declare` hand the request to `find_package`, and so to a provider [41]. **[Inference]** That makes "`find_package` names and imported targets" the common contract across all four backends.
- **CPS.**
  - **CMake 4.3** (2026-03-17) made CPS import (`find_package`) and export (`install/export(PACKAGE_INFO)`) official. The `project()` command gained `COMPAT_VERSION` and `SPDX_LICENSE` [40].
  - **CMake 4.5** (release candidate) adds CPS executable components and header file sets [40].
  - Generating CPS from `install(EXPORT)` is still behind an experimental flag [42].
  - The spec is at 0.15.0 [44].
  - Conan 2.25 through 2.34 added a full CPS round trip, components and `dyld_requires` (CMake ≥ 4.4) [19]. The vcpkg and xmake docs don't mention CPS.
- **C++20 modules** are not yet a reason to pick a manager. CMake supports named modules with Ninja and Visual Studio generators (from 3.28). `import std` moves out of experimental in 4.5. Header units are unsupported, and the Visual Studio generators cannot export or install module information [43]. **[Inference]** Module-based libraries will force rebuilds of module interfaces in the consumer, which favors source-building backends, but none of the 23 probe libraries depends on modules.

## 9. Catalog coverage (snapshots of 2026-10-28/29)

✓ = at the latest upstream release or tag; ↓ = behind; ✗ = absent. Upstream from [54].

| Library | Upstream latest (tag date) | vcpkg [16] | Conan Center [32] | xmake-repo [38] |
|---|---|---|---|---|
| SDL3 | 3.4.18 (2026-10-02) | ✓ `sdl3` 3.4.18 | ↓ `sdl` 3.4.16 | ↓ `libsdl3` 3.4.16 |
| SDL3_shadercross | no tags or releases [57] | `sdl3-shadercross` "3.0.0-preview2" (pinned commit) | ✗ | ✗ |
| Dear ImGui | 1.93.0 (2026-10-14) | ✓ 1.93.0; backend features incl. SDL3/SDL_GPU | ↓ 1.92.9b; SDL3 platform backend only | ↓ 1.92.9; `sdl3`, `sdl3_gpu` configs |
| Jolt | 5.6.0 (2026-07-11) | ✓ 5.6.0#1 (no double precision) | ↓ 5.2.0 | ✓ 5.6.0 (`double_precision`) |
| Box2D v3 | 3.2.0 (2026-10-18) | ✓ 3.2.0 | ↓ 3.1.1 | ↓ 3.1.1 |
| flecs | 4.1.6 (2026-06-28) | ✓ | ✓ | ✓ |
| EnTT | 4.0.0 (2026-07-22) | ✓ | ✓ | ✓ |
| fastgltf | 0.9.1 (2026-09-27) | ✓ | ✓ | ↓ 0.9.0 |
| cgltf | 1.15 (2025-02-09) | ✓ | ✓ | ✓ |
| ufbx | 0.23.1 (2026-09-25) | ✗ | ✗ | ↓ 0.23.0 |
| meshoptimizer | 1.3 (2026-09-25) | ✓ | ↓ 1.0 | ↓ 1.2 |
| KTX-Software | 4.4.2 (2025-10-04) | ✓ `ktx` | ✓ | ✓ |
| Basis Universal | 2.50 (2026-07-22) | ✓ `basisu` 2.50 (`!android`) | ↓ `libbasisu` 1.15.0 | ↓ 1.16.4 |
| ozz-animation | 0.17.0 (2026-08-01) | ✗ | ↓ 0.14.1 | ↓ 0.16.0 |
| recastnavigation | 1.6.0 (2023-05-21) | ✓ | ✓ | ✓ |
| miniaudio | 0.11.25 (2026-03-04) | ✓ | ✓ | ✓ |
| Tracy | 0.14.1 (2026-08-22) | ✓ | ✓ | ✓ |
| sokol | no releases (master) | ↓ snapshot 2023-10-07 | snapshot `cci.20251113` | snapshot `2026.09.14` |
| bgfx | bgfx.cmake 1.167.9593-603 (2026-10-28) | ↓ 1.129.8940-496 | ↓ 1.146.9306-550 | ↓ build 9392 |
| Slang | 2026.20 (2026-10-17) | ✓ `shader-slang` 2026.20 (x64/arm64; Windows, Linux, macOS only) | ✗ | ↓ 2026.16 |
| sol2 | 3.5.0 (2025-02-25) | ✓ | ✓ | ✓ |
| GameNetworkingSockets | 1.6.1 (2026-10-02) | ✓ | ↓ 1.4.1 | ↓ 1.6.0 |
| msdfgen | 1.13 (2025-11-30) | ✓ 1.13+new-skia-api | ↓ 1.12 | ✓ |
| **Totals** | | **21 carried / 18 current** | **20 / 9** | **22 / 10** |

Other ecosystems, checked by directory listing [51][52][53]:
- **Bazel Central Registry** (1,392 modules) has `sdl3`, `imgui`, `meshoptimizer` and `basis_universal`, and none of the physics, ECS, glTF, networking or audio libraries in the list.
- **Meson WrapDB** (369 wraps) has `sdl3`, `imgui` and `imgui-docking`, `box2d`, `entt`, `miniaudio` and `tracy`.
- **nixpkgs** has sdl3, imgui, box2d, entt, meshoptimizer, tracy and recastnavigation under `by-name`. Jolt and flecs were not at the paths I probed.

## 10. Other options (only as they inform the choice)

- **Bazel with bzlmod and the BCR** gives hermetic builds and hash-pinned modules. Game-library coverage is thin (see §9), and it would take studios off CMake. Not a fit as Automap's backend. **[Inference]**
- **Meson wraps** cover a handful of the probe libraries [52]; the Meson build system is uncommon in game engines. **[Inference]**
- **Nix** is strongest on reproducibility, but has no Windows or console story, which rules it out for game studios. **[Inference]**
- **build2 / cppget.org:** **[Unverified]**; the site was unreachable and the search budget was spent. It is unlikely to change the decision, given its separate build system.

## 11. Which backend to build on, and can Automap support several? [Inference]

Scores run from 1 (weak) to 5 (strong) and are judgment calls grounded in the sections above.

| Criterion | vcpkg | Conan 2 | xmake | CPM / FetchContent |
|---|---|---|---|---|
| Ubiquity and trust among C++ and game developers (stars as a proxy; survey data unverified) | 5 | 4 | 2 | 4 (FetchContent ships with CMake) |
| Catalog breadth and currency for game libraries (§9) | 5 | 2 | 3 | n/a: upstream directly |
| Fit with CMake | 5 | 4 | 2 | 5 |
| Private registries and overlays | 4 | 5 | 4 | 2 |
| Offline / air-gapped use | 4 | 5 | 3 | 3 |
| Binary caching | 4 | 5 | 3 | 1 |
| Reach to wasm, mobile and consoles | 4 (Xbox precedent; community triplets untested) | 4 | 4 (broad CI, no console) | 3 |
| Pinning and reproducibility | 4 (no lockfile; baselines plus overrides) | 5 | 4 | 3 |
| Supply-chain features | 3 | 4 | 1 | 1 |
| Variant expressiveness (non-additive build variants, ABI propagation) | 2 | 5 | 4 | 2 |
| Agent-friendliness: declarative manifest, schema, machine-readable CLI | 5 | 3 (`.txt` is limited; `.py` is code) | 3 | 2 |
| **Unweighted total (of 55)** | **45** | **46** | **33** | **26 (+n/a)** |

**Reading the table.** vcpkg and Conan are effectively tied, and they win on different axes.
- **vcpkg** wins on what Automap's *testbed and starter-engine tiers* and its *agent* need: a current catalog of the exact libraries, a pure-JSON manifest an agent can write and validate against a schema, and zero-friction CMake.
- **Conan** wins on what the *custom-integration tier* needs: private remotes, binaries as first-class artifacts, non-additive options, ABI propagation, signing and audit.
- **xmake** has the best variant coverage for these libraries, Jolt double precision and ImGui SDL_GPU included, but it sits outside CMake, which most game developers use.
- **CPM / FetchContent** is the universal fallback for unpackaged or bleeding-edge libraries (ufbx, ozz-animation, SDL_shadercross).

**Recommendation:**
- **vcpkg as the primary backend**, used through `vcpkg.json` and `vcpkg-configuration.json`. Automap ships an **Automap registry or overlay ports** for:
  - gaps in the catalog (ufbx, ozz-animation, fresh sokol and bgfx);
  - non-additive variants (a `joltphysics` overlay with double precision);
  - NDA platforms, as private overlay triplets.
- **Conan as a second backend** for custom-integration studios that already run Artifactory.
- **CPM / FetchContent** as an escape hatch.

**More than one backend is feasible if:**
1. Automap's manifest is backend-neutral: library identity, a mapping to vcpkg port, Conan reference and xmake name, the required variant, the ABI-define set, and the pairing evidence.
2. Recipes and glue target CMake `find_package` names and imported targets, never backend-specific variables.
3. Each pairing is verified per backend, because a variant can exist in one backend and not another (Probe A and Probe B).

The cost is a verification matrix that grows with each backend, so v1 should probably verify vcpkg only.

---

## Leads checked

| Lead | Verdict | Evidence |
|---|---|---|
| "C++ package managers (vcpkg, Conan, xmake/xrepo, CPM/FetchContent) solve fetching and building. They do not solve which libraries fit together, how to wire them, or which layer owns what." | **Corrected (partly).** The "how to wire" and "which layer owns what" parts are **confirmed**: no manifest or recipe API has a concept of allocator, threading, math-type, lifecycle or ownership seams (§4). "Which libraries fit together" is **too strong**: the managers do encode build-level fit through feature-to-dependency edges (`imgui[sdlgpu3-binding]` → `sdl3`), dependency-option checks (`validate()`), ABI-define propagation, `package_id_abi_options`, platform `supports`, and file-conflict rules. They also give partial wiring hints in exported CMake targets and `vcpkg print-usage` (2026-07) [2]. What they lack is semantic and runtime fit, plus cross-library pairing evidence. | [4][6][16][22][23][32][38][55] |
| Brief's framing: vcpkg "baselines, `version>=`, overrides"; Conan "version ranges and lockfiles"; `validate()` and `compatibility()`; package_id modes; vcpkg git and filesystem registries, overlay ports and triplets; Conan remotes and local recipes index; xmake custom repos | **Confirmed**, with three additions: vcpkg has no lockfile and no upper bounds; Conan's local-recipes-index is experimental; vcpkg's `x-gha` cache provider has been removed | [3][5][11][20][23][25][34] |

## Implications for Automap (all inference)

1. **Don't replace the package manager; sit above it.** Automap's unique content is the layer no manager can express (§4): seam contracts, pairings with evidence, recipes, and init order. Variant selection and ABI defines should be *compiled down* into backend artifacts (vcpkg features and overlay ports, Conan options and profiles), not re-implemented.
2. **A pairing has to name the backend and the variant**, not just a library and version. "ImGui 1.93.0 + SDL 3.4.18" means something different under vcpkg (SDL_GPU backend compiled for you) than under Conan (you compile it).
3. **Treat ABI-define sets as part of a contract.** Jolt's `JPH_*` defines are currently mirrored by hand in Conan's recipe and set per config in xmake. Automap can state the expected set and generate a `VerifyJoltVersionID()`-style startup assertion in its glue recipes.
4. **Use vcpkg overlays as Automap's variant mechanism**, because vcpkg rejects alternative features by policy. An Automap registry of overlay ports also closes the catalog gaps and lets Automap move faster than upstream vcpkg where needed (sokol, bgfx).
5. **A release-age cooldown is an Automap-level feature.** No C++ manager has one, and Dependabot's vcpkg cooldown only measures registry-commit age. Automap's curation can require both a minimum upstream release age and passing verification before a version enters a "known-good" pairing. That fits the planned freshness and governance branch of the decision tree.
6. **Console tier:** follow the Xbox pattern (private triplet, stub port that finds the SDK through environment variables, no SDK contents in the registry) for any NDA platform. Automap's private registries should hold only metadata and recipes, never SDK material.
7. **Agent surface:** vcpkg's JSON plus JSON Schema plus `x-update-baseline` is the easiest target for an agent to generate and validate. A Conan backend should emit `conanfile.txt` plus profiles rather than Python where it can.
8. **Watch CPS.** If CPS reaches 1.0 and vcpkg adopts it, CPS files could carry ABI defines and component requirements in a backend-neutral way that Automap reads rather than duplicates.

## Open questions

1. Does Automap's v1 commit to one backend (recommended: vcpkg), or promise several from day one? This is a product-scope call for the product owner.
2. Should Automap run its own vcpkg registry (overlay ports for missing or stale libraries and variants), with the maintenance burden that implies, or upstream fixes to microsoft/vcpkg?
3. Are CMake imported-target names consistent enough across vcpkg, Conan's CMakeConfigDeps and xrepo for backend-neutral glue (e.g. `SDL3::SDL3`, `imgui::imgui`, `Jolt::Jolt`)? Needs a spike; not verified here.
4. Test the per-port triplet route to Jolt double precision, then decide whether overlay ports or triplet fragments are the canonical variant mechanism.
5. How do studios with PlayStation and Switch actually integrate vcpkg or Conan? Ask target users; there is no public source.
6. Game-developer adoption data (JetBrains or other 2025–2026 surveys) could not be verified this session. Worth a follow-up search.
7. What default cooldown should Automap's curation use, and should security fixes bypass it, as Dependabot's does?

## Sources

All accessed 2026-10-29.
1. vcpkg registry releases (2026.10.16 published 2026-10-20; 2026.07.29; 2026.06.24). https://github.com/microsoft/vcpkg/releases
2. vcpkg-tool releases (2026-09-26, 2026-07-27, 2026-07-24 pre-release, 2026-07-13, 2026-05-27). https://github.com/microsoft/vcpkg-tool/releases
3. vcpkg docs, "Versioning reference" (ms.date 2025-01-16). https://learn.microsoft.com/vcpkg/users/versioning. Read from https://github.com/microsoft/vcpkg-docs/blob/main/vcpkg/users/versioning.md @ eda2ef2.
4. vcpkg docs, "vcpkg.json reference" (ms.date 2026-09-16): features, supports, platform expressions. vcpkg-docs `vcpkg/reference/vcpkg-json.md`.
5. vcpkg docs, "vcpkg-configuration.json reference". vcpkg-docs `vcpkg/reference/vcpkg-configuration-json.md`.
6. vcpkg docs, "Maintainer guide" ("Do not use features to implement alternatives"; file conflicts). vcpkg-docs `vcpkg/contributing/maintainer-guide.md`.
7. vcpkg docs, "Triplet variables" (ms.date 2024-02-14; per-port customization). vcpkg-docs `vcpkg/users/triplets.md`.
8. vcpkg docs, "Xbox" (ms.date 2024-01-10). vcpkg-docs `vcpkg/users/platforms/xbox.md`.
9. vcpkg docs, "Registries concepts" (ms.date 2024-11-10). vcpkg-docs `vcpkg/concepts/registries.md`.
10. vcpkg docs, "Offline Usage of vcpkg" (ms.date 2025-03-12). vcpkg-docs `vcpkg/concepts/offline.md`.
11. vcpkg docs, "Binary caching configuration". vcpkg-docs `vcpkg/reference/binarycaching.md`.
12. vcpkg docs, "Asset caching configuration". vcpkg-docs `vcpkg/users/assetcaching.md`.
13. vcpkg docs, "Software Bill of Materials". vcpkg-docs `vcpkg/reference/software-bill-of-materials.md`.
14. vcpkg docs, "Dependabot support for vcpkg" (ms.date 2025-08-12). vcpkg-docs `vcpkg/concepts/dependabot.md`.
15. vcpkg docs, "CMake integration". vcpkg-docs `vcpkg/users/buildsystems/cmake-integration.md`.
16. microsoft/vcpkg @ d4c0a46 (2026-10-28): `ports/*/vcpkg.json` and `portfile.cmake`, `triplets/`, `scripts/azure-pipelines/azure-pipelines.yml`. https://github.com/microsoft/vcpkg
17. vcpkg JSON Schemas. https://raw.githubusercontent.com/microsoft/vcpkg-tool/main/docs/vcpkg.schema.json and https://raw.githubusercontent.com/microsoft/vcpkg-tool/main/docs/vcpkg-configuration.schema.json
18. vcpkg docs, "vcpkg install" (`--allow-unsupported`). vcpkg-docs `vcpkg/commands/install.md`.
19. Conan changelog (2.34.0, 2026-10-28; 2.33, 2.32, 2.30, 2.26, 2.25, 2.14 entries). https://docs.conan.io/2/changelog.html. Read from https://github.com/conan-io/docs/blob/develop2/changelog.rst @ 25c21b6.
20. Conan docs, "Version ranges" and "Lockfiles" tutorials. conan-io/docs `tutorial/versioning/version_ranges.rst` and `lockfiles.rst`.
21. Conan docs, `requirements()` traits (incl. `options`, `consistent`). conan-io/docs `reference/conanfile/methods/requirements.rst`.
22. Conan docs, `validate()`. conan-io/docs `reference/conanfile/methods/validate.rst`.
23. Conan docs, binary model (`binary_model/dependencies.rst`), `package_id_abi_options` (`conanfile/attributes/binary_model.inc`), binary compatibility plugin (`extensions/binary_compatibility.rst`), `compatibility()` method.
24. Conan docs, `settings_user.yml` example (`examples/config_files/settings/settings_user.rst`), and the default settings.yml. https://github.com/conan-io/conan/blob/develop2/conan/internal/default_settings.py
25. Conan docs, "Local Recipes Index" (experimental). conan-io/docs `devops/devops_local_recipes_index.rst`.
26. Conan docs, "Backing up third-party sources" (`devops/backup_sources/sources_backup.rst`) and "Save and restore" (`devops/save_restore.rst`).
27. Conan docs, "Software Bills of Materials" (experimental; CycloneDX 1.4/1.6). conan-io/docs `security/sboms.rst`.
28. Conan docs, "Package signing" plugin (experimental). conan-io/docs `reference/extensions/package_signing.rst`.
29. Conan docs, "Using ConanCenter", the frozen center.conan.io note (default remote center2 since 2.9.2), and "Creating and hosting your own ConanCenter binaries". conan-io/docs `devops/using_conancenter.rst`, `common/conancenter_frozen.inc`, `devops/conancenter/hosting_binaries.rst`.
30. Conan docs, cross-build examples (Emscripten, Android NDK). conan-io/docs `examples/cross_build/`.
31. cmake-conan develop2 README ("CMake dependency provider for the Conan C and C++ package manager"). https://github.com/conan-io/cmake-conan/tree/develop2
32. conan-io/conan-center-index @ 7fd5753 (2026-10-29): `recipes/*/config.yml`, `recipes/imgui/all/conanfile.py`, `recipes/joltphysics/5.x/conanfile.py`. https://github.com/conan-io/conan-center-index
33. xmake releases (v3.1.2 2026-10-16; v3.1.1 2026-08-27; v3.1.0 2026-08-08). https://github.com/xmake-io/xmake/releases
34. xmake docs, "Using official packages" (versions, configs, sha256, lock, repositories). https://github.com/xmake-io/xmake-docs/blob/master/docs/guide/package-management/using-official-packages.md @ 9840261.
35. xmake docs, "Using third-party packages" (`vcpkg::`, `conan::`). xmake-docs `docs/guide/package-management/using-third-party-packages.md`.
36. xmake docs, "Using packages in CMake" (xrepo.cmake), "Network optimization", "Distribute private libraries". xmake-docs `docs/guide/package-management/`.
37. xmake docs, platform list (`docs/api/description/conditions.md`) and the v3.0.7 release post (install schemes). xmake-docs.
38. xmake-io/xmake-repo @ cf09fff (2026-10-29): `packages/*/*/xmake.lua`, `.github/workflows/`. https://github.com/xmake-io/xmake-repo
39. CPM.cmake README and releases (v0.43.2 2026-09-24; v0.43.0 2026-07-06). https://github.com/cpm-cmake/CPM.cmake
40. CMake release notes 4.3 (v4.3.0 tagged 2026-03-17), 4.4 (v4.4.0 2026-07-09; v4.4.5 2026-10-27), 4.5 (v4.5.0-rc2 2026-10-28). https://github.com/Kitware/CMake/tree/master/Help/release @ 9325f20; https://cmake.org/cmake/help/latest/release/
41. CMake docs, `cmake_language(SET_DEPENDENCY_PROVIDER)` (3.24) and the FetchContent module. Kitware/CMake `Help/command/cmake_language.rst`, `Modules/FetchContent.cmake`.
42. CMake experimental features guide (CPS from `install(EXPORT)`). Kitware/CMake `Help/dev/experimental.rst`.
43. CMake `cmake-cxxmodules(7)`. Kitware/CMake `Help/manual/cmake-cxxmodules.7.rst`.
44. Common Package Specification repository (version_info 0.15.0; last commit 2026-10-23). https://github.com/cps-org/cps
45. npm CLI changelog (`min-release-age` added in 11.10.0, 2026-02-11; latest 12.2.1, 2026-10-14). https://github.com/npm/cli/blob/latest/CHANGELOG.md
46. pnpm v11.0.0 release notes (2026-04-28; `minimumReleaseAge` default 1440). https://github.com/pnpm/pnpm/releases/tag/v11.0.0
47. uv relative `exclude-newer` (secondary sources). https://pydevtools.com/handbook/how-to/how-to-protect-against-python-supply-chain-attacks-with-uv/ ; https://newreleases.io/project/pypi/uv/release/0.9.17
48. Renovate `minimumReleaseAge` and `minimumReleaseAgeBehaviour` (default `timestamp-required`), the manager list, and the Conan datasource. https://github.com/renovatebot/renovate (`docs/usage/configuration-options.md`, `lib/config/options/index.ts`, `lib/modules/manager/`, `lib/modules/datasource/conan/index.ts`) @ bdf434d.
49. GitHub Docs, "Dependabot options reference": `cooldown`, default 3 days, vcpkg row. https://docs.github.com/en/code-security/reference/supply-chain-security/dependabot-options-reference. Read from github/docs `content/code-security/reference/supply-chain-security/dependabot-options-reference.md`.
50. Secondary comparison of pnpm 11, Yarn 4.10 and npm 11.10 release-age gates. https://lilting.ch/en/articles/pnpm-11-minimum-release-age-default-shai-hulud-defense
51. Bazel Central Registry (`modules/`). https://github.com/bazelbuild/bazel-central-registry
52. Meson WrapDB (`subprojects/`). https://github.com/mesonbuild/wrapdb
53. nixpkgs `pkgs/by-name/`. https://github.com/NixOS/nixpkgs
54. Upstream tags read with `git ls-remote` and tag-commit dates: libsdl-org/SDL, ocornut/imgui, jrouwe/JoltPhysics, erincatto/box2d, SanderMertens/flecs, skypjack/entt, spnda/fastgltf, jkuhlmann/cgltf, ufbx/ufbx, zeux/meshoptimizer, KhronosGroup/KTX-Software, BinomialLLC/basis_universal, guillaumeblanc/ozz-animation, recastnavigation/recastnavigation, mackron/miniaudio, wolfpld/tracy, floooh/sokol, bkaradzic/bgfx and bgfx.cmake, shader-slang/slang, ThePhD/sol2, ValveSoftware/GameNetworkingSockets, Chlumsky/msdfgen (all on github.com).
55. JoltPhysics v5.6.0: `Jolt/Jolt.cmake` (PUBLIC defines; "a mismatch causes link errors") and `Jolt/RegisterTypes.h` (`VerifyJoltVersionID`). https://github.com/jrouwe/JoltPhysics/tree/v5.6.0
56. GitHub repository pages (star counts): microsoft/vcpkg, conan-io/conan, xmake-io/xmake, cpm-cmake/CPM.cmake.
57. SDL_shadercross repository (no releases or tags). https://github.com/libsdl-org/SDL_shadercross
58. Dear ImGui v1.93.0 tree (no root `CMakeLists.txt`). https://github.com/ocornut/imgui/tree/v1.93.0
