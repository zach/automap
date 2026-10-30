#include "stub_window_setup.h"

#include <stub_window.h>

#include "automap/lifecycle.h"

bool am_stub_window_setup_init(am::Context& ctx) {
    if (!stub_window_init()) return false;
    am::lifecycle::init_step("stub-window:init");
    ctx.stub_window.window = stub_window_open("automap", 640, 480);
    if (ctx.stub_window.window == nullptr) return false;
    am::lifecycle::init_step("stub-window:open_window");
    return true;
}

void am_stub_window_setup_shutdown(am::Context& ctx) {
    stub_window_close(ctx.stub_window.window);
    ctx.stub_window.window = nullptr;
    am::lifecycle::shutdown_step("stub-window:close_window");
    stub_window_shutdown();
    am::lifecycle::shutdown_step("stub-window:shutdown");
}
