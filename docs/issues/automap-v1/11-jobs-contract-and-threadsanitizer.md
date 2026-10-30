# 11: Jobs contract and ThreadSanitizer

**What to build:** The jobs contract arrives with its shared interface and reference thread pool, sized from the choices header. `verify` runs the jobs contract's ThreadSanitizer procedure as a separate build on the platforms the rule names, applying the frame rule.

**Blocked by:** [10: automap extend](10-automap-extend.md)

**Status:** ready-for-agent

- [ ] Extending a composition with a library that has jobs facts, while the jobs slot is empty, writes the jobs interface with the reference thread pool (ADR-0057).
- [ ] The reference thread pool waits for in-flight jobs before it is destroyed.
- [ ] The seeded race patch makes `verify` report `TSAN_REPORT(...)`.
- [ ] A report whose every frame is outside code the composition builds is recorded under set-aside reports and doesn't fail the rule (ADR-0073); a fixture with an uninstrumented shared library tests this.
- [ ] On a platform outside the rule's list, the result is not-applicable with the reason (ADR-0076).
