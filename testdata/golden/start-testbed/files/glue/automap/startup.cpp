// Written by automap from the composition's recipes. Edit the manifest, not this file.
#include "automap/startup.h"

#include <cstdio>

bool am_bind_stub_ui_allocation(am::Context& ctx);
bool am_bind_stub_window_allocation(am::Context& ctx);
bool am_bind_stub_ui_diagnostics(am::Context& ctx);
bool am_bind_stub_window_diagnostics(am::Context& ctx);
bool am_stub_window_setup_init(am::Context& ctx);
bool am_stub_ui_setup_init(am::Context& ctx);
bool am_pair_stub_ui_stub_window_init(am::Context& ctx);
void am_pair_stub_ui_stub_window_shutdown(am::Context& ctx);
void am_stub_ui_setup_shutdown(am::Context& ctx);
void am_stub_window_setup_shutdown(am::Context& ctx);

namespace am {

namespace {
bool init_failed(const char* call) {
    std::fprintf(stderr, "INIT_FAILED(%s)\n", call);
    return false;
}
} // namespace

bool startup(Context& ctx) {
    // memory
    if (!am_bind_stub_ui_allocation(ctx)) return init_failed("am_bind_stub_ui_allocation");
    if (!am_bind_stub_window_allocation(ctx)) return init_failed("am_bind_stub_window_allocation");
    // diagnostics
    if (!am_bind_stub_ui_diagnostics(ctx)) return init_failed("am_bind_stub_ui_diagnostics");
    if (!am_bind_stub_window_diagnostics(ctx)) return init_failed("am_bind_stub_window_diagnostics");
    // platform
    if (!am_stub_window_setup_init(ctx)) return init_failed("am_stub_window_setup_init");
    // systems
    if (!am_stub_ui_setup_init(ctx)) return init_failed("am_stub_ui_setup_init");
    if (!am_pair_stub_ui_stub_window_init(ctx)) return init_failed("am_pair_stub_ui_stub_window_init");
    return true;
}

void shutdown(Context& ctx) {
    // systems
    am_pair_stub_ui_stub_window_shutdown(ctx);
    am_stub_ui_setup_shutdown(ctx);
    // platform
    am_stub_window_setup_shutdown(ctx);
}

} // namespace am
