#pragma once
#include "automap/context.h"

// Installs stub-ui's allocator callbacks so every block it allocates goes
// through the allocation facade. Runs in the memory phase.
bool am_bind_stub_ui_allocation(am::Context& ctx);
