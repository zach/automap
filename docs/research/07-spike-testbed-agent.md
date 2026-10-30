# 07: Spike — can an agent compose from map data alone?

**Question.** Given only a draft of the map (slots, contracts, shared interfaces, library facts, recipes with reference glue and probes), a draft router skill and a written stand-in for the CLI, can a fresh coding agent with no other context start an SDL3 and Dear ImGui testbed, then extend it with Jolt, reaching evidence at the probed level each time? Where does it go wrong, and is each failure a gap in the format or in the content?

**Run:** 2026-10-30, Linux x64, Clang 18.1.3, CMake 3.28, Ninja, Debug, SDL's offscreen video driver on Mesa lavapipe. The full composition, the draft map and the agent's own report (`RESULTS.md`, 27 friction entries) are archived on the branch [`spike/testbed-agent`](https://github.com/zach/automap/tree/spike/testbed-agent/spikes/testbed-agent).

**Sandbox substitution.** The sandbox's proxy returns 403 for GitHub archive downloads, so vcpkg could not fetch sources; the agent built from the same pinned commits through CMake FetchContent and git. This substitution is not part of the design.

## Outcome

| Job | Time | Evidence level |
| --- | --- | --- |
| start (SDL3 + Dear ImGui) | 4 min 50 s, cold build | probed, 14 of 14 rules |
| extend (Jolt, plus a falling-sphere demo in an ImGui window) | 11 min 43 s | probed, 22 of 22 rules, only with the software Vulkan driver's rasterizer made single-threaded (`LP_NUM_THREADS=0`); read literally, smoke |

The demo works: a sphere falls under Jolt, bounces, and its height shows in an ImGui window. Both jobs came in under v1's 15-minute target for request-to-testbed ([ADR-0063](../adr/0063-v1-success-measures.md)), the first from a cold build.

## Findings

Tags follow the agent's log: **content** (wrong or missing data or reference glue), **instructions** (CLI steps or skill), **format** (what the data model can express), **environment** (sandbox).

### 1. The map's content had build-breaking errors that one CI run would have caught (content)

- SDL 3.4 won't configure headless without `SDL_UNIX_CONSOLE_BUILD=ON`, which the facts and reference build file lacked.
- The probe harness (`probe.h`, `probe.cpp`) was listed by no contract, so following the steps exactly left it out and the build failed.
- Jolt builds without RTTI, so glue deriving from a Jolt class fails to link unless compiled with `-fno-rtti`; the facts gave no consequence for consumers.
- Probe runs wrote `imgui.ini` into the developer's repo, making runs depend on each other.
- The reference Jolt job-system adapter had a shutdown race: the adapter could be freed while a worker was between `Execute()` and `Release()`. No probe caught it.

### 2. Extend needed reference glue the map didn't carry (content)

- Nothing created the shared scheduler from the manifest's `workers` choice or named its lifecycle steps.
- Nothing showed Jolt's systems-phase setup that every user needs: `TempAllocator`, `PhysicsSystem::Init`, and the three layer interfaces. The agent wrote it from Jolt's HelloWorld sample.
- There was no build-file fragment for adding a library on extend.

### 3. Some probes gave less assurance than "probed" implies (content, format)

- SDL's "hooks before first allocation" probe can never fail: `SDL_GetNumAllocations()` returns -1 unless SDL is built with allocation counting, and the glue treated -1 as a pass. The agent proved it with a mutation (the exact mistake the recipe warns against still passed).
- Lifecycle `after` dependencies within a phase were declared but never checked.
- A cross-library ordering (`SDL_WaitForGPUIdle` before Dear ImGui's backend shutdown) could not be expressed, so it could not be checked.

### 4. The ThreadSanitizer rule was not operational (instructions, environment)

The rule didn't define the sanitizer build, how reports are detected, or how to treat reports whose every frame is inside uninstrumented system code. Every unmodified run reported two data races wholly inside Mesa's software Vulkan driver; none involved SDL, Dear ImGui, Jolt or glue. Real GPU drivers are uninstrumented too, so this is not only a sandbox problem.

### 5. Evidence could not say what was actually checked (format)

Evidence had no place for run conditions (build type, compiler version, environment variables, sanitizer options) and no result status other than pass or fail, so it couldn't record a rule as not applicable (Dear ImGui has no runtime log hook) or unverifiable (the SDL probe above).

### 6. Smaller format and instruction gaps

- Facts sat at library level, not per version.
- A manifest glue entry could name only one file, and had no way to say "not applicable".
- Recipe entry points used keys that aren't phases.
- Nothing tied manifest choices to the glue that implements them, so they can drift apart.
- The CLI stand-in left platform selection, file destinations, build type, which contracts are "in use", and evidence counting undefined.
- The skill's and CLI steps' lists of "nevers" differed.

## Implications (inference)

- The format held: no finding required a new concept. Every gap fits the existing model (facts, choices, rules, probes, recipes, evidence) with added fields or stricter content.
- Most failures were content errors that the design already intends to prevent by verifying reference glue in Automap's CI ([ADR-0016](../adr/0016-recipes-are-reference-glue-plus-probes.md), [ADR-0033](../adr/0033-automap-ci-verifies-declared-edges-latest-and-previous.md)). The spike shows what happens without that gate.
- Probes need their own verification: a probe that cannot fail is worse than none, because it produces evidence.
- Evidence needs richer results and recorded conditions for "probed" to mean exactly what was checked ([ADR-0022](../adr/0022-rules-are-probed-or-guidance.md)).
