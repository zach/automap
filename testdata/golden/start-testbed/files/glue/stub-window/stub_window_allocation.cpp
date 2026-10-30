#include "stub_window_allocation.h"

#include <cstddef>

#include <stub_window.h>

#include "automap/alloc.h"
#include "automap/lifecycle.h"

namespace {
void* trampoline_alloc(std::size_t size) { return am::alloc::allocate("stub-window", size); }
void trampoline_free(void* block) { am::alloc::release("stub-window", block); }
} // namespace

bool am_bind_stub_window_allocation(am::Context&) {
    stub_window_set_allocator(&trampoline_alloc, &trampoline_free);
    am::lifecycle::init_step("stub-window:set_allocator");
    return true;
}
