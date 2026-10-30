#include "stub_ui_setup.h"

#include <stub_ui.h>

#include "automap/lifecycle.h"
#include "automap/probe.h"

bool am_stub_ui_setup_init(am::Context&) {
#ifdef AUTOMAP_PROBE
    am::probe::check("build-flags.abi-defines-match", "stub-ui", stub_ui_abi_check(STUB_UI_ID_BITS),
                     "ABI_DEFINE_MISMATCH(STUB_UI_WIDE_IDS)");
#endif
    if (!stub_ui_create_context()) return false;
    am::lifecycle::init_step("stub-ui:create_context");
    return true;
}

void am_stub_ui_setup_shutdown(am::Context&) {
    stub_ui_destroy_context();
    am::lifecycle::shutdown_step("stub-ui:destroy_context");
}
