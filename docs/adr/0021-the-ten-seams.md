---
status: accepted
date: 2026-10-29
---

# Automap's contracts cover ten seams

## Context

Automap owns one contract per seam ([ADR-0011](0011-automap-owns-one-contract-per-seam.md)), so the list of seams fixes the scope of the contract vocabulary. The research built a seams catalog from library headers and Godot's production Jolt glue ([02c](../research/02c-catalog-simulation-runtime-glue.md)) and found two seams the brief's list lacked: determinism and hot-reload boundaries ([05](../research/05-developer-needs-2026.md)). Libraries expose profiling through the same kind of hooks as logging and asserts (Jolt's external profiler, flecs's perf trace, enkiTS's profiler callbacks). Version coupling, such as Tracy's client–server protocol, is a constraint between pins, not a seam. Whether the asset pipeline is a seam is decided in branch 10.

## Options

1. **The brief's list:** allocation, threading and jobs, math types, handles and ownership, lifecycle and init order, error handling, serialization, build options, platform constraints.
2. **The brief's list extended:** add determinism, fold hot-reload boundaries into lifecycle, and widen error handling to diagnostics (logging, asserts, profiling).

## Decision

Option 2. Automap's contracts cover ten seams:

1. **Allocation:** who owns memory, and when each library's hooks are installed.
2. **Jobs:** threads and schedulers.
3. **Math types:** one engine math library, with adapters to each library's own types.
4. **Handles:** IDs, generations and ownership across libraries.
5. **Lifecycle:** init and shutdown order, including which libraries can live in a hot-reloadable module.
6. **Diagnostics:** logging, asserts and profiling hooks.
7. **Serialization:** data that libraries produce, and when it must be rebuilt.
8. **Build flags:** ABI-affecting defines, C++ standard, exceptions, RTTI and C runtime.
9. **Platform:** targets, main-thread rules, web and console constraints.
10. **Determinism:** an opt-in composition-wide profile.

## Consequences

- Version coupling between libraries is expressed as constraints between pins in the lock, not as a contract.
- Branch 10 may add an asset seam.
