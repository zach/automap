# 34: Custom-integration evals

**What to build:** Journey 3 as eval tasks: adopt, a patched variant, and an NDA platform, using a fixture copy of Acme's engine and a stand-in NDA map (a Linux toolchain under a private platform id, signed with a test key).

**Blocked by:** [09: Eval harness and the testbed-start eval](09-eval-harness-and-testbed-start.md), [32: automap adopt](32-automap-adopt.md), [33: Per-platform variants, NDA maps and NDA signing](33-per-platform-variants-and-nda-maps.md)

**Status:** ready-for-agent

- [ ] `custom-adopt`: the session ends with a manifest, a private map describing the in-house libraries, and probed desktop evidence.
- [ ] `custom-variant`: the session adds the patched SDL3 variant and ends probed.
- [ ] `custom-nda`: the session resolves per-platform variants and ends with signed evidence stored in the stand-in map.
