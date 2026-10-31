# 28: SDL_shadercross and its pairing with SDL3

**What to build:** SDL_shadercross, pinned by commit, fills the shader-compiler slot, and the `sdl-shadercross--sdl3` pairing checks that each platform's shaders are in a format its SDL_GPU backend accepts and that resource counts come from reflection.

**Blocked by:** [08: SDL3 and Dear ImGui in the map](08-sdl3-and-dear-imgui-in-the-map.md), [27: Assets contract, fastgltf and stb_image](27-assets-contract-fastgltf-stb-image.md), [29: Handles, serialization and determinism; Jolt's variants through the Automap registry](29-handles-serialization-determinism-and-registry.md) (the curated sdl3-shadercross port depends on directx-dxc, which vcpkg doesn't support on macOS; the Automap registry carries a port that needs DXC only where it builds)

**Status:** done

- [x] A composition with SDL3 and SDL_shadercross reaches probed on all three platforms.
- [x] Seeded wrong-format output and hard-coded resource counts each fail their pairing rule.
