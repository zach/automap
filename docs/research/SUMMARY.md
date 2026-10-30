# Phase 1 research summary

Synthesis of the eight Phase 1 research notes, written 2026-10-29. Every claim here is cited in the note named in brackets; this file adds no new facts. Tags follow the notes: **[Verified]** means a cited primary source, **[Inference]** means our reading of it.

| Note | Topic |
| --- | --- |
| [01](01-package-managers.md) | vcpkg, Conan 2, xmake, CPM/FetchContent; which backend to build on |
| [02a](02a-catalog-platform-graphics-ui-audio.md) | SDL3/SDL_GPU, raylib, sokol, bgfx, WebGPU; shader tools; Dear ImGui and UI; audio incl. FMOD/Wwise |
| [02b](02b-catalog-assets-and-data.md) | Asset libraries, serialization and reflection, asset-pipeline patterns |
| [02c](02c-catalog-simulation-runtime-glue.md) | Physics, navigation, ECS, allocators, jobs, math, profiling, scripting, networking; seams catalog |
| [03](03-prior-art.md) | Bevy, Godot, sokol/fibs, Magnum, O3DE, Unreal, Our Machinery, Defold, Mach, Odin, templates |
| [04](04-agent-ecosystem.md) | Agent Skills, routing, MCP, AGENTS.md, game-dev skill packs, agents writing C++ |
| [05](05-developer-needs-2026.md) | Engine landscape, platforms and consoles, iteration, determinism, licensing, agent trust |
| [06](06-verification.md) | CI cost, test ladder, headless graphics, pruning, signed evidence, staleness |

## How far to trust this research

The session's network policy blocked most vendor and documentation sites (for example gdconf.com, godotengine.org, unrealengine.com, learn.microsoft.com, docs.conan.io, cmake.org, wiki.libsdl.org, fmod.com, audiokinetic.com, slsa.dev, arxiv.org), and the shared web-search budget ran out partway through. The agents compensated by reading primary sources from source repositories (library headers, changelogs and tags; the GitHub repos behind vcpkg, Conan, CMake, GitHub and Godot docs). Library and tooling facts are therefore strong. Survey figures and vendor terms are weaker: several rest on search-engine extracts and are labeled so in the notes. The gaps that matter are listed under **Follow-ups** below; none blocks Phase 2.

## The five findings most likely to change the hypothesis

### 1. Trust is the product: "known-good" must be graded, signed, time-boxed evidence about compositions

- **[Verified]** No ecosystem publishes "A works with B". vcpkg verifies that ports build on 13 triplets; ConanCenter calls itself "not a testing service"; Rust's crater checks builds and tests per crate [06]. Bevy's compatibility evidence is hand-written README tables, and 6 of 20 popular plugins still didn't accept Bevy 0.20 three weeks after release [03]. O3DE's gem metadata is mostly empty or wrong in its own canonical repo, including a gem requiring `o3de>=4.1.0` against a 2.x engine [03].
- **[Verified]** Developers trust agent output only when they can check it: about 48% trust output they can easily verify, and 78.7% say attribution matters (Stack Overflow 2026). 59% of game programmers see generative AI as harmful (GDC 2026) [05].
- **[Verified]** Public-repo GitHub Actions runs, logs and artifacts are deleted after at most 90 days since 2026-10-01, so a link to a CI run is not durable evidence [06].
- **[Verified]** Several seams already check themselves: Jolt packs 12 ABI-affecting defines into `JPH_VERSION_ID`; Tracy 0.14 fails at link time on mismatched options; `IMGUI_CHECKVERSION` compares struct sizes; MSVC `#pragma detect_mismatch` turns define mismatches into link errors [02c][06].
- **[Inference]** Compute is cheap: about $85 per full sweep of 25 libraries × 4 platforms × 2 versions on declared pairings, about $220/month steady state, $0 on public runners. Triage and golden-image upkeep are the real cost [06].

**What changes:** a pairing is not a curated assertion; it is a claim about a pinned composition lock on a platform, at a stated verification level (build → smoke → contract probes → behavioral → hardware), with a date, an expiry and a signed record kept in Automap's own storage. Contracts need executable probes, not just prose. The verification harness and the agent's feedback loop are the same artifact.

### 2. In C++ no library owns the seams, so Automap has to act as the "virtual core"

- **[Verified]** Every composition ecosystem that thrived has one core that owns the seams and the lifecycle vocabulary: Bevy's plugin `build → ready → finish → cleanup`; Godot's four initialization levels with reverse teardown and a load-time ABI check; Unreal's module loading phases × host types; Our Machinery's versioned API registry [03].
- **[Verified]** In C++ each library brings its own allocator hook style, job model, math types, handle layout, init order and error model; no two profiled libraries share a vector type [02c]. Godot 4.6's Jolt module is production glue for exactly these seams [02c].
- **[Inference]** Job systems are the hardest seam: Jolt (jobs + barriers), Box2D 3.2 (enqueue/finish), flecs (task threads) and PhysX (dispatcher) assume different scheduling semantics, so bridging every scheduler to every library is N×M unless a neutral adapter makes it N+M [02c].

