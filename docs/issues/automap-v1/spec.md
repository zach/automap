# Automap v1

**Status:** ready-for-agent

The product-wide spec is [`docs/spec/SPEC.md`](../../spec/SPEC.md); this file is the feature spec the tickets in this directory are cut from. Terms are defined in [`CONTEXT.md`](../../../CONTEXT.md).

## Problem Statement

A framework-first indie builds games on their own C++ engine, assembled from libraries, with a coding agent as the engine's primary steward. Every library brings its own allocator hooks, scheduler, math types, init order, logging and assert hooks, and ABI-affecting defines. Package managers fetch and build the libraries but know nothing about these seams, so each developer's agent writes the glue from scratch, guided by READMEs and examples, and nothing checks the result. Mistakes such as installing an allocator hook after the library's first allocation, letting a physics library start its own threads, shutting a GPU backend down before the device is idle, or mismatching an ABI define compile fine and fail later, intermittently, or only on another platform. Nobody publishes durable evidence that two libraries work together, and links to CI runs expire. The developer can't tell whether their engine is sound without reading C++ they may not want to read, and their agent has nothing to check its work against.

## Solution

Automap gives the developer's agent a map of the C++ engine ecosystem and a CLI that acts on it. The map describes slots, libraries, one contract per seam, and recipes that bind each library to the contracts, each with reference glue and probes. When the developer asks for a testbed, or to add a library, the agent runs the CLI. The CLI writes the composition's manifest and lock, the vcpkg and CMake files, and the glue into the developer's repo, then builds the engine and runs probes that check every rule. It records the outcome as evidence at a stated level, signed when produced in CI. The developer owns everything written; nothing from Automap is linked into the engine, and they can eject at any time. Automap's own CI verifies every binding and pairing on three desktop platforms and publishes signed evidence in the map, so the agent can tell which combinations are known-good before it starts. Private maps extend the same model to in-house libraries, patched libraries and NDA console platforms, with no special cases.

## User Stories

### Starting and extending

1. As a framework-first indie, I want to ask my agent for a C++ testbed with SDL3 and Dear ImGui in one sentence, so that I have a running window with a UI panel without assembling build files myself.
2. As a framework-first indie, I want the testbed to come with evidence that its integrations were checked, so that I can trust it without reading the glue.
3. As a developer's agent, I want one CLI command that creates a composition from a list of libraries, so that starting is deterministic rather than improvised.
4. As a developer's agent, I want the CLI to assign a library to every slot it can fill unless told otherwise, so that SDL3 fills windowing, input and GPU in one step.
5. As a developer's agent, I want to name a variant when I add a library, so that I get Dear ImGui's docking branch or Jolt's double precision deliberately.
6. As a framework-first indie, I want to add 3D physics to my testbed with one request, so that extending feels as easy as starting.
7. As a developer's agent, I want extending to write the new library's build fragment, bindings and setup glue and update the choices header, so that the library is wired into every seam it touches.
8. As a developer's agent, I want the CLI to generate the startup that calls every recipe's entry points in order, so that I never work out init order by hand.
9. As a developer's agent, I want extending to add any shared interface the new library needs, such as the jobs interface with its reference thread pool, so that one extend step adds exactly one library.
10. As a developer's agent, I want to search the map for libraries that fill a slot, with their evidence state on my platforms, so that I pick a known-good option.
11. As a developer's agent, I want to inspect any map entry with its facts and evidence, so that I understand a library's hooks before adapting glue.
12. As a framework-first indie, I want a warning before a copyleft or proprietary library enters my engine, with the reason, so that I don't ship something I can't license.
13. As a developer's agent, I want every command to have a JSON output mode with a published schema, so that I parse results instead of scraping text.
14. As a developer's agent, I want each failure as a stable, named failure code with a one-line verdict and a details file, so that I can act on it without reading logs first.
15. As a developer's agent, I want the CLI to update the build files through each library's build fragment, so that the build stays correct as libraries are added.
16. As a framework-first indie, I want every build file the CLI writes to be an ordinary committed file, so that I can read, edit and keep them.

### Verifying

