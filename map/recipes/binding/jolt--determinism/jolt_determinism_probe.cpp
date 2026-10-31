// Probes for Jolt's binding to the determinism contract.
// SPDX-License-Identifier: MIT-0
#include "automap/choices.h"

#if defined(AUTOMAP_PROBE) && defined(AUTOMAP_DETERMINISM_PROFILE_ON)
#include <cinttypes>
#include <cstdio>

#include "automap/probe.h"
#include "jolt_determinism.h"

namespace {

// The replay hash of the fixed simulation in a deterministic build of Jolt
// 5.6.0, the same on every platform.
constexpr std::uint64_t expected_replay_hash = 0x0a9c29e881ff7288ull;

// determinism.replay-hash-matches: the fixed simulation gives the recorded
// replay hash.
void am_probe_jolt_determinism(am::Context& ctx) {
    std::uint64_t hash = am_jolt_replay_hash(ctx);
    char reason[96] = "";
    if (hash != expected_replay_hash) std::snprintf(reason, sizeof reason, "The replay hash was %016" PRIx64 ".", hash);
    am::probe::check("determinism.replay-hash-matches", "jolt", hash == expected_replay_hash, "REPLAY_MISMATCH(jolt)", reason);
}
const am::probe::Registration registration("jolt--determinism", &am_probe_jolt_determinism);

} // namespace
#endif