**What changes:** the contract vocabulary (phases, ownership of allocators, jobs, logging and asserts, handle conventions) *is* the product, not metadata around it. This opens a product-shape question the hypothesis didn't have: is Automap only a map plus generated glue in the user's repo, or does it also ship a thin runtime layer (for example a scheduler contract and generated init/shutdown graph)?

### 3. Build on vcpkg; Automap owns the semantic layer and an overlay registry

- **[Verified]** vcpkg has the most current catalog for the target libraries (21 of 23 carried, 18 at the latest release), a JSON manifest with a published schema, and native CMake integration. Conan is strongest for private remotes, binaries, non-additive options and signing [01].
- **[Verified]** Package managers already encode build-level fit (feature → dependency edges, platform `supports`, ABI-define propagation), so the brief's "they don't solve which libraries fit together" is too strong. None can express allocator, threading, math-type or init-order seams [01].
- **[Verified]** Pins need more than `library@version`:
  - The same versions mean different builds per backend: vcpkg builds Dear ImGui's SDL_GPU backend for you, Conan leaves it to you [01].
  - vcpkg forbids non-additive features, so variants like Jolt double precision need an overlay port [01].
  - Several key libraries are untagged or rolling (sokol, bgfx, SDL_shadercross); vcpkg's sokol port is from 2023 [01][02a][03].
  - Libraries bundle stale copies of each other: KTX 5.0-rc2 vendors Basis Universal 2.1 while Basis is at 2.5 [02b].
- **[Verified]** No C++ package manager has a release-age cooldown; Dependabot's vcpkg cooldown measures registry-commit age, not upstream release age [01].

**What changes:** Automap sits above vcpkg (primary) rather than beside it, and must run its own overlay registry for variants, gaps and fresh commits. A pairing names backend, variant, commit or version, and bundled dependencies. Cooldowns and staleness are Automap curation features.

### 4. The agent surface is a CLI with a version-keyed index behind one router skill, not hundreds of installed skills

- **[Verified]** Skill triggering is description-only and budgeted: Claude Code gives the skill listing 1% of the context window and drops descriptions of the least-used skills on overflow; Claude tends to under-trigger [04].
- **[Verified/secondary]** Generic skills add little on average, and version-mismatched guidance made results worse in SWE-Skills-Bench [04].
- **[Verified]** The mature game-dev packs route explicitly: a fingerprint router (awesome-gamedev-agent-skills, now 77 skills across 12 engines) and a typed BM25 index with CI-gated routing evals (Summer Engine, 137 skills) [04]. Both target existing engines; no composition tooling was found [04].
- **[Verified]** First-party guidance prefers a CLI where one exists, with MCP for hosts without a shell [04]. AGENTS.md is a Linux Foundation standard that Claude Code now reads natively [04].
- **[Verified]** Upstream agent files are for contributors, not consumers, and several projects ban AI-generated contributions: SDL, SDL_shadercross, sokol, stb, PhysicsFS; Godot since July 2026 [02a][02b][04][05].

**What changes:** "per-library skills" become version-keyed content that a CLI index serves to one router skill, with routing evals as CI gates. MCP is an optional wrapper. Generated project files stay small. Agent edits live in user-owned glue and are never proposed upstream.

### 5. The pure C++ custom-engine audience is narrow, and the testbed audience expects web builds the SDL_GPU stack can't produce

