// The dear-imgui--sdl3 pairing.
// SPDX-License-Identifier: MIT-0
#include "imgui_sdl3_gpu.h"

#include <SDL3/SDL.h>
#include <imgui.h>
#include <imgui_impl_sdl3.h>
#include <imgui_impl_sdlgpu3.h>

#include "automap/lifecycle.h"

bool am_pair_imgui_sdl3_init(am::Context& ctx) {
    if (!ImGui_ImplSDL3_InitForSDLGPU(ctx.sdl3.window)) return false;
    am::lifecycle::init_step("dear-imgui--sdl3:ImGui_ImplSDL3_InitForSDLGPU");
    ImGui_ImplSDLGPU3_InitInfo info = {};
    info.Device = ctx.sdl3.gpu_device;
    info.ColorTargetFormat = SDL_GetGPUSwapchainTextureFormat(ctx.sdl3.gpu_device, ctx.sdl3.window);
    info.MSAASamples = SDL_GPU_SAMPLECOUNT_1;
    if (!ImGui_ImplSDLGPU3_Init(&info)) return false;
    am::lifecycle::init_step("dear-imgui--sdl3:ImGui_ImplSDLGPU3_Init");
    return true;
}

void am_pair_imgui_sdl3_shutdown(am::Context& ctx) {
    // The renderer backend releases GPU resources the device may still be using.
    SDL_WaitForGPUIdle(ctx.sdl3.gpu_device);
    am::lifecycle::shutdown_step("dear-imgui--sdl3:SDL_WaitForGPUIdle");
    ImGui_ImplSDL3_Shutdown();
    am::lifecycle::shutdown_step("dear-imgui--sdl3:ImGui_ImplSDL3_Shutdown");
    ImGui_ImplSDLGPU3_Shutdown();
    am::lifecycle::shutdown_step("dear-imgui--sdl3:ImGui_ImplSDLGPU3_Shutdown");
}

void am_pair_imgui_sdl3_process_event(const SDL_Event& event) {
    ImGui_ImplSDL3_ProcessEvent(&event);
}

void am_pair_imgui_sdl3_new_frame() {
    ImGui_ImplSDLGPU3_NewFrame();
    ImGui_ImplSDL3_NewFrame();
    ImGui::NewFrame();
}

bool am_pair_imgui_sdl3_prepare(SDL_GPUCommandBuffer* commands) {
    ImGui::Render();
    ImDrawData* draw_data = ImGui::GetDrawData();
    if (draw_data->DisplaySize.x <= 0.0f || draw_data->DisplaySize.y <= 0.0f) return false; // minimized
    // Upload vertex and index buffers before the render pass begins.
    ImGui_ImplSDLGPU3_PrepareDrawData(draw_data, commands);
    return true;
}

void am_pair_imgui_sdl3_draw(SDL_GPUCommandBuffer* commands, SDL_GPURenderPass* pass) {
    ImGui_ImplSDLGPU3_RenderDrawData(ImGui::GetDrawData(), commands, pass);
}

void am_pair_imgui_sdl3_render(SDL_GPUCommandBuffer* commands, SDL_GPUTexture* target) {
    if (!am_pair_imgui_sdl3_prepare(commands) || target == nullptr) return;
    SDL_GPUColorTargetInfo color = {};
    color.texture = target;
    color.load_op = SDL_GPU_LOADOP_LOAD;
    color.store_op = SDL_GPU_STOREOP_STORE;
    SDL_GPURenderPass* pass = SDL_BeginGPURenderPass(commands, &color, 1, nullptr);
    am_pair_imgui_sdl3_draw(commands, pass);
    SDL_EndGPURenderPass(pass);
}

bool am_pair_imgui_sdl3_begin_frame(am::Context&) {
    bool open = true;
    SDL_Event event;
    while (SDL_PollEvent(&event)) {
        am_pair_imgui_sdl3_process_event(event);
        if (event.type == SDL_EVENT_QUIT || event.type == SDL_EVENT_WINDOW_CLOSE_REQUESTED) open = false;
    }
    am_pair_imgui_sdl3_new_frame();
    return open;
}

void am_pair_imgui_sdl3_end_frame(am::Context& ctx) {
    SDL_GPUCommandBuffer* commands = SDL_AcquireGPUCommandBuffer(ctx.sdl3.gpu_device);
    SDL_GPUTexture* swapchain = nullptr;
    SDL_WaitAndAcquireGPUSwapchainTexture(commands, ctx.sdl3.window, &swapchain, nullptr, nullptr);
    bool drawing = am_pair_imgui_sdl3_prepare(commands);
    if (swapchain != nullptr && drawing) {
        SDL_GPUColorTargetInfo target = {};
        target.texture = swapchain;
        target.clear_color = SDL_FColor{0.10f, 0.12f, 0.15f, 1.0f};
        target.load_op = SDL_GPU_LOADOP_CLEAR;
        target.store_op = SDL_GPU_STOREOP_STORE;
        SDL_GPURenderPass* pass = SDL_BeginGPURenderPass(commands, &target, 1, nullptr);
        am_pair_imgui_sdl3_draw(commands, pass);
        SDL_EndGPURenderPass(pass);
    }
    SDL_SubmitGPUCommandBuffer(commands);
}
