#pragma once
#include "automap/context.h"

// Creates stub-ui's context. Runs in the systems phase.
bool am_stub_ui_setup_init(am::Context& ctx);

// Destroys stub-ui's context.
void am_stub_ui_setup_shutdown(am::Context& ctx);