17. As a developer's agent, I want one command that builds the composition, runs it headless for a set number of frames, runs every probe and writes evidence, so that I know whether my changes broke anything.
18. As a developer's agent, I want verification to say the highest level reached (builds, smoke, probed), so that I know how far the composition got.
19. As a developer's agent, I want each rule's result marked pass, fail, not-applicable or unverifiable, with reasons, so that I know exactly what was checked.
20. As a developer's agent, I want a ThreadSanitizer run on platforms where it exists, ignoring reports wholly inside uninstrumented drivers but recording them, so that real races in glue fail and driver noise doesn't.
21. As a developer's agent, I want the lifecycle probe to check init and shutdown order across all six phases and every pairing's ordering constraints, so that order mistakes fail at startup instead of crashing at exit.
22. As a developer's agent, I want the build-flags probes to catch mismatched ABI defines and disagreeing C++ standard, RTTI, exceptions or C runtime, so that ODR and ABI bugs fail loudly.
23. As a developer's agent, I want evidence recorded with its run conditions (build type, compiler, environment, sanitizers), so that it can be reproduced.
24. As a framework-first indie, I want evidence produced in my CI signed with my CI's identity, and local evidence marked unsigned, so that I know which records to trust.
25. As a framework-first indie, I want my CI to verify my engine on Windows, Linux and macOS, so that I find platform-specific breakage early.
26. As a developer's agent, I want the router skill to carry a reference CI workflow for multi-platform verification, so that I can set it up without inventing one.

### Trust in the map

27. As a developer's agent, I want every binding and pairing in the map to carry signed evidence from Automap's CI per platform, so that I can tell known-good from untested.
28. As a developer's agent, I want evidence marked current, stale or broken, so that I don't rely on claims older than 30 days or on runs that failed.
29. As a framework-first indie, I want a new library release to enter the map only after 7 days and known-good evidence, so that my engine doesn't pick up a broken release on day one.
30. As a framework-first indie, I want security releases to skip the waiting period but never the evidence, so that urgent fixes arrive fast and still checked.
31. As a framework-first indie, I want every probe to have been shown to fail on a seeded violation, so that "probed" means a check that can catch the mistake actually ran.
32. As a developer's agent, I want the map released as immutable, dated map releases recorded in my lock, so that my composition resolves the same way months later.

### Upgrading

33. As a framework-first indie, I want my agent to upgrade one library, or everything, to the newest map release with one command, so that staying current is cheap.
34. As a developer's agent, I want upgrade to apply changed reference glue where my glue still matches the old reference, and report where it doesn't, so that I adapt only what needs it.
35. As a developer's agent, I want upgrade to end with verification and name the failing rules, so that I fix glue against concrete failure codes.
36. As a framework-first indie, I want nothing to change in my engine until my own agent runs an upgrade, so that I control when change happens.

### Adopting an existing engine

37. As a studio with a hand-built engine, I want my agent to adopt it into Automap, so that my existing engine gets contracts, probes and evidence without being rewritten.
38. As a developer's agent, I want adopt to match my CMake targets to map libraries and report the ones no map describes, so that I know what to describe in a private map.
39. As a studio, I want my in-house scheduler and renderer described as ordinary libraries in a private map, so that they are checked by the same probes as Jolt and SDL3.
40. As a developer's agent, I want adopt to write a manifest that points at the glue that already exists, so that I adapt it in place rather than replacing it.
41. As a studio, I want success judged by probes passing, not by resemblance to reference glue, so that my existing architecture is respected.
42. As a studio, I want an in-house library joined to the build as ordinary CMake targets and pinned by its path and content hash, so that nothing about it needs a package registry.

### Private and NDA content

43. As a studio, I want a private map with the public map's schema, in a repo I control, that the CLI never uploads, so that proprietary content stays private.
44. As a studio, I want to carry a patched library as a private variant with its own pin and evidence, so that a fix I can't send upstream is still tracked and checked.
45. As a developer's agent, I want a private map to extend a public library only by adding versions, variants and (for NDA platforms) supported targets, so that private content can't silently change public facts.
46. As a studio shipping on a console, I want that console's content in a private map of its own, so that staff without that NDA see nothing of it.
47. As a studio, I want to choose a library's variant per platform, such as SDL's console port on the console and the public source elsewhere, so that one composition covers every platform I ship on.
48. As a studio, I want evidence for the console produced on my devkits, signed with my own key and never written to a public transparency log, so that NDA platform names never leak.

### Ejecting and working offline

49. As a framework-first indie, I want to eject in one step and be left with a plain CMake and vcpkg project that builds and passes its smoke test, so that I'm never locked in.
50. As a framework-first indie, I want composing, building and verifying to work offline once caches are filled, so that I can work on a plane.
51. As a framework-first indie, I want to vendor everything my composition needs into my repo, so that builds are fully reproducible without the network.

