#ifdef AUTOMAP_PROBE
#include <stub_ui.h>
#include <stub_window.h>

#include "automap/probe.h"
#include "stub_ui_stub_window.h"

// stub-ui--stub-window.input-forwarded: a synthetic event pushed into
// stub-window reaches stub-ui within one frame.
namespace {
void am_probe_stub_ui_stub_window_input(am::Context&) {
    stub_window_push_event(stub_window_event{7, 12, 34});
    stub_window_event event;
    while (stub_window_poll(&event)) am_pair_stub_ui_stub_window_event(event);
    stub_ui_new_frame();
    bool seen = stub_ui_last_event_id() == 7;
    am::probe::check("stub-ui--stub-window.input-forwarded", "", seen, "INPUT_NOT_FORWARDED(stub-ui--stub-window)");
}
const am::probe::Registration registration("stub-ui--stub-window", &am_probe_stub_ui_stub_window_input);
} // namespace
#endif
