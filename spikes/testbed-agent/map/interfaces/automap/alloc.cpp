// Automap shared interface for the allocation contract 0.1.0 (DRAFT, spike).
// Backend for the "system" allocator choice.
#include "alloc.h"

#include <atomic>
#include <cstdlib>
#include <cstring>

namespace am::alloc {
namespace {
std::atomic<std::uint64_t> g_counts[static_cast<int>(Client::count)];
void note(Client c) { g_counts[static_cast<int>(c)].fetch_add(1, std::memory_order_relaxed); }
} // namespace

const char* client_name(Client c) {
    switch (c) {
    case Client::sdl3: return "sdl3";
    case Client::dear_imgui: return "dear-imgui";
    case Client::jolt: return "jolt";
    default: return "other";
    }
}

void* allocate(Client c, std::size_t size) { note(c); return std::malloc(size); }

void* allocate_zeroed(Client c, std::size_t count, std::size_t size) { note(c); return std::calloc(count, size); }

void* reallocate(Client c, void* block, std::size_t new_size) { note(c); return std::realloc(block, new_size); }

void release(Client, void* block) { std::free(block); }

void* allocate_aligned(Client c, std::size_t size, std::size_t alignment) {
    note(c);
#if defined(_WIN32)
    return _aligned_malloc(size, alignment);
#else
    void* p = nullptr;
    if (alignment < sizeof(void*)) alignment = sizeof(void*);
    return posix_memalign(&p, alignment, size) == 0 ? p : nullptr;
#endif
}

void release_aligned(Client, void* block) {
#if defined(_WIN32)
    _aligned_free(block);
#else
    std::free(block);
#endif
}

std::uint64_t routed_count(Client c) { return g_counts[static_cast<int>(c)].load(std::memory_order_relaxed); }

} // namespace am::alloc
