# 05: Results that say why

**What to build:** A failed build's result carries its cause; a blocked download is `FETCH_FAILED` with the URL; a failed start and a seed run describe themselves plainly. Field notes [12, 13, 18, 19, 48].

**Blocked by:** None

**Status:** ready-for-agent

- [ ] A `BUILD_FAILED` failure's message ends with the first error lines of the build log (compiler, CMake or vcpkg), at most a few lines.
- [ ] When the build log shows vcpkg failing to download, the failure is `FETCH_FAILED` with the first URL that failed and a pointer to `automap vendor` and `automap fetch`, from `verify`, `start`, `extend` and `upgrade`.
- [ ] A start, extend or upgrade whose build fails summarizes as "…; it didn't build on <platform>: <code>", never "the composition is none".
- [ ] `verify --seed` reports no `level`.
- [ ] SPEC's failure-code table has `FETCH_FAILED` from verify.
