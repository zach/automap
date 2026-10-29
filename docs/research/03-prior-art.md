# Prior art in engine composition models

**Scope.** This note looks at how existing engines and frameworks let people build a game engine out of separately developed parts. It covers Bevy, Godot, sokol/fips/fibs, Magnum/Corrade, Cute Framework, raylib, Flecs, O3DE, Unreal, Our Machinery, Defold, Mach, zig-gamedev, Odin and C3 vendor collections, and starter templates. For each one it records:

- the unit of composition;
- how dependencies and compatibility are declared;
- how the parts are wired together at runtime;
- how the parts are distributed;
- its status in 2026;
- the evidence for why it thrived or stalled.

The aim is to test Automap's starting hypothesis, which is not assumed: that a curated, machine-readable map of slots, libraries, integration contracts and known-good pairings is what C++ engine composition is missing.

**Accessed:** 2026-10-29.

**Method and limits.** Most project websites could not be reached from this session: bevy.org, godotengine.org, docs.o3de.org, dev.epicgames.com, defold.com, machengine.org and odin-lang.org all failed DNS. Primary sources were therefore read from the projects' own source repositories through `raw.githubusercontent.com`. These include website-source repositories (bevy-website, godot-website, o3de.org, defold/doc), engine source at release tags, and the crates.io sparse index. Epic's documentation was read only as search-result excerpts. The web-search budget for the session ran out before the planned search for 2025–2026 newcomers could run, so that sub-question is only partly covered (see **Open questions**).

---

## Key findings

1. **[Verified] Every ecosystem that thrived has one owner of the seams.**
   - Bevy's `Plugin` trait defines the lifecycle (`build` → `ready` → `finish` → `cleanup`) [4].
   - Godot's GDExtension defines a C ABI, four initialization levels, and load-time checks of `compatibility_minimum`/`compatibility_maximum` [19][20].
   - Unreal gives every module a host type (`EHostType`) and a loading phase (`ELoadingPhase`) [61][62].
   - Our Machinery had a versioned API registry [71].

   In each case, third parties only declare compatibility against contracts the core owns.

2. **[Verified] Compatibility evidence in the Rust/Bevy ecosystem is declared, not tested.**
   - Bevy's official guidance is a hand-written README table mapping plugin versions to Bevy versions [7].
   - The Bevy Assets site infers a crate's "Bevy version" from its crates.io dependency on `bevy` [9].
   - arewegameyet.rs has no version or compatibility field at all [11].
   - Measured on 2026-10-29, 21 days after Bevy 0.20, the latest release of 6 of 20 hand-picked popular plugins still did not accept 0.20 [10].

3. **[Verified] Bevy breaks compatibility every 3.5–5 months, and the core absorbs its ecosystem.**
   - Releases: 0.16 (2025-04-24), 0.17 (2025-09-30), 0.18 (2026-01-13), 0.19 (2026-06-19), 0.20 (2026-10-08). Each one ships a migration guide [1][2][3].
   - Picking, and now a pan-orbit camera, moved into core [1][5]. The ecosystem crate `bevy_mod_picking` was left at Bevy 0.14 [10].

4. **[Verified] Godot 4.6 shipped LibGodot.** It adds a C entry point, `libgodot_create_godot_instance`, which returns a `GodotInstance` with `start`/`iteration`/`stop`. Only one instance is allowed per process, and the first release supports Linux, Windows and macOS only [14][21]. **Godot 4.7** shipped HDR output and replaced the Asset Library with the Asset Store [17][18]. Both halves of the lead are confirmed.

5. **[Verified] Curated, versioned "known-good maps" exist outside Rust and are closer to what Automap needs than arewegameyet.**
   - Odin's `vendor:` collection records, for each package, the upstream version its bindings and binaries match, the license, the upstream URL, and notes on DLLs it pulls in. The Odin team curates it [83].
   - raylib's `BINDINGS.md` records which raylib version each binding targets. Bindings more than two versions behind are moved to an "older or unmaintained" section; 75 of 171 rows target raylib ≤4.x [47].
   - rlImGui keeps one branch per (raylib, Dear ImGui) version pair [49].
   - vcpkg's curated registry has baselines and a CI build-state file across tested triplets, which is evidence that ports build together [36][37].

6. **[Verified] Version-pinned registries go stale for C libraries that are released by commit rather than by version.**
   - vcpkg's `sokol` port is pinned to a 2023-10-07 commit [35]. Since then, upstream sokol logged 111 changelog entries in 2025–26, 17 of them flagged as breaking [29].
   - Magnum's last tagged release is 2020.06, followed by roughly 2,300 lines of changelog with no release, and vcpkg still ships 2020.06 [38][35].

7. **[Verified] Compatibility metadata written by hand decays unless something checks it.** O3DE's `gem.json` supports PEP 440 version ranges for engines, gems and individual engine APIs [55]. In the canonical `o3de-extras` gem repository:
   - none of the 20 gems sets `engine_api_dependencies`;
   - several leave `compatible_engines` empty;
   - one declares `o3de>=4.1.0`, although O3DE's engine version is 2.x [58][54].

   The engine's own API versions have not changed between 24.09 and 26.05 [54].

8. **[Verified] There are working precedents for the console/NDA tier.**
   - Defold's open-source build server (Extender) has `consoles`, `nintendo` and `playstation` profiles [77].
   - Defold releases its Switch-enabled engine source only after Nintendo confirms the developer is approved [78].
   - Fab requires source code for code plugins and a separate upload per engine version (secondary sources [67]).
   - vcpkg git and filesystem registries get the same versioning and caching features as the curated registry [36].

9. **[Verified] Upstream attitudes to coding agents differ sharply.**
   - SDL's `AGENTS.md` bans generative-AI contributions [88].
   - sokol forbids LLM generation in its headers and documentation [30].
   - Cute Framework and Defold ship agent guides [45][79].

   All of these files are aimed at people contributing to the library. None tells a consumer how to compose the library into an engine.

10. **[Inference] In C++, no core defines the seams, so Automap would have to act as the core.** Automap would own the contract vocabulary and the verification that pairings work. The unit of compatibility evidence should be a tested *composition*, not a library's self-declaration.

---

## 1. Rust / Bevy

**Composition model.**

- **Unit:** the unit is a `Plugin`, meaning any `fn(&mut App)` or a type that implements the trait.
- **Lifecycle:** "the app calls `Plugin::build` immediately … once the app started, it will wait for all registered `Plugin::ready` to return `true` … then call all registered `Plugin::finish` … and call all registered `Plugin::cleanup`" [4]. `ready`/`finish` exist for setup that has to wait on asynchronous work, for example "the initialization of a renderer" [4].
- **Uniqueness:** a plugin can be added only once unless it overrides `is_unique()`. Duplicates are detected by `name()`, which defaults to the type name [4].
- **Plugin groups:** `DefaultPlugins` is an ordered `plugin_group!`. Its ordering constraints live in source comments, not in data. Examples: "Load this before AssetPlugin to properly register http asset sources", "WinitPlugin needs to be after AssetPlugin because of custom cursors", and `ImagePlugin` loads "after renderer initialization so that it knows about the supported compressed texture formats" [5].
- **Feature gating:** each plugin in the group is gated by a cargo feature [5]. Bevy 0.20 has 177 top-level cargo features [6]. Since 0.18 it also has tiered `PROFILE` features (`2d`, `3d`, `ui`) built from `COLLECTION` features (`default_app`, `default_platform`, `audio`, `scene`, …) [6].

