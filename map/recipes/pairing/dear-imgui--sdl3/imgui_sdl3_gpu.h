// The dear-imgui--sdl3 pairing: Dear ImGui's SDL3 platform backend and
// SDL_GPU renderer backend.
// SPDX-License-Identifier: MIT-0
#pragma once
#include "automap/context.h"

union SDL_Event;
struct SDL_GPUCommandBuffer;
struct SDL_GPURenderPass;
struct SDL_GPUTexture;

// Initializes both backends for sdl3.window and sdl3.gpu_device. Runs in the
// systems phase, after the ImGui context exists and SDL has claimed the
// window for the device.
bool am_pair_imgui_sdl3_init(am::Context& ctx);

// Waits for the GPU to go idle, then shuts both backends down, before the
// ImGui context and the GPU device are destroyed.
void am_pair_imgui_sdl3_shutdown(am::Context& ctx);

// The frame in pieces, for a game that polls its own events and owns its
// render passes:
//
//     SDL_Event event;
//     while (SDL_PollEvent(&event)) {
//         am_pair_imgui_sdl3_process_event(event);
//         // ... the game's own handling ...
//     }
//     am_pair_imgui_sdl3_new_frame();
//     // ... ImGui:: calls ...
//     // ... the game's render passes into `target` ...
//     am_pair_imgui_sdl3_render(commands, target);

// Passes one SDL event to ImGui.
void am_pair_imgui_sdl3_process_event(const SDL_Event& event);

// Begins an ImGui frame on both backends. Draw with ImGui after it.
void am_pair_imgui_sdl3_new_frame();

// Ends the ImGui frame and draws it over what `target` holds, in a render
// pass of its own on `commands`. Call it after the game's passes into
// `target`, before the command buffer is submitted. Draws nothing when the
// window is minimized.
void am_pair_imgui_sdl3_render(SDL_GPUCommandBuffer* commands, SDL_GPUTexture* target);

// For a game that draws ImGui inside one of its own render passes instead:
// am_pair_imgui_sdl3_prepare() ends the ImGui frame and uploads its buffers,
// and must run before that pass begins; it returns false when there is
// nothing to draw. am_pair_imgui_sdl3_draw() then draws into the pass.
bool am_pair_imgui_sdl3_prepare(SDL_GPUCommandBuffer* commands);
void am_pair_imgui_sdl3_draw(SDL_GPUCommandBuffer* commands, SDL_GPURenderPass* pass);

// The whole frame, for a testbed with nothing else to draw: begin_frame
// polls every SDL event, passes each to ImGui and begins an ImGui frame,
// returning false once the window is closed; end_frame clears the window's
// swapchain, draws ImGui into it and submits.
bool am_pair_imgui_sdl3_begin_frame(am::Context& ctx);
void am_pair_imgui_sdl3_end_frame(am::Context& ctx);
