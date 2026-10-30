// Reference glue: the Dear ImGui and SDL3 pairing (DRAFT, spike).
#include "imgui_sdl3_gpu.h"

#include <SDL3/SDL.h>
#include <imgui.h>
#include <imgui_impl_sdl3.h>
#include <imgui_impl_sdlgpu3.h>

#include "automap/lifecycle.h"
#include "automap/probe.h"

using am::lifecycle::Phase;

namespace {
bool g_init_ok = false;
int g_frames_presented = 0;
} // namespace

bool imgui_sdl3_gpu_init(SDL_Window* window, SDL_GPUDevice* device) {
    IMGUI_CHECKVERSION();
    ImGui::CreateContext();
    ImGui::GetIO().ConfigFlags |= ImGuiConfigFlags_NavEnableKeyboard | ImGuiConfigFlags_DockingEnable;
    am::lifecycle::init_step(Phase::systems, "dear-imgui", "ImGui::CreateContext");

    bool platform_ok = ImGui_ImplSDL3_InitForSDLGPU(window);
    ImGui_ImplSDLGPU3_InitInfo info = {};
    info.Device = device;
    info.ColorTargetFormat = SDL_GetGPUSwapchainTextureFormat(device, window);
    info.MSAASamples = SDL_GPU_SAMPLECOUNT_1;
    bool renderer_ok = ImGui_ImplSDLGPU3_Init(&info);
    am::lifecycle::init_step(Phase::systems, "dear-imgui+sdl3", "ImGui_ImplSDL3_InitForSDLGPU+ImGui_ImplSDLGPU3_Init");

    g_init_ok = platform_ok && renderer_ok;
    return g_init_ok;
}

void imgui_sdl3_gpu_process_event(const SDL_Event* event) { ImGui_ImplSDL3_ProcessEvent(event); }

void imgui_sdl3_gpu_new_frame() {
    ImGui_ImplSDLGPU3_NewFrame();
    ImGui_ImplSDL3_NewFrame();
    ImGui::NewFrame();
}

void imgui_sdl3_gpu_render(SDL_Window* window, SDL_GPUDevice* device, const float clear_color[4]) {
    ImGui::Render();
    ImDrawData* draw_data = ImGui::GetDrawData();
    const bool minimized = draw_data->DisplaySize.x <= 0.0f || draw_data->DisplaySize.y <= 0.0f;

    SDL_GPUCommandBuffer* command_buffer = SDL_AcquireGPUCommandBuffer(device);
    SDL_GPUTexture* swapchain = nullptr;
    SDL_WaitAndAcquireGPUSwapchainTexture(command_buffer, window, &swapchain, nullptr, nullptr);

    if (swapchain != nullptr && !minimized) {
        // Mandatory: upload vertex and index buffers before the render pass begins.
        ImGui_ImplSDLGPU3_PrepareDrawData(draw_data, command_buffer);

        SDL_GPUColorTargetInfo target = {};
        target.texture = swapchain;
        target.clear_color = SDL_FColor{clear_color[0], clear_color[1], clear_color[2], clear_color[3]};
        target.load_op = SDL_GPU_LOADOP_CLEAR;
        target.store_op = SDL_GPU_STOREOP_STORE;

        SDL_GPURenderPass* pass = SDL_BeginGPURenderPass(command_buffer, &target, 1, nullptr);
        ImGui_ImplSDLGPU3_RenderDrawData(draw_data, command_buffer, pass);
        SDL_EndGPURenderPass(pass);
        ++g_frames_presented;
    }
    SDL_SubmitGPUCommandBuffer(command_buffer);
}

void imgui_sdl3_gpu_shutdown() {
    ImGui_ImplSDL3_Shutdown();
    ImGui_ImplSDLGPU3_Shutdown();
    am::lifecycle::shutdown_step(Phase::systems, "dear-imgui+sdl3", "ImGui_ImplSDL3_Shutdown+ImGui_ImplSDLGPU3_Shutdown");
    ImGui::DestroyContext();
    am::lifecycle::shutdown_step(Phase::systems, "dear-imgui", "ImGui::DestroyContext");
}

void imgui_sdl3_gpu_register_probes() {
    am::probe::check("pairing.dear-imgui-sdl3.backends-initialized", "dear-imgui+sdl3", g_init_ok, "PAIRING_INIT_FAILED(dear-imgui, sdl3)");
    am::probe::check("pairing.dear-imgui-sdl3.frames-presented", "dear-imgui+sdl3", g_frames_presented > 0, "NO_FRAMES_PRESENTED(dear-imgui, sdl3)");
}
