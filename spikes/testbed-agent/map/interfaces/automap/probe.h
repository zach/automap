// Automap probe harness (DRAFT, spike). Probes record results; report() prints one JSON line.
#pragma once
#include <string>

namespace am::probe {

// rule: the contract rule id; library: who the result is about; failure_code: reported when !ok.
void check(const std::string& rule, const std::string& library, bool ok, const std::string& failure_code);

// Prints {"verdict": "pass"|"fail", "results": [...]} to stdout. Returns 0 if every check passed.
int report();

} // namespace am::probe