**Dependency and compatibility declaration.** Cargo semver requirements on `bevy` or on individual `bevy_*` sub-crates are the only machine-readable declaration. Popular plugins often depend on `bevy_app` rather than `bevy` [10], so checking compatibility means looking at the whole `bevy_*` family. Beyond that, the official "Building Bevy's Ecosystem" guide:

- recommends a hand-written README table (`| bevy | bevy_awesome_plugin |`);
- recommends a badge saying whether the plugin tracks Bevy `main` or released versions;
- asks plugin authors to use `default-features = false`, because "Bevy features enabled in your plugin cannot be disabled by someone using your plugin" [7].

Feature unification is therefore a build-flag contract that the guide documents in prose.

**Distribution and maps.**

- Distribution is through crates.io.
- **Bevy Assets** is a curated list. Each entry is a TOML file with name, description, link and an optional crate [8]. The site generator fills in `bevy_versions` from the crates.io database dump or from the GitHub/GitLab manifest. The code comments say version merging is a TODO ("this just returns version1 if it's Some") [9].
- **arewegameyet.rs** is a Zola site of per-category `data.toml` files. Each item has `name`, `description`, `categories`, and optional `crate_url`, `repository_url` and `homepage_url`, or it is pulled from crates.io or GitHub by `source` [11]. It has no field for engine-version compatibility or for known-good pairings [11].

**Status in 2026.**

- Bevy 0.20 was released on 2026-10-08, with 227 contributors and 817 PRs [1]. Its highlights are BSN scene syntax, adoption of WESL shaders, and Solari on Metal [1].
- The main branch is `0.20.0-dev` and crates.io lists 0.20.1 [10].
- The official template `bevy_new_2d` still pins `bevy = "0.19"` [12].
- `bevy_cli` is an unofficial alpha (`cli-v0.1.0-alpha.2`) and is not published to crates.io [13].

**Ecosystem lag, measured.** The crates.io index was checked on 2026-10-29 for 20 popular plugins. The sample is a convenience sample, not a random one.

| Group | Count | Crates |
|---|---|---|
| Latest release requires Bevy 0.20 | 14 | `bevy_egui`, `avian3d`, `bevy_rapier3d`, `lightyear`, `leafwing-input-manager`, … |
| Latest release requires Bevy 0.19 | 4 | `bevy_hanabi`, `bevy_tweening`, `bevy_seedling`, `bevy_framepace` |
| Older | 2 | `iyes_perf_ui` (0.16); `bevy_mod_picking` (0.14), whose functionality now ships as `DefaultPickingPlugins` in core [5] |

Source: [10].

**Why it works / costs.**

- [Verified] Bevy succeeds as a composition ecosystem because it owns *all* the seams: the `App`, the ECS, schedules, asset types and math types. Every plugin targets one versioned axis [4][7].
- [Verified] The cost is that every minor release breaks that axis, and the ecosystem catches up crate by crate [1][10].
- [Inference] Compatibility is effectively "it compiles with `bevy ^0.N`". That is an easy contract to state, because there is one core, and it says nothing about behavioral interactions between plugins, such as ordering constraints like the comments in `DefaultPlugins`.

---

## 2. Godot: GDExtension, LibGodot, Asset Store

**GDExtension composition model.** The unit is a shared library described by a `.gdextension` INI file. In 4.7-stable the loader parses exactly these sections and keys [19]:

- `[configuration]`: `entry_symbol`, `compatibility_minimum` (required, and must be ≥4.1), `compatibility_maximum` (optional), `reloadable`, `autodetect_library_prefix`;
- `[libraries]`: one entry per feature-tag combination;
- `[dependencies]`: native shared objects to copy, keyed by feature tags;
- `[icons]`.

Compatibility is checked when the extension loads. The engine refuses to load the extension and reports "only compatible with Godot version X or later" [19].

There is **no extension-to-extension dependency or ordering** in the format. `[dependencies]` lists native libraries, not other extensions [19].

**Wiring.** Extensions register classes through four initialization levels: `CORE`, `SERVERS`, `SCENE`, `EDITOR`. The manager initializes them level by level, upward, and tears them down in reverse [20].

**Compatibility contract.** "GDExtensions targeting an earlier version of Godot should work in later minor versions, but not vice-versa." Extensions must also match the engine's floating-point precision and use an `extension_api.json` generated by the target build [22].

- godot-cpp is versioned separately from Godot from 10.x on. Its `api_version` option targets Godot 4.3 or later [23].
- The official template encodes build variants in the library keys, for example `macos.single.debug` and `windows.x86_64.double.release`. Its `compatibility_minimum` is `"4.1"`, and its `godot-cpp` submodule still tracks branch `4.3` [24].
- Godot 4.7 added a GDExtensions section to Project Settings. Before that, "there was unfortunately no way to know exactly which project extensions were actually present … except by manually locating the project's `.gdextension` files" [17].

**LibGodot.** It was introduced in the 4.6 dev-2 snapshot (2025-10-20) as "a functional baseline (GH-110863) … basic support for a `GodotInstance` to function as an entry point" [15]. It shipped in 4.6 as "Godot as a library!":

> "embed the engine directly into your own applications … control startup, manage the engine loop … The initial release supports Linux, Windows, and macOS" [14]

The API is `libgodot_create_godot_instance(argc, argv, init_func)`. The `init_func` is a GDExtension initialization function supplied by the host [21]. `GodotInstance` exposes `start`, `iteration`, `stop`, `pause`, `resume`, `focus_in` and `focus_out`, and "Only one may be created per process" [21].

Release timing: 4.6 RC1 was 2026-01-14 and RC2 2026-01-20 [16]. Stable is reported as 2026-01-26 (secondary [27]).

**Asset Store.**

- The Asset Store was announced live on 2026-05-22, "fully integrated in Godot 4.7". It adds reviews and ratings, publisher analytics, multiple versions per asset, changelogs and tags [18]. Its roadmap includes buying and selling assets [18].
- The old Asset Library "doesn't actually host any of the assets it distributes". Assets were not auto-migrated, and the Library will become read-only [18].
- The Foundation built its own store because a hands-off approach led to confusion over which stores it endorsed, and to resold free assets [18].
- In the 4.7 editor, the store "replaces the Asset Library" and adds threaded background work [17]. 4.7 RC3 was 2026-06-15 [25], 4.7.1 was 2026-07-14 [25], and 4.8 beta 1 was 2026-10-15 [25].

**Why it succeeded.**

- [Verified] Godot keeps one owned ABI that stays forward-compatible within 4.x [22].
- [Verified] It has rapid adoption, at about 2 million downloads per release [26].
- [Verified] The foundation decided to own its catalog [18].

[Inference] Godot avoids Bevy's churn by making the seam a forward-compatible C ABI rather than a source-level API. That is the same choice Our Machinery made (§9).

---

## 3. sokol, fips → fibs

**sokol composition model.** sokol is a set of STB-style single-file C headers. "The core headers are standalone and can be used independently from each other" [28].

`sokol_gfx.h` "does *not* handle window creation or 3D API context initialization". Instead, the caller supplies an environment and swapchain, and `sokol_glue.h` forwards them from `sokol_app.h` when both are used [28]. This is the "bring your own everything" pattern: it makes the seams explicit structs rather than hiding them.

The README states the integration contracts in plain terms:

- define `SOKOL_IMPL` in exactly one translation unit;
- select exactly one backend define, which must match across `sokol_gfx.h` and `sokol_app.h` and match the shaders;
- shader headers must come "from a `sokol-shdc` version compatible with the Sokol headers";
- compile the implementation as Objective-C on Apple platforms [28].

