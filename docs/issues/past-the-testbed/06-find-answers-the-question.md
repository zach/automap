# 06: find answers the question

**What to build:** `automap find <query>` matches libraries by id, name, description, and the ids, names and descriptions of the slots they fill, in every map layer; when nothing matches it lists the slots; `find --json` gives one short entry per library. Field notes [6, 46].

**Blocked by:** None

**Status:** done

- [x] `find text` finds a private map's library that fills a private `text` slot; `find font` finds it through the slot's or library's description.
- [x] A query that matches nothing says so and lists every slot id with the count of libraries that fill it.
- [x] A `find --json` entry has `id`, `name`, `version`, `variants`, `slots`, `license`, `license_class`, `match` and `evidence_by_platform`: per platform, `<level> <state>`, the lowest level and the worst state across the library's bindings and pairings; no per-binding rows.
- [x] `show library` still carries every evidence row.
- [x] The cli-result schema, its examples and SPEC §5.4 describe the entry.

## Comments

- Ids and names match by substring; slot and variant descriptions match by whole words, singular or plural, so `find text` finds a slot that draws text and not one that loads textures. Matches sort closest first, and each entry says how it matched.
- On the public map `find --json` is 261 lines for thirteen libraries, from about 600, and each entry has the library's slots.
