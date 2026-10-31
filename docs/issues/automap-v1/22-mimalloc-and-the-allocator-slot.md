# 22: mimalloc and the allocator slot

**What to build:** mimalloc enters the map and fills the allocator slot; extending with it sets `allocation.allocator` to `allocator-slot` and backs the allocator facade with mimalloc.

**Blocked by:** [15: Automap's CI verifies declared edges and signs evidence into the map](15-automap-ci-signs-evidence-into-the-map.md)

**Status:** done

- [x] A composition with mimalloc reaches probed on all three platforms in Automap's CI.
- [x] Every library's allocations still route through the facade, now served by mimalloc.

## Comments

- mimalloc 3.5.3 is the version vcpkg carries at the map's baseline; newer releases arrive through curation.
- The allocation binding defines the facade's `am_allocator_slot_*` functions with `mi_malloc_aligned`, `mi_realloc` and `mi_free`; mimalloc has no hooks, so `allocation.hooks-before-first-allocation` is not applicable to it.
- Its probe needs both a probe-build count of the blocks mimalloc served and `mi_is_in_heap_region`: on Windows the heap check alone missed the seed.
- mimalloc's output goes to the diagnostics facade through `mi_register_output`, with its own probe and seed.
- A choice naming a slot (`<slot>-slot`) now needs a library in that slot, for every slot.
- The enkiTS jobs binding's tasks now delete themselves through a completion action; reaping completed tasks could free one another thread was still adding.
