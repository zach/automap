# 08: SDL3 and Dear ImGui in the map

**What to build:** SDL3 and Dear ImGui (docking) enter the public map with library entries, library recipes (build fragments, setup glue), their bindings, the `dear-imgui--sdl3` pairing recipe with its rules, ordering constraints and probes, and the testbed entry point. A Linux job in Automap's CI starts the testbed on Mesa lavapipe with SDL's offscreen driver and reaches probed.

**Blocked by:** [07: Slice-1 contracts and shared interfaces](07-slice-1-contracts-and-shared-interfaces.md)

**Status:** ready-for-agent

- [ ] `automap start --with sdl3 --with dear-imgui:docking` on a clean Linux runner ends probed.
- [ ] Every probe in these recipes fails its seeded violation in CI.
- [ ] Facts validate against the contracts; pins match upstream tags and vcpkg's source hashes; SDL's headless Linux build options are in its platform facts.
- [ ] Probe mode writes no `imgui.ini`; the testbed's normal mode runs a frame loop with one Dear ImGui window until closed.
- [ ] Dear ImGui's diagnostics binding is not-applicable, with the reason, in evidence.
