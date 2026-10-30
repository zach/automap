#ifdef AUTOMAP_PROBE
#include "automap/choices.h"
#include "automap/probe.h"

// build-flags.language-settings-agree: the glue compiles with the C++
// standard the composition chooses.
namespace {
void am_probe_stub_window_language() {
    bool agree = __cplusplus / 100 % 100 == AUTOMAP_BUILD_FLAGS_CXX_STANDARD % 100;
    am::probe::check("build-flags.language-settings-agree", "", agree, "LANG_SETTINGS_MISMATCH(cxx_standard)");
}
const am::probe::Registration registration("stub-window", &am_probe_stub_window_language);
} // namespace
#endif
