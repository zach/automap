// mimalloc's binding to the allocation contract: mimalloc in the allocator
// slot serves every block the allocation facade hands out.
// SPDX-License-Identifier: MIT-0
#include <atomic>
#include <cstddef>
#include <cstdint>

#include <mimalloc.h>

#ifdef AUTOMAP_PROBE
// Probe builds only: blocks mimalloc has served the facade.
std::atomic<std::uint64_t> am_mimalloc_served{0};
#define AM_MIMALLOC_SERVED() am_mimalloc_served.fetch_add(1, std::memory_order_relaxed)
#else
#define AM_MIMALLOC_SERVED() (void)0
#endif

void* am_allocator_slot_allocate(std::size_t size, std::size_t alignment) {
    AM_MIMALLOC_SERVED();
    return mi_malloc_aligned(size, alignment);
}

void* am_allocator_slot_reallocate(void* block, std::size_t size) {
    AM_MIMALLOC_SERVED();
    return mi_realloc(block, size);
}

void am_allocator_slot_release(void* block) { mi_free(block); }
