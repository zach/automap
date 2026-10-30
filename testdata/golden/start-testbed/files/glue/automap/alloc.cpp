#include "automap/alloc.h"

#include <cstdlib>
#include <map>
#include <mutex>
#include <set>
#include <string>

namespace am::alloc {
namespace {

struct State {
    std::mutex mutex;
    std::map<std::string, std::uint64_t> routed;
    std::set<std::string> foreign;
#ifdef AUTOMAP_PROBE
    std::set<void*> live;
#endif
};

State& state() {
    static State* s = new State; // never destroyed: libraries may free during exit
    return *s;
}

} // namespace

void* allocate(const char* library, std::size_t size) {
    void* block = std::malloc(size == 0 ? 1 : size);
    State& s = state();
    std::lock_guard<std::mutex> lock(s.mutex);
    ++s.routed[library];
#ifdef AUTOMAP_PROBE
    s.live.insert(block);
#endif
    return block;
}

void release(const char* library, void* block) {
    if (block == nullptr) return;
    {
        State& s = state();
        std::lock_guard<std::mutex> lock(s.mutex);
#ifdef AUTOMAP_PROBE
        if (s.live.erase(block) == 0) s.foreign.insert(library);
#else
        (void)library;
#endif
    }
    std::free(block);
}

std::uint64_t routed_count(const char* library) {
    State& s = state();
    std::lock_guard<std::mutex> lock(s.mutex);
    auto it = s.routed.find(library);
    return it == s.routed.end() ? 0 : it->second;
}

bool released_foreign_block(const char* library) {
    State& s = state();
    std::lock_guard<std::mutex> lock(s.mutex);
    return s.foreign.count(library) != 0;
}

} // namespace am::alloc
