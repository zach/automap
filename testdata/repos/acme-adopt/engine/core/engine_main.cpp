// Acme's engine: its own startup, which brings up its libraries in its own
// order, runs frames, and shuts down. With --probe <results-file>, it also
// runs every probe and writes the results.
// SPDX-License-Identifier: MIT-0
#include <cstdlib>
#include <cstdio>
#include <cstring>

#include <acme_jobs.h>
#include <acme_renderer.h>
#include <stub_physics.h>

#include "automap/context.h"
#include "automap/probe.h"

bool am_bind_stub_window_allocation(am::Context& ctx);
bool am_bind_stub_window_diagnostics(am::Context& ctx);
bool am_stub_window_setup_init(am::Context& ctx);
bool am_jobs_setup_init(am::Context& ctx);
bool am_bind_stub_physics_jobs(am::Context& ctx);
bool am_stub_physics_setup_init(am::Context& ctx);
void am_stub_physics_setup_shutdown(am::Context& ctx);
void am_jobs_setup_shutdown(am::Context& ctx);
void am_stub_window_setup_shutdown(am::Context& ctx);

namespace {

bool engine_init(am::Context& ctx) {
    // Memory and diagnostics first, then the window, the scheduler, and
    // physics on the scheduler.
    return am_bind_stub_window_allocation(ctx) && am_bind_stub_window_diagnostics(ctx) &&
           am_stub_window_setup_init(ctx) && am_jobs_setup_init(ctx) &&
           am_bind_stub_physics_jobs(ctx) && am_stub_physics_setup_init(ctx);
}

void engine_shutdown(am::Context& ctx) {
    am_stub_physics_setup_shutdown(ctx);
    am_jobs_setup_shutdown(ctx);
    am_stub_window_setup_shutdown(ctx);
}

} // namespace

int main(int argc, char** argv) {
    int frames = 120;
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
    for (int frame = 0; frame < frames; ++frame) {
        stub_physics_step();
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
