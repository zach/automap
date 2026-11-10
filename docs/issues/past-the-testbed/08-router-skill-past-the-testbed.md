# 08: The router skill past the testbed

**What to build:** The router skill covers what comes after the testbed: making it a game, release builds and shipping, shaders, private maps, changing a composition, restricted networks and signed evidence from a workflow of one's own. Field notes [13, 15, 25, 35, 38, 50].

**Blocked by:** 04, 07

**Status:** ready-for-agent

- [ ] `SKILL.md` routes to two guides the CLI writes beside it, `game.md` and `private-maps.md`, and refreshes with it when a newer CLI runs.
- [ ] `game.md` says: rename the executable in the manifest and replace the testbed's `CMakeLists.txt` and sources; the executable must honour `--frames <n>` and `--probe <file>` (with the code that does it); it is run from the build root, so assets are copied beside it; how to configure a release build with the CLI's vcpkg; that the DLLs vcpkg copies beside the executable are what ships; where shaders come from.
- [ ] `private-maps.md` says: `map.json` and the manifest entry; start from `automap show recipe` of a similar public recipe; a library, a library recipe with a build fragment, a binding with a probe and a seeded violation; commit before composing; `automap upgrade <library>` after a change.
- [ ] `SKILL.md` says how to change a composition (the commands that exist), that a blocked download is answered with `automap vendor` on an open network, and what a workflow of one's own needs for signed evidence.
- [ ] The `AGENTS.md` section's build line works without `VCPKG_ROOT` set: it names where the CLI's own vcpkg is.
- [ ] SPEC §5.7 describes the skill's files.
