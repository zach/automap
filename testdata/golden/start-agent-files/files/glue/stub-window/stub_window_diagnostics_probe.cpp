#ifdef AUTOMAP_PROBE
#include <stub_window.h>

#include "automap/diag.h"
#include "automap/probe.h"

// diagnostics.routed-through-facade: a message stub-window logs reaches the
// diagnostics facade.
namespace {
void am_probe_stub_window_diagnostics(am::Context&) {
    auto before = am::diag::message_count("stub-window");
    stub_window_log_test_message();
    bool routed = am::diag::message_count("stub-window") > before;
    am::probe::check("diagnostics.routed-through-facade", "stub-window", routed, "DIAG_NOT_ROUTED(stub-window)");
}
const am::probe::Registration registration("stub-window--diagnostics", &am_probe_stub_window_diagnostics);
} // namespace
#endif
