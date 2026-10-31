// Jolt's binding to the math-types contract: adapters between Jolt's math
// types and GLM's, for a composition with GLM in the math slot. Both are
// right-handed with no fixed up axis in the conversion, and both store
// matrices column-major, so the adapters copy components without changing
// conventions. glm::quat's constructor takes w first; Jolt's takes it last.
// SPDX-License-Identifier: MIT-0
#pragma once
#if defined(AUTOMAP_LIBRARY_GLM)
#include <Jolt/Jolt.h>

#include <glm/glm.hpp>
#include <glm/gtc/quaternion.hpp>

namespace am::math {

inline glm::vec3 to_glm(JPH::Vec3Arg v) { return glm::vec3(v.GetX(), v.GetY(), v.GetZ()); }
inline JPH::Vec3 to_jolt(const glm::vec3& v) { return JPH::Vec3(v.x, v.y, v.z); }

inline glm::vec4 to_glm(JPH::Vec4Arg v) { return glm::vec4(v.GetX(), v.GetY(), v.GetZ(), v.GetW()); }
inline JPH::Vec4 to_jolt(const glm::vec4& v) { return JPH::Vec4(v.x, v.y, v.z, v.w); }

inline glm::quat to_glm(JPH::QuatArg q) { return glm::quat(q.GetW(), q.GetX(), q.GetY(), q.GetZ()); }
inline JPH::Quat to_jolt(const glm::quat& q) { return JPH::Quat(q.x, q.y, q.z, q.w); }

inline glm::mat4 to_glm(JPH::Mat44Arg m) {
    glm::mat4 out;
    for (int c = 0; c < 4; ++c) out[c] = to_glm(m.GetColumn4(c));
    return out;
}
inline JPH::Mat44 to_jolt(const glm::mat4& m) {
    return JPH::Mat44(to_jolt(m[0]), to_jolt(m[1]), to_jolt(m[2]), to_jolt(m[3]));
}

#if defined(JPH_DOUBLE_PRECISION)
// World positions are double precision: RVec3 converts to glm::dvec3.
inline glm::dvec3 to_glm(JPH::RVec3Arg v) { return glm::dvec3(v.GetX(), v.GetY(), v.GetZ()); }
inline JPH::RVec3 to_jolt(const glm::dvec3& v) { return JPH::RVec3(v.x, v.y, v.z); }
#endif

} // namespace am::math
#endif
