#include "automap/probe.h"

#include <cstdio>
#include <vector>

#include "automap/lifecycle.h"

namespace am::probe {
namespace {

struct Result {
    std::string rule, library, failure_code;
    bool ok;
};

std::vector<Result>& results() {
    static std::vector<Result> r;
    return r;
}

struct Registered {
    std::string recipe;
    Probe probe;
    When when;
};

std::vector<Registered>& registry() {
    static std::vector<Registered> r;
    return r;
}

std::string quote(const std::string& s) {
    std::string out = "\"";
    for (char c : s) {
        if (c == '"' || c == '\\') out += '\\';
        if (static_cast<unsigned char>(c) < 0x20) {
            char buf[8];
            std::snprintf(buf, sizeof buf, "\\u%04x", c);
            out += buf;
            continue;
        }
        out += c;
    }
    return out + "\"";
}

void write_steps(std::FILE* f, const char* key, const std::vector<std::string>& steps) {
    std::fprintf(f, "  %s: [", quote(key).c_str());
    for (std::size_t i = 0; i < steps.size(); ++i) std::fprintf(f, "%s%s", i ? ", " : "", quote(steps[i]).c_str());
    std::fprintf(f, "]");
}

} // namespace

Registration::Registration(const char* recipe, Probe probe, When when) {
    registry().push_back({recipe, probe, when});
}

void run(When when) {
    for (const Registered& r : registry()) {
        if (r.when == when) r.probe();
    }
}

void check(const std::string& rule, const std::string& library, bool ok, const std::string& failure_code) {
    results().push_back({rule, library, failure_code, ok});
}

bool write_results(const std::string& path) {
    std::FILE* f = std::fopen(path.c_str(), "w");
    if (f == nullptr) return false;
    bool all_ok = true;
    std::fprintf(f, "{\n  \"results\": [");
    for (std::size_t i = 0; i < results().size(); ++i) {
        const Result& r = results()[i];
        all_ok = all_ok && r.ok;
        std::fprintf(f, "%s\n    {\"rule\": %s, \"library\": %s, \"status\": %s", i ? "," : "", quote(r.rule).c_str(),
                     quote(r.library).c_str(), r.ok ? "\"pass\"" : "\"fail\"");
        if (!r.ok) std::fprintf(f, ", \"failure_code\": %s", quote(r.failure_code).c_str());
        std::fprintf(f, "}");
    }
    std::fprintf(f, "\n  ],\n");
    write_steps(f, "init_steps", lifecycle::init_steps());
    std::fprintf(f, ",\n");
    write_steps(f, "shutdown_steps", lifecycle::shutdown_steps());
    std::fprintf(f, "\n}\n");
    std::fclose(f);
    return all_ok;
}

} // namespace am::probe
