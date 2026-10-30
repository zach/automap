// The testbed's executable: starts the composition, runs it for a number of
// frames (--frames <n>, default 120), and shuts it down.
#include <cstdlib>
#include <cstring>

#include "automap/startup.h"

#if __has_include(<stub_ui.h>)
#include <stub_ui.h>
#endif

int main(int argc, char** argv) {
    int frames = 120;
    for (int i = 1; i + 1 < argc; ++i) {
        if (std::strcmp(argv[i], "--frames") == 0) frames = std::atoi(argv[i + 1]);
    }
    am::Context ctx;
    if (!am::startup(ctx)) return 1;
    for (int frame = 0; frame < frames; ++frame) {
#if __has_include(<stub_ui.h>)
        stub_ui_new_frame();
#endif
    }
    am::shutdown(ctx);
    return 0;
}
