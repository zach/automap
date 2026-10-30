# 30: Starter-engine reference composition and journey 2 evals

**What to build:** All thirteen catalog libraries form the starter-engine reference composition in Automap's CI, and journey 2's extend steps become eval tasks.

**Blocked by:** [20: The starter-upgrade eval](20-starter-upgrade-eval.md), [21: enkiTS and the jobs slot](21-enkits-and-the-jobs-slot.md), [22: mimalloc and the allocator slot](22-mimalloc-and-the-allocator-slot.md), [23: spdlog and Tracy as diagnostics backends](23-spdlog-and-tracy.md), [24: GLM and the math adapters](24-glm-and-math-adapters.md), [25: flecs](25-flecs.md), [26: miniaudio](26-miniaudio.md), [27: Assets contract, fastgltf and stb_image](27-assets-contract-fastgltf-stb-image.md), [28: SDL_shadercross and its pairing with SDL3](28-sdl-shadercross-and-pairing.md), [29: Handles, serialization and determinism; Jolt's variants through the Automap registry](29-handles-serialization-determinism-and-registry.md)

**Status:** ready-for-agent

- [ ] The starter engine reaches probed on all three platforms in Automap's CI.
- [ ] Each `starter-extend-*` eval passes: from the previous step and a one-line request for the next library, the session ends probed with no human input.
