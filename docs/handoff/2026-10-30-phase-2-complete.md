# Handoff: Phase 2 complete, Phase 3 next

Date: 2026-10-30. Branch: `exploration` on `zach/automap`, pushed.

## Where things stand

- **Phase 0 and Phase 1** are done. Research is in `docs/research/`; start with `docs/research/SUMMARY.md`.
- **Phase 2 (grilling)** is done: every branch of the decision tree in `docs/PLAN.md`, including branch 15 (the agent-driven build plan), is closed. Decisions are ADR-0001 to ADR-0069 in `docs/adr/`; terms are in `CONTEXT.md`.
- **`docs/OPEN_QUESTIONS.md`** has no open items. Two items are deferred past v1: OQ-008 (web testbed) and OQ-011 (lock identity across build backends).
- **Next: Phase 3**, the planned spike from the brief, then **Phase 4** (spec synthesis). See below.

## What the next session should do

1. Read `CLAUDE.md`, `docs/EXPLORATION_BRIEF.md`, `docs/PLAN.md`, then skim the ADR titles in `docs/adr/` and `CONTEXT.md`.
2. **Phase 3 spike** (pending the product owner's go-ahead): draft the manifest, map entries (facts, bindings, the Dear ImGui and SDL3 pairing) and recipes for SDL3, Dear ImGui and Jolt in `spikes/<name>/`; give a fresh agent with no conversation history only those drafts and a one-line request (start an SDL3 and Dear ImGui testbed, then extend it with Jolt); have it build and run on Linux x64 in this container; record where it went wrong in an ADR; delete or archive the spike code. The drafts double as starting points for the spec's data-format examples.
3. **Phase 4:** write `docs/spec/SPEC.md` (the 12 sections in the brief, section 5) only from ADRs and `CONTEXT.md`; run `to-spec` and `to-tickets` (feature directory under `docs/issues/`; slices per ADR-0061); run the definition-of-done check in brief section 6, including a fresh-agent test of whether slice 1 can begin without foundational questions.

## Environment facts for the spike (checked 2026-10-30)

- Toolchain present: CMake 3.28.3, Ninja, GCC 13.3, Clang 18.1.3, Vulkan loader (`libvulkan1`). 4 cores, 15 GB RAM.
- No Vulkan driver is installed, but `apt-get update` works and `mesa-vulkan-drivers` (lavapipe) 25.2.8 is installable.
- The network policy blocks many vendor documentation sites; GitHub (`github.com`, `raw.githubusercontent.com`) and Ubuntu mirrors work. `npx` package execution was blocked earlier in the session.

## Process rules that matter (all in `CLAUDE.md`)

- One question at a time, each with a recommendation; say when a question needs the product owner's judgment.
- Documents state the current version only: rewrite ADRs in place when a decision changes; no "amended" or "previously".
- An ADR records only what the product owner agreed to; ask about any new consequence rather than writing it in.
- Don't assume agents are weaker stewards of code than humans.
- Commit and push docs as decisions land (the environment's stop hook enforces it).

## Suggested skills

- `prototype` for the Phase 3 spike.
- `domain-modeling` while drafting schemas, to keep names aligned with `CONTEXT.md`.
- `to-spec` and `to-tickets` in Phase 4.
- `handoff` at the end of the next session.
