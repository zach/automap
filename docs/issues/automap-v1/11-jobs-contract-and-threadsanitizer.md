# 11: Jobs contract and ThreadSanitizer

**What to build:** The jobs contract arrives with its shared interface and reference thread pool, sized from the choices header. `verify` runs the jobs contract's ThreadSanitizer procedure as a separate build on the platforms the rule names, applying the frame rule.

**Blocked by:** [10: automap extend](10-automap-extend.md)

**Status:** done

- [x] Extending a composition with a library that has jobs facts, while the jobs slot is empty, writes the jobs interface with the reference thread pool (ADR-0057).
- [x] The reference thread pool waits for in-flight jobs before it is destroyed.
- [x] The seeded race patch makes `verify` report `TSAN_REPORT(...)`.
- [x] A report whose every frame is outside code the composition builds is recorded under set-aside reports and doesn't fail the rule (ADR-0073); a fixture with an uninstrumented shared library tests this.
- [x] On a platform outside the rule's list, the result is not-applicable with the reason (ADR-0076).

## Comments

- The public map's jobs contract ships `automap/jobs.h` and `jobs.cpp`: `am::jobs::submit`, `wait` (which runs queued jobs on the waiting thread), `workers` and `submitted_count`, backed by the reference thread pool sized by `AUTOMAP_JOBS_WORKERS`, or by the jobs slot's binding. The startup calls its setup in the jobs phase; at shutdown the workers drain the queue and are joined, so no job outlives the pool.
- The fixture map gains `stub-physics`, whose jobs facts bring the jobs contract in when it is added; its binding routes its tasks through the interface. The jobs contract's race seed is exercised in that composition, not the public testbed, which has no jobs library until Jolt.
- ThreadSanitizer report parsing is a pure function, tested on real GCC output and on clang's frame layout; the uninstrumented-library case runs end to end in CI, where clang's ThreadSanitizer runtime is installed.
- In a composition where no library declares an ABI-affecting define, `build-flags.abi-defines-match` is not-applicable with that reason.