### The agent's instructions

52. As a developer's agent, I want an Automap section in the repo's `AGENTS.md`, between markers, holding the build, test and run commands, the manifest and lock locations, the slots in use, and the instruction to go through the router skill and CLI, so that every session starts oriented.
53. As a framework-first indie, I want the rest of my `AGENTS.md` left untouched, and a `CLAUDE.md` import line added only if I have a `CLAUDE.md`, so that my own instructions stay mine.
54. As a developer's agent, I want the router skill written into every place my agents read skills from, at the CLI's version, so that its guidance always matches the CLI.
55. As a framework-first indie, I want my agent held to four nevers (no hand-edited lock, no fabricated evidence, no weakened probes, no upstream contributions to projects that ban them), with the CLI refusing where it can detect a violation, so that the evidence stays honest.
56. As a developer's agent, I want the CLI to refuse to act on a lock whose canonical hash doesn't match, so that a hand edit is caught immediately.

### Automap's own operation

57. As Automap's maintainer, I want a watcher to notice upstream releases and curation agents to update facts, recipes and probes, so that the map stays fresh without human toil.
58. As Automap's maintainer, I want Automap's CI to verify every binding, every pairing with a recipe, and the reference compositions, for each library's latest and previous version and each variant, on every input change and at least every 30 days, so that evidence is always current.
59. As a contributor, I want to propose facts, recipes, probes or new versions as pull requests that pass the same gates as curation agents, so that anyone can improve the map.
60. As Automap's maintainer, I want humans to decide only what the map covers and contract major versions, so that agents do the rest.
61. As Automap's maintainer, I want an eval suite of fresh agent sessions doing the five jobs along the three journeys on every map release, so that success is measured without telemetry.
62. As Automap's maintainer, I want the four v1 measures (testbed in 15 minutes warm, 90% first-attempt success, freshness, 100% eject) computed from the evals and the map, so that the v1 release decision rests on evidence.

## Implementation Decisions

### Modules

