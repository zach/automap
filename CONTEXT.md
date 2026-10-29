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

**Slot** _(provisional)_:
A named role in an engine that one library, or the developer's own code, fills: for example windowing, physics or audio.
_Avoid_: category, module, component, layer

**Library** _(provisional)_:
A third-party, in-house or proprietary body of code that can fill one or more slots.
_Avoid_: package, port, dependency (those name a build backend's artifacts, not the thing being composed)

**Seam**:
A kind of point where libraries, or a library and the engine around it, must agree to work together: for example allocation, jobs, math types, handles, lifecycle, errors, serialization, build flags or platform.
_Avoid_: integration point, interface, boundary

**Contract**:
Automap's definition of what must hold at one seam: the vocabulary and rules every library in a composition is bound to there. Automap owns one contract per seam.
_Avoid_: interface, API, spec, requirement

**Binding** _(provisional)_:
The connection of one library to one contract, made by a recipe and backed by evidence.
_Avoid_: adapter, conformance, implementation

**Pairing** _(provisional)_:
A direct integration between specific libraries that doesn't go through a contract, such as Dear ImGui's SDL3 backend, backed by evidence that it works.
_Avoid_: compatibility, combo, integration

**Glue**:
Code in the developer's repo that wires libraries together at their seams. The developer owns it, whoever wrote it.
_Avoid_: boilerplate, integration code, wrapper

**Recipe** _(provisional)_:
Instructions, and the glue code they produce, for binding a library to a contract or wiring a pairing together.
_Avoid_: template, snippet, tutorial, guide

**Composition** _(provisional)_:
One engine described as a set of slots, the libraries filling them, and the recipes wiring them together.
_Avoid_: stack, project, build, configuration

### Growth

**Tier** _(provisional)_:
The level of ambition a composition works at: testbed, starter engine or custom integration.
_Avoid_: level, stage, plan

**Testbed** _(provisional)_:
The smallest tier: a composition that opens a window and runs a render and input loop, made for experiments.
_Avoid_: sandbox, demo, prototype, hello world

**Starter engine** _(provisional)_:
The middle tier: a composition organized into layers (platform and GPU, core runtime, engine systems, world and tools, game).
_Avoid_: framework, template engine

**Custom integration** _(provisional)_:
The top tier: a composition with swappable implementations behind contracts, which may include private and proprietary libraries.
_Avoid_: advanced mode, enterprise tier
