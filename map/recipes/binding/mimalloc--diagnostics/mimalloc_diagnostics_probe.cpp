// Probes for mimalloc's binding to the diagnostics contract.
// SPDX-License-Identifier: MIT-0
#ifdef AUTOMAP_PROBE
#include <mimalloc.h>

#include "automap/diag.h"
#include "automap/probe.h"

namespace {

// diagnostics.routed-through-facade: mimalloc's output reaches the facade.
// mimalloc has no query for its output function, so the probe has it print
// its options and counts what the facade receives.
void am_probe_mimalloc_diagnostics(am::Context&) {
    auto before = am::diag::message_count("mimalloc");
    mi_options_print();
    bool routed = am::diag::message_count("mimalloc") > before;
    am::probe::check("diagnostics.routed-through-facade", "mimalloc", routed, "DIAG_NOT_ROUTED(mimalloc)");
}
const am::probe::Registration registration("mimalloc--diagnostics", &am_probe_mimalloc_diagnostics);

} // namespace
#endif
