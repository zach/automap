---
status: accepted
date: 2026-10-30
---

# The build's tests attach at the CLI, run as a process

## Context

The build's test strategy has four layers: test-driven Go tests of CLI logic, golden tests of its JSON output, the map testing itself, and the eval suite as acceptance ([ADR-0068](0068-test-strategy-four-layers.md)). Tests need a seam to attach to. Agents use Automap only through the CLI and its JSON output ([ADR-0041](0041-cli-json-output-and-failure-codes.md)), and the CLI's real work, `verify`, drives CMake and vcpkg to build and run C++ ([ADR-0034](0034-vcpkg-is-the-v1-build-backend.md), [ADR-0035](0035-glue-targets-cmake-only.md)). Building the real catalog takes minutes per composition.

## Options

1. **Many internal seams:** unit tests against each Go package, with the build backend faked.
2. **One high seam:** run the CLI as a process against fixture maps and fixture repos, with real builds of tiny stub libraries.

## Decision

Option 2.

- The main test seam is the CLI run as a process: a test runs `automap <command> --json` against a fixture map and a fixture repo, then checks the JSON result and the files written. Layers 1 and 2 of the test strategy both sit there.
- The fixture map holds two tiny stub C++ libraries with their own recipes, probes and seeded violations, so `verify` runs real CMake and vcpkg builds in seconds.
- Go tests below the CLI cover only pure functions with no I/O, such as lock canonicalization and resolution.
- Acceptance stays with the eval suite.

## Consequences

- Tests exercise the same interface agents use.
- Builds of the real catalog's libraries run in Automap's CI and the eval suite, not in the CLI's tests.
