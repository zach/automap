#include <stub_physics.h>

#include "automap/context.h"
#include "automap/lifecycle.h"

// Creates stub-physics's world. Runs in the systems phase.
bool am_stub_physics_setup_init(am::Context&) {
    if (!stub_physics_create_world(4096)) return false;
    am::lifecycle::init_step("stub-physics:create_world");
    return true;
}

// Destroys the world.
void am_stub_physics_setup_shutdown(am::Context&) {
    stub_physics_destroy_world();
    am::lifecycle::shutdown_step("stub-physics:destroy_world");
}
