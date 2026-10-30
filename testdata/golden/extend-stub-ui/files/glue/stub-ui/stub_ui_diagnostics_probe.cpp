#ifdef AUTOMAP_PROBE
#include <stub_ui.h>

#include "automap/diag.h"
#include "automap/probe.h"

// diagnostics.routed-through-facade: a failed stub-ui assertion reaches the
// diagnostics facade.
namespace {
void am_probe_stub_ui_diagnostics(am::Context&) {
    auto before = am::diag::message_count("stub-ui");
    stub_ui_fail_test_assertion();
    bool routed = am::diag::message_count("stub-ui") > before;
    am::probe::check("diagnostics.routed-through-facade", "stub-ui", routed, "DIAG_NOT_ROUTED(stub-ui)");
}
const am::probe::Registration registration("stub-ui--diagnostics", &am_probe_stub_ui_diagnostics);
} // namespace
#endif
