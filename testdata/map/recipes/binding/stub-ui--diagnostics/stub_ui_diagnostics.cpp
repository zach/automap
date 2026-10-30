#include <stub_ui.h>

#include "automap/context.h"
#include "automap/diag.h"
#include "automap/lifecycle.h"

namespace {
bool forward_assert(const char* expression, const char* file, int line) {
    return am::diag::assert_failed("stub-ui", expression, nullptr, file, line);
}
} // namespace

bool am_bind_stub_ui_diagnostics(am::Context&) {
    stub_ui_set_assert_handler(&forward_assert);
    am::lifecycle::init_step("stub-ui:set_assert_handler");
    return true;
}
