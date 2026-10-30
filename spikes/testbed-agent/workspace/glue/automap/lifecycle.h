// Automap shared interface for the lifecycle contract 0.1.0 (DRAFT, spike).
#pragma once
#include <string>

namespace am::lifecycle {

enum class Phase : int { memory = 1, diagnostics = 2, platform = 3, jobs = 4, systems = 5, game = 6 };

const char* phase_name(Phase p);

// Glue calls these as each library's init and shutdown steps run.
void init_step(Phase phase, const char* library, const char* step);
void shutdown_step(Phase phase, const char* library, const char* step);

// Used by probes. On failure, detail names the offending pair of steps.
bool init_order_ok(std::string* detail);
bool shutdown_order_ok(std::string* detail);

} // namespace am::lifecycle
