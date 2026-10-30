# 22: mimalloc and the allocator slot

**What to build:** mimalloc enters the map and fills the allocator slot; extending with it sets `allocation.allocator` to `allocator-slot` and backs the allocator facade with mimalloc.

**Blocked by:** [15: Automap's CI verifies declared edges and signs evidence into the map](15-automap-ci-signs-evidence-into-the-map.md)

**Status:** ready-for-agent

- [ ] A composition with mimalloc reaches probed on all three platforms in Automap's CI.
- [ ] Every library's allocations still route through the facade, now served by mimalloc.
