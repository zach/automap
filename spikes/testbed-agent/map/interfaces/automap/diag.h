// Automap shared interface for the diagnostics contract 0.1.0 (DRAFT, spike).
#pragma once
#include <cstdint>

namespace am::diag {

enum class Level { trace, info, warn, error };

void log(Level level, const char* source, const char* message);

// Returns true if the caller should break into the debugger.
bool assert_failed(const char* source, const char* expression, const char* message, const char* file, int line);

// Number of log messages and assertion failures received from a source.
std::uint64_t message_count(const char* source);

} // namespace am::diag
