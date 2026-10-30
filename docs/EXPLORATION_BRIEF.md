# Automap exploration prompt: an agent-ready engine composer for C++ game development

> Hand this file to Claude Code at the root of a new, empty repository. It runs a multi-session product exploration that ends with a spec an agent team can build from. The project is named **Automap**.

---

## 1. Your mission

You are leading a product exploration with me (Zach) as the product owner. The subject is **Automap**, a system that gives game developers a package-manager-like experience for **composing a custom C++ game engine from many libraries**, and that doubles as a map a coding agent can follow to assemble, extend and maintain that engine.

The exploration is finished when this repository contains a spec that is:

1. **Buildable.** An agent team with only this repo (spec, ADRs, glossary, research, issues) can start implementation without coming back to me with foundational questions.
2. **Immediately useful.** The first vertical slice delivers something a developer would actually use within days, such as a working testbed from a one-line request.
3. **Scalable.** The same model serves a throwaway testbed, a layered starter engine, and unique, very complex integrations, including private and proprietary libraries.
4. **Agent-native.** It is designed from the start for a fully agent-driven project, both in how the product is built and in how its users' agents consume it.
5. **Grounded in 2026 reality.** It meets what game developers actually need and expect in 2026, verified by research rather than assumed.

You do not write product code during this exploration. Throwaway prototypes are allowed only to answer questions that words cannot settle (see Phase 3).

---

## 2. Method: grill-with-docs

This exploration is built around Matt Pocock's `/grill-with-docs` skill from `mattpocock/skills`.

### Setup

1. Install the skills: `npx skills@latest add mattpocock/skills`. Select at least `setup-matt-pocock-skills`, `grill-with-docs` (and its dependencies), `prototype`, `handoff`, the PRD/spec skill (`to-prd`, or `to-spec` in newer versions), and `to-issues`.
2. Run `/setup-matt-pocock-skills`. Use local markdown issues under `docs/issues/` unless I say otherwise. Glossary goes in `CONTEXT.md` at the repo root, ADRs in `docs/adr/`.
3. If the skills can't be installed, emulate them faithfully using the rules below, and tell me you are doing so.

### Grilling rules

These apply for the whole exploration:

- **One question at a time.** Walk the decision tree branch by branch, resolving dependencies between decisions in order.
- **Recommend an answer with every question**, with a one- or two-sentence rationale grounded in the research. I should be able to reply "yes" and move on.
- **Don't ask what you can find out.** If research, the docs folder or a quick experiment can answer it, do that first and bring me the finding, not the question.
- **Write docs as decisions land, not at the end.**
  - Every new or sharpened term goes into `CONTEXT.md` immediately.
  - Every decision with real alternatives gets an ADR in `docs/adr/` (context, options, decision, consequences).
- **Keep a running list** of open questions and parked branches in `docs/OPEN_QUESTIONS.md`.
- **Use `/handoff` at the end of each session** so the next session resumes cleanly.
- **Challenge me.** If one of my answers conflicts with an earlier decision, the research, or what developers actually need, say so before recording it.

---

## 3. Context you are starting from

This summarizes a prior research conversation, current as of late October 2026. Treat every claim as a lead to verify, not a fact.

### Engine landscape
- **Godot** is the indie momentum story. 4.6 made Jolt the default physics engine and introduced LibGodot for embedding the engine as a library; 4.7 added HDR output and a new Asset Store.
- **Unreal** shipped 5.8 as the last major UE5 release. UE6, with Early Access targeted for late 2027, moves gameplay programming to Verse, which calls into transactionalized C++. C++ is being pushed down a layer.
- **SDL 3.4** (January 2026) improved interop between its GPU API and its 2D renderer. **raylib 6.0** (April 2026) added a software renderer.

### The custom-engine stack
The typical layering, bottom to top:
- **Platform and GPU:** SDL3 + SDL_GPU, or raylib.
- **Core runtime:** arenas, jobs, containers, logging.
- **Engine systems:** renderer frontend, asset system, physics and audio wrappers.
- **World and tools:** entity model, plus Dear ImGui editor and debug tools.
- **Game layer:** gameplay, content, scripts.

Libraries cover algorithms; custom code concentrates in gameplay, tools, the renderer frontend, the asset pipeline, and data definition.

### Candidate libraries by slot
- **Graphics:** SDL_GPU, sokol_gfx, bgfx; Slang or glslang + SPIRV-Cross for shaders.
- **Assets:** fastgltf/cgltf, ufbx, stb_image, meshoptimizer, KTX/Basis Universal, xatlas, ozz-animation, msdfgen.
- **Simulation:** Jolt, Box2D v3, Recast/Detour, flecs/EnTT.
- **Tools and UI:** Dear ImGui (ImGuizmo, ImPlot), RmlUi, Clay.
- **Audio:** miniaudio.
- **Glue:** Tracy, Lua + sol2, GameNetworkingSockets/ENet, zstd/LZ4, xxHash/BLAKE3, SQLite, FlatBuffers.