The `util/` headers are **first-party glue recipes** for other libraries: `sokol_imgui.h` for Dear ImGui, `sokol_nuklear.h`, `sokol_fontstash.h`, and `sokol_spine.h` for the Spine C runtime [28]. Bindings for Zig, Odin, Nim, Rust, D and C3 are regenerated automatically [28].

**Versioning and status.**

- sokol is very active: there are changelog entries through 27-Oct-2026, and 26-Oct-2026 notes "some minor breaking API cleanup in sokol_gfx.h" [28][29].
- In 2025–26 the changelog has 111 dated entries, and 17 of them mention breaking changes [29].
- vcpkg packages sokol by date (`2023-10-07`, a pinned commit) [35]. [Inference] Users of the package manager are therefore about three years behind.
- `AGENTS.md` forbids LLM generation in headers and documentation, and allows it only in tests, bindgen and shader-generation scripts [30].

**fips.** fips is a Python/CMake meta-build that imported "fipsified" wrapper repositories. Its README labels its list of fipsified projects "OBSOLETE", and its changelog ends at 08-Nov-2023 [31]. It has been replaced by **fibs**, "a Typescript-based cmake-wrapper and task-runner … spiritual successor to fips" [32]:

- dependencies are git imports stored inside the project;
- target platforms and IDE integrations can live in ordinary dependencies [32];
- sokol-samples now builds with `./fibs build` [33] and imports curated wrappers from `fibs-libs`, "Collection of fibs project wrappers for common 3rd-party libraries", which includes sokol, stb, glfw3, box3d and microui [33][34];
- each wrapper declares a **typed option schema**, for example sokol's `backend` enum with values `glcore`/`gles3`/`d3d11`/`metal`/`wgpu`/`vulkan`, and `useEGL` [34].

[Inference] fibs-libs is a small, single-owner, machine-readable recipe registry with typed build flags, which is close to Automap's "build flags" contract.

---

## 4. Magnum and Corrade's plugin manager

**Model.** Corrade's `PluginManager` loads static or dynamic plugins that implement an interface.

- Each plugin interface returns a `pluginInterface()` string, and "Plugins that don't define the exact same interface string won't be loaded". The recommended form includes a version, for example `…AbstractFood/1.0` [39].
- A global `CORRADE_PLUGIN_VERSION` (currently 7) guards the ABI [39].
- Plugin metadata (`.conf`) declares `depends=` (plugins that "must be loaded before this plugin") and `provides=` (aliases). For static plugins the metadata is compiled in [40].
- Magnum's asset importers and converters in `magnum-plugins` are the main users [42].

**Status.**

- The last tagged release is **2020.06** [41]. The master changelog's "Changes since 2020.06" section runs about 2,300 lines [38], and copyright headers extend to 2026 [38], so development continues on master.
- vcpkg ships `magnum`/`corrade` 2020.06 [35].
- [Inference] Rolling development on master with no tagged releases makes package-manager distribution go stale, as with sokol.
- **[Unverified]** Adoption numbers and maintainer headcount. The planned search could not run because the search budget was exhausted.

**Lesson.** The interface-string-plus-version pattern and the `depends`/`provides` metadata are a sound in-process service registry for one C++ codebase [39][40]. They did not create a cross-vendor ecosystem: plugins are Magnum's own.

---

## 5. "Batteries-included" C/C++ frameworks (the opposite pole)

**Cute Framework (CF).**

- CF is a C/C++ framework "for making 2D and 3D games … without baggage, gnarly dependencies" [43]. Its `AGENTS.md` still says "2D" [45].
- CF 1.1.1 [44] pins its whole stack in CMake:
  - SDL3 3.4.0 (a system SDL must be in `3.4.0...<4.0.0`);
  - Box2D `v3.1.1` and Box3D `v0.1.0`, compiled into CF itself;
  - PhysFS at a fixed commit;
  - optional wgpu-native release zips with SHA-256 checks [44].
- A project template provides the user-facing CMake [43].

The unit is the whole framework. The owner picks and tests one known-good composition and ships it as one library.

**raylib.**

- raylib 6.0 was released 2026-04-23 and 6.2 on 2026-10-15. Master reports `6.5-dev` [46].
- An official parser emits `raylib_api.json` from `raylib.h` for binding generators [48].
- `BINDINGS.md` is a curated compatibility table with a stated policy: bindings "should be reasonably up to date with the last release, out of date bindings may be removed", and bindings "more than 2 versions old" are listed separately. 75 of 171 rows target raylib ≤4.x [47].
- Ecosystem glue in raylib-extras is pinned by pair. rlImGui says: "Tags and releases are made for major releases of raylib or significant changes to Dear ImGui", with branches `Raylib_6_0` (ImGui 1.92.7) and `Raylib_5_5` [49]. `game-premake` and `raylib-quickstart` are starter templates [50].

**Flecs modules and flecs-hub** (an addition to the lead list).

- A Flecs module is an import function that defines namespaced components and systems, loaded with `ECS_IMPORT(world, Module)` [51].
- flecs-hub packages declare `use` dependencies in bake's `project.json`. They include *shared component packages* (`flecs.components.transform`, `.graphics`, `.input`) and per-OS link flags. For example, `flecs.systems.sokol` uses the component packages plus `flecs.game` and links GL/X11 on Linux [52].

[Inference] This is the closest C/C++ example of "a dominant core defines the seams": the Flecs world and its shared component types fix the math, transform and input contracts.

---

## 6. O3DE Gems

**Model.** A Gem is a directory with `gem.json`. The default template includes `dependencies`, `compatible_engines`, `engine_api_dependencies`, `platforms`, `version`, `repo_uri` and `type: Code` [53].

- Version fields use semver. Version ranges follow PEP 440, for example `o3de>=2.0.0`, `o3de-sdk==1.2.0`, `Atom~=2.0.0`. If the fields are empty, the gem is "assumed to be compatible with all engines and have no gem dependencies" [55].
- `engine.json` carries engine-wide `api_versions` (`editor`, `framework`, `launcher`, `tools`) that gems can depend on [54][55][56].
- `versions_data` in a remote `repo.json` records per-version download URIs and changes [56]. Remote repositories are therefore an existing private-registry mechanism.
- Large features are composed out of sub-gems. The `Atom` gem lists RHI, RPI and tools as external subdirectories and dependencies [54].

**Status.**

- 26.05.0 was released on 2026-05-27 [57]. Engine `version` went from 2.3.0 (24.09) to 2.6.0 (26.05), with 2.8.0 on development, while `api_versions` stayed unchanged across all four releases [54].
- The canonical `o3de-extras` repository lists 20 gems, and 11 of them are robotics, ROS2, warehouse or simulation gems [58].
- A secondary summary of O3DE's "2025 in review" says the 2026 focus is gaming, robotics and film [59].

**Why it stalled as a game-engine ecosystem.**

- [Verified] Expressive compatibility metadata is not maintained in practice: `engine_api_dependencies` is unused in the canonical repo, and one entry claims `o3de>=4.1.0` against a 2.x engine [58].
- [Verified] The catalog that does exist leans toward robotics [58].
- **[Unverified]** Contributor, funding and shipped-game trends. The search budget ran out before this could be checked.
- [Inference] A rich compatibility *schema* without a large third-party author base, and without automated checks that the declarations are true, gives little evidence.

---

