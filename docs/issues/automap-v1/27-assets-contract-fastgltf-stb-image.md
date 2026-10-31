# 27: Assets contract, fastgltf and stb_image

**What to build:** The assets contract arrives with its file-I/O shared interface. fastgltf fills model import and stb_image fills image decode, each reading files only through the interface.

**Blocked by:** [15: Automap's CI verifies declared edges and signs evidence into the map](15-automap-ci-signs-evidence-into-the-map.md)

**Status:** done

- [x] A composition with both reaches probed on all three platforms.
- [x] A seeded direct file read fails the every-read-through-the-interface rule; a seeded nondeterministic output fails the identical-output rule.
- [x] stb_image's compile-time allocation hooks route to the facade.

## Comments

- The assets contract 1.0.0 has two probed rules: `assets.reads-through-interface` (`ASSET_READ_BYPASSED`) and `assets.output-identical` (`ASSET_OUTPUT_DIFFERS`), with `assets.reload-on-change` as guidance. Its shared interface, `automap/assets.h`, reads whole files with the C runtime and counts reads per library; a game that keeps assets elsewhere replaces `read()`.
- fastgltf 0.9.1 (curated): `am_fastgltf_load()` reads the file through the interface, parses without `LoadExternalBuffers`, and loads relative buffers through the interface into `sources::Array`. Its seeds turn `LoadExternalBuffers` on and name meshes after their addresses.
- stb_image (the stb port of 2026-08-02, header v2.30): `am_stb_image_load()` decodes bytes read through the interface. The allocation binding is the implementation's one translation unit, with `STBI_MALLOC`, `STBI_REALLOC` and `STBI_FREE` on the facade; `hooks-before-first-allocation` is not-applicable to compile-time hooks. The stb port has only a find module, so the build fragment defines `stb::image`.
- The probes write their sample files (a glTF with an external buffer, a 3x2 PNG) to the temp directory. `testbed-assets` joins the reference compositions.
- The SPEC's catalog gives fastgltf a math-types binding; it isn't part of this ticket and isn't in the map yet.
