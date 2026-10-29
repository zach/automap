# Automap

Automap is a curated, machine-readable map for composing a custom C++ game engine from many libraries, used by game developers and their coding agents to assemble, extend and maintain that engine.

> Terms marked _(provisional)_ are not yet decided; Phase 2 branch 3 (domain model) sharpens them into schema-ready definitions. Open ambiguities are tracked in `docs/OPEN_QUESTIONS.md`.

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

### Product

**Map**:
Automap's versioned data describing slots, libraries, contracts, pairings and recipes, plus the evidence records behind them.
_Avoid_: catalog, database, registry, index

**CLI**:
Automap's command-line tool: the one deterministic way to query the map and act on a composition, designed for agents first.
_Avoid_: tool, client, engine

**Build backend**:
An existing package manager that Automap delegates fetching and building to, such as vcpkg.
_Avoid_: backend (ambiguous with rendering and platform backends), package manager (when naming this role)

**Router skill**:
The single agent skill that teaches a coding agent to drive the CLI.
_Avoid_: main skill, entry skill, meta-skill

### The map

**Slot**:
A named role in an engine, such as windowing, physics-3d or audio, that a library fills. A slot has no Automap-defined API: the engine calls the filling library's own API. In a composition each slot holds at most one library.
_Avoid_: category, module, component, layer

**Library**:
A body of code that can fill one or more slots, whether third-party, proprietary or the developer's own. The developer's own code is a library like any other; its source may be a directory in their repo.
_Avoid_: package, port, dependency (those name a build backend's artifacts, not the thing being composed)

**Seam**:
A kind of point where libraries, or a library and the engine around it, must agree to work together: for example allocation, jobs, math types, handles, lifecycle, errors, serialization, build flags or platform.
_Avoid_: integration point, interface, boundary

**Contract**:
Automap's definition of what must hold at one seam: the vocabulary and rules every library in a composition is bound to there. Automap owns one contract per seam.
_Avoid_: interface, API, spec, requirement

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
What Automap provides for one binding or pairing: reference glue, the probes that check any glue doing the same job, and notes on adapting the reference glue.
_Avoid_: template, snippet, tutorial, guide

**Reference glue**:
Glue that Automap has verified for a binding or pairing. The CLI writes it into new compositions; agents adapt it where glue already exists.
_Avoid_: template, boilerplate, sample code

**Probe**:
An executable check that glue does its job for a binding or pairing, such as "allocator hook installed before first allocation". Probes, not resemblance to reference glue, decide whether glue works.
_Avoid_: test (too broad), assertion, validator

**Composition**:
One engine as Automap describes it: which library fills each slot, and where the glue for each binding and pairing lives. It consists of a manifest and a lock in the developer's repo.
_Avoid_: stack, project, build, configuration

**Manifest**:
The file in the developer's repo, edited by the agent, that declares a composition: the library in each slot and where each binding's and pairing's glue lives.
_Avoid_: config, project file; don't use "manifest" alone for a build backend's file (name it, e.g. `vcpkg.json`)

**Lock**:
The file the CLI writes beside the manifest, holding the exact pins the composition resolved to.
_Avoid_: lockfile (for a build backend's own), snapshot, baseline

**Pin**:
The exact resolved identity of one library in a lock, including its variant.
_Avoid_: version (a pin is more than a version), reference

**Variant**:
A named set of build options for a library, such as "Jolt, double precision". Bindings, pairings and evidence are per variant, and every library has at least one.
_Avoid_: feature (a vcpkg term), option, configuration, flavor

**Testbed**:
A small composition that opens a window and runs a render and input loop, made for experiments. Starting a new engine typically produces one.
_Avoid_: sandbox, demo, prototype, hello world, tier