## 7. Unreal: plugins, modules, loading phases, Fab

**Model.**

- **Plugin:** a plugin is found through its `.uplugin` descriptor [60]. Plugins declare dependencies "by enabling other Plugins within their own .uplugin files". Plugins and modules sit at hierarchical levels and "can only depend on other Plugins or Modules at the same level or higher" [60].
- **Module:** a module has a `Build.cs` (a `ModuleRules` subclass) that declares module dependencies, plus a module class with startup and shutdown [63]. In a real plugin [64]:
  - `PublicDependencyModuleNames` and `PrivateDependencyModuleNames` split dependencies into those exposed to dependents and those used privately;
  - `IMPLEMENT_MODULE(FSocketIOClientModule, SocketIOClient)` registers `StartupModule()`/`ShutdownModule()`;
  - the descriptor sets `"EngineVersion": "5.8"`, gives each module `"Type": "Runtime"|"Editor"`, sets `"LoadingPhase": "PreDefault"` for runtime modules and `"Default"` for editor modules, and lists per-module platforms [64].
- **Loading phases** (`ELoadingPhase`, earliest to latest), a mature init-order vocabulary [61]:
  - `EarliestPossible`: for example, compression plugins needed to read paks;
  - `PostConfigInit`: after config, before the engine is fully up;
  - `PostSplashScreen`;
  - `PreEarlyLoadingScreen`: before CoreUObject;
  - `PreLoadingScreen`;
  - `PreDefault`;
  - `Default`: after game modules;
  - `PostDefault`;
  - `PostEngineInit`;
  - `None`: never auto-load.
- **Host types** (`EHostType`) are an orthogonal filter on where a module loads: `Runtime`, `RuntimeNoCommandlet`, `RuntimeAndProgram`, `CookedOnly`, `UncookedOnly`, `DeveloperTool`, `Editor`, `EditorNoCommandlet`, `EditorAndProgram`, `Program`, `ServerOnly`, `ClientOnly`, `ClientOnlyNoCommandlet`. `Developer` is deprecated "due to ambiguities" [62].
- **Failure mode:** phase mismatches between related plugins are a known problem. Users asked why the Oculus plugin used `PostConfigInit` while related plugins used `Default` [65].

**Distribution.**

- Fab: Epic's structure requirements require a `.uplugin` and a `Source/` directory for code plugins [66].
- Community tooling reports that Fab requires source code, an `EngineVersion` per upload, and a new upload for each engine version (secondary [67]). One seller's post-mortem found their packaging had only edited `EngineVersion` without compiling for each version [67].
- **[Unverified]** Fab's launch date (October 2024, from prior knowledge, not re-checked).

**Status.**

- UE 5.7 was released 2025-11-12 [68].
- UE 5.8 was released in June 2026 and described as the last 5.x release; it reportedly includes an MCP plugin (secondary [68]). A third-party plugin already targets `"EngineVersion": "5.8"` [64].
- UE6 is targeting early access by the end of 2027 (secondary [68]).

**Lesson.**

- [Verified] Unreal's phase × host-type matrix is the most complete init-order and lifecycle model surveyed [61][62].
- [Inference] It works because one owner defines the phases and every module must sit in one. Compatibility is still per exact engine version, and binary plugins are rebuilt for every release [67].

---

## 8. Our Machinery (The Machinery)

**Model.** The engine was "a collection of little things that can be configured and assembled". APIs are C structs of function pointers, published to and looked up from a central `api_registry` by name [69].

- C was chosen for the API layer because "C has a standard ABI (in practice) and C++ doesn't", while implementations could be C++ [69].
- A plugin is a DLL exporting `load_xxx(reg, reload)` and `unload_xxx(reg, reload)`. During hot-reload, the new DLL loads before the old one unloads, so state can be transferred [70].
- Since 2021.9, APIs carry semver: `get(name, version)` and `set(name, version, api, size)`. Major versions must match exactly [71].
- `get` and `set` can resolve **in any order**. The registry returns a placeholder that is filled when the API is `set`, and checks all `get` calls only after every plugin has loaded. Plugins with unmet dependencies are reported and their registrations rolled back [71].
- *The Truth* was the shared data model. Features such as backward compatibility of files were implemented once in the data model rather than in each system [72].

**Why it ended.**

- [Verified] In August 2022 users were told the license would end in 14 days under a newly invoked EULA section. They were asked to delete source and binaries and were refunded. The only explanation was "it's no longer possible for us to continue in the current direction" [73][74].
- **[Unverified]** The underlying cause. No founder account was found.
- [Inference] The design was technically strong. Its failure was commercial and organizational: a closed composition platform with one vendor and no escrow can disappear entirely.

**Lessons for Automap.**

- A versioned API registry with dependency checks deferred until everything has loaded is a good runtime pattern for the "lifecycle" and "handles" contracts [71].
- A composition map whose existence depends on one closed vendor is a single point of failure [73].

---

## 9. Defold native extensions

**Model.**

- An extension is a project folder with `ext.manifest`, `src/`, `include/`, `lib/<platform>`, `manifests/<platform>` (Gradle, CocoaPods, manifest stubs) and `res/` [75].
- Entry points are declared with `DM_DECLARE_EXTENSION(symbol, name, app_init, app_final, init, update, on_event, final)`. The documented order is [75]:
  - start-up: engine systems start → `app_init` → `init` (all Defold APIs ready) → script `init`;
  - per frame: `update` → script `update`;
  - shutdown: script `final` → `final` → `app_final`.
- Build configuration is a **three-level manifest merge**: `merge(game.appmanifest, ext.manifest, build.yml)`. The engine's `build.yml` pins SDKs, compilers and default flags, the app manifest can strip engine features, and the extension manifest adds flags, defines, libraries and frameworks per platform [76].

**Distribution.**

- A free cloud build server compiles a project-specific engine, so "any team member that builds and runs the project will get a project specific engine executable with all native extensions baked in" [75].
- The server (Extender) can be run locally with Docker. Its profiles include `consoles`, `nintendo` and `playstation` [77].
- Switch support is gated: once Nintendo confirms the developer, Defold grants source access to the Switch extension and engine, plus a CLI that can bundle for the console [78].

**Status.** The dev branch is at version 1.14.2. The repository ships `AGENTS.md` [79].

**Lesson.** [Verified] Defold shows the private, NDA-gated tier working in production. Console code lives in access-controlled repositories and on private build servers, while the public extension format is the same for everyone [77][78].

---

## 10. Zig gamedev, Mach, Odin `vendor:`, C3 vendor

**Mach.**

- Mach is at version 0.5.0. It pins a **nominated toolchain**, `mach_zig_version = "2026.4.10-mach"`, and fetches every dependency from its own mirror `pkg.hexops.org` with content hashes [80].
- Development moved to `code.hexops.org`, and GitHub is "purely for visibility" [80].
- Modules declare `mach_module` and `mach_systems`. Init order is an explicit, compile-time-checked `mach.schedule(.{ .{App,.init}, .{Physics,.init}, … })` [81].
- Object IDs encode a generation to catch many use-after-free errors. This is a concrete example of a handle contract [81].

**zig-gamedev.** The libraries (`zaudio`, `zbullet`, `zflecs`, `zglfw`, `zgpu`, …) live in separate repositories. The development repository pins each one by commit tarball and hash in `build.zig.zon` and sets `minimum_zig_version = "0.15.2"` [82]. [Inference] The samples repository works as the verified composition, which amounts to a lockfile for the whole set.

**Odin `vendor:`.**

