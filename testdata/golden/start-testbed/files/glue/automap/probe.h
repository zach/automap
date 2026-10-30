// Automap probe harness (fixture map). Probe glue registers its probes;
// the executable runs them with --probe and the harness writes every
// result, with the recorded lifecycle steps, to the results file the CLI
// reads. MIT-0.
#pragma once
#include <string>

namespace am::probe {

// rule: the rule id; library: who the result is about ("" for a pairing or
// the composition); failure_code: reported when !ok.
void check(const std::string& rule, const std::string& library, bool ok, const std::string& failure_code);

using Probe = void (*)();

// When a registered probe runs: after startup and the frames, or after
// shutdown (for checks that need the libraries to have released
// everything).
enum class When { running, after_shutdown };

// Registers a probe. Probe glue declares one static Registration per probe.
struct Registration {
    Registration(const char* recipe, Probe probe, When when = When::running);
};

// Runs every probe registered for when.
void run(When when);

// Writes every recorded result and lifecycle step as JSON to path. Returns
// whether every check passed.
bool write_results(const std::string& path);

} // namespace am::probe
