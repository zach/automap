---
status: accepted
date: 2026-10-29
---

# v1 ships no MCP server

## Context

First-party guidance from Anthropic and GitHub prefers a CLI where one exists and reserves MCP for hosts without a shell; connected MCP servers cost context even when unused; and many MCP servers merely re-expose a CLI ([04](../research/04-agent-ecosystem.md)). The persona's agents (Claude Code, Codex, Copilot) all have shells. The CLI already has schema'd JSON output and stable failure codes ([ADR-0041](0041-cli-json-output-and-failure-codes.md)).

## Options

1. **An MCP server alongside the CLI in v1.**
2. **No MCP server in v1.**

## Decision

Option 2. v1 ships no MCP server. Every shell-capable agent drives the CLI directly. An MCP wrapper is reconsidered only if Automap needs to reach hosts without a shell.

## Consequences

- The CLI is the only agent-facing interface to keep in step with its schemas and failure codes.
