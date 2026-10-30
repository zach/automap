#pragma once
#include "automap/context.h"

// Installs stub-window's allocator hooks so every block it allocates goes
// through the allocation facade. Runs in the memory phase.
bool am_bind_stub_window_allocation(am::Context& ctx);
