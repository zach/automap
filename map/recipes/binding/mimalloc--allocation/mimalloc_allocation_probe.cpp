// Probes for mimalloc's binding to the allocation contract.
// SPDX-License-Identifier: MIT-0
#ifdef AUTOMAP_PROBE
#include <atomic>
#include <cstddef>
#include <cstdint>

#include <mimalloc.h>

#include "automap/alloc.h"
#include "automap/probe.h"

extern std::atomic<std::uint64_t> am_mimalloc_served;

namespace {

// allocation.routed-through-facade: blocks the facade serves, plain,
// zeroed, reallocated and aligned, are mimalloc's: mimalloc served each one
// and its heap holds it.
void am_probe_mimalloc_allocation(am::Context&) {
    std::uint64_t before = am_mimalloc_served.load();
    void* plain = am::alloc::allocate("mimalloc-probe", 24);
    void* zeroed = am::alloc::allocate_zeroed("mimalloc-probe", 4, 8);
    void* moved = am::alloc::reallocate("mimalloc-probe", am::alloc::allocate("mimalloc-probe", 16), 4096);
    void* aligned = am::alloc::allocate_aligned("mimalloc-probe", 100, 64);
    bool served = am_mimalloc_served.load() - before >= 5; // four allocations and a reallocation
    bool held = mi_is_in_heap_region(plain) && mi_is_in_heap_region(zeroed) && mi_is_in_heap_region(moved) &&
                mi_is_in_heap_region(aligned);
    am::alloc::release("mimalloc-probe", plain);
    am::alloc::release("mimalloc-probe", zeroed);
    am::alloc::release("mimalloc-probe", moved);
    am::alloc::release_aligned("mimalloc-probe", aligned);
    am::probe::check("allocation.routed-through-facade", "mimalloc", served && held, "ALLOC_NOT_ROUTED(mimalloc)");
}
const am::probe::Registration registration("mimalloc--allocation", &am_probe_mimalloc_allocation);

} // namespace
#endif
