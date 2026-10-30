# 27: Assets contract, fastgltf and stb_image

**What to build:** The assets contract arrives with its file-I/O shared interface. fastgltf fills model import and stb_image fills image decode, each reading files only through the interface.

**Blocked by:** [15: Automap's CI verifies declared edges and signs evidence into the map](15-automap-ci-signs-evidence-into-the-map.md)

**Status:** ready-for-agent

- [ ] A composition with both reaches probed on all three platforms.
- [ ] A seeded direct file read fails the every-read-through-the-interface rule; a seeded nondeterministic output fails the identical-output rule.
- [ ] stb_image's compile-time allocation hooks route to the facade.
