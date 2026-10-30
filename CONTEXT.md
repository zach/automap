# Automap

Automap is a curated, machine-readable map for composing a custom C++ game engine from many libraries, used by game developers and their coding agents to assemble, extend and maintain that engine.

## Language

### People

**Framework-first indie**:
The primary persona: a solo developer or small team who builds and ships games on their own engine assembled from libraries, with a coding agent as the steward of its code. Reading the C++ is optional.
_Avoid_: user, customer, engine developer

**Steward**:
Whoever writes and maintains an engine's code day to day. By default the developer's coding agent; the human sets direction.
_Avoid_: maintainer, owner, author

### Jobs

**Extend**:
The core job: add or swap a library in an engine, with seam glue the developer owns and evidence that the resulting pairings work.
_Avoid_: install, integrate, add package

**Start**:
Extending an empty composition: the first libraries added to a new engine, typically producing a testbed.
_Avoid_: scaffold, generate, bootstrap, new project

**Upgrade**:
Moving a composition's libraries to newer versions while keeping its glue and evidence valid.
_Avoid_: update, bump

**Adopt**:
Describing an existing engine that Automap did not create as a composition, so it can be extended like any other.
_Avoid_: import, onboard, migrate

**Audit**:
Reporting what a composition ships: licenses, provenance and versions. Comes after v1, beyond license warnings at extend time.
_Avoid_: compliance check, scan

**Port**:
Bringing a composition to a new platform. Comes after v1.
_Avoid_: migrate, cross-compile

**Eject**:
Leaving Automap: one CLI step that vendors any Automap registry ports into the developer's repo and removes the manifest and lock, leaving a plain CMake and vcpkg project.
_Avoid_: uninstall, export, detach

### Product

**Map**:
Versioned data describing slots, libraries, contracts, pairings and recipes, plus the evidence behind them. The public map is Automap's; private maps layer over it.
_Avoid_: catalog, database, registry, index

**Map release**:
An immutable, dated release of the public map. Every lock records the map release it was resolved against.
_Avoid_: snapshot, baseline (a vcpkg term), map version

**Private map**:
A map with the public map's schema, kept in the developer's repo or another repo they control and layered over the public map by the CLI. It holds in-house and proprietary libraries, private variants (including patched libraries), recipes and evidence, and is never uploaded. A private map can extend a library from a map beneath it by adding variants, and an NDA platform's private map can also add its platform to a library's supported targets.
_Avoid_: overlay (a vcpkg term), private registry, fork

**CLI**:
Automap's command-line tool: the one deterministic way to query the map and act on a composition, designed for agents first.
_Avoid_: tool, client, engine

**Failure code**:
A stable, named identifier for one kind of failure the CLI reports, such as `DEFINE_MISMATCH(JPH_DOUBLE_PRECISION)` or `INIT_ORDER(jolt before jobs)`.
_Avoid_: error message, exit code

**Build backend**:
An existing package manager that Automap delegates fetching and building to, such as vcpkg.
_Avoid_: backend (ambiguous with rendering and platform backends), package manager (when naming this role)

**Watcher**:
Automap's process that notices upstream releases, tags and commits for every library in the public map.
_Avoid_: bot, crawler, monitor

**Curation agent**:
An agent that updates the public map's facts, recipes and probes. Nothing it writes enters the map without evidence.
_Avoid_: curator (humans decide coverage), maintainer bot

**Eval suite**:
Automap's CI suite in which fresh agent sessions carry out the v1 jobs along the worked journeys, on every map release. The source of every success measure.
_Avoid_: benchmark, test suite, telemetry

**Fixture map**:
A small map of stub libraries, with their own recipes, probes and seeded violations, that the CLI's own tests run against so builds take seconds.
_Avoid_: mock map, test map

**Automap registry**:
Automap's own vcpkg registry of overlay ports, for variants vcpkg can't express, libraries vcpkg lacks, and fresh commits of libraries vcpkg packages stale.
_Avoid_: the map (the map describes; the registry builds), overlay, fork

**Router skill**:
The single agent skill that teaches a coding agent to drive the CLI. It ships inside the CLI, which writes it into the developer's repo at its own version.
_Avoid_: main skill, entry skill, meta-skill

### The map

**Slot**:
A named role in an engine, such as windowing, physics-3d or audio, that a library fills. A slot has no Automap-defined API: the engine calls the filling library's own API. In a composition each slot holds at most one library.
_Avoid_: category, module, component, layer

