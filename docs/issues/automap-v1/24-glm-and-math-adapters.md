# 24: GLM and the math adapters

**What to build:** GLM fills the math slot. Jolt's and Dear ImGui's math-types bindings gain adapters to GLM, with round-trip probes, and the position-precision rule checks the composition's choice.

**Blocked by:** [12: Jolt in the map](12-jolt-in-the-map.md), [15: Automap's CI verifies declared edges and signs evidence into the map](15-automap-ci-signs-evidence-into-the-map.md)

**Status:** ready-for-agent

- [ ] `automap extend glm` on a composition with Jolt and Dear ImGui ends probed on all three platforms.
- [ ] Each adapter converts its sample values there and back without loss; a seeded lossy adapter fails `math-types.adapters-round-trip`.
- [ ] A seeded precision mismatch fails `math-types.precision-agrees`.
