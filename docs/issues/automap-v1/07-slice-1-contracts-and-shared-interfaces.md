# 07: Slice-1 contracts and shared interfaces

**What to build:** The public map gets its first real content: the allocation, diagnostics, lifecycle, build-flags and platform contracts at 1.0.0, the allocator facade (tracking live blocks in probe builds) and the diagnostics facade as shared interface sources, the probe harness, and a seeded-violation patch for every probed composition rule.

**Blocked by:** [05: Probe harness and the probed level](05-probe-harness-and-probed-level.md)

**Status:** done

- [x] Each contract validates and has the probed and guidance rules listed in SPEC §8.2, which meet ADR-0027's floor.
- [x] In probe builds, releasing a block the facade never allocated fails `allocation.hooks-before-first-allocation` for that library.
- [x] Every probed composition rule has a seeded violation (in its contract, or in the fixture's recipes where the mistake belongs in their files), and `verify --seed` catches each in the fixture composition.
- [x] Shared interfaces and the probe harness are MIT-0 (ADR-0055) and compile with the fixture composition on Linux.

## Comments

- The fixture map has no contracts or probe harness of its own: tests assemble it from `testdata/map` plus `map/contracts` and `map/probe-harness`, so every fixture build and seed run exercises the public map's shared interfaces.
- The allocation facade serves plain and aligned blocks and reallocation, and routes to the allocator slot's binding when the allocation choice is `allocator-slot`; the diagnostics facade routes logs and profile zones to the logging and profiling slots' bindings when chosen. In probe builds the allocation facade tracks live blocks.
- On its own, `map/` doesn't load yet: `lifecycle.*` and `build-flags.*` are composition rules whose seeds belong in recipe files, and the public map has no recipes until ticket 08. The fixture's recipes seed them meanwhile.
