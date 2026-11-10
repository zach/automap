# 17: Text and data files in the public map

**What to build:** The public map gains a `text` slot filled by stb_truetype and a `data-files` slot filled by nlohmann/json, from Spintacular's private map. Field notes [24, 30].

**Blocked by:** 01

**Status:** ready-for-human

- [ ] The product owner agrees the two slots and libraries as map coverage.
- [ ] Both libraries, their recipes, stb_truetype's allocation binding with its probe and seeded violation are in `map/`, MIT-0, verified on all three platforms.
- [ ] stb_truetype and stb_image share the `stb` port in one composition.
