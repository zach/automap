---
status: accepted
date: 2026-10-29
---

# Evidence is graded on five levels; known-good means probed or higher

## Context

Verification evidence, not human review, is what makes a change trustworthy ([ADR-0005](0005-agent-is-primary-steward.md)). Package ecosystems verify only that packages build and link; none publishes "A works with B" ([06](../research/06-verification.md)). The research recommended graded evidence, with a working precedent on hosted CI runners for every rung: build matrices, headless smoke runs with dummy video and audio drivers, define probes and sanitizers, determinism hashes and golden images with perceptual tolerances, and real-GPU runners.

## Options

1. **Binary evidence:** a claim is known-good or not.
2. **A graded ladder with a known-good threshold.**

## Decision

Option 2. Evidence is graded on five levels:

1. **builds:** configures and builds on the platform.
2. **smoke:** links, runs a set number of frames headless, and shuts down cleanly.
3. **probed:** every probed rule passes, including sanitizer runs.
4. **behavioral:** determinism hashes, serialization round-trips and golden-image renders all pass.
5. **hardware:** passes on real GPUs, devices or consoles.

Every claim states the highest level it passed. **Known-good** means at least probed.

## Consequences

- "Known-good" covers the seams, which is what Automap uniquely checks, not just buildability.
- An agent can tell a claim that was only built from one that ran its probes.
