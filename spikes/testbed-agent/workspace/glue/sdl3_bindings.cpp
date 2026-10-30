// Reference glue: SDL3 bindings to the allocation and diagnostics contracts (DRAFT, spike).
#include "sdl3_bindings.h"

#include <SDL3/SDL.h>

#include "automap/alloc.h"
#include "automap/diag.h"
#include "automap/lifecycle.h"
#include "automap/probe.h"

using am::alloc::Client;
using am::lifecycle::Phase;

namespace {
int g_allocations_at_install = -1;

void* SDLCALL am_malloc(size_t size) { return am::alloc::allocate(Client::sdl3, size); }
void* SDLCALL am_calloc(size_t count, size_t size) { return am::alloc::allocate_zeroed(Client::sdl3, count, size); }
void* SDLCALL am_realloc(void* block, size_t size) { return am::alloc::reallocate(Client::sdl3, block, size); }
void SDLCALL am_free(void* block) { am::alloc::release(Client::sdl3, block); }

void SDLCALL am_log(void*, int, SDL_LogPriority priority, const char* message) {
    am::diag::Level level = priority >= SDL_LOG_PRIORITY_ERROR ? am::diag::Level::error
                          : priority >= SDL_LOG_PRIORITY_WARN  ? am::diag::Level::warn
                          : priority >= SDL_LOG_PRIORITY_INFO  ? am::diag::Level::info
                                                               : am::diag::Level::trace;
    am::diag::log(level, "sdl3", message);
}

SDL_AssertState SDLCALL am_assert(const SDL_AssertData* data, void*) {
    bool should_break = am::diag::assert_failed("sdl3", data->condition, nullptr, data->filename, data->linenum);
    return should_break ? SDL_ASSERTION_BREAK : SDL_ASSERTION_IGNORE;
}
} // namespace

void sdl3_bind_allocation() {
    g_allocations_at_install = SDL_GetNumAllocations();
    SDL_SetMemoryFunctions(am_malloc, am_calloc, am_realloc, am_free);
    am::lifecycle::init_step(Phase::memory, "sdl3", "SDL_SetMemoryFunctions");
}

void sdl3_bind_diagnostics() {
    SDL_SetLogOutputFunction(am_log, nullptr);
    SDL_SetAssertionHandler(am_assert, nullptr);
    am::lifecycle::init_step(Phase::diagnostics, "sdl3", "SDL_SetLogOutputFunction+SDL_SetAssertionHandler");
}

void sdl3_register_probes() {
    am::probe::check("allocation.hooks-before-first-allocation", "sdl3", g_allocations_at_install <= 0, "ALLOC_HOOK_LATE(sdl3)");

    SDL_malloc_func m = nullptr; SDL_calloc_func c = nullptr; SDL_realloc_func r = nullptr; SDL_free_func f = nullptr;
    SDL_GetMemoryFunctions(&m, &c, &r, &f);
    bool routed = m == am_malloc && c == am_calloc && r == am_realloc && f == am_free && am::alloc::routed_count(Client::sdl3) > 0;
    am::probe::check("allocation.routed-through-facade", "sdl3", routed, "ALLOC_NOT_ROUTED(sdl3)");

    SDL_LogOutputFunction log_fn = nullptr; void* log_ud = nullptr;
    SDL_GetLogOutputFunction(&log_fn, &log_ud);
    am::probe::check("diagnostics.log-routed", "sdl3", log_fn == am_log, "DIAG_NOT_ROUTED(sdl3)");

    void* assert_ud = nullptr;
    am::probe::check("diagnostics.assert-routed", "sdl3", SDL_GetAssertionHandler(&assert_ud) == am_assert, "ASSERT_NOT_ROUTED(sdl3)");

    int linked = SDL_GetVersion();
    bool versions_ok = linked == SDL_VERSION && SDL_VERSIONNUM_MINOR(linked) % 2 == 0;
    am::probe::check("build-flags.abi-match", "sdl3", versions_ok, "DEFINE_MISMATCH(sdl3)");
}
