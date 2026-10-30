#include "stub_ui_stub_window.h"

#include <stub_ui.h>

#include "automap/lifecycle.h"

bool am_pair_stub_ui_stub_window_init(am::Context& ctx) {
    if (!stub_ui_attach(ctx.stub_window.window)) return false;
    am::lifecycle::init_step("stub-ui--stub-window:attach");
    stub_ui_enable_input();
    am::lifecycle::init_step("stub-ui--stub-window:route_input");
    return true;
}

void am_pair_stub_ui_stub_window_shutdown(am::Context&) {
    stub_ui_disable_input();
    am::lifecycle::shutdown_step("stub-ui--stub-window:unroute_input");
    stub_ui_detach();
    am::lifecycle::shutdown_step("stub-ui--stub-window:detach");
}

void am_pair_stub_ui_stub_window_event(const stub_window_event& event) {
    stub_ui_feed_event(event.id, event.x, event.y);
}