**Library**:
A body of code that can fill one or more slots, whether third-party, proprietary or the developer's own. The developer's own code is a library like any other; its source may be a directory in their repo.
_Avoid_: package, port, dependency (those name a build backend's artifacts, not the thing being composed)

**Contribution policy**:
A library's upstream rule on AI-generated contributions (banned, restricted, allowed or unknown), recorded as a fact. The agent never opens an issue or pull request upstream where it is banned.
_Avoid_: contributing guide, CoC

**Seam**:
A kind of point where libraries, or a library and the engine around it, must agree to work together. Automap's seams are allocation, jobs, math types, handles, lifecycle, diagnostics, serialization, build flags, platform, determinism and assets.
_Avoid_: integration point, interface, boundary

**Contract**:
Automap's definition of what must hold at one seam: the facts libraries declare, the choices compositions make, the rules over them, and optionally a shared interface. Automap owns one contract per seam, and each is versioned independently.
_Avoid_: interface, API, spec, requirement

**Fact**:
Something a library declares about one seam, such as its allocator hook style or its ABI-affecting defines. Facts live in the map and can differ between a library's versions and variants.
_Avoid_: metadata, property, attribute

**Choice**:
A decision a composition makes at one seam, such as which engine math library to use or whether determinism is on. Choices live in the manifest.
_Avoid_: setting, option, config

**Phase**:
One of the six fixed stages of an engine's startup, in order: memory, diagnostics, platform, jobs, systems, game. Each library's init steps belong to a phase; shutdown runs in exact reverse.
_Avoid_: stage, level, loading phase, init order

**Ordering constraint**:
A requirement in a pairing's recipe that one lifecycle step run before another, across the two libraries, at init or shutdown. The lifecycle probe checks it.
_Avoid_: dependency (used for `after` within a library), sequencing

**Choices header**:
The header the CLI writes from the manifest, committed with the build files, through which glue reads the composition's choices instead of hard-coding them.
_Avoid_: config header, generated config

**Shared interface**:
An interface Automap delivers as glue source: the allocator facade, the jobs interface, the diagnostics facade (log, assert and profile) and the assets file-I/O interface, plus the probe harness. Contracts define all of them except the probe harness. Never an Automap library the engine links.
_Avoid_: adapter, abstraction layer, API

**Reference thread pool**:
The small scheduler in the jobs interface's setup glue that runs jobs when no library fills the jobs slot. Its worker count comes from the composition's choices.
_Avoid_: default scheduler, built-in job system

**Rule**:
One requirement that glue must meet, stated by a contract or by a pairing's recipe. A rule is either probed (it has a probe, and passing it counts as evidence) or guidance (labeled, and never counts as evidence). A probed rule may name the platforms its probe can run on; elsewhere its result is not-applicable.
_Avoid_: constraint, check, requirement

**Binding**:
The connection of one library to one contract, made by glue that follows a recipe and backed by evidence.
_Avoid_: adapter, conformance, implementation

**Pairing**:
A direct integration between exactly two libraries that doesn't go through a contract, such as Dear ImGui's SDL3 backend, backed by evidence that it works.
_Avoid_: compatibility, combo, integration

**Glue**:
Code in the developer's repo that wires libraries together at their seams. The developer owns it, whoever wrote it.
_Avoid_: boilerplate, integration code, wrapper

**Recipe**:
What Automap provides for one binding, pairing or library: reference glue (including build fragments and setup glue), the probes that check any glue doing the same job, and notes on adapting the reference glue.
_Avoid_: template, snippet, tutorial, guide

**Reference glue**:
Glue that Automap has verified for a binding or pairing. The CLI writes it into new compositions; agents adapt it where glue already exists.
_Avoid_: template, boilerplate, sample code

**Build fragment**:
The part of a composition's build file that adds one library: how it is fetched, its options, its targets, and any flags its consumers need. Part of every library recipe's reference glue.
_Avoid_: CMake snippet, build recipe

**Setup glue**:
Reference glue that creates the minimal objects the engine needs to use a library, or a shared interface from the composition's choices. Never game logic.
_Avoid_: boilerplate, scaffolding, sample code

**Entry point**:
A call in a recipe's glue that the composition's startup makes in one phase, at init or at shutdown.
_Avoid_: hook (a library's own), callback, init function

**Probe harness**:
The shared interface every composition includes, through which probes record their results for the CLI and glue records each lifecycle step it runs.
_Avoid_: test framework, test runner

**Probe**:
An executable check that glue satisfies one rule, such as "allocator hook installed before first allocation". Probes, not resemblance to reference glue, decide whether glue works. A probe counts only once it has been shown to fail on a seeded violation.
_Avoid_: test (too broad), assertion, validator

**Seeded violation**:
A known mistake Automap's CI introduces on purpose to show that a probe reports failure.
_Avoid_: mutation (too general), negative test

**Evidence**:
A record that one binding, pairing or composition passed verification at a stated evidence level on one platform. Automap's published evidence is signed and stored in the map; evidence for a developer's own composition lives in their repo beside the lock.
_Avoid_: proof, certification, badge, test result

**Result status**:
How one rule fared in a piece of evidence: pass, fail, not-applicable (with a reason) or unverifiable (with a reason).
_Avoid_: outcome, verdict

**Run conditions**:
What a piece of evidence was produced under: build type, compiler version, environment variables and sanitizer options.
_Avoid_: environment, settings, config

**Set-aside report**:
A sanitizer report whose every frame is in code the composition doesn't build, such as an uninstrumented GPU driver. It is recorded in the evidence but doesn't fail the rule.
_Avoid_: suppression, ignored report, false positive

**Evidence level**:
How much a piece of evidence shows, from least to most: builds, smoke, probed, behavioral, hardware. Every claim states the highest level it passed. Probed requires every applicable probed rule to pass, with none unverifiable.
_Avoid_: tier, grade, rung

**Evidence state**:
Whether evidence can be relied on: current (produced within the last 30 days), stale (older, with a re-run queued) or broken (the latest run failed; re-checked on schedule).
_Avoid_: valid, expired, outdated

**Known-good**:
Backed by current evidence at the probed level or higher.
_Avoid_: verified, certified, compatible, tested

**Platform**:
An OS, a CPU architecture and a toolchain, such as Windows x64 with MSVC 2026. Evidence is always for exactly one platform.
_Avoid_: target, triplet (a vcpkg term), OS

**NDA platform**:
A platform whose SDKs and APIs are available only under its holder's non-disclosure agreement, such as a console. Its content lives in a private map of its own.
_Avoid_: closed platform, console (not every NDA platform is one)

**Cooldown**:
The minimum age, 7 days, an upstream release must reach before it can enter the map. Security fixes may skip it; nothing skips the evidence.
_Avoid_: quarantine, embargo, delay

**Composition**:
One engine as Automap describes it: which library fills each slot, and where the glue for each binding and pairing lives. It consists of a manifest and a lock in the developer's repo.
_Avoid_: stack, project, build, configuration

**Reference composition**:
A composition Automap's CI maintains and verifies on every platform, such as the testbed.
_Avoid_: sample project, template, demo

**Manifest**:
The file in the developer's repo, edited by the agent, that declares a composition: the library in each slot, the composition's choices, and where each binding's and pairing's glue lives.
_Avoid_: config, project file; don't use "manifest" alone for a build backend's file (name it, e.g. `vcpkg.json`)

**Lock**:
The file the CLI writes beside the manifest, holding the exact pins the composition resolved to and the map release it was resolved against.
_Avoid_: lockfile (for a build backend's own), snapshot, baseline

**Pin**:
The exact resolved identity of one library in a lock: the library, its version (or commit, for libraries without release tags), its variant, the registry it came from at a commit, and a content hash of its source. A library whose variant differs on some platform has a pin for each such platform.
_Avoid_: version (a pin is more than a version), reference

**Variant**:
A named set of build options for a library, such as "Jolt, double precision", or a build from different source, such as a patched fork or a console port, with its own pin. Bindings, pairings and evidence are per variant, and every library has at least one. A composition chooses each library's variant once or per platform.
_Avoid_: feature (a vcpkg term), option, configuration, flavor

**Console variant**:
A variant of a library for an NDA platform, distributed by its maintainers to that platform's licensees. The public map says only that it exists and who distributes it; the variant itself lives in that platform's private map.
_Avoid_: console port (fine in prose, but not as the term), platform fork

**Testbed**:
A small composition that opens a window and runs a render and input loop, made for experiments. Starting a new engine typically produces one.
_Avoid_: sandbox, demo, prototype, hello world, tier

**Journey**:
One of the spec's three worked examples of the jobs in sequence: testbed, starter engine and custom integration. A journey is an example, not a kind of composition.
_Avoid_: tier, use case, scenario