- **Map model.** Loads a map directory and its metadata (public, then private maps in manifest order), validates every file against the published JSON Schemas, and applies the layering rules: a private file with an existing library id may add versions, variants and (for an NDA platform's map) supported targets, nothing else. Interface: load a set of maps; query slots, libraries, versions, variants with merged facts, contracts, recipes and evidence by id.
- **Resolver.** A pure function from a manifest and a loaded map to a lock: chooses each library's version (the newest in the map release; `upgrade` names the libraries it moves), per-platform variants, pins with content hashes and licenses, contract versions, private-map commits. Reports resolution failures as failure codes.
- **Lock codec.** Pure: RFC 8785 canonical encoding and `lock_hash`; detects hand edits.
- **Composition writer.** Writes and updates the manifest, the vcpkg manifest and configuration (curated registry at a baseline, the Automap registry, private overlay ports), the per-library build fragments and the include file the developer's CMake project includes once, the shared interfaces and probe harness, each recipe's reference glue and setup glue, the choices header, the startup (which calls every recipe's entry points in phase order with one context, checking that every context field read is provided earlier), and (for testbeds) the testbed's executable source. It never overwrites glue the manifest says the developer owns without the recipe's reference matching what is there.
- **Agent files writer.** Maintains the `AGENTS.md` section between markers, the `CLAUDE.md` import line, and the router skill in each agent's skill directory.
- **Verifier.** Configures and builds through CMake with vcpkg's toolchain, runs the composition's executable headless with `--probe`, reads the probe harness's results file (rule results and recorded lifecycle steps), checks lifecycle order and pairing ordering constraints, runs the separate ThreadSanitizer build where the rule's platforms include the host and applies the frame rule, computes the level, and writes evidence.
- **Evidence.** Builds in-toto v1 Statements with the Automap predicate, signs them through the CI's identity when available (GitHub artifact attestations), marks local runs unsigned, computes states (current, stale, broken) and fetches published evidence for a lock's pins.
- **Fetcher and cache.** Reads the map from a directory when told to (recording `git:<commit>` or `unreleased`), otherwise downloads map releases (archive plus checksum), registry ports, library sources (through vcpkg) and published evidence; serves everything offline afterwards; vendors into the repo on request.
- **Commands.** `start`, `extend`, `upgrade`, `adopt`, `eject`, `verify`, `fetch`, `vendor`, `find`, `show`, each with `--json` output conforming to the CLI result schema and exit statuses 0 (ok), 1 (rule or build failure), 2 (usage), 3 (refused).
- **Map content.** The eleven contracts at 1.0.0 with their shared interfaces (allocation facade, jobs interface with reference thread pool, diagnostics facade, assets file-I/O interface), the probe harness, the testbed's executable source, and the thirteen catalog libraries with library recipes, bindings and the two pairings (`dear-imgui--sdl3`, `sdl-shadercross--sdl3`).
- **Automap registry.** A vcpkg git registry on its own branch for variants vcpkg can't express (Jolt double precision and deterministic), missing libraries and fresh commits.
- **Automap CI.** Verifies declared edges on the three platforms (lavapipe on Linux, WARP on Windows, paravirtual Metal on macOS), runs each probe's seeded violation, signs evidence into the map, re-verifies on input change and every 30 days, cuts map releases, and runs the eval suite.
- **Watcher and curation agents.** Detect upstream releases and open map changes; the cooldown and evidence gate decides merging.
- **Eval suite.** Runs fresh agent sessions on the journey tasks and computes the four measures.

### Formats

The eight schemas and their examples are the contract between modules; they are in the product spec's data formats section and validated by its example checker. Notable rules the schemas can't express: lock canonical form and hash; private-map extension rules; per-platform pins exactly where variants differ; platform-limited rules are not-applicable elsewhere; the probed level requires no fail and no unverifiable among applicable results; every probed composition rule and every recipe probe carries a seeded violation as a patch.

### Architectural decisions

- vcpkg is the only build backend; glue and build files target CMake only and refer to libraries by package and imported target.
- Nothing from Automap is a runtime dependency; shared interfaces are glue source under MIT-0.
- Only fetching needs the network; private maps are never uploaded; no telemetry; no MCP server.
- The CLI is Go, one static binary per platform; the router skill is embedded in it.
- One repository; the registry on its own branch; tags `cli-v<semver>` and `map-YYYY.MM.DD`.

### Specific interactions

- `start` and `extend` end by running `verify` for the host platform and report its level.
- `upgrade` moves to the newest map release by default and ends with `verify`.
- `adopt` writes the manifest only once every target maps to a library in some map; until then it reports `UNMAPPED_TARGET(<target>)` for each.
- `eject` vendors Automap registry ports and removes the manifest and lock; it refuses if the lock was hand-edited.

## Testing Decisions

- **A good test checks external behaviour only:** what the CLI prints with `--json`, what files it writes, what it exits with, and what the built engine does under `--probe`. Tests never reach into Go internals except pure functions.
- **The main seam is the CLI run as a process** against a fixture map and a fixture repo. Golden tests compare JSON results and written files against expected outputs and the published schemas. Test-driven development happens at this seam.
- **The fixture map holds two tiny stub C++ libraries** with facts for the service contracts, recipes with reference glue, probes and seeded-violation patches, a pairing between them, and a library recipe each, so `verify` runs real CMake and vcpkg builds in seconds. Seeded-violation tests apply each patch and expect its failure code.
- **Go unit tests below the CLI** cover pure functions only: lock canonicalization and hashing, resolution, evidence level computation, layering rules.
- **The map tests itself** in Automap's CI: every recipe's probes on declared edges, every seeded violation, on every platform the rule runs on.
- **Acceptance is the eval suite:** each ticket that completes a slice names the eval task that must pass.
- **Prior art:** the product spec's example checker encodes the cross-file rules the map model must enforce; its examples become fixtures. The spike archived on the `spike/testbed-agent` branch shows working glue, a probe harness and the ThreadSanitizer procedure for the real testbed.

## Out of Scope

- Audit beyond extend-time license warnings, and porting to new platforms.
- Build backends other than vcpkg (Conan is considered after v1) and build systems other than CMake.
- Web platforms and a web-capable testbed; consoles in Automap's public CI.
- Godot and LibGodot extenders; consumers in languages other than C++.
- An MCP server, telemetry, commercial services.
- Engine-facing slot APIs.

## Further Notes

- The first three slices are fixed: start a testbed on Linux; extend with Jolt on Linux; three platforms and published evidence. Upgrade, the rest of the catalog, adopt with private maps, and eject follow.
- Human checkpoints: new or changed ADRs, map coverage and contract majors, and the v1 release. Everything else merges on evidence and evals.
- The product spec lists the details it fills in that no ADR decided; treat them as defaults, open to change until the build relies on them.
