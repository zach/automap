// The testbed's executable: starts the composition, runs it, shuts it down.
#include "automap/startup.h"

int main() {
    am::Context ctx;
    if (!am::startup(ctx)) return 1;
    am::shutdown(ctx);
    return 0;
}
