// Automap shared interface for the diagnostics contract 0.1.0 (DRAFT, spike).
// Backend for the "stderr" diagnostics choice.
#include "diag.h"

#include <cstdio>
#include <map>
#include <mutex>
#include <string>

namespace am::diag {
namespace {
std::mutex g_mutex;
std::map<std::string, std::uint64_t>& counts() { static std::map<std::string, std::uint64_t> m; return m; }
const char* level_name(Level l) {
    switch (l) {
    case Level::trace: return "trace";
    case Level::info: return "info";
    case Level::warn: return "warn";
    default: return "error";
    }
}
} // namespace

void log(Level level, const char* source, const char* message) {
    std::lock_guard<std::mutex> lock(g_mutex);
    ++counts()[source];
    std::fprintf(stderr, "[%s] %s: %s\n", source, level_name(level), message);
}

bool assert_failed(const char* source, const char* expression, const char* message, const char* file, int line) {
    std::lock_guard<std::mutex> lock(g_mutex);
    ++counts()[source];
    std::fprintf(stderr, "[%s] assert failed: %s (%s) at %s:%d\n", source, expression, message ? message : "", file, line);
    return false;
}

std::uint64_t message_count(const char* source) {
    std::lock_guard<std::mutex> lock(g_mutex);
    auto it = counts().find(source);
    return it == counts().end() ? 0 : it->second;
}

} // namespace am::diag
