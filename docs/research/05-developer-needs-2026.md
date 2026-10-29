# What indie and small-team game developers need in 2026

**Scope.** This report covers what indie and small-team developers need and expect in late 2026, as background for Automap, a curated map of slots, libraries, contracts, pairings, recipes and compositions for building custom C++ engines with the help of a coding agent. It covers the engine landscape and why teams go custom, platform targets (consoles and web), iteration speed and hot reload, determinism, licensing comfort, how coding agents are used and trusted, and first-hand pain points of assembling an engine from libraries.

**Method and limits.** Direct page fetches were blocked in this research environment: the egress proxy refused gdconf.com, unrealengine.com, godotengine.org and wikipedia.org. Every claim below was therefore checked through search-engine extracts of the cited pages, not by reading each page in full. Where a figure appears only in secondary write-ups, the text says so. Near the end of the work the web-search budget ran out, so Unity's agent tooling and Phaser's "Game Agent" were not checked.

Accessed: 2026-10-29

---

## Key findings

1. **[Verified]** **Custom engines are rare by release count but not by revenue. The indie market is Unity plus Godot.**
   - Survey data: GDC's 2026 State of the Game Industry survey found Unreal 42% and Unity 30% as primary engine [1][2]. Secondary summaries put proprietary/in-house engines at 19% (13% in 2025) and Godot at 5% overall, but 11% among newer indie studios [1][3][4][8].
   - Steam data: custom engines made 13% (by another cut, under 10%) of 2024 Steam releases, but about 41–42% of Steam revenue, mostly from AAA [10][11].
2. **[Verified]** **Godot is the indie momentum story; minimal custom stacks also produce hits.**
   - Godot was used for 47% of GMTK Game Jam 2026 entries, against 16% in 2022 [13]. Slay the Spire II moved from Unity to Godot and sold more than 3 million copies in its first week [46].
   - Minimal stacks: Mewgenics (in-house "Glaiel Game Engine") sold 1M copies in a week [45], Balatro (LÖVE) sold 5M by January 2025 [47], and Animal Well runs on an engine written from scratch [44]. What these developers say they want is control, a small surface area and no forced upgrades [41][43][44].
3. **[Verified]** **Console support for open-source libraries lives in private, NDA-gated channels run by maintainers, vendors or porting houses.**
   - SDL keeps Switch and PlayStation backends in separate private repositories, given free under zlib only to developers who are under NDA with that platform [49].
   - Godot cannot ship console code and points users to W4 Games and porting houses [51][52]. Middleware console builds are released only after the platform holder confirms developer status [53].
   - LGPL is effectively unusable on consoles [78].
4. **[Verified]** **Developers use AI a lot but trust it less, and game developers are among the least trusting.**
   - Stack Overflow 2026: about 66% of developers use coding assistants or agents. Only about 48% trust output that they can easily verify, 78.7% say attribution matters, and 6.6% would trust AI for important decisions [20][21].
   - GDC 2026: 52% of game professionals think generative AI harms the industry, up from 30% a year earlier; among game programmers the figure is 59% [1][5].
   - ISO C++ survey 2026: 77.5% of C++ developers report faulty output and about 70% generally distrust it [22].
5. **[Verified]** **Key upstream projects now refuse agent-written code.** SDL banned LLM-generated contributions in April 2026, citing licensing and provenance, and ships an `AGENTS.md` that refuses AI contributions [27]. Godot banned almost all AI-authored code in July 2026, saying that "AI cannot take responsibility" and that disclosure is required [28].
6. **[Verified]** **Agents still struggle with deeply integrated real-time C++.**
   - GameEngineBench (July 2026): the best agent reaches 55.5% pass@1 on 110 C++ tasks in UE5 projects, and 31 tasks are unsolved by every configuration [25].
   - Practitioners say the fix is to make the game legible and checkable to the agent: serialized command replay, headless runs and validation layers in CI [26][30].
7. **[Verified]** **WebGPU now ships in all major browsers, but web builds still need a WebGL2 fallback.** It shipped in Chrome/Edge 113, Firefox 141 (Windows) and Safari 26 [56], with gaps on Linux and Android in Firefox [57]. Godot 4.7's web export still uses WebGL2 [34]. Game jams effectively require a browser build [61].
8. **[Verified]** **Managing third-party libraries is the top C++ pain point.** About 47% of C++ developers call it a major pain, ahead of build times [23]. Indies cope by vendoring and using single-header libraries [82][83]. Shader cross-compilation is a separate, recurring pain [84].
9. **[Inference]** **Engine vendors are becoming agent-native.** UE 5.8 ships an experimental MCP plugin and UE6 plans a built-in MCP server [35][38]. Automap's distinctive value is therefore not "an agent can drive an engine". It is "an agent can compose, cite and verify a multi-library stack, and keep its edits in user-owned glue code."
10. **[Inference]** **Determinism is a niche need, and a hard one to retrofit.** The libraries that support it make it opt-in, and turning it on costs SIMD and parallelism [72][73]. Automap should record it as a property of a library's contract, not a footnote.

---

## 1. Engine landscape and why teams go custom

### Survey data

**GDC State of the Game Industry 2026** (2,300+ respondents, published 29 January 2026) [1][2]:
- **Primary engine:** Unreal 42%, Unity 30%. This is the first year Unreal led clearly; the previous two surveys were roughly tied [3].
- **By studio size:** Unreal is reported more often at AA (59%) and AAA (47%) studios. Most developers at older indie studios still use Unity, and Godot is at 11% among newer indie studios [1].
- **Secondary summaries only:** proprietary/in-house 19% and Godot 5% overall [3][4]. **[Unverified]** against GDC's own report, which could not be fetched.
- **For comparison, GDC 2025:** Unity 32%, Unreal 32%, proprietary 13%, Godot 4% [8].
- **GDC 2024:** 7% had already switched engines and 28% were considering it, mostly in response to Unity's Runtime Fee [9].

**JetBrains State of Game Development 2025** (drawn from the Developer Ecosystem survey) [14]:
- Engines: Unity leads overall, 55% of AAA respondents use Unreal, and Godot reaches 31% among hobbyists.
- Languages: 40% used C# for game development in the past year, against 29% for C++.
- Rider is now the top IDE among indies.

**Perforce/JetBrains 2025 Game Technology report** (521 respondents) [15]: 36% of AAA respondents build at least part of their games on proprietary engines, up from 27%.

### Steam data