- The collection ships with the compiler. Its packages and binaries are "curated and maintained by the Odin team. Please don't open pull requests for `vendor:` without first consulting them" [83].
- Each of roughly 50 packages records:
  - the upstream version the bindings and binaries match, for example SDL3 3.4.18, Box2D 3.2.0, raylib 6.0, and wgpu 29.0.1.1;
  - the license;
  - the upstream URL;
  - footnotes on DLLs the package pulls in and their licenses [83].

[Inference] **Yes, `vendor:` is a known-good-map analogue,** and in one respect a stronger one than arewegameyet: it records exact versions, and one owner vouches that the bindings match. It covers library-to-language compatibility only, not library-to-library integration contracts such as allocators, threading or math types.

**C3 vendor.** It is a smaller list with version pins, for example Raylib 6, Raylib 5.5, and Raygui pinned to a commit. `c3c vendor-fetch <lib>` downloads the library and adds it to `project.json` [84].

---

## 11. Starter templates and "engine from libraries" guides

| Template | Pinning strategy | Staleness signal |
|---|---|---|
| SDL_gpu_examples | Builds against a fresh SDL `main` clone; HLSL compiled with SDL_shadercross [85] | No pin, so the build breaks whenever SDL changes |
| vkguide (`all-chapters-2`) | Copies vk-bootstrap, VMA, glm, Dear ImGui, fastgltf, fmt and SDL into `third_party/` [86] | Still uses `imgui_impl_sdl2` in 2026 [86] |
| Ravbug sdl3-sample | Git submodules for SDL, SDL_ttf, SDL_mixer, SDL_image; publishes a tested platform × generator matrix [87] | The matrix is a manual claim |
| godot-cpp-template | Submodule on branch `4.3`, `compatibility_minimum = "4.1"` [24] | Forward-compatible ABI tolerates the lag [22] |
| bevy_new_2d | `bevy = "0.19"` [12] | One release behind three weeks after 0.20 |
| raylib-extras game-premake / quickstart; rlImGui | Premake; per-version branches [49][50] | Explicit pairs |
| Cute Framework template | Copy CMakeLists; CF pins everything [43][44] | Owner-maintained |

[Inference] Templates are snapshots. Each fixes one composition at one moment, with no signal when it goes stale beyond its own commit date. The best of them publish a tested matrix ([87]) or per-pair branches ([49]). None publishes machine-readable compatibility.

**2025–2026 newcomers (partial).** Found incidentally during this research:

- **fibs** (sokol's new build system) [32];
- **LibGodot** [14];
- **Box3D**, Erin Catto's 3D physics library, already pinned by CF (`v0.1.0`) and Odin (`master snapshot`) [44][83];
- **Bevy CLI** (alpha) [13];
- Mach's move to self-hosted infrastructure [80].

**[Unverified]** A systematic search for new "compose an engine from libraries" products could not run because the search budget was exhausted.

---

## 12. Comparison

| System | Unit | Dependency declaration | Compatibility evidence | Wiring / lifecycle | Registry / distribution | Private-extension story | Outcome (2026) |
|---|---|---|---|---|---|---|---|
| Bevy | `Plugin` / `PluginGroup` + cargo features | Cargo semver on `bevy`/`bevy_*` [10] | README tables by hand; Bevy Assets infers version from the dependency [7][9] | `build`→`ready`→`finish`→`cleanup`; group order in comments [4][5] | crates.io; Bevy Assets; arewegameyet (no compatibility data) [8][11] | Git dependencies tracking a rev [7] | Thriving; breaking release every 3.5–5 months; 6 of 20 sampled plugins lagging at +21 days [1][10] |
| Godot GDExtension | Shared library + `.gdextension` | None between extensions [19] | Load-time min/max check; forward-compatible ABI; float-precision match [19][22] | 4 init levels, torn down in reverse [20] | Asset Store, which replaced the Library in 4.7 [17][18] | Any folder of files; LibGodot for embedding [21] | Thriving; one owned ABI |
| sokol | Header | None (bring your own) | Prose rules: one implementation TU, matching backend define, matching shdc version [28] | Caller-driven `*_setup`/`*_shutdown` | GitHub; vcpkg port pinned to 2023 [35] | Copy the headers | Active, single owner; package managers stale |
| fips / fibs | Wrapped git import | `imports` with typed option schema (fibs) [33][34] | Owner's samples CI | n/a (build) | fips-libs (obsolete) → fibs-libs [31][34] | Any git URL | fips superseded by fibs |
| Magnum / Corrade | Plugin (static/dynamic) | `depends=`/`provides=` metadata [40] | Versioned interface string + `CORRADE_PLUGIN_VERSION` [39] | Manager load/instance/unload | GitHub; vcpkg at 2020.06 [35] | Custom plugin directories | No tagged release since 2020.06 [38] |
| Cute Framework / raylib | Whole framework | Owner pins its dependencies [44] | Owner tests; raylib bindings table [47] | Framework-owned main loop | GitHub / templates | Fork | Healthy, narrow scope |
| Flecs hub | ECS module / package | bake `use` [52] | Owner CI (assumed) | `ECS_IMPORT` [51] | GitHub org | Any package | **[Unverified]** activity level |
| O3DE | Gem | PEP 440 ranges for gems, engines and engine APIs [55] | Hand-authored, mostly empty or wrong [58] | Gem modules (not examined) | Remote `repo.json` repositories [56][58] | Private remote repositories [56] | Releasing twice a year; ecosystem leans to robotics [57][58] |
| Unreal | Plugin → modules | `.uplugin` plugins; `Build.cs` module dependencies [60][63] | Exact `EngineVersion`; rebuilt per version [64][67] | `ELoadingPhase` × `EHostType`; startup/shutdown [61][62] | Fab [66] | Project/engine `Plugins/` folders [60] | Dominant; per-version rebuild cost |
| Our Machinery | DLL plugin + C API structs | Versioned `get`/`set` [71] | Registry checks after load; ABI rules [71] | Load/unload, hot-reload, deferred resolution [70][71] | Closed SDK | n/a | Ended 2022 [73] |
| Defold | Native extension folder | Gradle/CocoaPods per platform [75] | Engine `build.yml` pins toolchains [76] | `app_init`/`init`/`update`/`on_event`/`final`/`app_final` [75] | Cloud build server; asset portal [75] | Self-hosted Extender; console profiles [77][78] | Stable, niche |
| Mach / zig-gamedev | Module / library package | `build.zig.zon` hash pins [80][82] | Nominated toolchain; dev-repo samples [80][82] | `mach.schedule` explicit order [81] | Own package mirror [80] | Any URL + hash | Pre-1.0 |
| Odin / C3 vendor | Binding package | Ships with the compiler [83][84] | Upstream version per package, team-maintained [83] | n/a | Compiler distribution | Own collections | Healthy; curated |
| vcpkg (substrate) | Port | Manifest + baseline [36] | CI build state per triplet [37] | n/a | Curated + git/filesystem registries [36] | Private registries [36] | Ubiquitous; stale ports for rolling-release libraries [35] |

---

## 13. Lessons

**What made composition ecosystems thrive.**

1. **A dominant core defines the seams.** Bevy's `App`/ECS/types, Godot's ClassDB and C ABI, Unreal's module system, and Flecs's world plus shared component packages each mean a third-party unit targets exactly one axis [4][19][60][52]. Compatibility becomes one-dimensional ("works with core N") instead of N×N.
2. **The core owns the lifecycle vocabulary.** All the successful models have named phases with a defined order and reverse teardown: Unreal's phases and host types, Godot's four levels, Bevy's two-phase `build`/`finish`, and Defold's app/engine/frame entry points [61][20][4][75]. Without such a vocabulary, ordering constraints end up in comments, as they partly do even in Bevy's `DefaultPlugins` [5].
3. **A forward-compatible ABI reduces churn.** Godot (4.x extensions keep working on later 4.y) and Our Machinery (major-version-exact semver over C structs) let old binaries keep running [22][71]. Bevy's source-level API, broken every release, makes the whole ecosystem re-release every few months [1][10].
4. **Somebody owns the catalog and its staleness policy.** Godot built its own store after a hands-off approach caused confusion [18]. Odin's team curates `vendor:` [83]. raylib demotes bindings more than two versions old [47].
5. **The owner ships glue.** sokol's `util/` headers [28], rlImGui's per-pair branches [49], and Cute Framework's pinned stack [44] are known-good pairings maintained by someone accountable.
6. **Toolchains are taken off the user's hands.** Defold's build server [75] and Mach's nominated Zig version [80] remove the "which compiler and SDK?" failure class.

**What made them stall.**

1. **No owner of the seams.** fips's list of fipsified wrappers became "OBSOLETE" [31]. C++ package managers record versions but no integration contracts, and their snapshots of rolling-release libraries go stale [35].
2. **Breaking churn without automated catch-up.** Bevy's ecosystem lags each release; even the official template trails [10][12].
3. **Metadata nobody verifies.** O3DE's expressive schema is mostly empty or wrong in its own canonical repository [58].
4. **Bus factor and vendor lock.** Magnum has had no tagged release since 2020 [38]. Our Machinery's licence termination left nothing behind [73].
5. **Mismatched lifecycle declarations across related units.** Unreal plugins in the same domain chose different loading phases [65].

**C++: no single core defines the seams.** In C++ the slots (windowing, GPU, physics, audio, ECS, UI) are filled by independent libraries. Each library owns its own versions of the contracts: allocator hooks, threading model, math types, handle and ownership conventions, init order, error style, and build flags. sokol is explicit that it does not own windowing or context creation [28]. The consequences for a map follow, all **[Inference]**:

- Compatibility is pairwise or worse, not "against core N". The evidence has to be about **pairings and compositions**, like rlImGui's per-(raylib, ImGui) branches [49] and sokol's `util/` glue [28], not about libraries in isolation.
- Some party has to define the shared vocabulary that Bevy, Godot or Unreal would otherwise provide. That covers phases, host types, handle and generation conventions, and allocator and thread ownership. Unreal's `ELoadingPhase`/`EHostType`, Godot's four levels, Bevy's `ready`/`finish`, and Our Machinery's deferred resolution are the best templates to borrow from [61][62][20][4][71].
- Declarations alone will not be trusted, given O3DE's evidence [58]. The map's credibility has to come from **compositions built and run in CI**, as vcpkg does for buildability [37] and as zig-gamedev and CF do for their own stacks [82][44].

---

## Leads checked

| Lead | Verdict | Evidence |
|---|---|---|
| "Rust's Bevy plugins plus cargo (mapped by arewegameyet.rs) are the nearest analogue, and C++ has no equivalent." | **Corrected (partly true).** Bevy plus cargo is the nearest analogue for *runtime composition with a single core* [4][5]. arewegameyet is a categorized directory with no version or compatibility fields [11]. Bevy's own compatibility evidence is README tables and versions inferred from dependencies [7][9]. Closer "known-good map" analogues are Odin `vendor:` [83], raylib `BINDINGS.md` [47] and the vcpkg curated registry with baselines and CI [36][37]. C/C++ does have build-level known-good sets (vcpkg, fibs-libs, flecs-hub, CF's pinned stack) [34][37][44][52]. What it lacks is a shared seam layer and a map of integration contracts. | [4][5][7][9][11][34][36][37][44][47][52][83] |
| "Godot 4.6 introduced LibGodot for embedding the engine as a library; 4.7 added HDR output and a new Asset Store." | **Confirmed**, with detail. LibGodot shipped in 4.6 for Linux, Windows and macOS, one instance per process [14][21]. 4.7 added HDR output on Windows, macOS, iOS, visionOS and Linux (Wayland) [17], and editor integration of the Asset Store, which was announced live on 2026-05-22 [17][18]. | [14][17][18][21] |
| Prior-art list: Bevy plugins/cargo, Godot GDExtension and LibGodot, sokol and fips, Magnum, Cute Framework, O3DE Gems, Unreal plugins and modules, starter templates. | **Confirmed but incomplete.** fips is superseded by fibs [31][32]. The list should add Odin and C3 vendor [83][84], Defold (manifest merge, build server, console gating) [75]–[78], the Flecs module ecosystem [51][52], vcpkg registries and baselines [36][37], raylib bindings and pairs [47][49], Mach and zig-gamedev [80]–[82], and Our Machinery [69]–[71]. | as cited |

