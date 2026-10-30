// Automap probe harness (DRAFT, spike).
#include "probe.h"

#include <cstdio>
#include <vector>

namespace am::probe {
namespace {
struct Result { std::string rule, library; bool ok; std::string failure_code; };
std::vector<Result> g_results;
} // namespace

void check(const std::string& rule, const std::string& library, bool ok, const std::string& failure_code) {
    g_results.push_back({rule, library, ok, ok ? std::string() : failure_code});
}

int report() {
    bool all = !g_results.empty();
    for (auto& r : g_results) all = all && r.ok;
    std::printf("{\"verdict\":\"%s\",\"results\":[", all ? "pass" : "fail");
    for (size_t i = 0; i < g_results.size(); ++i) {
        auto& r = g_results[i];
        std::printf("%s{\"rule\":\"%s\",\"library\":\"%s\",\"ok\":%s,\"failure_code\":%s%s%s}", i ? "," : "",
                    r.rule.c_str(), r.library.c_str(), r.ok ? "true" : "false",
                    r.ok ? "null" : "\"", r.ok ? "" : r.failure_code.c_str(), r.ok ? "" : "\"");
    }
    std::printf("]}\n");
    return all ? 0 : 1;
}

} // namespace am::probe
