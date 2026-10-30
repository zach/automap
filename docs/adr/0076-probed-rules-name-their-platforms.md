---
status: accepted
date: 2026-10-30
---

# A probed rule may name the platforms its probe can run on

## Context

The jobs contract's probed rules include a clean ThreadSanitizer run ([ADR-0027](0027-v1-probe-floor-and-classification.md), [ADR-0073](0073-tsan-reports-count-when-they-touch-composition-code.md)). ThreadSanitizer doesn't exist for MSVC, which lists it as future work, and the research recommended running it on Linux only ([06](../research/06-verification.md)). The probed level requires every applicable probed rule to pass, with none unverifiable ([ADR-0072](0072-result-statuses-and-run-conditions.md)). On `windows-x64-msvc` the ThreadSanitizer rule could only be unverifiable, so no Windows composition using the jobs contract could ever be known-good. Glue is the same source on every platform.

## Options

1. **Unverifiable on Windows:** Windows compositions with jobs never reach probed.
2. **Platform-limited probes:** a probed rule names the platforms its probe can run on; elsewhere its result is not-applicable, with the reason.

## Decision

Option 2. A probed rule may name the platforms its probe can run on. On any other platform its result is recorded as not-applicable, with the reason. `jobs.tsan-clean` runs on `linux-x64-clang` and `macos-arm64-appleclang`, the v1 platforms where ThreadSanitizer exists.

## Consequences

- Windows compositions can reach the probed level.
- Evidence on Windows shows that the ThreadSanitizer rule didn't run there and why.
