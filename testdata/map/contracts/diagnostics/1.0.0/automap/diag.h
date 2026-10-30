// Automap diagnostics facade, contract diagnostics 1.0.0 (fixture map).
// Every library's log and assert hooks route through it. MIT-0.
#pragma once
#include <cstdint>

namespace am::diag {

enum class Level { trace, info, warn, error };

void log(Level level, const char* library, const char* message);

// Returns whether the caller should break into the debugger.
bool assert_failed(const char* library, const char* expression, const char* message, const char* file, int line);

// Log messages and assertion failures received from a library.
std::uint64_t message_count(const char* library);

} // namespace am::diag