---

## Implications for Automap (Inference)

1. **Automap has to be the "virtual core".** Since no C++ library owns the seams, Automap's contract vocabulary is the product. It should borrow Unreal's phase × host-type matrix [61][62], Godot's level-ordered init with reverse teardown [20], Bevy's split between build and later finish for asynchronous setup [4], and Our Machinery's versioned, order-independent service resolution [71]. Library-specific conventions such as sokol's single-TU implementation and backend define [28], or Godot's float precision [22], should then map onto that vocabulary.
2. **Make the composition the unit of evidence.** A "known-good pairing" record should cite a CI-built composition: exact versions or commits, toolchain, flags and platforms. It should not rest on library self-declarations, which decay (O3DE [58]) or lag (Bevy [10]). The vcpkg CI baseline [37] and per-pair branches [49] are the models to follow.
3. **Build on existing package substrates; don't replace them.** vcpkg already has curated and private registries, baselines and binary caching [36]. Automap's added value is the contract and pairing layer. It must also handle rolling-release libraries that vcpkg packages by date and lets go stale (sokol [35]), for example by recording commit pins the way CF and zig-gamedev do [44][82].
4. **Glue recipes are first-class artifacts with owners.** sokol's `util/` headers, rlImGui and Flecs shared component packages show glue maintained by someone accountable [28][49][52]. Each Automap recipe needs an owner, a tested pair, and a staleness policy along the lines of raylib's "more than two versions old" rule [47].
5. **Typed build-flag schemas are proven.** fibs-libs declares backend and option enums per library [34], and Defold merges engine, app and extension manifests [76]. Automap's manifest could follow either pattern.
6. **The private and console tier is feasible but needs access control in the infrastructure.** Defold gates console source behind platform-holder verification and runs console builds on self-hosted servers [77][78]. vcpkg private registries reuse the public format [36]. Automap's private registries should keep one schema and only control who can see what.
7. **Agent-facing, consumer-side guidance is unclaimed ground, and some upstreams are hostile.** Existing `AGENTS.md` files target contributors [30][45][79], and SDL bans AI contributions [88]. Per-library skills should be maintained outside the upstream repositories and must not depend on upstream cooperation.
8. **Avoid the Our Machinery failure mode.** The map and recipes should be openly licensed and easy to fork, so the map outlives any single maintainer [73].

---

## Open questions

