# 08: The router skill past the testbed

**What to build:** The router skill covers what comes after the testbed: making it a game, release builds and shipping, shaders, private maps, changing a composition, restricted networks and signed evidence from a workflow of one's own. Field notes [13, 15, 25, 35, 38, 50].

**Blocked by:** 04, 07

**Status:** done

- [x] `SKILL.md` routes to two guides the CLI writes beside it, `game.md` and `private-maps.md`, and refreshes with it when a newer CLI runs.
- [x] `game.md` says: rename the executable in the manifest and replace the testbed's `CMakeLists.txt` and sources; the executable must honour `--frames <n>` and `--probe <file>` (with the code that does it); it is run from the build root, so assets are copied beside it; how to configure a release build with the CLI's vcpkg; that the DLLs vcpkg copies beside the executable are what ships; where shaders come from.
- [x] `private-maps.md` says: `map.json` and the manifest entry; start from `automap show recipe` of a similar public recipe; a library, a library recipe with a build fragment, a binding with a probe and a seeded violation; commit before composing; `automap upgrade <library>` after a change.
- [x] `SKILL.md` says how to change a composition (the commands that exist), that a blocked download is answered with `automap vendor` on an open network, and what a workflow of one's own needs for signed evidence.
- [x] The `AGENTS.md` section's build line works without `VCPKG_ROOT` set: it points at the configure command on the first line of `verify`'s build log, which names the CLI's own vcpkg.
- [x] SPEC §5.7 describes the skill's files.

## Comments

- `internal/agentfiles` embeds three templates; `Files` and `StaleSkills` write and refresh all three in every skill directory, `extend` and `adopt` merge them, `eject` removes the directory, and the eval runner installs all three.
- The guides state what the CLI does at this ticket. Tickets 09 (window settings, the Dear ImGui frame in pieces), 13 (`remove`, `resolve`), 14 (uncommitted private maps) and 15 (the shader build function) each update the sentences they change.
