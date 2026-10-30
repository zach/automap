// Automap probe harness (fixture map). Probes record one result per rule
// and library; the harness writes them with the recorded lifecycle steps
// to the results file the CLI reads. MIT-0.
#pragma once
#include <string>

namespace am::probe {

// rule: the rule id; library: who the result is about ("" for a pairing or
// the composition); failure_code: reported when !ok.
void check(const std::string& rule, const std::string& library, bool ok, const std::string& failure_code);

// Writes every recorded result and lifecycle step as JSON to path. Returns
// whether every check passed.
bool write_results(const std::string& path);

} // namespace am::probe