- How do top Bevy crates test compatibility with Bevy `main` before a release, and could Automap reuse that pattern for C++ pairings?
- Could Automap generate per-library "contract cards" from sources such as raylib's `raylib_api.json` [48] or sokol's bindgen inputs?
- Is a stable C ABI at the seams (Godot, Our Machinery) achievable for Automap-generated glue, or will glue stay source-level C++?
- How does Defold distribute console SDKs into private Extender images, and what does the platform holder's licence permit [77][78]?
- What does the UE 5.8 MCP plugin expose, and does it change how agents author `.uplugin`/`Build.cs` files [68]? This is only covered by secondary sources.
- What are the adoption numbers for O3DE, Magnum, Cute Framework and flecs-hub? Unverified because the search budget ran out.
- Are there 2025–2026 products that compose engines from libraries beyond those found incidentally (§11)? This needs a follow-up search.
- When did Fab launch, and what do its current code-plugin rules say? Only community sources were checked.

---

## Sources

All accessed 2026-10-29. GitHub paths were read through `raw.githubusercontent.com`. The canonical `github.com/<owner>/<repo>/blob/<ref>/<path>` URL is given.

1. Bevy 0.20 release post (dated 2026-10-08). https://github.com/bevyengine/bevy-website/blob/main/content/news/2026-10-08-bevy-0.20/index.md (also https://bevy.org/news/bevy-0-20/)
2. Bevy 0.18 release post (dated 2026-01-13). https://github.com/bevyengine/bevy-website/blob/main/content/news/2026-01-13-bevy-0.18/index.md
3. Bevy news post paths for 0.16 (2025-04-24), 0.17 (2025-09-30) and 0.19 (2026-06-19). https://github.com/bevyengine/bevy-website/tree/main/content/news
4. Bevy `Plugin` trait, v0.20.0. https://github.com/bevyengine/bevy/blob/v0.20.0/crates/bevy_app/src/plugin.rs
5. Bevy `DefaultPlugins`, v0.20.0. https://github.com/bevyengine/bevy/blob/v0.20.0/crates/bevy_internal/src/default_plugins.rs
6. Bevy `Cargo.toml` features, v0.20.0; profiles first appear in v0.18.0. https://github.com/bevyengine/bevy/blob/v0.20.0/Cargo.toml
7. "Building Bevy's Ecosystem" (plugin guidelines). https://github.com/bevyengine/bevy-website/blob/main/content/learn/quick-start/plugin-development.md (https://bevy.org/learn/quick-start/plugin-development/)
8. Bevy Assets README. https://github.com/bevyengine/bevy-assets/blob/main/README.md
9. Bevy Assets generator, metadata extraction. https://github.com/bevyengine/bevy-website/blob/main/generate-assets/src/lib.rs
10. crates.io sparse index entries for `bevy` and 20 plugins (read 2026-10-29). https://index.crates.io/
11. arewegameyet README and schema. https://github.com/rust-gamedev/arewegameyet/blob/master/README.md
12. `bevy_new_2d` Cargo.toml. https://github.com/TheBevyFlock/bevy_new_2d/blob/main/Cargo.toml
13. Bevy CLI README (alpha 2). https://github.com/TheBevyFlock/bevy_cli/blob/main/README.md
14. Godot 4.6 release page and feature data (`libgodot`). https://github.com/godotengine/godot-website/blob/master/_data/release_4_6/features.yml ; https://godotengine.org/releases/4.6/
15. "Dev snapshot: Godot 4.6 dev 2" (2025-10-20). https://github.com/godotengine/godot-website/blob/master/collections/_article/dev-snapshot-godot-4-6-dev-2.md
16. Godot 4.6 RC1 (2026-01-14) and RC2 (2026-01-20) articles. https://github.com/godotengine/godot-website/tree/master/collections/_article
17. Godot 4.7 release page and feature data (`asset_store_update`, `hdr_support`, `gdextensions_view_inside_settings`). https://github.com/godotengine/godot-website/blob/master/_data/release_4_7/features.yml ; https://godotengine.org/releases/4.7/
18. "Introducing the Godot Asset Store" (2026-05-22). https://github.com/godotengine/godot-website/blob/master/collections/_article/introducing-the-godot-asset-store.md
19. Godot GDExtension library loader, 4.7-stable. https://github.com/godotengine/godot/blob/4.7-stable/core/extension/gdextension_library_loader.cpp
20. Godot `gdextension.h` and `gdextension_manager.cpp`, 4.7-stable. https://github.com/godotengine/godot/tree/4.7-stable/core/extension
21. LibGodot: `libgodot.h`, `godot_instance.h`, `doc/classes/GodotInstance.xml`, 4.7-stable. https://github.com/godotengine/godot/blob/4.7-stable/core/extension/libgodot.h
22. Godot docs, "About godot-cpp" (version compatibility). https://github.com/godotengine/godot-docs/blob/master/tutorials/scripting/cpp/about_godot_cpp.rst
23. godot-cpp README (10.x versioning, `api_version`). https://github.com/godotengine/godot-cpp/blob/master/README.md
24. godot-cpp-template (README, `.gitmodules`, `project/bin/example.gdextension`). https://github.com/godotengine/godot-cpp-template
25. Godot 4.7 RC3 (2026-06-15), 4.7.1 (2026-07-14) and 4.8 beta 1 (2026-10-15) articles; master `version.py` (4.8 beta). https://github.com/godotengine/godot-website/tree/master/collections/_article ; https://github.com/godotengine/godot/blob/master/version.py
26. "Godot growth stats 2026" (2026-05-06). https://github.com/godotengine/godot-website/blob/master/collections/_article/godot-growth-stats-2026.md
27. Phoronix, "Godot 4.6 Released" (secondary, gives 2026-01-26). https://phoronix.com/news/Godot-4.6-Released
28. sokol README (master, latest note 26-Oct-2026). https://github.com/floooh/sokol/blob/master/README.md
29. sokol CHANGELOG. https://github.com/floooh/sokol/blob/master/CHANGELOG.md
30. sokol AGENTS.md. https://github.com/floooh/sokol/blob/master/AGENTS.md
31. fips README and CHANGELOG (last entry 08-Nov-2023). https://github.com/floooh/fips
32. fibs README. https://github.com/floooh/fibs/blob/main/README.md
33. sokol-samples README, `fibs.ts`, `fibs-scripts/imports.ts`. https://github.com/floooh/sokol-samples
34. fibs-libs README and `sokol.ts`. https://github.com/floooh/fibs-libs
35. vcpkg ports: `sokol` (2023-10-07, pinned REF), `magnum`/`corrade` (2020.06), `raylib` (6.2), `sdl3` (3.4.18), `joltphysics` (5.6.0), `flecs` (4.1.6). https://github.com/microsoft/vcpkg/tree/master/ports
36. vcpkg docs: versioning (baselines); registries concepts (2024-11-10). https://github.com/microsoft/vcpkg-docs/blob/main/vcpkg/users/versioning.md ; https://github.com/microsoft/vcpkg-docs/blob/main/vcpkg/concepts/registries.md
37. vcpkg `scripts/ci.baseline.txt`. https://github.com/microsoft/vcpkg/blob/master/scripts/ci.baseline.txt
38. Magnum changelog ("Changes since 2020.06"). https://github.com/mosra/magnum/blob/master/doc/changelog.dox
39. Corrade `AbstractPlugin.h` and `AbstractManager.h` (`CORRADE_PLUGIN_VERSION 7`). https://github.com/mosra/corrade/tree/master/src/Corrade/PluginManager
40. Corrade `PluginMetadata` documentation (search excerpt). https://doc.magnum.graphics/corrade/classCorrade_1_1PluginManager_1_1PluginMetadata.html
41. "Magnum 2020.06 released" (search excerpt). https://blog.magnum.graphics/announcements/2020.06/
42. magnum-plugins README. https://github.com/mosra/magnum-plugins/blob/master/README.md
43. Cute Framework README. https://github.com/RandyGaul/cute_framework/blob/master/README.md
44. Cute Framework CMakeLists.txt (1.1.1 and its pins). https://github.com/RandyGaul/cute_framework/blob/master/CMakeLists.txt
45. Cute Framework AGENTS.md. https://github.com/RandyGaul/cute_framework/blob/master/AGENTS.md
46. raylib CHANGELOG (6.2 on 2026-10-15, 6.0 on 2026-04-23) and `src/raylib.h`. https://github.com/raysan5/raylib/blob/master/CHANGELOG
47. raylib BINDINGS.md. https://github.com/raysan5/raylib/blob/master/BINDINGS.md
48. raylib rlparser README. https://github.com/raysan5/raylib/blob/master/tools/rlparser/README.md
49. rlImGui README. https://github.com/raylib-extras/rlImGui/blob/main/README.md
50. raylib-extras game-premake README (links raylib-quickstart). https://github.com/raylib-extras/game-premake/blob/main/README.md
51. Flecs Manual, "Modules". https://github.com/SanderMertens/flecs/blob/master/docs/Manual.md
52. flecs-hub `flecs-systems-sokol/project.json`; flecs-game README. https://github.com/flecs-hub/flecs-systems-sokol ; https://github.com/flecs-hub/flecs-game
53. O3DE DefaultGem template `gem.json`. https://github.com/o3de/o3de/blob/development/Templates/DefaultGem/Template/gem.json
54. O3DE `engine.json` at tags 2409.0, 2505.0, 2510.0 and 2605.0, and on development; `Gems/Atom/gem.json`. https://github.com/o3de/o3de
55. O3DE docs, "Gem Versioning and Compatibility". https://github.com/o3de/o3de.org/blob/main/content/docs/user-guide/gems/gem-versioning.md
56. O3DE docs, gem manifest reference. https://github.com/o3de/o3de.org/blob/main/content/docs/user-guide/programming/gems/manifest.md
57. O3DE release notes index (26.05.0, 2026-05-27). https://github.com/o3de/o3de.org/blob/main/content/docs/release-notes/_index.md
58. o3de-extras `repo.json`. https://github.com/o3de/o3de-extras/blob/development/repo.json
59. "O3DE 2025 In Review and A Look Into 2026" (search excerpt only). https://o3de.org/o3de-2025-in-review-and-a-look-into-2026/
60. Epic, "Plugins in Unreal Engine" (search excerpt). https://dev.epicgames.com/documentation/unreal-engine/plugins-in-unreal-engine
61. Epic API, `ELoadingPhase::Type` (search excerpt). https://dev.epicgames.com/documentation/en-us/unreal-engine/API/Runtime/Projects/ELoadingPhase__Type
62. Epic API, `EHostType::Type` (search excerpt). https://dev.epicgames.com/documentation/en-us/unreal-engine/API/Runtime/Projects/EHostType__Type
63. Epic, "Unreal Engine modules" and "Module properties" (search excerpts). https://dev.epicgames.com/documentation/unreal-engine/unreal-engine-modules ; https://dev.epicgames.com/documentation/unreal-engine/module-properties-in-unreal-engine
64. getnamo SocketIOClient-Unreal: `.uplugin` (EngineVersion 5.8), `SocketIOClient.Build.cs`, module source. https://github.com/getnamo/SocketIOClient-Unreal
65. Unreal AnswerHub thread on plugin LoadingPhase inconsistency (search excerpt). https://answers.unrealengine.com/questions/535446/view.html
66. Epic, "Asset File Format and Structure Requirements in Fab" (search excerpt). https://dev.epicgames.com/documentation/fab/asset-file-format-and-structure-requirements-in-fab
67. Community Fab packaging evidence (secondary): GhzGarage/MantleCore issue #33, https://github.com/GhzGarage/MantleCore/issues/33 ; "One Plugin, Six Engine Versions", https://www.pixelsdesign.it/blog/unreal-engine-plugin-multiple-engine-versions.html
68. Unreal releases: "Unreal Engine 5.7 is now available" (2025-11-12), https://www.unrealengine.com/de/news/unreal-engine-5-7-is-now-available ; UE 5.8, UE6 and MCP plugin coverage (secondary, June 2026), https://godisageek.com/2026/06/epic-ue6-uefn-payouts/
69. Our Machinery, "Little Machines Working Together (Part 1)" (2017-05-16), archive. https://github.com/ruby0x1/machinery_blog_archive/blob/main/post/little-machines-working-together-part-1/index.html
70. Our Machinery, "Little Machines Working Together (Part 2)" (2017-06-05), archive. https://github.com/ruby0x1/machinery_blog_archive/blob/main/post/little-machines-working-together-part-2/index.html
71. Our Machinery, "API Versioning" (2021-09-22), archive. https://github.com/ruby0x1/machinery_blog_archive/blob/main/post/api-versioning/index.html
72. Our Machinery, "The Story behind The Truth" (2018-11-07), archive. https://github.com/ruby0x1/machinery_blog_archive/blob/main/post/the-story-behind-the-truth-designing-a-data-model/index.html
73. Game World Observer, "Our Machinery terminates its game engine…" (2022-08-01). https://gameworldobserver.com/2022/08/01/our-machinery-terminates-machinery-game-engine-changes-eula
74. Game Developer, "An indie game engine is ordering users to destroy its source code" (2022). https://www.gamedeveloper.com/production/developer-our-machinery-is-ending-its-game-engine-and-asks-users-to-destroy-source-code
75. Defold manual, "Native extensions". https://github.com/defold/doc/blob/master/docs/en/manuals/extensions.md
76. Defold manual, "Extension, Application and Engine manifest files". https://github.com/defold/doc/blob/master/docs/en/manuals/extensions-ext-manifests.md
77. Defold manual, local Extender setup (profiles). https://github.com/defold/doc/blob/master/docs/en/manuals/extender-local-setup.md
78. Defold manual, "Nintendo Switch development". https://github.com/defold/doc/blob/master/docs/en/manuals/nintendo-switch.md
79. Defold `AGENTS.md` and `VERSION` (1.14.2), dev branch. https://github.com/defold/defold/blob/dev/AGENTS.md
80. Mach README and `build.zig.zon` (0.5.0, `mach_zig_version` 2026.4.10-mach). https://github.com/hexops/mach
81. Mach `src/module.zig`. https://github.com/hexops/mach/blob/main/src/module.zig
82. zig-gamedev dev repo README and `build.zig.zon` (min Zig 0.15.2). https://github.com/zig-gamedev/zig-gamedev
83. Odin `vendor/README.md`. https://github.com/odin-lang/Odin/blob/master/vendor/README.md
84. C3 vendor README. https://github.com/c3lang/vendor/blob/main/README.md
85. SDL_gpu_examples README. https://github.com/TheSpydog/SDL_gpu_examples/blob/main/README.md
86. vkguide README and `all-chapters-2` CMake (`third_party/CMakeLists.txt`). https://github.com/vblanco20-1/vulkan-guide
87. Ravbug sdl3-sample README and `.gitmodules`. https://github.com/Ravbug/sdl3-sample
88. SDL `AGENTS.md`. https://github.com/libsdl-org/SDL/blob/main/AGENTS.md
