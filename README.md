# Automap

Automap is an agent-ready map and CLI for composing custom C++ game engines from libraries. A developer's coding agent uses it to start an engine or add a library to one, with seam glue the developer owns and evidence that the result works.

The **map** describes libraries, the slots they fill (windowing, physics, jobs, allocator, …), one contract per seam between libraries (allocation, jobs, math types, lifecycle, diagnostics, …), and recipes that bind each library to those contracts with reference glue and probes. The **CLI** writes a composition into the developer's repo (a manifest and lock, the vcpkg and CMake files, and the glue), builds it, runs the probes, and records the outcome as evidence. Nothing from Automap is linked into the engine, and `automap eject` leaves a plain CMake and vcpkg project.

## Status

Pre-release, building toward 0.1.0. The public map covers thirteen libraries on Linux, Windows and macOS: SDL3, Dear ImGui, Jolt, enkiTS, mimalloc, spdlog, Tracy, GLM, flecs, miniaudio, fastgltf, stb_image and SDL_shadercross. Automap's CI verifies every binding and pairing on all three platforms and publishes signed evidence in the map.

## Quick start

You need CMake 3.25+, Ninja, a C++20 compiler (clang on Linux, MSVC on Windows, Apple Clang on macOS) and Go 1.24 to build the CLI. vcpkg is optional: the CLI uses the one at `VCPKG_ROOT`, or bootstraps the release the map names into its cache.

```console
$ go install github.com/zach/automap/cmd/automap@latest
$ mkdir my-engine && cd my-engine
$ automap start --with sdl3 --with dear-imgui:docking --json
$ automap extend jolt --json
```

`start` writes the composition and a testbed executable and verifies it on your platform; `extend` adds a library to the slots it fills and verifies again. Every command takes `--json` and prints one result: a summary, any failing rules with their failure codes, the files it wrote and the evidence file. `start` also writes an `AGENTS.md` section and a router skill, so a coding agent in the repo knows when to run which command.

| Command | Does |
| --- | --- |
| `automap start --with <library>[:<variant>]...` | Creates a composition in the current directory, then verifies it. |
| `automap extend <library>[:<variant>]` | Adds a library, then verifies. |
| `automap upgrade [<library>...]` | Moves to a newer map release, re-resolves, verifies. |
| `automap adopt` | Brings an existing CMake and vcpkg project under Automap. |
| `automap verify` | Builds, runs the smoke test and probes, writes evidence. |
| `automap find [<query>] [--slot <slot>]`, `automap show <kind> <id>` | Search and inspect the map, with evidence states. |
| `automap fetch`, `automap vendor` | Fill the caches for offline work; copy everything a build needs into the repo. |
| `automap eject` | Leaves Automap: a plain CMake and vcpkg project remains. |

Run `automap` with no arguments for the full usage.

## How it works

- A **composition** is a manifest (the libraries, their slots, variants and choices) and a lock (exact pins, canonical JSON, written only by the CLI).
- **Glue** is ordinary source in the developer's repo. The CLI writes each recipe's reference glue; the developer or their agent may rewrite it. Success is judged by the probes, not by resemblance to the reference.
- **Verification** builds the composition, runs a headless smoke test, then a probe build that checks every rule of every contract in play. The outcome is a signed in-toto statement at a level: builds, smoke or probed.
- The **public map** is released as immutable, dated releases; evidence for exact pins is read live from the map's main branch. Private maps layer over it for in-house and patched libraries.

## Repository

| Path | What lives there |
| --- | --- |
| `cmd/automap/`, `internal/` | The CLI and its tests. |
| `map/` | The public map: libraries, slots, contracts, recipes, probe harness and evidence (MIT-0). |
| `ci/` | Automap's CI for the map: verification, the Automap registry's ports, the watcher and curation agents, the reference compositions. |
| `evals/` | The eval suite: fresh agent sessions doing Automap's jobs, graded through the CLI. |
| `schemas/` | The published JSON Schemas. |
| `docs/spec/SPEC.md` | The product spec; `docs/adr/` holds the decisions and `CONTEXT.md` the glossary. |

[`AGENTS.md`](AGENTS.md) has the build and test commands and the working rules; the repo is built by coding agents, ticket by ticket, from [`docs/issues/automap-v1/`](docs/issues/automap-v1/).

## License

MIT. Everything in `map/` is MIT-0, so glue copied into your repo carries no attribution requirement.
