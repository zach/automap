# 08: SDL3 and Dear ImGui in the map

**What to build:** SDL3 and Dear ImGui (docking) enter the public map with library entries, library recipes (build fragments, setup glue), their bindings, the `dear-imgui--sdl3` pairing recipe with its rules, ordering constraints and probes, and the testbed's executable source. A Linux job in Automap's CI starts the testbed on Mesa lavapipe with SDL's offscreen driver and reaches probed.

**Blocked by:** [07: Slice-1 contracts and shared interfaces](07-slice-1-contracts-and-shared-interfaces.md)

**Status:** done

- [x] `automap start --with sdl3 --with dear-imgui:docking` on a clean Linux runner ends probed.
- [x] Every probe in these recipes fails its seeded violation in CI.
- [x] Facts validate against the contracts; pins match upstream tags and vcpkg's source hashes; SDL3's default variant enables the sdl3 port's `vulkan` feature, which SDL_GPU needs for its Vulkan backend; its platform facts record its headless environment and Linux system dependencies.
- [x] SDL3's library recipe provides the window and GPU device in the context; the pairing's entry points read them (ADR-0080).
- [x] Probe mode writes no `imgui.ini`; the testbed's normal mode runs a frame loop with one Dear ImGui window until closed.
- [x] Dear ImGui's diagnostics binding is not-applicable, with the reason, in evidence.

## Comments

- Automap's CI has a public-map job (`.github/workflows/ci.yml`): vcpkg at the map's baseline, the libraries' system packages and Mesa lavapipe, then `TestTheTestbedStartsProbed`, which starts the testbed from `map/`, requires probed with Dear ImGui's diagnostics result not-applicable with its reason and no `imgui.ini` in the repo, and runs all eleven seeds the testbed carries. It is gated by `AUTOMAP_TEST_PUBLIC_MAP=1`, so the CLI's own tests stay on the fixture map (ADR-0077).
- SDL3's default variant turns the port's default features off and enables `vulkan`, `x11` and `wayland`: the default `ibus` feature pulls in D-Bus and more system packages for IME support a testbed doesn't need. Its platform facts list the Linux system packages, including the autotools vcpkg needs for libxcrypt.
- Every probed composition rule now has a seed in the public map itself: `lifecycle.init-order` in Dear ImGui's allocation binding, `lifecycle.shutdown-reverse` in the pairing, `build-flags.abi-defines-match` in Dear ImGui's library recipe (`ImDrawIdx`), `build-flags.language-settings-agree` in SDL3's library recipe, and `platform.targets-supported` in its contract.
- Probes take the composition's context; the harness records crash-safe failures (Dear ImGui's ABI check asserts on a mismatch) and checks the C++ standard itself.
