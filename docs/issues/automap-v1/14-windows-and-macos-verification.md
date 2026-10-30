# 14: Windows and macOS verification

**What to build:** `verify` works on `windows-x64-msvc` and `macos-arm64-appleclang` as well as Linux: MSVC and Apple Clang toolchains, WARP for D3D12 on Windows, the paravirtual Metal device on macOS, and each library's headless settings from its platform facts. The build-flags probe also emits `#pragma detect_mismatch` on MSVC.

**Blocked by:** [11: Jobs contract and ThreadSanitizer](11-jobs-contract-and-threadsanitizer.md)

**Status:** done

- [x] The fixture composition reaches probed on all three hosted runners.
- [x] On Windows, `jobs.tsan-clean` is not-applicable with the reason (ADR-0076).
- [x] Run conditions record the compiler version and runner image on each platform.
- [x] A seeded ABI-define mismatch fails at link time on MSVC and at probe time elsewhere.

## Comments

- CI's "Fixture builds" jobs run the whole test suite, builds included, on hosted Windows (MSVC through the Visual Studio developer environment, Ninja) and macOS (Apple Clang, with ThreadSanitizer). Files check out with LF everywhere (`.gitattributes`), so golden files and seed patches match byte for byte.
- `verify` picks the compiler by platform and records its version: `cl`'s banner on MSVC, `clang++ --version` on Apple Clang. Run conditions record the hosted runner's image.
- The fixture's stub libraries build static under every triplet (`vcpkg_check_linkage(ONLY_STATIC_LIBRARY)`), since they export no DLL symbols. stub-ui's header declares `STUB_UI_WIDE_IDS` with `#pragma detect_mismatch`; on MSVC the seeded mismatch fails the link with `LNK2038`, which `verify` reports as `ABI_DEFINE_MISMATCH(STUB_UI_WIDE_IDS)` at level none, and elsewhere the build-flags probe reports the same code.
- The probe harness reads MSVC's C++ standard from `_MSVC_LANG`. ThreadSanitizer frames on macOS name the module with its architecture (`testbed:arm64`), which the report parser strips.
- The public map's SDL3 and Dear ImGui testbed on WARP and paravirtual Metal is verified on those runners by ticket 15's matrix.
