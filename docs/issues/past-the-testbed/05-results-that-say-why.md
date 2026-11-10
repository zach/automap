# 05: Results that say why

**What to build:** A failed build's result carries its cause; a blocked download is `FETCH_FAILED` with the URL; a failed start and a seed run describe themselves plainly. Field notes [12, 13, 18, 19, 48].

**Blocked by:** None

**Status:** done

- [x] A `BUILD_FAILED` failure's message ends with the first error lines of the build log (compiler, CMake or vcpkg), at most a few lines.
- [x] When the build log shows vcpkg failing to download, the failure is `FETCH_FAILED` with the URL that failed and a pointer to `automap vendor`, from `verify`, `start`, `extend` and `upgrade`.
- [x] A start, extend or upgrade whose build fails summarizes as "…; it didn't build on <platform>: <code>", and a verify as "The composition didn't build on <platform>: <code>", never "is none".
- [x] `verify --seed` reports no `level`.
- [x] SPEC's failure-code table has `FETCH_FAILED` from verify.

## Comments

- `verify.DiagnoseBuild` is a pure function of a failed build's log: the URL vcpkg was downloading when a download failed, and the first error with the lines that belong to it. Its test data includes the real log of a build behind a proxy that refuses GitHub's archive URLs. `verify.Run` turns it into `FETCH_FAILED` or a `BUILD_FAILED` whose message quotes the error; `fetch` names the URL the same way.
- The router skill's list of failure codes has `FETCH_FAILED` and the `vendor` route.
