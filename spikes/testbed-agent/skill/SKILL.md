---
name: automap
description: Compose, extend and verify a C++ game engine built from libraries, using Automap's map of slots, contracts, pairings and recipes. Use whenever the user asks to start an engine or testbed, add or swap a library, or check that the engine's libraries are wired together correctly.
---

# Automap router skill (DRAFT, spike)

Automap describes an engine as a **composition**: which library fills each **slot**, and where the **glue** for each **binding** (a library connected to one of Automap's per-seam **contracts**) and each **pairing** (two libraries integrated directly) lives. The **manifest** declares it; the **lock** pins it; **probes** decide whether glue works; **evidence** records what passed.

## Which job?

| The user wants to | Job |
| --- | --- |
| a new engine or testbed | `automap start <slot>=<library> ...` |
| add or swap a library | `automap extend <slot>=<library>` |
| know whether the engine is wired correctly | `automap verify` |

In this spike there is no CLI binary: perform the job's steps from `CLI-STEPS.md` by hand, exactly as written.

## How to work

1. Read the manifest and lock if they exist. Never guess a library's version: the lock says what is in use.
2. Read the map entries the job touches: the libraries' facts, the contracts, and the recipes. Facts and recipes are tied to library versions; use the ones matching the lock.
3. Start from reference glue. Adapt it where the engine's existing glue already does the job; the probes, not resemblance to the reference, decide whether glue works.
4. Run verification after every change and act on failure codes: each names a rule and a library.
5. Report the evidence level reached and any failure codes, in one line.

## Never

1. Never edit the lock by hand.
2. Never create or claim evidence except through verification.
3. Never weaken, skip or delete a probe to get a pass. Fix the glue or report the failure.
4. Never open an issue or pull request on an upstream project whose contribution policy bans AI-generated contributions (see each library's `contribution_policy`).
