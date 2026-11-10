# 11: Map facts match what vcpkg builds

**What to build:** SDL3's default variant states the features vcpkg builds, and Automap's CI compares every reference composition's install plan with its variants' stated features. Field notes [14].

**Blocked by:** None

**Status:** done

- [x] SDL3's default variant builds what it says: IBus is off, because Dear ImGui comes from an Automap registry port that asks for SDL3 without its default features, and `vulkan` is on every platform, as SDL_shadercross's port asks.
- [x] A test in the public-map CI job runs `vcpkg install --dry-run` for each reference composition and for each of its libraries alone, and fails when a library's port gets a feature in the composition that it doesn't get alone.
- [x] The check passes, and fails when Dear ImGui goes back to vcpkg's port or SDL3's variant limits `vulkan` to Linux.
- [x] CI runs it on every push, on all three platforms, in the public-map job's `core` group.

## Comments

- vcpkg builds a port with the union of the features every port in the plan asks of it, and a top-level `default-features: false` doesn't override another port's request. vcpkg's `imgui[sdl3-binding]` depends on `sdl3` with its defaults, so every composition with Dear ImGui built SDL3 with IBus, and with it D-Bus, libsystemd and libxcrypt: fourteen more ports on Linux, autotools on the machine, and a `pip install` during libsystemd's build, which fails offline.
- Stating IBus in SDL3's variant was tried first. It made the fact true and made every SDL3 composition pay for it: `fetch` and then an offline `verify` of SDL3 with Jolt failed in CI on libsystemd's `pip install`. The Automap registry now carries Dear ImGui's port with its SDL3 dependencies asking for no default features, and SDL3's variant is lean again: the day-one composition plans 4 ports where it planned 18.
- `TestALibrarysPortBuildsTheSameFeaturesInEveryReferenceComposition` compares dry-run install plans, which need the registries and no downloads. On Windows and macOS it also showed SDL_shadercross's port switching on SDL3's `vulkan` feature, which the variant now states for every platform.
- SDL3's system packages no longer include autotools, which only libxcrypt's build needed.
- SDL_shadercross's note about DXC at run time is ticket 15's.
