# 17: Text and data files in the public map

**What to build:** The public map gains a `text` slot filled by stb_truetype and a `data-files` slot filled by nlohmann/json, from Spintacular's private map. Field notes [24, 30].

**Blocked by:** 01

**Status:** ready-for-human

- [x] The product owner agrees the two slots and libraries as map coverage ([ADR-0084](../../adr/0084-the-map-covers-text-and-data-files.md)).
- [ ] Both libraries, their recipes, stb_truetype's allocation binding with its probe and seeded violation are in `map/`, MIT-0, verified on all three platforms.
- [ ] stb_truetype and stb_image share the `stb` port in one composition.

## Comments

- stb_truetype is a library of the `stb` port in the public map, not the in-repo wrapper Spintacular had to use: with ticket 01, it and stb_image share the port's entry. Its build fragment defines `stb::truetype` over the port's headers, as stb_image's defines `stb::image`.
- `ci/compositions.json` gains `testbed-text-data` (SDL3, Dear ImGui, stb_image, stb_truetype, nlohmann/json), and `TestTheTestbedExtendsWithTextAndDataFilesProbed` runs in the public-map job's `media` group on all three platforms with stb_truetype's seed. The starter engine stays v1's thirteen libraries.
