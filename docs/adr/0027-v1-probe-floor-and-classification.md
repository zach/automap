---
status: accepted
date: 2026-10-29
---

# Every contract has at least one probed rule in v1

## Context

The brief asks which seams are machine-checkable and which are guidance. Every rule is probed or guidance ([ADR-0022](0022-rules-are-probed-or-guidance.md)), so the question becomes which rules get probes first. The research rated each seam from real headers ([02c](../research/02c-catalog-simulation-runtime-glue.md), [06](../research/06-verification.md)): build flags are cheap to probe because Jolt (`JPH_VERSION_ID`), Tracy (link-time option mangling) and Dear ImGui (`IMGUI_CHECKVERSION`) already self-check, and MSVC's `#pragma detect_mismatch` turns mismatched defines into link errors; ThreadSanitizer is the practical probe for jobs. Prose-only guidance adds little on its own ([04](../research/04-agent-ecosystem.md)).

## Options

1. **No floor:** each contract gets probes as they're written; some may ship as guidance only.
2. **One probed rule per contract, from the research's classification.**
3. **Every rule probed before v1.**

## Decision

Option 2. In v1 every contract has at least one probed rule. The starting classification:

| Contract | v1 probed rules | v1 guidance |
| --- | --- | --- |
| Build flags | ABI defines match; C++ standard, exceptions, RTTI and C runtime agree | — |
| Platform | every library supports every target | — |
| Lifecycle | phase order | — |
| Serialization | rebuild keys include the library's version and variant | — |
| Allocation | hooks installed before the first allocation; all allocations route through the facade | semantic rules |
| Diagnostics | hooks route through the facade | semantic rules |
| Jobs | library work runs on the shared interface; ThreadSanitizer run is clean | semantic rules |
| Math types | precision agrees across libraries; adapters convert there and back without loss | semantic rules |
| Determinism (when chosen) | floating-point flags and library defines set; replay hashes match | semantic rules |
| Handles | ID bit widths | ownership and caching rules |
| Assets | output identical across runs with the same input, version and settings; every file read goes through the file-I/O interface | semantic rules |

## Consequences

- No contract ships as prose alone.
- The spec's content plan lists each contract's probed and guidance rules from this classification.
