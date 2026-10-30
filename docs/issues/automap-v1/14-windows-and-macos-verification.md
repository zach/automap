# 14: Windows and macOS verification

**What to build:** `verify` works on `windows-x64-msvc` and `macos-arm64-appleclang` as well as Linux: MSVC and Apple Clang toolchains, WARP for D3D12 on Windows, the paravirtual Metal device on macOS, and each library's headless settings from its platform facts. The build-flags probe also emits `#pragma detect_mismatch` on MSVC.

**Blocked by:** [11: Jobs contract and ThreadSanitizer](11-jobs-contract-and-threadsanitizer.md)

**Status:** ready-for-agent

- [ ] The fixture composition reaches probed on all three hosted runners.
- [ ] On Windows, `jobs.tsan-clean` is not-applicable with the reason (ADR-0076).
- [ ] Run conditions record the compiler version and runner image on each platform.
- [ ] A seeded ABI-define mismatch fails at link time on MSVC and at probe time elsewhere.
