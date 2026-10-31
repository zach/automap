# 24: GLM and the math adapters

**What to build:** GLM fills the math slot. Jolt's and Dear ImGui's math-types bindings gain adapters to GLM, with round-trip probes, and the position-precision rule checks the composition's choice.

**Blocked by:** [12: Jolt in the map](12-jolt-in-the-map.md), [15: Automap's CI verifies declared edges and signs evidence into the map](15-automap-ci-signs-evidence-into-the-map.md)

**Status:** done

- [x] `automap extend glm` on a composition with Jolt and Dear ImGui ends probed on all three platforms.
- [x] Each adapter converts its sample values there and back without loss; a seeded lossy adapter fails `math-types.adapters-round-trip`.
- [x] A seeded precision mismatch fails `math-types.precision-agrees`.

## Comments

- GLM 1.0.3 is the curated port at the map's baseline; the map records it as MIT (it is MIT or the Happy Bunny License).
- The math-types contract's 1.1.0 makes `position_precision` optional, for libraries with no world positions (Dear ImGui, GLM), and names the `math` slot on `math-types.adapters-round-trip`. A rule's `slot` makes its result not-applicable while that slot is empty, so Automap's CI skips the adapters' seeds there; `math-types.precision-agrees` is not-applicable when no library exposes world positions.
- The adapters are inline `am::math::to_glm` / `to_jolt` / `to_imgui` functions in the libraries' math-types bindings, compiled when `AUTOMAP_LIBRARY_GLM` is defined. Their probes round-trip sample values bit for bit; the seeds pass a quaternion to `glm::quat` in x, y, z, w order and drop an ImVec4's w.
- The precision seed (the contract's) runs in the GLM composition too and reports `PRECISION_MISMATCH(jolt)`.
- `testbed-jolt-glm` joins the reference compositions.
