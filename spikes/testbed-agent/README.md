# PROTOTYPE — throwaway spike: can an agent compose from map data alone?

**Status: throwaway.** This directory exists only to answer one question for the Automap exploration. It is deleted or archived once the result is recorded in an ADR.

## The question

Given only a draft of Automap's map data (slots, contracts, shared interfaces, library facts, recipes with reference glue and probes) and a draft router skill, can a fresh coding agent with no other context:

1. **start** a testbed composition (SDL3 + Dear ImGui), then
2. **extend** it with Jolt,

so that it builds, runs headless, and passes every probed rule? Where does it go wrong, and is each failure a gap in the *format* (what the map can express) or in the *content* (a wrong fact or broken reference glue)?

## Layout

| Path | What it is |
| --- | --- |
| `map/` | Draft public map: `FORMAT.md`, `slots.json`, `contracts/`, `interfaces/`, `libraries/`, `recipes/`. Read-only for the agent under test. |
| `skill/SKILL.md` | Draft router skill. |
| `CLI-STEPS.md` | What the CLI would do for each job. No CLI exists, so the agent performs these steps by hand. |
| `workspace/` | Where the agent under test builds the composition (created by the agent). |
| `RESULTS.md` | The agent's friction log and probe results (written by the agent). |

## Sandbox substitutions (not part of the design)

- **Build backend:** this sandbox's proxy returns 403 for GitHub archive downloads, so vcpkg can't fetch sources. The spike uses CMake `FetchContent` from each library's git repository at the pinned tag and commit instead. The map still records the vcpkg port each library would use.
- **Platform:** Linux x64 only, with GCC 13 or Clang 18. Headless runs use SDL's `offscreen` video driver and Mesa's software Vulkan driver (lavapipe), already installed (`VK_DRIVER_FILES=/usr/share/vulkan/icd.d/lvp_icd.json`).
- **Evidence:** written unsigned, since there is no CI identity here.
