// Reference glue: the Dear ImGui and SDL3 pairing, using Dear ImGui's SDL3 platform backend and SDL_GPU renderer backend (DRAFT, spike).
#pragma once

union SDL_Event;
struct SDL_Window;
struct SDL_GPUDevice;

bool imgui_sdl3_gpu_init(SDL_Window* window, SDL_GPUDevice* device); // systems phase, after the GPU device claims the window
void imgui_sdl3_gpu_process_event(const SDL_Event* event);
void imgui_sdl3_gpu_new_frame();
void imgui_sdl3_gpu_render(SDL_Window* window, SDL_GPUDevice* device, const float clear_color[4]);
void imgui_sdl3_gpu_shutdown(); // systems phase, after SDL_WaitForGPUIdle
void imgui_sdl3_gpu_register_probes();
