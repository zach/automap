# 07: Slice-1 contracts and shared interfaces

**What to build:** The public map gets its first real content: the allocation, diagnostics, lifecycle, build-flags and platform contracts at 1.0.0, the allocator facade (tracking live blocks in probe builds) and the diagnostics facade as shared interface sources, the probe harness, and a seeded-violation patch for every probed composition rule.

**Blocked by:** [05: Probe harness and the probed level](05-probe-harness-and-probed-level.md)

**Status:** ready-for-agent

- [ ] Each contract validates and has the probed and guidance rules of ADR-0027's table.
- [ ] In probe builds, releasing a block the facade never allocated fails `allocation.hooks-before-first-allocation` for that library.
- [ ] Applying each composition rule's seeded-violation patch to the fixture composition makes `verify` report that rule's failure code.
- [ ] Shared interfaces and the probe harness are MIT-0 (ADR-0055) and compile with the fixture composition on Linux.
