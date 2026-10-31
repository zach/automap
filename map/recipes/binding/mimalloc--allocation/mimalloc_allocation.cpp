// mimalloc's binding to the allocation contract: mimalloc in the allocator
// slot serves every block the allocation facade hands out.
// SPDX-License-Identifier: MIT-0
#include <cstddef>

#include <mimalloc.h>

void* am_allocator_slot_allocate(std::size_t size, std::size_t alignment) { return mi_malloc_aligned(size, alignment); }

void* am_allocator_slot_reallocate(void* block, std::size_t size) { return mi_realloc(block, size); }

void am_allocator_slot_release(void* block) { mi_free(block); }
