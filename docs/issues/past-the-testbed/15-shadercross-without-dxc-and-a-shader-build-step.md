# 15: SDL_shadercross without DXC at run time, and a shader build step

**What to build:** The default SDL_shadercross variant doesn't link DXC, so a game ships no `dxcompiler.dll`; the build fragment offers a CMake function that compiles HLSL to SPIR-V at build time. Field notes [35, 42].

**Blocked by:** 11

**Status:** ready-for-agent

- [ ] The Automap registry's port links DXC into the library only with its `dxc` feature; the default variant builds without it on every platform, and `SDL3_shadercross.dll` on Windows imports no `dxcompiler.dll` (checked in CI).
- [ ] The pairing's probe still passes on all three platforms with DXBC on Direct3D 12, SPIR-V on Vulkan and MSL on Metal.
- [ ] A `dxc` variant keeps HLSL input and DXIL for the developer who wants them, with the note that `dxcompiler.dll` and `dxil.dll` ship.
- [ ] `automap_compile_shaders(<target> <stage>:<file>...)` in the build fragment compiles HLSL to SPIR-V headers with the `dxc` tool vcpkg installs, where vcpkg has it (Windows, x64 Linux); the recipe's notes say what to do on macOS.
- [ ] A public-map CI test builds a composition that uses the function.
- [ ] The recipe notes and the library's facts state what is built.
