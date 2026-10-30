#pragma once
#include <stub_window.h>

#include "automap/context.h"

// Attaches stub-ui to the window in stub-window.window, then routes its
// input. Runs in the systems phase, after stub-ui's context exists.
bool am_pair_stub_ui_stub_window_init(am::Context& ctx);

// Unroutes stub-ui's input, then detaches it, before the window closes.
void am_pair_stub_ui_stub_window_shutdown(am::Context& ctx);

// Passes one polled stub-window event to stub-ui. The game's event loop calls
// this for every event, before starting the next UI frame.
void am_pair_stub_ui_stub_window_event(const stub_window_event& event);
