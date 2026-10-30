---
status: accepted
date: 2026-10-30
---

# A ThreadSanitizer report fails the jobs rule only if it touches code the composition builds

## Context

The jobs contract's probed rules include a clean ThreadSanitizer run ([ADR-0027](0027-v1-probe-floor-and-classification.md)). In the spike ([07](../research/07-spike-testbed-agent.md)), every unmodified run reported the same two data races, every frame of them inside Mesa's uninstrumented software Vulkan driver and none in SDL, Dear ImGui, Jolt or glue; whether the composition reached probed depended on one reading of the rule. Real GPU drivers are uninstrumented too. The rule also didn't define the sanitizer build, the run, or how reports are detected. Evidence records its run conditions ([ADR-0072](0072-result-statuses-and-run-conditions.md)).

## Options

1. **Driver settings per platform,** such as making the software rasterizer single-threaded.
2. **Suppression lists per platform,** maintained in the map with a justification for each entry.
3. **A frame rule:** a report fails the rule only if a frame in either of its stacks is in code the composition builds.

## Decision

Option 3. A ThreadSanitizer report fails the jobs rule only if at least one frame in either of its stacks is in code the composition builds: glue, game code, or libraries built from source. Reports wholly inside uninstrumented system code are recorded in the evidence but don't fail the rule. The jobs contract spells out the exact sanitizer build and run, and how reports are detected.

## Consequences

- The rule works the same on every platform and driver, with no per-driver settings or lists to maintain.
- Every race touching code the developer owns or builds still fails the rule.
- Evidence shows which reports were set aside and why.
