#include "automap/diag.h"

#include <cstdio>
#include <map>
#include <mutex>
#include <string>

#include "automap/choices.h"

namespace am::diag {
namespace {

std::mutex mutex;
std::map<std::string, std::uint64_t> counts;

const char* level_name(Level level) {
    switch (level) {
    case Level::trace: return "trace";
    case Level::info: return "info";
    case Level::warn: return "warn";
    case Level::error: return "error";
    }
    return "?";
}

} // namespace

void log(Level level, const char* library, const char* message) {
    std::lock_guard<std::mutex> lock(mutex);
    ++counts[library];
#ifdef AUTOMAP_DIAGNOSTICS_LOG_STDERR
    std::fprintf(stderr, "[%s] %s: %s\n", level_name(level), library, message ? message : "");
#endif
}

bool assert_failed(const char* library, const char* expression, const char* message, const char* file, int line) {
    std::lock_guard<std::mutex> lock(mutex);
    ++counts[library];
    std::fprintf(stderr, "[assert] %s: %s %s (%s:%d)\n", library, expression ? expression : "", message ? message : "",
                 file ? file : "?", line);
    return false;
}

std::uint64_t message_count(const char* library) {
    std::lock_guard<std::mutex> lock(mutex);
    auto it = counts.find(library);
    return it == counts.end() ? 0 : it->second;
}

} // namespace am::diag
