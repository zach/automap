// mimalloc's binding to the diagnostics contract: its statistics, warnings
// and errors go to the facade.
// SPDX-License-Identifier: MIT-0
#include <cstddef>
#include <cstring>

#include <mimalloc.h>

#include "automap/context.h"
#include "automap/diag.h"
#include "automap/lifecycle.h"

namespace {

// The line mimalloc is writing on this thread: it writes a prefix, such as
// "mimalloc: warning: ", and the message separately.
thread_local char line[512];
thread_local std::size_t length = 0;

} // namespace

// Collects mimalloc's output and logs each line, at the level its prefix
// names.
void am_mimalloc_output(const char* text, void*) {
    for (; *text != '\0'; ++text) {
        if (*text != '\n') {
            if (length < sizeof line - 1) line[length++] = *text;
            continue;
        }
        line[length] = '\0';
        if (length > 0) {
            am::diag::Level level = am::diag::Level::info;
            if (std::strstr(line, "error:") != nullptr) level = am::diag::Level::error;
            else if (std::strstr(line, "warning:") != nullptr) level = am::diag::Level::warn;
            am::diag::log(level, "mimalloc", line);
        }
        length = 0;
    }
}

// Routes mimalloc's output through the diagnostics facade. Runs in the
// diagnostics phase.
bool am_bind_mimalloc_diagnostics(am::Context&) {
    mi_register_output(&am_mimalloc_output, nullptr);
    am::lifecycle::init_step("mimalloc:mi_register_output");
    return true;
}