### The gap
- **C++ package managers** (vcpkg, Conan, xmake/xrepo, CPM/FetchContent) solve fetching and building. They do not solve which libraries fit together, how to wire them, or which layer owns what.
- **2026 agent skill packs for game dev** all assume an existing engine; none cover composing one. Examples include awesome-gamedev-agent-skills (66 skills plus a router across 10 engines) and Summer Engine's skill library.
- **Rust's Bevy plugins plus cargo** (mapped by arewegameyet.rs) are the nearest analogue, and C++ has no equivalent.
- **The hard part is the seams:** allocator ownership, job systems and threading, math types, handles, asset lifecycle, serialization, and build flags.

### Asset systems
Proven patterns include:
- GUID sidecars (Unity `.meta`)
- An import/cook split
- Hash-keyed derived-data caches (Unreal DDC/Zen)
- Dependency graphs
- Handle-based async loading with hot reload
- Reachability-driven packaging

Reference systems: Godot's importer, O3DE's Asset Processor, Naughty Dog's DC data compiler, and Our Machinery's "The Truth."

### Starting hypothesis (to be grilled, not assumed)
Automap is a curated, machine-readable **map of slots, libraries, integration contracts and known-good pairings, with glue recipes and verified example compositions**. It would be delivered as:
- an agent router skill plus per-library skills,
- a manifest format, and
- optionally a CLI or MCP server that scaffolds and maintains compositions.

It would scale across three tiers:
1. **Testbed:** window, render and input loop in minutes.
2. **Starter engine:** the layered architecture above.
3. **Custom integration:** swappable implementations behind contracts, plus private registries.

### My context
- I'm an engineering manager with a game-industry background.
- If a `docs/reference/` folder exists, read everything in it before Phase 1.

---

## 4. Phases

### Phase 0: Orientation (session 1, short)
- Complete setup, read `docs/reference/` if present, and seed `CONTEXT.md` with the terms already in play: slot, contract, pairing, recipe, composition, tier, testbed.
- Show me a one-screen plan of the phases and the decision tree in Phase 2, then start research.

### Phase 1: Research (before grilling the big branches)
Use parallel subagents. Write findings to `docs/research/`, one file per topic, citing sources with access dates. Separate verified facts from inference. Topics:

1. **Package-manager capabilities.** For vcpkg (manifests, registries, overlays, features, baselines), Conan 2, xmake/xrepo, and CPM/FetchContent: what each can express about compatibility, features and options, and what it can't. Which works best as a backend we build on rather than replace?
2. **Library catalog.** For each slot: candidates, license, language and API flavor, maintenance health, platform reach (desktop, web/wasm, mobile, consoles), and integration-relevant facts. Those facts are allocator hooks, threading model, math-type expectations, error handling, build options, and existing backends (for example, Dear ImGui's SDL_GPU backend).
3. **Prior art.** Bevy plugins/cargo, Godot GDExtension and LibGodot, sokol and fips, Magnum, Cute Framework, O3DE Gems, Unreal plugins and modules, and engine starter templates. What composition models exist, and why did each succeed or stall?
4. **Agent ecosystem.** The current Agent Skills format and conventions, router-skill patterns, MCP server patterns, AGENTS.md/CLAUDE.md conventions, and existing game-dev skill packs: their structure, freshness strategy, and gaps.
5. **Developer needs in 2026.** What indie and small-team developers expect, from public discussion, postmortems and surveys:
   - platform targets, including the reality of console NDAs for private overlays,
   - iteration speed and hot reload,
   - determinism, licensing comfort, and web export,
   - how they actually use coding agents, and what makes them distrust agent-generated engine code.
6. **Verification.** How to prove a composition works: CI build matrices, smoke tests, golden-image rendering tests, sanitizers. What is the cost of verifying N×M pairings, and how do you prune that space?

Close Phase 1 with a short `docs/research/SUMMARY.md` and the five findings most likely to change the hypothesis. Walk me through those before grilling begins.

### Phase 2: Grilling
Work the tree roughly in this order. Answer from research where possible, and grill me on the rest.

1. **Who and why.** Primary users (solo indie, small studio, tools programmer, agent-only builders), their jobs-to-be-done, and what "immediately useful" means to each. Pick a primary persona.
2. **Product shape.** Knowledge map, scaffolder, package-manager layer, maintenance agent, or a staged combination. What is the minimum that is still valuable on day one?
3. **Domain model.** Define slot, library, contract, pairing, recipe, composition, tier and anything else that emerges, precisely enough to become schema. This is the heart of the glossary.
4. **Contracts.** How integration requirements are expressed. Cover allocation, threading and jobs, math types, handles and ownership, lifecycle and init order, error handling, serialization, build options, and platform constraints. Decide which are machine-checkable and which are guidance for agents.
5. **Compatibility and trust.** How pairings are declared, verified, versioned and retired. What evidence backs a "known-good" claim, and how stale can it get?
6. **Build integration.** Which backend(s) to build on, how compositions become builds, and how to pin versions and respect supply-chain hygiene (for example, release-age cooldowns before adopting new versions).
7. **Tiers and growth.** How a composition grows from testbed to starter engine to custom integration without a rewrite, and what the "eject" story is when a user outgrows the product.
8. **Agent interface.** Skills, router, manifest, CLI, MCP server, and generated AGENTS.md for the user's project. What the agent reads, writes, and must never do.
9. **Extensibility and privacy.** Private and in-house libraries, studio overlays, console platforms under NDA, and offline use.
10. **The data and asset slots.** Whether data definition and the asset pipeline are first-class slots with contracts, given they are where most custom engines diverge.
11. **Freshness and governance.** Who curates, how community contributions work, and how the map stays current as libraries ship.
12. **Product and distribution.** Open source, commercial or hybrid, and how the product earns developer trust.
13. **Scope of v1.** Slots, libraries, platforms and tiers in the first release, and the first vertical slice.
14. **Success measures.** What we measure to know it's working: for example, time from request to running testbed, and the share of agent-generated compositions that build first time.

### Phase 3: Prototype spikes (only as needed)
When a question can't be settled in words (for example, "can the manifest express an allocator contract an agent will reliably honor?"), use `/prototype` for a throwaway spike in `spikes/<name>/`. Record the result in an ADR, then delete or archive the code.

One spike is probably worth doing regardless. From a draft manifest and recipes, have a fresh agent session generate an SDL3 + SDL_GPU + Dear ImGui + Jolt testbed, and record where it went wrong.

### Phase 4: Spec synthesis
1. Produce `docs/spec/SPEC.md` (see section 5), drawing only on decisions recorded in ADRs and `CONTEXT.md`. Anything not decided goes in the spec's open-questions section, not silently into the design.
2. Run `/to-prd` (or `/to-spec`), then `/to-issues` to split the work into vertical slices under `docs/issues/`. The first slice must be a thin end-to-end path, such as one tier-1 testbed composed, built and verified in CI.
3. Do a final review pass with me against the definition of done (section 6).

---

## 5. Required spec contents

`docs/spec/SPEC.md` must cover:

1. **Problem, users and jobs-to-be-done**, with the primary persona and non-goals.
2. **Product overview:** what it is, the tiers, and the user journey from first request to complex custom integration.
3. **Domain model**, linking to `CONTEXT.md`.
4. **Data formats:** manifest, slot, library, contract, pairing and recipe schemas, with complete examples for at least three slots and one private library.
5. **Architecture:** components, boundaries, build-backend integration, and the agent interface (skills, router, CLI or MCP surfaces).
6. **Verification and trust:** CI matrix strategy, test types, evidence records, and how staleness is detected.
7. **Extensibility:** private registries and overlays, proprietary and NDA platforms, and the eject path.
8. **Content plan for v1:** the exact slots, libraries, pairings and recipes, with acceptance tests for each.
9. **Operations:** curation workflow, contribution model, versioning, supply-chain hygiene.
10. **Agent-driven development plan:** how this product itself gets built by agents. Cover repo conventions, CLAUDE.md/AGENTS.md, the skills the build team uses, the test strategy (TDD where it fits), and human checkpoints.
11. **Milestones and the first three vertical slices**, each with acceptance criteria.
12. **Risks, open questions and the ADR index.**

---

## 6. Definition of done

Before declaring the exploration complete, check each item and show me the evidence:

- [ ] A fresh agent session given only this repo could begin slice 1 without asking me anything foundational. Test this by spawning a subagent with no conversation history, and record its questions.
- [ ] Every schema in the spec has a complete, valid example.
- [ ] Every "known-good" pairing in the v1 content plan names its verification method.
- [ ] Testbed, starter-engine and custom-integration tiers each have a worked example journey.
- [ ] Private and proprietary libraries are supported in the design, demonstrated with a hypothetical in-house library as the example, without special cases for it.
- [ ] `CONTEXT.md` defines every term the spec relies on, and the spec uses those terms consistently.
- [ ] Every significant decision has an ADR; `docs/OPEN_QUESTIONS.md` contains only items explicitly deferred past v1.
- [ ] Research claims the design depends on are cited and dated.

---

## 7. Working agreements

- Keep sessions focused: research, a block of grilling, then docs updates and a `/handoff`.
- Prefer short messages with one question each. Summaries go in the docs, not the chat.
- When you need my taste or judgment (product positioning, what "feels right" to game developers), say so explicitly instead of hiding it behind a recommendation.
- Begin with Phase 0 now.