**Video Game Insights (2024 Steam data)** [10][11]:
- **Releases:** Unity 51%, Unreal 28%, Godot 5%, GameMaker 4%. Custom engines fell from 71% of releases in 2012 to 13% in 2024; 80.lv's cut of the same data says under 10%.
- **Revenue:** custom engines still earned about 41–42%, Unreal 31% and Unity 26%. This was the first year since 2014 that custom engines took less than half of Steam revenue.

**GameDiscover scans** [12]: Unity at about 49–51% of releases and Unreal rising to about 20% by 2025. Methods differ between trackers, so no single 2025 share figure is confirmed.

**Godot jam data** [13]: GMTK Game Jam 2026 was 47% Godot (about 4,900 games) against 34% Unity. Godot's share rose from 16% (2022) to 19%, 37%, 39% and then 47%. This is a jam sample, not a market measure.

### Why indies choose a custom engine or framework (first-hand)

- **Control and fit.** Noel Berry (Celeste) wrote that he doesn't need most of a big engine's features [41]. He builds on C#, SDL3 (including its GPU abstraction), Dear ImGui editors and FMOD. He also objects to being forced into engine updates late in development, especially for console builds [41][42].
- **Learning and a game-specific engine.** Elias Farhan's August 2026 series on porting *Soup Raiders* to a native engine separates "Godot extended with C++" from a real custom engine [43]. He cites curiosity and learning, and builds an engine for one game, not a Unity competitor. The concrete target is Switch at 60 Hz with loads under 3 s.
- **Escaping bloat.** Billy Basso left off-the-shelf engines because they had "a lot of bloat" and were "laggy" [44]. He also abandoned an earlier general-purpose 3D engine after about three years because it solved problems in advance instead of producing a game.
- **Licensing trust.** Unity cancelled the Runtime Fee on 12 September 2024. It went back to seat pricing (Pro up 8% to $2,200) [81]. The episode still drives choices: Mega Crit moved Slay the Spire II to Godot and ported about two years of work [46].

### Commonly cited costs

- **Anecdote:** time lost to "engine work instead of game work". Basso and the classic "make games, not engines" advice both describe this [44].
- **Inference** from the sections below: console porting effort, tooling and editors, and maintaining many dependencies.

### Recent successes and their stacks

| Game | Stack | Outcome | Source status |
|---|---|---|---|
| Mewgenics (Feb 2026) | In-house "Glaiel Game Engine" | 250k copies in about 12 h, 1M in a week; console ports dated 8 Sep 2026 | Engine language and libraries **[Unverified]** [45] |
| Balatro | Lua on LÖVE | 5M copies by January 2025 | [47] |
| Tiny Glade | Started on Bevy, then custom rendering, still a modified Bevy | 616k copies in its first month | Rendering API conflicting (WGPU vs Vulkan) [48] |
| Slay the Spire II (EA 2026) | Godot, after moving from Unity | Peak 574k concurrent players | [46] |
| Animal Well | From-scratch engine | Language **[Unverified]** | [44] |

Hades II's engine could not be verified.

---

## 2. Platform targets

### Survey data (GDC 2026) [3][4]

- **Most recent project's platforms:** PC 83%, PS5 47%, Xbox Series 40%, Steam Deck 28% (new this year), Android/iOS 23%, Linux 8%.
- **Next project:** 40% are interested in Steam Deck and 39% in Switch 2.
- **What drives platform choice:** audience access (78%), business model (44%), discoverability (43%).
- **Web:** the 2026 summaries don't break it out. In GDC 2025, 16% were preparing projects for the web, up from 9% [8].

### Console reality

**SDL** [49][50]:
- The SDL3 wiki says the Switch, PS4 and PS5 ports live in separate private repositories that track mainline. They are "available for free, under the zlib license, to anyone under NDA" with that platform holder, through Ryan Gordon (icculus). Commercial games already ship on them [49].
- One repository per console means a developer with one platform's NDA does not see the others.
- Xbox is different. SDL3's GDK README says all the code needed is public; only Microsoft's GDKX libraries are restricted to licensed developers [50].

**Godot** [51][52]:
- Godot's docs say console SDKs are under NDA, so platform code cannot be published as open source. The only console it officially supports is Steam Deck [51].
- Console builds come from you, a contractor, or third parties: W4 Games, Lone Wolf Technology, Pineapple Works and RAWRLAB [51].
- W4 offers Switch, Xbox Series and PS5 for Godot 4.3+. Secondary coverage gives tiers from $800/yr for one platform up to $10,000/yr, with Premium Support at $7,000/yr [52].

**Middleware** [53]: Audiokinetic validates PlayStation status through Sony's DevNet and gets platform-holder approval for Xbox before releasing Wwise console integrations. A Defold forum notes that the Switch FMOD integration is under Nintendo's NDA. The common pattern: **the platform holder authorizes, then the vendor releases privately.**

**Hardware access** [55]: in 2025, developers widely reported trouble getting Switch 2 dev kits; some were told to ship on Switch 1 and rely on backwards compatibility.

**Language runtime** [41]: Berry notes that modern C# NativeAOT now compiles for all major console architectures.

**[Inference] What this means for a public tool:**
- A public map can say *that* a console variant exists and *who* distributes it: maintainer, vendor or porting house. It cannot contain the variant itself, or recipes that reveal platform APIs.
- Private overlays have to be layered on and stored separately from the public map. They need per-platform scoping, because developers often hold one platform's NDA and not the others.
- Even the act of resolving an overlay must not leak NDA content into public logs or into agent context that may be shared.
- Steam Deck and Linux (via Proton or native builds) give 28% of developers a "console-like" target without an NDA.

### Web export

**Browsers:**
- WebGPU is supported in Chrome/Edge 113+, Firefox 141+ (Windows, July 2025) and Safari 26 (macOS/iOS/iPadOS/visionOS 26, September 2025). web.dev declared support across major browsers on 25 November 2025 [56][58].
- Gaps [57][88]:
  - Firefox: macOS only on Apple Silicon (from 145). Linux is Nightly-only with stable shipping targeted for 2026, and Android is behind a flag. Whether Firefox WebGPU has reached stable on Linux by October 2026 is **[Unverified]**.
  - Chrome: Android needs Android 12+. A secondary source says Linux support arrived in 144 (Intel) and 147 (NVIDIA).

**Toolchain:**
- Emscripten supports WebGPU through Dawn's `emdawnwebgpu` port (`webgpu.h`), with a "remote" port from Emscripten 4.0.10 [59].
- SDL 3.4 improved Emscripten support [39]. raylib 6.0 added an Emscripten backend and a CPU software renderer [40].

