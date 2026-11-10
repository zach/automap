# 11: Map facts match what vcpkg builds

**What to build:** SDL3's default variant states the features vcpkg builds, and Automap's CI compares every reference composition's install plan with its variants' stated features. Field notes [14].

**Blocked by:** None

**Status:** done

- [x] SDL3's default variant names `ibus` among its Linux features, its description says IBus is on and why, and the system dependencies still hold.
- [x] A test in the public-map CI job runs `vcpkg install --dry-run` for each reference composition and for each of its libraries alone, and fails when a library's port gets a feature in the composition that it doesn't get alone.
- [x] The check passes on `main` and fails when SDL3's variant drops `ibus`.
- [x] CI runs it on every push, on all three platforms, in the public-map job's `core` group.

## Comments

- vcpkg builds a port with the union of the features every port in the plan asks of it, and a top-level `default-features: false` doesn't override another port's request: `imgui[sdl3-binding]` depends on `sdl3` with its defaults, so every composition with Dear ImGui built SDL3 with IBus and D-Bus while the variant said IBus was off. A variant can't promise a feature stays off, so SDL3's variant now states the port's Linux defaults plus `vulkan`.
- `TestALibrarysPortBuildsTheSameFeaturesInEveryReferenceComposition` compares dry-run install plans, which need the registries and no downloads; before the fix it failed for SDL3 in six reference compositions and for nothing else.
- SDL_shadercross's note about DXC at run time is ticket 15's.
