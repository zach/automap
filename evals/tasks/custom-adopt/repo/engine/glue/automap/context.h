// Acme's engine context: what its glue shares between entry points.
#pragma once

namespace JPH { class JobSystem; }
namespace JPH { class PhysicsSystem; }
namespace JPH { class TempAllocator; }
struct SDL_GPUDevice;
struct SDL_Window;

namespace am {

// The context every entry point takes: the fields recipes provide and read.
struct Context {
    struct {
        ::SDL_Window* window = nullptr;
        ::SDL_GPUDevice* gpu_device = nullptr;
    } sdl3;
    struct {
        ::JPH::JobSystem* job_system = nullptr;
        ::JPH::PhysicsSystem* physics_system = nullptr;
        ::JPH::TempAllocator* temp_allocator = nullptr;
    } jolt;
};

} // namespace am