**Engines:**
- Godot 4.7 web export still uses WebGL2 through the Compatibility renderer. A secondary source says 4.7 adds wasm64 export, which removes the 4 GB memory limit **[single-source]** [34].
- Godot 4 C# projects still cannot export to the web [60].

**Demand:**
- Anecdote: itch.io jam rules often require browser play ("If it doesn't run in a browser, it doesn't count"), and judges skip download-only entries [61].
- Market data: about 15,000 HTML5 games released in H1 2025; 55% were made with Unity and 16.4% with Construct [62].
- **[Inference]:** web matters most at the jam, learner and demo tier. WebGL2 remains a fallback that must be supported.

### Mobile

Mobile is a minority target among GDC respondents (23%) [3]. No 2026 data specific to indie mobile expectations was found.

---

## 3. Iteration speed and hot reload

No survey data on hot-reload adoption was found **[Unverified]**. What follows is vendor documentation and first-hand practice.

**Common techniques:**
- **Patching running binaries.** Unreal's Live Coding is an integration of Molecular Matters' Live++ [67]. JetBrains promotes Live++ for custom C++ engines too [68]. Live++ pricing could not be verified.
- **Gameplay-DLL reload.**
  - `cr.h` saves and restores static state with `CR_STATE`. It requires host and guest to share one C runtime (`/MD`) so that both use the same allocator [63].
  - A popular Odin + raylib template ships hot reload pre-configured (about 400 GitHub stars). It keeps the existing game memory if the size of `Game_Memory` is unchanged, and offers F5 to force a reload and F6 to reset [70].
- **Data or command reload with no compile step.** Kukshtel serializes game commands as text. An agent (or a person) can write them and send them to a running game instance, then replay them from a saved start state [26].

**Pain points (first-hand):**
- Statics reset or are duplicated across reloads. An Unreal plugin static returns to its default after hot reload [66]. In one case a statically linked library (SFML) existed as two copies, one in the host executable and one in the DLL [64].
- Virtual-function pointers in existing objects go stale after a reload [64].
- Interfaces across the DLL boundary have to stay ABI-stable [64].
- On Windows, the PDB/DLL file is locked while loaded. The workaround is to copy the DLL and give each build a unique PDB name [65].

**Engines:** Unity is turning domain reload off by default (Unity 6.7 docs) ahead of CoreCLR, which has no AppDomains. This trades iteration speed for leftover static state [69].

**[Inference]** Hot reload is an expected baseline. A composition's contract therefore needs to say which libraries are safe to keep in the host versus the reloadable module, and how allocators and statics behave.

---

## 4. Determinism

**Who needs it:** deterministic lockstep (RTS and factory games), rollback netcode (fighting and action games), replays, and reproducible tests. No survey data on how common this need is was found **[Unverified]**.

**Evidence (first-hand and docs):**
- **Factorio:** runs lockstep multiplayer, where every machine must simulate each tick identically. It finds divergences by computing a CRC of the whole map every tick and replaying to the first mismatch [74].
- **Godot:** its built-in physics is not deterministic. David Snopek wrote SG Physics 2D (MIT) and an MIT rollback addon that covers input, state save/load, mismatch detection and RNG [75][91].
- **Floating point:** results are deterministic only under the same instruction set, compiler and FPU settings. Debug and release builds can diverge [76].

**What libraries offer:**
- **Box2D v3:** cross-platform determinism tested on x64 and ARM with MSVC, GCC and Clang. Multithreading made it harder; ordering is restored by merging per-thread bit arrays [71].
- **Jolt:** cross-platform determinism is opt-in through `JPH_CROSS_PLATFORM_DETERMINISTIC`, which disables fused multiply-add, plus a determinism log for debugging [72].
- **Rapier:** `enhanced-determinism` cannot be combined with its SIMD or parallel features. It requires strict IEEE-754-2008 and its own trig functions [73].

**What users need from libraries:**
- a documented determinism mode and what it costs;
- deterministic RNG;
- fast state snapshots;
- checksum hooks;
- a guarantee that the job system and threading do not reorder results.

**[Inference]** Determinism is a property of the whole composition. One non-deterministic slot (physics, RNG, a threaded job system, or a math library compiled with fast-math) breaks it. That makes it a natural thing for an integration contract to check.

---

## 5. Licensing comfort

**Defaults:**
- The stacks indies cite are overwhelmingly permissive: SDL (zlib), Godot (MIT), Box2D and Box3D (MIT) [49][77], and stb (public domain or MIT) [82].
- Unity's fee episode raised the value of licenses that cannot be revoked [9][81].

**LGPL/GPL on consoles:**
- Ryan Gordon could not use OpenAL Soft (LGPL) on Switch because users cannot replace the library there. He wrote a zlib replacement, MojoAL [78].
- Forum advice on statically linking LGPL code is cautious and unsettled [79].
- **[Inference]** GPL and LGPL are effectively off the table for console-bound projects, and indies avoid them for that reason.

**Proprietary SDKs:** these are accepted when there is an indie tier. FMOD's licensing page lists an Indie tier for development budgets under $600k ("Free or $2,000"); older coverage gave different caps [54]. Wwise's current free tier is **[Unverified]**.

**Tooling:** vcpkg writes an SPDX SBOM for every installed package and, since July 2025, `vcpkg license-report` prints the SPDX licenses of all installed packages [80]. No survey of indie expectations for compliance tooling was found **[Unverified]**. Developers who vendor sources keep license files by hand.

**AI provenance is now a licensing issue:**
- SDL's ban is justified by the risk that AI output brings in code under unknown or incompatible licenses. Contributors must affirm authorship under zlib [27].
- Valve's January 2026 update exempts code assistants from Steam's AI disclosure; disclosure is still required for AI-generated content [29]. The constraint on agent-written code therefore comes from upstream projects and from developers' own trust, not from the store.

---

## 6. Coding agents in game development

### Usage (survey data)

**GDC 2026** [1][5]:
- 36% of respondents use generative AI at work; 30% at studios, against 58% at publishers, support and marketing firms.
- Tools: ChatGPT 74%, Gemini 37%, Copilot 22%.
- Uses: research and brainstorming 81%. A secondary source puts code assistance at 47% *of AI users* [6] **[secondary only]**.
- Only 5% use it for player-facing features.

**JetBrains game dev 2025** [14]: "nearly half" regularly use AI for feature implementation.

