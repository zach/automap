// Automap shared interface for the allocation contract 0.1.0 (DRAFT, spike).
// Copied verbatim into the composition's glue directory; the developer owns it.
#pragma once
#include <cstddef>
#include <cstdint>

namespace am::alloc {

// One client per library that routes allocations through the facade.
enum class Client : int { sdl3, dear_imgui, jolt, other, count };

const char* client_name(Client c);

void* allocate(Client c, std::size_t size);
void* allocate_zeroed(Client c, std::size_t count, std::size_t size);
void* reallocate(Client c, void* block, std::size_t new_size);
void release(Client c, void* block);

void* allocate_aligned(Client c, std::size_t size, std::size_t alignment);
void release_aligned(Client c, void* block);

// Number of allocations the facade has served for this client.
std::uint64_t routed_count(Client c);

} // namespace am::alloc
