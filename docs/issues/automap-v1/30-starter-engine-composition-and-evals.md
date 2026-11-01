# 30: Starter-engine reference composition and journey 2 evals

**What to build:** All thirteen catalog libraries form the starter-engine reference composition in Automap's CI, and journey 2's extend steps become eval tasks.

**Blocked by:** [20: The starter-upgrade eval](20-starter-upgrade-eval.md), [21: enkiTS and the jobs slot](21-enkits-and-the-jobs-slot.md), [22: mimalloc and the allocator slot](22-mimalloc-and-the-allocator-slot.md), [23: spdlog and Tracy as diagnostics backends](23-spdlog-and-tracy.md), [24: GLM and the math adapters](24-glm-and-math-adapters.md), [25: flecs](25-flecs.md), [26: miniaudio](26-miniaudio.md), [27: Assets contract, fastgltf and stb_image](27-assets-contract-fastgltf-stb-image.md), [28: SDL_shadercross and its pairing with SDL3](28-sdl-shadercross-and-pairing.md), [29: Handles, serialization and determinism; Jolt's variants through the Automap registry](29-handles-serialization-determinism-and-registry.md)

**Status:** done

- [x] The starter engine reaches probed on all three platforms in Automap's CI.
- [x] Each `starter-extend-*` eval passes: from the previous step and a one-line request for the next library, the session ends probed with no human input.

## Comments

The starter engine reaches probed on Linux, Windows and macOS in CI's `starter` group. All nine `starter-extend-*` tasks passed against map release `map-2026.10.31.1` in the Evals workflow's runs [41252513908](https://github.com/zach/automap/actions/runs/41252513908), [41252516144](https://github.com/zach/automap/actions/runs/41252516144) and [41252518014](https://github.com/zach/automap/actions/runs/41252518014), each probed with no human input, and every one's composition built and passed smoke without Automap once ejected (eject measure 9 of 9).
