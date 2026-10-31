// Acme's engine: its own startup, which brings up its libraries in its own
// order, runs frames, and shuts down. --frames <n> stops after n frames;
// --probe <results-file> also runs every probe and writes the results.
// SPDX-License-Identifier: MIT-0
#include <cstdio>
#include <cstdlib>
#include <cstring>

#include <SDL3/SDL.h>
#include <Jolt/Jolt.h>

#include <Jolt/Physics/PhysicsSystem.h>
#include <acme_jobs.h>
#include <acme_renderer.h>

#include "automap/context.h"
#include "automap/probe.h"

bool am_bind_jolt_allocation(am::Context& ctx);
bool am_bind_sdl3_allocation(am::Context& ctx);
bool am_bind_jolt_diagnostics(am::Context& ctx);
bool am_bind_sdl3_diagnostics(am::Context& ctx);
bool am_sdl3_setup_init(am::Context& ctx);
bool am_jobs_setup_init(am::Context& ctx);
bool am_bind_jolt_jobs(am::Context& ctx);
bool am_jolt_setup_init(am::Context& ctx);
void am_jolt_setup_shutdown(am::Context& ctx);
void am_unbind_jolt_jobs(am::Context& ctx);
void am_jobs_setup_shutdown(am::Context& ctx);
void am_sdl3_setup_shutdown(am::Context& ctx);

namespace {

bool engine_init(am::Context& ctx) {
    // Memory and diagnostics first, then the window, the scheduler, and
    // physics on the scheduler.
    return am_bind_jolt_allocation(ctx) && am_bind_sdl3_allocation(ctx) &&
           am_bind_jolt_diagnostics(ctx) && am_bind_sdl3_diagnostics(ctx) &&
           am_sdl3_setup_init(ctx) && am_jobs_setup_init(ctx) && am_bind_jolt_jobs(ctx) &&
           am_jolt_setup_init(ctx);
}

void engine_shutdown(am::Context& ctx) {
    am_jolt_setup_shutdown(ctx);
    am_unbind_jolt_jobs(ctx);
    am_jobs_setup_shutdown(ctx);
    am_sdl3_setup_shutdown(ctx);
}

} // namespace

int main(int argc, char** argv) {
    int frames = -1;
    const char* results = nullptr;
    for (int i = 1; i + 1 < argc; ++i) {
        if (std::strcmp(argv[i], "--frames") == 0) frames = std::atoi(argv[i + 1]);
        if (std::strcmp(argv[i], "--probe") == 0) results = argv[i + 1];
    }
    if (results != nullptr) am::probe::begin(results);
    am::Context ctx;
    if (!engine_init(ctx)) {
        std::fprintf(stderr, "acme: engine init failed\n");
        am::probe::write_results();
        return 1;
    }
    bool open = true;
    for (int frame = 0; open && (frames < 0 || frame < frames); ++frame) {
        SDL_Event event;
        while (SDL_PollEvent(&event)) {
            if (event.type == SDL_EVENT_QUIT || event.type == SDL_EVENT_WINDOW_CLOSE_REQUESTED) open = false;
        }
        ctx.jolt.physics_system->Update(1.0f / 60.0f, 1, ctx.jolt.temp_allocator, ctx.jolt.job_system);
    }
    std::printf("acme: jobs %d, renderer %d\n", acme::jobs::version(), acme::renderer::version());
    if (results != nullptr) am::probe::run(am::probe::When::running, ctx);
    engine_shutdown(ctx);
    if (results != nullptr) {
        am::probe::run(am::probe::When::after_shutdown, ctx);
        am::probe::write_results();
    }
    return 0;
}
