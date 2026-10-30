// The testbed's executable: starts the composition, runs it for a number of
// frames (--frames <n>, default 120), and shuts it down. With
// --probe <results-file>, it also runs every probe and writes the results.
#include <cstdlib>
#include <cstring>

#include "automap/probe.h"
#include "automap/startup.h"

#if __has_include(<stub_ui.h>)
#include <stub_ui.h>
#endif

int main(int argc, char** argv) {
    int frames = 120;
    const char* results = nullptr;
    for (int i = 1; i + 1 < argc; ++i) {
        if (std::strcmp(argv[i], "--frames") == 0) frames = std::atoi(argv[i + 1]);
        if (std::strcmp(argv[i], "--probe") == 0) results = argv[i + 1];
    }
    am::Context ctx;
    if (!am::startup(ctx)) {
        if (results != nullptr) am::probe::write_results(results);
        return 1;
    }
    for (int frame = 0; frame < frames; ++frame) {
#if __has_include(<stub_ui.h>)
        stub_ui_new_frame();
#endif
    }
    if (results != nullptr) am::probe::run(am::probe::When::running);
    am::shutdown(ctx);
    if (results != nullptr) {
        am::probe::run(am::probe::When::after_shutdown);
        am::probe::write_results(results);
    }
    return 0;
}
