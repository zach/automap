#pragma once
#include "automap/context.h"

// Initializes stub-window and opens one window, stored in the context as
// stub-window.window. Runs in the platform phase.
bool am_stub_window_setup_init(am::Context& ctx);

// Closes the window and shuts stub-window down.
void am_stub_window_setup_shutdown(am::Context& ctx);
