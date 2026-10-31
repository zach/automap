// Probes for Jolt's binding to the math-types contract.
// SPDX-License-Identifier: MIT-0
#if defined(AUTOMAP_PROBE) && defined(AUTOMAP_LIBRARY_GLM)
#include <cfloat>
#include <cstring>

#include "automap/probe.h"
#include "jolt_math_glm.h"

namespace {

// Sample values that a conversion through another type or precision would
// change: thirds, tiny and huge magnitudes, the smallest normal float,
// negative zero.
const float samples[] = {1.0f / 3.0f, -2.5e-7f, 3.0e38f, FLT_MIN, -0.0f, 12345.678f, -1.0f, 0.5f,
                         7.0f / 9.0f, -1.0e-30f, 2.0f, 1.0e7f + 1.0f, -3.25f, 0.1f, 6.0f, -0.75f};

template <typename T> bool same(const T& a, const T& b) { return std::memcmp(&a, &b, sizeof(T)) == 0; }

bool round_trips() {
    bool ok = true;
    for (int i = 0; i + 4 <= 16; i += 4) {
        const float* s = samples + i;
        JPH::Vec3 v3(s[0], s[1], s[2]);
        JPH::Vec4 v4(s[0], s[1], s[2], s[3]);
        JPH::Quat q(s[0], s[1], s[2], s[3]);
        ok = ok && same(am::math::to_jolt(am::math::to_glm(v3)), v3) && same(am::math::to_jolt(am::math::to_glm(v4)), v4) &&
             am::math::to_jolt(am::math::to_glm(q)) == q;
        glm::quat g = am::math::to_glm(q);
        ok = ok && g.x == s[0] && g.y == s[1] && g.z == s[2] && g.w == s[3];
    }
    JPH::Mat44 m(JPH::Vec4(samples[0], samples[1], samples[2], samples[3]), JPH::Vec4(samples[4], samples[5], samples[6], samples[7]),
                 JPH::Vec4(samples[8], samples[9], samples[10], samples[11]), JPH::Vec4(samples[12], samples[13], samples[14], samples[15]));
    glm::mat4 g = am::math::to_glm(m);
    ok = ok && am::math::to_jolt(g) == m && g[1][2] == samples[6];
#if defined(JPH_DOUBLE_PRECISION)
    JPH::RVec3 r(1.0 / 3.0, -1.0e300, 4503599627370497.0);
    ok = ok && same(am::math::to_jolt(am::math::to_glm(r)), r);
#endif
    return ok;
}

// math-types.adapters-round-trip: each adapter between Jolt's types and
// GLM's converts the sample values there and back without loss, and lands
// each component where GLM keeps it.
void am_probe_jolt_math(am::Context&) {
    am::probe::check("math-types.adapters-round-trip", "jolt", round_trips(), "ADAPTER_LOSSY(jolt)");
}
const am::probe::Registration registration("jolt--math-types", &am_probe_jolt_math);

} // namespace
#endif