- **[Verified]** JetBrains: 40% of game developers used C# in the past year against 29% for C++. Custom engines made about 13% of 2024 Steam releases but about 41–42% of revenue, mostly AAA. Godot was 47% of GMTK Game Jam 2026 entries [05].
- **[Verified]** Framework-first indie hits exist and say what they want: control, a small surface, no forced upgrades (Celeste's developer on SDL3 + SDL_GPU + Dear ImGui + FMOD; Mewgenics; Animal Well) [05]. Managing third-party libraries is C++ developers' top pain (about 47% call it major) [05].
- **[Verified]** SDL_GPU has no web backend (Metal, D3D12, Vulkan and a private console slot only). WebGPU ships in all major browsers with gaps, and WebGL2 remains the needed fallback. Game jams effectively require browser builds [02a][05].

**What changes:** the primary persona has to be chosen deliberately in branch 1, not assumed. The planned SDL3 + SDL_GPU + Dear ImGui + Jolt testbed is desktop-only; a web-capable testbed is a different composition (sokol, raylib, or SDL plus WebGPU through Dawn).

## Other findings that shape later branches

| Finding | Notes | Branch |
| --- | --- | --- |
| Console support is gated by platform holders, then released privately by maintainers, vendors or porting houses; developers often hold one platform's NDA and not others. Public precedents: vcpkg Xbox triplets plus an SDK stub port; Defold's Extender console profiles; Jolt's `PlatformBlue.h`. Nothing public for PlayStation or Switch package-manager integration. Never write NDA platform names to public transparency logs. | [01][02a][03][05][06] | 9 |
| The asset pipeline is better modeled as a cross-cutting contract layer (stable IDs, versioned deterministic importers, declared dependencies, cache keys, handles with reload, packaging manifests) than as a slot; no reusable cross-platform cooker exists. Serialization and reflection are ordinary slots. The authoring data model stays custom. | [02b] | 10 |
| Determinism is a property of the whole composition: Box2D v3 deterministic by default but flag-sensitive; Jolt opt-in with matching defines everywhere; one fast-math TU breaks it. | [02c][05][06] | 4 |
| Hot reload is an expected baseline; the contract must say which libraries live in the host vs. the reloadable module and how allocators and statics behave. | [02c][05] | 4, 7 |
| Version churn breaks glue: Box2D 3.2 (2026-10-19) changed callback signatures; Lua 5.5 changed `lua_newstate` and sol2 doesn't support it; EnTT 4 and fastgltf need C++20; Dear ImGui lists breaking changes every release. The C++ standard floor is a per-composition decision. | [02a][02b][02c] | 5, 11 |
| Maintenance health varies sharply: dormant (xatlas, SoLoud, Optick, cereal, bitsery, refl-cpp, microui), stalled (sol2), and many single-maintainer libraries. | [02a][02b][02c] | 11 |
| Upstream machine-readable material exists to ingest: Box2D API JSON, `raylib_api.json`, Jolt's *SBS* serialization-break markers, Tracy's MCP server, DirectXTex `SKILL.md` files. | [02b][02c][03] | 8, 11 |
| Licensing: permissive defaults (zlib, MIT); LGPL is effectively unusable on consoles; vcpkg already emits SPDX SBOMs and license reports. | [01][05] | 4, 12 |
| Engine vendors are going agent-native (UE 5.8 experimental MCP plugin; UE6 plans one), so "an agent can drive an engine" is not a differentiator; "an agent can compose, cite and verify a multi-library stack" is. | [04][05] | 2, 12 |
| C++26 static reflection is in the final draft but GCC-only and experimental in practice; not a baseline. | [02b] | 10 |

## Leads from the brief

| Lead | Verdict |
| --- | --- |
| Godot 4.6: Jolt default, LibGodot; 4.7: HDR, new Asset Store | Confirmed. Jolt is the default for new 3D projects only; the Asset Store is beta in 4.7. [03][05] |
| Unreal 5.8 last UE5 release; UE6 EA late 2027 moves gameplay to Verse | Confirmed with nuance: 5.8 is the *planned* final UE5; Epic says Verse "transactionalizes C++". "C++ pushed down a layer" is interpretation. [05] |
| SDL 3.4 (Jan 2026) GPU/2D interop; raylib 6.0 (Apr 2026) software renderer | Confirmed. Current: SDL 3.4.18, raylib 6.2. [02a] |
| Package managers don't solve which libraries fit together | Partly corrected: they encode build-level fit; they lack semantic fit and pairing evidence. [01] |
| Agent skill packs all assume an existing engine | Confirmed; the named pack is now 77 skills across 12 engines; Summer Engine is a Godot-based engine with 137 skills. [04] |
| Bevy + cargo (mapped by arewegameyet) is the nearest analogue; C++ has none | Partly corrected: arewegameyet has no compatibility data; Odin `vendor:`, raylib's bindings table and vcpkg baselines are closer "known-good map" analogues. [03] |
| Library and slot lists | Mostly confirmed. Corrections: xatlas dormant; sol2 stalled and lacks Lua 5.5; SoLoud dormant; glslang's HLSL path deprecated; Slang's Metal/WGSL experimental; fips superseded by fibs. [02a][02b][02c][03] |

## Follow-ups

Allowing these hosts in the environment's network settings would let a short follow-up close the gaps: gdconf.com, godotengine.org, docs.godotengine.org, unrealengine.com, dev.epicgames.com, learn.microsoft.com, docs.conan.io, cmake.org, wiki.libsdl.org, fmod.com, audiokinetic.com, lua.org, liveplusplus.tech, superluminal.eu, slsa.dev, arxiv.org.

Unverified items worth closing before the spec: FMOD and Wwise current terms and SDK versions; Live++ and Superluminal platforms and terms; Unreal DDC/Zen and Our Machinery's "The Truth" details; a final sweep for engine-composition tooling outside the skills ecosystem; game-developer package-manager adoption data.
