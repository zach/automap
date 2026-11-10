# 16: Choices per build type and a shippable profiler

**What to build:** A composition can hold a library in development builds only, and Tracy has a variant that can ship: on demand and localhost only, with a probe that fits it. Field notes [43].

**Blocked by:** None

**Status:** ready-for-human

- [ ] The product owner agrees an ADR for choices that differ by build type, and the on-demand Tracy variant as map coverage.
- [ ] The manifest can say a slot's library is for development builds only; a release build links nothing from it and its facade calls go nowhere.
- [ ] Tracy's `on-demand` variant records only while a profiler is connected and listens on localhost; the diagnostics probe checks it as such.
