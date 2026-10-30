// Automap shared interface for the lifecycle contract 0.1.0 (DRAFT, spike).
#include "lifecycle.h"

#include <vector>

namespace am::lifecycle {
namespace {
struct Step { Phase phase; std::string name; };
std::vector<Step> g_init, g_shutdown;

bool ordered(const std::vector<Step>& steps, bool ascending, std::string* detail) {
    for (size_t i = 1; i < steps.size(); ++i) {
        bool bad = ascending ? steps[i].phase < steps[i - 1].phase : steps[i].phase > steps[i - 1].phase;
        if (bad) {
            if (detail) *detail = steps[i - 1].name + " before " + steps[i].name;
            return false;
        }
    }
    return true;
}
} // namespace

const char* phase_name(Phase p) {
    switch (p) {
    case Phase::memory: return "memory";
    case Phase::diagnostics: return "diagnostics";
    case Phase::platform: return "platform";
    case Phase::jobs: return "jobs";
    case Phase::systems: return "systems";
    default: return "game";
    }
}

void init_step(Phase phase, const char* library, const char* step) {
    g_init.push_back({phase, std::string(library) + ":" + step});
}

void shutdown_step(Phase phase, const char* library, const char* step) {
    g_shutdown.push_back({phase, std::string(library) + ":" + step});
}

bool init_order_ok(std::string* detail) { return ordered(g_init, true, detail); }
bool shutdown_order_ok(std::string* detail) { return ordered(g_shutdown, false, detail); }

} // namespace am::lifecycle
