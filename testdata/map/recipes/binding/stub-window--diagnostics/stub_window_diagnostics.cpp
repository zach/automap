#include <stub_window.h>

#include "automap/context.h"
#include "automap/diag.h"
#include "automap/lifecycle.h"

namespace {
void forward_log(const char* message) { am::diag::log(am::diag::Level::info, "stub-window", message); }
} // namespace

bool am_bind_stub_window_diagnostics(am::Context&) {
    stub_window_set_log(&forward_log);
    am::lifecycle::init_step("stub-window:set_log");
    return true;
}
