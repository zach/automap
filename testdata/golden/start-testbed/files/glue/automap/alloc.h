// Automap allocation facade, contract allocation 1.0.0 (fixture map).
// Every library's allocator hooks route through it, tagged with the
// library's id. MIT-0.
#pragma once
#include <cstddef>
#include <cstdint>

namespace am::alloc {

void* allocate(const char* library, std::size_t size);
void release(const char* library, void* block);

// Allocations the facade has served for a library.
std::uint64_t routed_count(const char* library);

// Probe builds only: whether a library released a block the facade never
// allocated, which means its hooks were installed too late.
bool released_foreign_block(const char* library);

} // namespace am::alloc
