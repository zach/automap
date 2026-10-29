---
status: accepted
date: 2026-10-29
---

# The domain model has no tiers

## Context

The brief describes three tiers (testbed, starter engine, custom integration) and uses them in two ways: as a scalability test ("the same model serves a throwaway testbed, a layered starter engine, and unique, very complex integrations") and in the definition of done ("testbed, starter-engine and custom-integration tiers each have a worked example journey"). It also lists "the tiers" in the spec's product overview.

In the model as decided, a tier has no behavior. Growth is extending ([ADR-0002](0002-core-job-is-extend.md)); the CLI treats a three-slot composition and a forty-slot one the same way; private and proprietary libraries are ordinary libraries ([ADR-0013](0013-developer-code-is-a-library.md)); and libraries are swappable at the seam contracts in every composition ([ADR-0012](0012-slots-have-no-automap-api.md)).

## Options

1. **Tier as a declared property:** the manifest declares a tier and the CLI checks or gates on it.
2. **Tier as a descriptive name:** an undeclared label for how far a composition has grown.
3. **No tier term:** keep "testbed" as a term; use "starter engine" and "custom integration" only as names of worked example journeys in the spec.

## Decision

Option 3. "Tier" is not a domain term. "Testbed" stays in the glossary because the day-one minimum starts one ([ADR-0010](0010-day-one-minimum-start-then-extend-once.md)). "Starter engine" and "custom integration" survive only as the names of two worked example journeys in the spec. The three journeys (testbed, starter engine, custom integration) meet the brief's scalability check and definition of done, and the spec's product overview describes those journeys.

## Consequences

- Nothing in the manifest, lock or CLI refers to a tier; there are no tier flags or tier-gated features.
- The spec's product overview presents three worked example journeys in place of tiers.