**Google Cloud / Harris Poll** [24] (vendor-commissioned; n=615; US, South Korea, Nordics; June–July 2025): 90% use AI in workflows and 87% use "agents", but 63% worry about data ownership.

**General developer population:**
- **Stack Overflow 2026** (more than 30k respondents) [20][21]: 65.9% use coding assistants or agents, and 26.2% run automated agent workflows. Among agents named, Claude Code leads (66%) ahead of Copilot (59%). Use is mostly for code in familiar areas (67%) and debugging (61%).
- **JetBrains Developer Ecosystem 2025** (24,534 respondents) [17]: 85% regularly use AI tools and 62% rely on at least one coding assistant, agent or AI editor.
- **Stack Overflow 2025** [18][19]: 46% distrusted the accuracy of AI output against 33% who trusted it. 45% named "almost right, but not quite" as their top frustration, and 66% said they spend extra time fixing such code. 31% used agents and 38% had no plans to.

**C++ developers:**
- **ISO C++ 2026** (n=1,434) [22]: 39.8% frequently use AI to write code, up from 30.9%; 42% rarely or never use it.
- **JetBrains State of C++ 2025** [16]: C++ developers are less likely than others to try AI coding tools (44% vs 50%). Security and IP concerns are the main blockers.

### Sentiment

- 52% of game professionals see generative AI as harmful (game programmers 59%), up from 18% two years earlier [1][5].
- The GDC Trends Report frames acceptance as growing for planning and routine work, with augmentation preferred over replacement [7].

### Why they distrust agent-written engine code

**Benchmark:** GameEngineBench reports 55.5% pass@1 at best on UE5 C++ tasks. The authors conclude that agents "continue to struggle with deeply integrated C++ development for real-time interactive software" [25].

**Anecdote:**
- **Performance-naïve or structurally wrong code.** HN commenters say agent code looks reasonable at the micro scale but misses something important, and that agents fail to see "that a data structure is inside out" [31].
- **Threading and GPU-sync bugs.** One agent-assisted Zig/Vulkan project had about 600 synchronization hazards (write-after-write and write-after-read) that only the validation layer exposed. The fix was a barrier plus sync validation in CI [30].
- **Hallucination and provenance.** SDL's policy warns that AI "frequently hallucinates issues" and that its output carries licensing risk [27].
- **Responsibility and understanding.** Godot doubts that heavy AI users understand their code well enough to fix it [28].
- **Games aren't legible to the agent.** Scenes, assets and runtime state are opaque to it, unlike most web code [26]. A secondary survey says agents work best with text-serialized engines such as Godot and worst with binary-asset engines [90] **[anecdote]**.

### What would make them trust it

**Survey:**
- Easy verification: about 48% trust output when they can verify it [21].
- Attribution: 78.7% call it important [21].
- Stack Overflow's CEO frames skepticism as a safeguard: developers need to know where an answer came from and whether it can be checked [20].

**Practice (anecdote):**
- small tasks, tight diff review, small PRs and strict CI [89];
- validation layers in CI [30];
- deterministic replay and headless runs that print checkable text [26].

**[Inference]** Trust comes from cited, verified, small, reversible changes inside code the user owns. It does not come from bigger generated diffs.

### Vendor moves

- UE 5.8 ships an experimental first-party MCP plugin. A hands-on review found it reached some editor systems but not others [38].
- Epic says UE6 will include an MCP server [35].
- Phaser announced an engine built around AI agents (June 2026). This is **[Unverified detail]**: only the search title was seen.

---

## 7. Pain points of assembling an engine from libraries

**Survey data:**
- ISO C++ surveys put "managing libraries my application depends on" first, with about 47% calling it a major pain (2022–2023), ahead of build times, CMake and CI setup [23].
- JetBrains: 28% of C++ developers cite dependency constraints as the reason they don't move to a newer C++ standard [16].

**First-hand and docs:**
- **Build and dependency management.**
  - FetchContent is called both "the simplest and most problematic" option. vcpkg turns FetchContent off by default, and transitive dependencies cause conflicts [85].
  - The stb FAQ explains single-header libraries: Windows has no standard library directories, and libraries built against a different C runtime cause link conflicts [82].
  - Forrest Smith argues that dependencies belong in version control [83]. **[Inference]** Vendoring is the norm for shipping indies.
- **Shader toolchains.**
  - Each SDL_GPU backend needs its own shader format: SPIR-V, DXIL/DXBC or MSL [84].
  - SDL_shadercross translates offline. Its compile APIs were reworked, DXIL output needs a build with DXC, and Metal needs Apple's tools [84].
- **Integration seams.**
  - Jolt requires its allocator to be registered before any other call. Every update takes a temp allocator and a job system, and Jolt recommends implementing its `JobSystem` interface on your own scheduler [72][86].
  - Library stacks such as glm, EnTT, Jolt and taskflow need math-type and threading glue between them [87].
  - Hot reload adds rules about which library lives in which module [63][64].
- **Version churn.**
  - Libraries move quickly: SDL 3.4 added GPU and 2D-renderer interop and native PNG [39]; raylib 6.0 added a software renderer and new platform backends [40].
  - Developers want to avoid being *forced* to upgrade late in a project [41].
- **How people cope today:** templates (e.g. the Odin + raylib hot-reload template) [70], personal frameworks reused across games (Berry, Glaiel) [41][45], copying from tutorials and sample projects, and vendoring [82][83]. **[Inference]** No current tool records "known-good pairings" across libraries; that knowledge lives in individuals' repos.

---

## Leads checked

| Lead | Verdict | Notes |
|---|---|---|
| Godot 4.6 made Jolt the default physics engine | **Confirmed** | Default for *new* 3D projects; existing projects keep their settings. Released 26–27 Jan 2026 [32]. |
| Godot 4.6 introduced LibGodot | **Confirmed** | Initial support on Linux, Windows and macOS; web not confirmed [32]. |
| Godot 4.7 added HDR output and a new Asset Store | **Confirmed, with a caveat** | HDR for 2D and 3D, on Linux only under Wayland. The Asset Store replaces the Asset Library but is marked beta. Released about 18–19 Jun 2026 [33]. |
| "Godot is the indie momentum story" | **Supported** | GMTK 47% [13]; 11% of newer indie studios [1]; Slay the Spire II [46]. Only about 5% of all GDC respondents [4]. |
| Unreal shipped 5.8 as the last major UE5 release | **Corrected (nuance)** | Released 17 Jun 2026 as the *planned* final UE5 release; Epic does not rule out a 5.9 [36]. |
| UE6 Early Access targeted for late 2027 | **Confirmed** | "End of 2027", "-ish" [35][36]. |
| UE6 moves gameplay programming to Verse, "which calls into transactionalized C++"; "C++ pushed down a layer" | **Corrected (wording)** | Epic's words: moving the gameplay programming model to Verse, "which transactionalizes C++" [35]. "Pushed down a layer" is an interpretation; C++ is not being removed and secondary sources disagree on scope [37]. |
| SDL 3.4 (January 2026) improved GPU and 2D-renderer interop | **Confirmed** | Announced 1 Jan 2026; adds `SDL_CreateGPURenderer`, GPU render state and texture interop [39]. |
| raylib 6.0 (April 2026) added a software renderer | **Confirmed** | rlsw, a CPU OpenGL 1.1+ backend. The 23 Apr date comes from a changelog mirror only [40]. |

