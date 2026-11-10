# 06: find answers the question

**What to build:** `automap find <query>` matches libraries by id, name, description, and the ids, names and descriptions of the slots they fill, in every map layer; when nothing matches it lists the slots; `find --json` gives one short entry per library. Field notes [6, 46].

**Blocked by:** None

**Status:** ready-for-agent

- [ ] `find text` finds a private map's library that fills a private `text` slot; `find font` finds it through the slot's or library's description.
- [ ] A query that matches nothing says so and lists every slot id with the count of libraries that fill it.
- [ ] A `find --json` entry has `id`, `name`, `version`, `variants`, `slots`, `license` (SPDX and class) and `evidence`: per platform, the worst state and the lowest level across the library's bindings; no per-binding rows.
- [ ] `show library` still carries every evidence row.
- [ ] The cli-result schema, its examples and SPEC §5.4 describe the entry.
