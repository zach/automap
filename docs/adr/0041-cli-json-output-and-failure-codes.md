---
status: accepted
date: 2026-10-29
---

# Every CLI command has schema'd JSON output and stable failure codes

## Context

The CLI is designed for agents first ([ADR-0006](0006-product-shape-map-cli-router-skill.md)), so its output format is the core of the agent interface. Research ([06](../research/06-verification.md), [04](../research/04-agent-ecosystem.md)): Anthropic's parallel-agent C compiler project found that short, greppable test output, with details logged to files, decided whether agents made progress; vcpkg's agent skills key on stable failure keywords such as `BUILD_FAILED` and `POST_BUILD_CHECKS_FAILED`; first-party guidance recommending a CLI over MCP assumes machine-readable output.

## Options

1. **Human-readable output only.**
2. **JSON output with a published schema, and a stable, named failure vocabulary.**

## Decision

Option 2. Every CLI command supports a JSON output mode with a published schema. Failures use stable, named **failure codes**, such as `DEFINE_MISMATCH(JPH_DOUBLE_PRECISION)`, `INIT_ORDER(jolt before jobs)` or `GOLDEN_DIFF`. Each failure is a one-line verdict; full details are written to files.

## Consequences

- Failure codes are part of the CLI's public interface and are versioned with it.
- Agents can act on failures without parsing prose.
