---
name: automap
description: Compose, extend and verify a C++ game engine built from libraries, with Automap's map of slots, contracts, pairings and recipes and its automap CLI. Use whenever the user asks to start an engine or testbed, add or swap a library, or check that the engine's libraries are wired together correctly.
metadata:
  automap-cli-version: "0.0.0-dev"
---

# Automap

Automap describes an engine as a **composition**: which library fills each **slot**, and where the **glue** lives for each **binding** (a library connected to one of Automap's contracts) and each **pairing** (two libraries integrated directly). The **manifest** (`automap.manifest.json`) declares it; the **lock** (`automap.lock.json`) pins it; **probes** decide whether glue works; **evidence** records what passed. This skill was written by automap 0.0.0-dev; the CLI rewrites it when a newer CLI runs.

Always pass `--json` and act on the result: `ok`, a one-line `summary`, the `level` reached, and `failures`, each with a failure `code` and a `details_file`.

## Which command

| The user wants to | Run |
| --- | --- |
| a new testbed with some libraries | `automap start --with <library>[:<variant>]... --json` |
| to know whether the engine is wired correctly | `automap verify --json` |
| to know which libraries fill a slot | `automap find --slot <slot> --json` |
| a library's facts, hooks and variants | `automap show library <id> --json` |

`start` writes the manifest, lock, vcpkg files, build fragments, glue and startup, then verifies on this machine. Each library fills every slot it can unless `--slot <slot>=<library>` says otherwise. A library's variant comes after a colon, such as `dear-imgui:docking`.

## Reading a verify result

The `level` is the highest one reached: `none` (didn't build), `builds`, `smoke` (ran 120 headless frames and shut down cleanly) or `probed` (every applicable probed rule passed). Aim for `probed`.

- `BUILD_FAILED`: read the details file, the build log; fix the build.
- `SMOKE_FAILED`: the executable crashed or didn't shut down cleanly. `INIT_FAILED(<call>)` names the startup call that failed.
- A rule's failure code, such as `ALLOC_HOOK_LATE(<library>)`, `INIT_ORDER(<step> before <step>)` or `ABI_DEFINE_MISMATCH(<define>)`, names the rule and library. Read the details file, then fix the glue it points at.
- An `UNVERIFIABLE(<rule>)` warning means no probe could check that rule; it keeps the composition below `probed`.

## Adapting glue

Glue under `glue/` and the build fragments under `cmake/automap/` are the developer's: read and change them freely, starting from the recipe's notes (`automap show recipe <id> --json`). The probes, not resemblance to the reference glue, decide whether glue works. Never edit the files the CLI writes from the manifest: the choices header and the startup.

## Never

1. Never edit the lock by hand: only the CLI writes automap.lock.json.
2. Never create or claim evidence except by running the CLI.
3. Never weaken, skip or delete a probe to get a pass: fix the glue, or report the failure.
4. Never open an issue or pull request on an upstream project whose contribution policy bans AI-generated contributions (each library's contribution_policy in the map).

