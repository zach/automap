// Reference glue: the testbed composition's entry point (DRAFT, spike).
// Game phase code. `testbed --probe` runs a fixed number of frames headless and prints the probe report.
#include <cstdlib>
#include <cstring>
#include <string>

#include <SDL3/SDL.h>
#include <SDL3/SDL_main.h>
#include <imgui.h>

#include "automap/lifecycle.h"
#include "automap/probe.h"
#include "imgui_bindings.h"
#include "imgui_sdl3_gpu.h"
#include "sdl3_bindings.h"

using am::lifecycle::Phase;

int main(int argc, char** argv) {
    bool probe_mode = false;
    for (int i = 1; i < argc; ++i)
        if (std::strcmp(argv[i], "--probe") == 0) probe_mode = true;
    const int frame_limit = probe_mode ? 120 : -1;
    if (probe_mode) setenv("SDL_VIDEO_DRIVER", "offscreen", 1); // not SDL_SetHint: that would allocate before the memory phase

    // memory
    sdl3_bind_allocation();
    imgui_bind_allocation();

    // diagnostics
    sdl3_bind_diagnostics();

    // platform
    if (!SDL_Init(SDL_INIT_VIDEO)) { SDL_Log("SDL_Init failed: %s", SDL_GetError()); return 2; }
    am::lifecycle::init_step(Phase::platform, "sdl3", "SDL_Init");
    SDL_Window* window = SDL_CreateWindow("Automap testbed", 1280, 720, SDL_WINDOW_RESIZABLE);
    if (!window) { SDL_Log("SDL_CreateWindow failed: %s", SDL_GetError()); return 2; }
    am::lifecycle::init_step(Phase::platform, "sdl3", "SDL_CreateWindow");
    SDL_GPUDevice* device = SDL_CreateGPUDevice(SDL_GPU_SHADERFORMAT_SPIRV | SDL_GPU_SHADERFORMAT_DXIL | SDL_GPU_SHADERFORMAT_MSL |
                                                    SDL_GPU_SHADERFORMAT_METALLIB,
                                                false, nullptr);
    if (!device || !SDL_ClaimWindowForGPUDevice(device, window)) { SDL_Log("GPU setup failed: %s", SDL_GetError()); return 2; }
    am::lifecycle::init_step(Phase::platform, "sdl3", "SDL_CreateGPUDevice+SDL_ClaimWindowForGPUDevice");

    // systems
    imgui_sdl3_gpu_init(window, device);

    // game
    const float clear_color[4] = {0.10f, 0.12f, 0.15f, 1.0f};
    bool running = true;
    for (int frame = 0; running && (frame_limit < 0 || frame < frame_limit); ++frame) {
        SDL_Event event;
        while (SDL_PollEvent(&event)) {
            imgui_sdl3_gpu_process_event(&event);
            if (event.type == SDL_EVENT_QUIT) running = false;
        }
        imgui_sdl3_gpu_new_frame();
        ImGui::Begin("Automap testbed");
        ImGui::Text("Frame %d, GPU driver: %s", frame, SDL_GetGPUDeviceDriver(device));
        ImGui::End();
        imgui_sdl3_gpu_render(window, device, clear_color);
    }

    if (probe_mode) {
        sdl3_register_probes();
        imgui_register_probes();
        imgui_sdl3_gpu_register_probes();
    }

    // shutdown, in reverse phase order
    SDL_WaitForGPUIdle(device);
    imgui_sdl3_gpu_shutdown();
    SDL_ReleaseWindowFromGPUDevice(device, window);
    SDL_DestroyGPUDevice(device);
    am::lifecycle::shutdown_step(Phase::platform, "sdl3", "SDL_DestroyGPUDevice");
    SDL_DestroyWindow(window);
    SDL_Quit();
    am::lifecycle::shutdown_step(Phase::platform, "sdl3", "SDL_Quit");

    if (!probe_mode) return 0;
    std::string detail;
    am::probe::check("lifecycle.init-order", "composition", am::lifecycle::init_order_ok(&detail), "INIT_ORDER(" + detail + ")");
    detail.clear();
    am::probe::check("lifecycle.shutdown-reverse", "composition", am::lifecycle::shutdown_order_ok(&detail), "SHUTDOWN_ORDER(" + detail + ")");
    return am::probe::report();
}