No lead was outright false.

---

## Implications for Automap (inference)

### Candidate personas

**P1. Jam developer, learner or testbed builder**
- Mostly Godot, raylib, SDL or LÖVE users, and students.
- *Immediately useful:* one manifest → a running window, input, audio, Dear ImGui and a **web build** in minutes, with hot reload already set up. The agent can add a slot and the result still builds.
- *Must-have:* permissive licenses only, WebGL2 fallback, zero-config builds.

**P2. Framework-first shipping indie**
- The Berry, Glaiel and Basso archetype; also people porting games to a native engine, like Farhan.
- *Immediately useful:*
  - known-good pairings with pinned versions;
  - glue recipes for the seams: allocator, job system, math types, shaders, hot-reload module boundaries;
  - a vendoring-friendly output (source in the repo, no service dependency);
  - an SBOM or license report;
  - upgrade notes for library bumps.
- *Must-have:* no lock-in, reproducible offline builds, small and readable glue code the developer owns.

**P3. Small studio with custom tech going to consoles**
- Includes rollback and lockstep games.
- *Immediately useful:*
  - a private overlay format that adds console variants (SDL private forks, W4, vendor builds) per platform without leaking NDA content;
  - a CI matrix;
  - determinism flags recorded as part of each library's contract.
- *Must-have:* overlays stored separately per platform; a public map that only names who distributes console support.

**P4. Engine extender**
- Godot users adding GDExtension modules or embedding Godot through LibGodot; swapping physics engines.
- *Useful:* treat Godot (via LibGodot) as an option in a slot, and record contracts at the GDExtension boundary.
- Open question whether this is in scope.

**P5. Agent-led developer**
- *Immediately useful:*
  - a router skill and per-library skills that **cite** upstream docs and commits;
  - checks the agent must run: build, validation layers, headless smoke test, deterministic replay;
  - a rule that agent edits stay in user glue and never become upstream patches, honoring `AGENTS.md` refusals like SDL's.

### Must-have expectations (cross-persona)

1. **Provenance on every recipe and pairing:** an upstream source link, the version it was checked against, and the date it was last verified. This maps directly to the 78.7% who say attribution matters [21].
2. **Verified example compositions** built in CI on Windows, Linux, macOS and web, with smoke tests.
3. **Permissive-license defaults** that flag LGPL and GPL, plus a license report.
4. **Reproducible, vendor-friendly output.** No mandatory registry service at build time.
5. **Small, reviewable diffs** confined to user-owned glue, and an explicit "do not upstream agent code" policy.
6. **Recipes for hot reload and shader pipelines.**

### Nice-to-have

- An MCP server or CLI. Engines already ship MCP, so this is not a differentiator.
- Performance benchmarks per pairing.
- Determinism certification runs.
- A WebGPU-first web path (the WebGL2 fallback stays a must).
- Console overlays, built on the tier-3 design from the start but shipped later.

### Possible changes to the starting hypothesis

- **(a) Provenance and verification are the product, not an add-on.** Upstream projects and C++ developers distrust agent code, and trust depends on verification and attribution.
- **(b) The private-overlay design is set by platform holders, not by Automap.** It should be designed in from day one, even if it ships late.
- **(c) The pure C++ custom-engine audience is small; framework-first and Godot-adjacent developers are larger.** C# outranks C++ in indie usage [14]. A C++-only map risks a narrow market.

---

## Open questions

