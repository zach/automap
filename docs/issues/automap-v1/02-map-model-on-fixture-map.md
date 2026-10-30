# 02: Map model on the fixture map

**What to build:** The CLI loads the public map and any private maps, validates every file, enforces the cross-file rules, and answers queries. `automap show` and `automap find` print map entries with their facts. A fixture map with two tiny stub C++ libraries (facts for allocation, diagnostics, lifecycle, build flags and platform; bindings; a pairing between them; library recipes) is the test bed for the slices that follow.

**Blocked by:** [01: Repo setup and the CLI seam](01-repo-setup-and-cli-seam.md)

**Status:** ready-for-agent

- [ ] Every file under the product spec's examples loads, and each rule the spec's example checker enforces is enforced by the map model, reported as `MAP_INVALID(<file>)` with the reason.
- [ ] Private maps layer in manifest order; an extension that changes anything other than adding versions, variants or (in an NDA platform's map) supported targets is rejected (ADR-0075).
- [ ] `automap show <library|slot|contract|recipe> --json` and `automap find --slot <slot> --json` match golden outputs against the fixture map.
- [ ] The fixture map's recipes carry probes and seeded-violation patches, so later tickets can test them.
