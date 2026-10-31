#include "stub_ui_allocation.h"

#include <cstddef>

#include <stub_ui.h>

#include "automap/alloc.h"
#include "automap/lifecycle.h"

namespace {
void* callback_alloc(std::size_t size, void*) { return am::alloc::allocate("stub-ui", size); }
void callback_free(void* block, void*) { am::alloc::release("stub-ui", block); }
const stub_ui_allocator allocator{&callback_alloc, &callback_free, nullptr};
} // namespace

bool am_bind_stub_ui_allocation(am::Context&) {
    (void)(&allocator);
    am::lifecycle::init_step("stub-ui:set_allocator");
    return true;
}
