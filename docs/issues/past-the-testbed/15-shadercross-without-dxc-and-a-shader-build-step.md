# 15: SDL_shadercross without DXC at run time, and a shader build step

**What to build:** The default SDL_shadercross variant doesn't link DXC, so a game ships no `dxcompiler.dll`; the build fragment offers a CMake function that compiles HLSL to SPIR-V at build time. Field notes [35, 42].

**Blocked by:** 11

**Status:** ready-for-agent

- [ ] The Automap registry's port links DXC into the library only with its `dxc` feature; the default variant builds without it on every platform, and `SDL3_shadercross.dll` on Windows imports no `dxcompiler.dll` (checked in CI).
- [ ] The pairing's probe still passes on all three platforms with DXBC on Direct3D 12, SPIR-V on Vulkan and MSL on Metal.
- [ ] The port's `dxc` feature keeps HLSL input and DXIL for the developer who wants them, through a private variant; the recipe's notes say so, and that `dxcompiler` and `dxil` then ship.
- [ ] `automap_compile_shaders(<target> <stage>:<file>...)` in the build fragment compiles HLSL to SPIR-V files beside the executable with the `dxc` tool vcpkg installs, where vcpkg has it (Windows, x64 Linux); on macOS it uses a `dxc` on the `PATH` or copies a committed `.spv`.
- [ ] A public-map CI test builds a composition that uses the function.
- [ ] The recipe notes and the library's facts state what is built.

## Comments

- SDL_shadercross's source loads `d3dcompiler_47.dll` at run time for DXBC whether or not it is built with DXC; DXC is only called for HLSL input and DXIL output, which Automap's glue never asks for. Built with DXC, the library links `dxcompiler` on every platform, which is why the DLL had to ship.
- No `dxc` variant is in the map: nobody has asked for HLSL at run time, a variant would add jobs on three platforms, one of which can't build it, and a private variant with the port's feature serves the developer who wants it.
- The port lands on `main` on its own first (port version 2, with the map still pinning version 1), so that `Publish registry` puts it on the registry branch before the map change's CI builds it: this session may not push to the registry branch itself.