1. How large is the audience that would build a C++ engine from libraries, against those using Godot/LibGodot, or C#, Odin, Zig or Rust frameworks? Should Automap cover non-C++ bindings (for example C# over SDL3, as Berry does)?
2. How can console compositions be verified when Automap has no NDA access? Could porting houses or W4 act as overlay maintainers?
3. Can a public map legally say "SDL Switch backend exists, contact the maintainer" without breaching an NDA? (The SDL wiki already does [49].) What metadata about NDA platforms is safe to publish?
4. Will library maintainers object to a tool that drives agents over their APIs, given SDL's and Godot's policies? Can Automap win them over by showing it never generates upstream patches?
5. What counts as a "known-good" pairing, and how often must it be re-verified as SDL, raylib and Jolt ship new versions?
6. How should determinism be expressed and tested across a whole composition, beyond per-library flags?
7. For the web tier, should the baseline be WebGL2 (Godot's choice) or WebGPU, given the gaps in Firefox on Linux and Android?
8. There is no first-hand survey of how indies use agents *on engine-level C++* specifically. A small user study would close this gap.
9. Topics left unverified because search ran out: Unity's agent tooling, Phaser's Game Agent, Live++ pricing, the Mewgenics and Animal Well tech stacks, and Firefox WebGPU on Linux in stable.

---

## Sources

All accessed 2026-10-29. Each entry gives the title, URL and publication date ("n.d." = undated or a living page). They were read through search-engine extracts because direct fetches were blocked.

1. GDC, "GDC 2026 State of the Game Industry reveals impact of layoffs, generative AI, and more", https://gdconf.com/article/gdc-2026-state-of-the-game-industry-reveals-impact-of-layoffs-generative-ai-and-more/ — Jan 2026.
2. Business Wire, "2026 State of the Game Industry Report Reveals Widening Effect of Layoffs…", https://www.businesswire.com/news/home/20260129438528/en/ — 2026-01-29.
3. GameDev Reports, "GDC: The State of the Game Industry - 2026", https://gamedevreports.substack.com/p/gdc-the-state-of-the-game-industry-fcf — early 2026.
4. GamingOnLinux, "GDC 2026 report: 36% of devs use GenAI, 28% target Steam Deck and 8% target Linux", https://www.gamingonlinux.com/2026/01/gdc-2026-report-36pct-of-devs-use-genai-28pct-target-steam-deck-and-8pct-target-linux/ — Jan 2026.
5. Game Developer, "One third of game workers use generative AI, but half think it's bad for the industry", https://www.gamedeveloper.com/business/one-third-of-game-workers-use-generative-ai-but-half-think-it-s-bad-for-the-industry — Jan 2026.
6. Ziva, "GDC 2026 generative AI game development data", https://ziva.sh/blogs/gdc-2026-generative-ai-game-development-data — 2026 (vendor blog, secondary).
7. GameDev.net, "GDC Trends Report 2026: As use of generative AI rises, devs face 'infrastructure problem'", https://www.gamedev.net/news/gdc-trends-report-2026-as-use-of-generative-ai-rises-devs-face-infrastructure-problem-r3601/ — May 2026.
8. Games Market, "Latest trends and insights from the GDC 2025 State of the Game Industry report", https://www.gamesmarket.global/industry-survey-latest-trends-and-insights-from-the-gdc-2025-state-of-the-game-industry-report-6c5cef296d946f9d85f6692b09bef4e1/ — Jan 2025.
9. WN Hub, "Over the past year, a third of game developers have changed their game engine or thought about doing so", https://wnhub.io/news/engines/item-42759 — Jan 2024.
10. Video Game Insights, "The Big Game Engines Report of 2025" (PDF), https://vginsights.com/assets/reports/The_Big_Game_Engines_Report_of_2025.pdf — early 2025; summary: https://gamedevreports.substack.com/p/video-game-insights-game-engines.
11. 80.lv, "Less than 10% of games released on Steam in 2024 were made with proprietary engines", https://80.lv/articles/less-than-10-of-games-released-on-steam-in-2024-were-made-with-proprierary-engines/ — 2025.
12. GameDiscover, "How's PC game engine usage changing?", https://newsletter.gamediscover.co/p/hows-pc-game-engine-usage-changing — n.d. (2025/26).
13. Game World Observer, "Unity is no longer the leader — Godot has become the most popular engine at GMTK Game Jam 2026", https://gameworldobserver.com/2026/07/29/unity-is-no-longer-the-leader-godot-has-become-the-most-popular-engine-at-gmtk-game-jam-2026 — 2026-07-29.
14. JetBrains, "The State of Game Development 2025", https://lp.jetbrains.com/the-state-of-gamedev-2025/ — late 2025; excerpt: https://blog.jetbrains.com/dotnet/2026/01/29/game-dev-in-2025-excerpts-from-the-state-of-game-development-report/ — 2026-01-29.
15. Perforce & JetBrains, "2025 State of Game Technology Report", https://www.perforce.com/resources/p4/game-technology-report — Sep 2025; coverage: https://wnhub.io/news/analytics/item-48857.
16. JetBrains, "The State of C++ 2025", https://lp.jetbrains.com/the-state-of-cpp-2025/ — 2025/26.
17. JetBrains Research, "State of Developer Ecosystem 2025", https://blog.jetbrains.com/research/2025/10/state-of-developer-ecosystem-2025/ — Oct 2025.
18. Stack Overflow, "Developers remain willing but reluctant to use AI: the 2025 Developer Survey results are here", https://stackoverflow.blog/2025/07/29/developers-remain-willing-but-reluctant-to-use-ai-the-2025-developer-survey-results-are-here/ — 2025-07-29.
19. DevOps.com, "Stack Overflow survey shows AI adoption for devs", https://devops.com/stack-overflow-survey-shows-ai-adoption-for-devs/ — mid-2025.
20. Stack Overflow, "The results of the 2026 Developer Survey are here", https://stackoverflow.blog/2026/10/06/the-results-of-the-2026-developer-survey-are-here/ — 2026-10-06.
21. IT Brief UK, "Stack Overflow survey shows developers trust AI with caveats", https://itbrief.co.uk/story/stack-overflow-survey-shows-developers-trust-ai-with-caveats — Oct 2026.
22. The Register, "C++ survey finds AI use rising, though trust is in short supply", https://www.theregister.com/devops/2026/05/07/c-survey-finds-ai-use-rising-though-trust-is-in-short-supply/5234708 — 2026-05-07; Heise, https://www.heise.de/en/news/C-Developers-Use-AI-More-Often-But-Remain-Skeptical-11289474.html — May 2026.
23. Standard C++ Foundation, "CppDevSurvey 2023 summary" (PDF), https://isocpp.org/files/papers/CppDevSurvey-2023-summary.pdf — 2023.
24. Google Cloud, "90% of Games Developers Already Using AI in Workflows", https://www.googlecloudpresscorner.com/2025-08-18-90-of-Games-Developers-Already-Using-AI-in-Workflows,-According-to-New-Google-Cloud-Research — 2025-08-18.
25. arXiv 2607.03525, "GameEngineBench: Evaluating Coding Agents on Real C++ Runtime Environments", https://arxiv.org/abs/2607.03525 — Jul 2026.
26. Kyle Kukshtel, "Your Game Is A Harness", https://kylekukshtel.com/ai-harness-engineering-game-development-claude-code — 2026-10-10; HN: https://news.ycombinator.com/item?id=50072120.
27. Phoronix, "SDL Adds Policy To Forbid LLM/AI Generated Code Contributions", https://phoronix.com/news/SDL-Says-No-To-AI-LLMs — Apr 2026; libsdl: https://discourse.libsdl.org/t/sdl-policy-added-to-the-pr-template-and-an-agents-md-refusing-ai-contributions/67256 — Apr 2026.
28. Game Developer, "Godot to ban almost all AI coding contributions", https://www.gamedeveloper.com/business/godot-to-ban-almost-all-ai-coding-contributions — ~2026-07-01; follow-up: https://gamedeveloper.com/business/godot-confirms-it-tolerates-some-ai-assistance-but-rejects-vibe-coded-accusations — Jul 2026.
29. Gigazine, "Steam updates AI disclosure guidelines", https://gigazine.net/gsc_news/en/20260120-steam-updates-ai-disclosure-guidelines — 2026-01-20.
30. zigguro PR #1315, "fix(vulkan): clear synchronization-validation hazards", https://github.com/ChristianPresley/zigguro/pull/1315 — 2026 (n.d.).
31. Hacker News, "Two things LLM coding agents are still bad at", https://news.ycombinator.com/item?id=45523537 — 2025; "Why LLMs can't really build software", https://news.ycombinator.com/item?id=44900116 — 2025.
32. Godot, "Godot 4.6", https://godotengine.org/releases/4.6/ — Jan 2026; Phoronix, https://phoronix.com/news/Godot-4.6-Released — 2026-01-26.
33. Phoronix, "Godot 4.7 Released With HDR Output Support", https://www.phoronix.com/news/Godot-4.7-Released — Jun 2026; Digital Production, https://digitalproduction.com/2026/06/23/godot-4-7-turns-up-hdr/ — 2026-06-23.
34. Cinevva, "Godot 4.7 released", https://app.cinevva.com/news/2026-06-19-godot-4-7-released — 2026-06-19 (secondary).
35. Epic Games, "State of Unreal 2026: Top news from the show", https://www.unrealengine.com/news/state-of-unreal-2026-top-news-from-the-show — 2026-06-17.
36. GameFromScratch, "Unreal Engine 5.8 Released", https://gamefromscratch.com/unreal-engine-5-8-released/ — Jun 2026; Inven Global, https://www.invenglobal.com/articles/23006/unreal-engine-58-released-enhanced-large-scale-open-world-and-ai-workflows — Jun 2026.
37. GamesBeat, "Unreal Engine 6 will combine UE5 and UEFN into a unified engine", https://gamesbeat.com/unreal-engine-6-will-combine-ue5-and-uefn-into-a-unified-engine-state-of-unreal/ — Jun 2026.
38. Puget Systems, "Unreal Engine MCP Hands-On: Testing AI Inside the Editor", https://www.pugetsystems.com/blog/2026/07/09/unreal-engine-mcp-hands-on-testing-ai-inside-the-editor/ — 2026-07-09.
39. libsdl, "Announcing SDL 3.4.0", https://discourse.libsdl.org/t/announcing-sdl-3-4-0/65438 — 2026-01-01; Phoronix, https://www.phoronix.com/news/SDL-3.4-Released — 2026-01-01.
40. raysan5, "raylib 6.0 released", https://raysan5.itch.io/raylib/devlog/1497342/raylib-60-released — Apr 2026.
41. Noel Berry, "Making Video Games in 2025 (without an engine)", https://noelberry.ca/posts/making_games_in_2025/ — 2025-05-18; HN: https://news.ycombinator.com/item?id=44038209.
42. Hanselminutes #1014, "Making Video Games in 2025 with Celeste's Noel Berry", https://hanselminutes.com/1014/making-video-games-in-2025-with-celestes-noel-berry — 2025-09-11.
43. Elias Farhan, "Soup Raiders Goes Native: What You Gain by Building Your Own Game Engine", https://eliasfarhan.ch/gamedev/cpp/2026/08/14/srnative-01-why-a-custom-engine.html — 2026-08-14.
44. Game Developer, "Why Animal Well's home-brewed engine was key to its success", https://gamedeveloper.com/design/why-animal-well-s-home-brewed-engine-was-key-to-its-success — 2024.
45. Game World Observer, "Mewgenics recouped its development costs three hours after release", https://gameworldobserver.com/2026/02/11/mewgenics-recouped-its-development-costs-three-hours-after-release — 2026-02-11; Game Informer, https://gameinformer.com/2026/02/17/mewgenics-hits-1-million-copies-sold — 2026-02-17; Wikipedia, https://en.wikipedia.org/wiki/Mewgenics — n.d.
46. Wikipedia, "Slay the Spire II", https://en.wikipedia.org/wiki/Slay_the_Spire_II — n.d.; 80.lv, https://80.lv/articles/slay-the-spire-2-is-sticking-with-godot-leaving-unity-behind/ — 2024.
47. MobileSyrup, "Balatro 5 million copies sold", https://mobilesyrup.com/2025/01/21/balatro-5-million-copies-sold-update/ — 2025-01-21; Wikipedia, "Löve (game framework)", https://en.wikipedia.org/wiki/L%C3%B6ve_(game_framework) — n.d.
48. 80.lv, "Tiny Glade developers discuss Bevy…", https://80.lv/articles/exclusive-tiny-glade-developers-discuss-bevy-proceduralism-publishers-cozy-games/ — 2024; Gigazine, https://gigazine.net/gsc_news/en/20241120-tiny-glade-600k-sold-reason — 2024-11-20.
49. SDL Wiki, "README-switch", https://wiki.libsdl.org/SDL3/README-switch and "README-ps5", https://wiki.libsdl.org/SDL3/README-ps5 — n.d. (living).
50. SDL Wiki, "README-gdk", https://wiki.libsdl.org/SDL3/README-gdk — n.d. (living).
51. Godot Docs, "Console support in Godot", https://docs.godotengine.org/en/3.5/tutorials/platform/consoles.html and 4.x source https://docs.godotengine.org/de/4.3/_sources/tutorials/platform/consoles.rst.txt — n.d.
52. W4 Games, "W4 Consoles", https://www.w4games.com/w4consoles — n.d.; GameFromScratch, https://gamefromscratch.com/w4-games-release-godot-w4-console-pricing/ — n.d.
53. Audiokinetic Q&A, "Where to get Unity integration packages", https://www.audiokinetic.com/qa/2136/where-to-get-unity-integration-packages — n.d.; Defold forum, "FMOD for Switch support?", https://forum.defold.com/t/fmod-for-switch-support/65709 — n.d.
54. FMOD, "Licensing", https://fmod.com/licensing — n.d. (living).
55. Nintendo Life, "Developers reportedly still struggling to get Switch 2 dev kits", https://www.nintendolife.com/news/2025/08/developers-reportedly-still-struggling-to-get-switch-2-dev-kits — Aug 2025.
56. web.dev, "WebGPU is now supported in major browsers", https://web.dev/blog/webgpu-supported-major-browsers — 2025-11-25.
57. GPU for the Web, "Implementation Status", https://github.com/gpuweb/gpuweb/wiki/Implementation-Status — n.d. (living).
58. Linuxiac, "WebGPU Lands in Firefox 141 on Windows…", https://linuxiac.com/?p=177809 — Jul 2025.
59. Emscripten, "Using WebGPU in Emscripten", https://emscripten.org/docs/porting/multimedia_and_graphics/WebGPU-support.html — n.d.
60. Godot Forum, "Is there an update on exporting C# projects to Web", https://forum.godotengine.org/t/is-there-an-update-on-exporting-c-projects-to-web/128821 — 2025/26.
61. itch.io, "Cthulhu Jam" rules, https://itch.io/jam/cthulhujam — Dec 2025.
62. GameDev Reports, "Playgama: Most Popular Engines for HTML5 Games in H1'25", https://gamedevreports.substack.com/p/playgama-most-popular-engines-for — 2025.
63. cr.h, "A Simple C Hot Reload Header-only Library", https://github.com/island-org/cr — n.d.
64. Handmade Network, "hot reloading problem (SFML)", https://hero.handmade.network/forums/code-discussion/t/7015/p/21433 and "Hot-reloading code with virtual tables", https://hero.handmade.network/forums/code-discussion/t/2677/p/13331 — n.d.
65. Microsoft Q&A, "Copying a DLL fails with error 32", https://learn.microsoft.com/en-us/answers/questions/1601703/copying-a-dll-fails-with-error-32 — 2024.
66. Unreal Engine Issues, "UE-61259", https://issues.unrealengine.com/issue/UE-61259 — n.d.
67. Epic Games, "Using Live Coding to recompile Unreal Engine applications at runtime", https://dev.epicgames.com/documentation/en-us/unreal-engine/using-live-coding-to-recompile-unreal-engine-applications-at-runtime — n.d.
68. JetBrains, "Hot reloading C++ with Live++ and JetBrains Rider", https://blog.jetbrains.com/dotnet/2024/09/24/hot-reloading-cplusplus-with-liveplusplus-and-jetbrains-rider/ — 2024-09-24.
69. Unity Docs, "Configurable Enter Play Mode" (6000.7), https://docs.unity.com/en-us/engine/6000.7/manual/programming-environment/code-reload-serialization/code-reload/code-reloading-editor/configurable-enter-play-mode — n.d.; Unity Discussions, "Path to CoreCLR, 2026: Upgrade Guide", https://discussions.unity.com/t/path-to-coreclr-2026-upgrade-guide/1714279 — 2026.
70. odin-raylib-template README (fork of Karl Zylinski's Odin + raylib hot-reload template), https://next.tangled.org/wlls.dev/odin-raylib-template/blob/main/README.md — n.d.
71. Erin Catto, "Determinism", https://box2d.org/posts/2024/08/determinism/ — Aug 2024.
72. Jolt Physics, Build README, https://raw.githubusercontent.com/jrouwe/JoltPhysics/master/Build/README.md and HelloWorld.cpp, https://raw.githubusercontent.com/jrouwe/JoltPhysics/master/HelloWorld/HelloWorld.cpp — n.d.
73. Rapier, "Determinism", https://rapier.rs/docs/user_guides/rust/determinism — n.d.
74. Factorio, "Friday Facts #188 — Bug, Bug, Desync", https://factorio.com/blog/post/fff-188 — c. 2017; "Friday Facts #47 — CRC fun", https://factorio.com/blog/post/fff-47 — c. 2014.
75. David Snopek, "Retro Tank Party v1.2: rollback netcode", https://dsnopek.itch.io/retro-tank-party/devlog/330853/retro-tank-party-v12-rollback-netcode — c. 2022; Godot Forums, SG Physics 2D, https://godotforums.org/d/28338-getting-started-with-sg-physics-2d-and-deterministic-physics-in-godot-good-for-rollback-netcode — c. 2022.
76. Glenn Fiedler, "Floating Point Determinism", https://gafferongames.com/post/floating_point_determinism/ — 2010.
77. Box2D, "Announcing Box3D", https://box2d.org/posts/2026/06/announcing-box3d/ — Jun 2026; Phoronix, https://phoronix.com/news/Box3D-Open-Source-3D-Physics — 2026-07-02.
78. Ryan C. Gordon, "Project MojoAL", https://www.patreon.com/icculus/posts/project-mojoal-17871211 — c. 2018.
79. OpenFL Community, "OpenAL android static/dynamic linking and LGPL license", https://community.openfl.org/t/openal-android-static-dynamic-linking-and-lgpl-license/10573 — c. 2018.
80. Microsoft Learn, "vcpkg license-report", https://learn.microsoft.com/en-us/vcpkg/commands/license-report — n.d.; Microsoft C++ Blog, "What's new in vcpkg (July 2025)", https://devblogs.microsoft.com/cppblog/whats-new-in-vcpkg-july-2025/ — Jul/Aug 2025.
81. Unity, "Unity is canceling the Runtime Fee", https://unity.com/blog/unity-is-canceling-the-runtime-fee — 2024-09-12.
82. Sean Barrett, stb README / FAQ, https://github.com/nothings/stb — n.d.
83. Forrest Smith, "Dependencies Belong in Version Control", https://forrestthewoods.com/blog/dependencies-belong-in-version-control — 2023.
84. Moonside Games, "Introducing SDL_shadercross", https://moonside.games/posts/introducing-sdl-shadercross/ — c. 2024/25; libsdl, "A bit confusion over how SDL3 handles shaders", https://discourse.libsdl.org/t/a-bit-confusion-over-how-sdl3-handles-shaders/59922 — 2025; "SDL_gpu_shadercross: Rework Compile APIs", https://discourse.libsdl.org/t/sdl-gpu-shadercross-rework-compile-apis/55454 — c. 2024/25.
85. CMake Discourse, "FetchContent vs vcpkg, conan?", https://discourse.cmake.org/t/fetchcontent-vs-vcpkg-conan/6578 — n.d.; Medium, "Battling transitive C++ dependencies with FetchContent", https://medium.com/@nerudaj/battling-transitive-c-dependencies-with-fetchcontent-3ee0300a7973 — n.d.
86. Jan Wedekind, "Getting started with the Jolt Physics Engine", https://www.wedesoft.de/simulation/2024/09/26/jolt-physics-engine/ — 2024-09-26.
87. darmok, "C++ game engine combining open-source libraries & tools that I like", https://github.com/miguelibero/darmok — n.d.
88. Wikipedia, "WebGPU", https://en.wikipedia.org/wiki/WebGPU — n.d.
89. Althera Games, "AI Agents for Game Development 2026: A Practical Indie Workflow", https://altheragames.com/en/blog/ai-agents-game-development-2026 — 2026 (studio blog, opinion).
90. chierhu (Medium), "Claude Code for Game Development: A Comprehensive Survey of Real-World…", https://chierhu.medium.com/claude-code-for-game-development-7a88fcd19992 — 2026 (secondary).
91. Godot Asset Library, "Godot Rollback Netcode (Godot 4)", https://godotengine.org/asset-library/asset/2450 — n.d.
