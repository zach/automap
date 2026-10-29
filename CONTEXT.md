# Automap

Automap is a curated, machine-readable map for composing a custom C++ game engine from many libraries, used by game developers and their coding agents to assemble, extend and maintain that engine.

> **Status:** seeded in Phase 0 from the starting hypothesis. Terms marked _(provisional)_ are not yet decided; Phase 2 branch 3 (domain model) sharpens them into schema-ready definitions. Ambiguities found while seeding are tracked in `docs/OPEN_QUESTIONS.md`.

## Language

### People

**Framework-first indie**:
The primary persona: a solo developer or small team, comfortable in C/C++, who builds and ships games on their own engine assembled from libraries, working with a coding agent.
_Avoid_: user, customer, engine developer

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

### The map

**Slot** _(provisional)_:
A named role in an engine that one library, or the developer's own code, fills: for example windowing, physics or audio.
_Avoid_: category, module, component, layer

**Library** _(provisional)_:
A third-party, in-house or proprietary body of code that can fill one or more slots.
_Avoid_: package, port, dependency (those name a build backend's artifacts, not the thing being composed)

**Seam** _(provisional)_:
A point where two libraries, or a library and the engine around it, must agree to work together: allocator ownership, threading, math types, handles, lifecycle, error handling, serialization, build flags or platform.
_Avoid_: integration point, interface, boundary

**Contract** _(provisional)_:
The stated requirements at one or more seams that whatever fills a slot must meet, and that the engine around it must honor.
_Avoid_: interface, API, spec

**Pairing** _(provisional)_:
A claim, backed by evidence, that specific libraries at stated versions work together on stated platforms.
_Avoid_: compatibility, combo, integration

**Recipe** _(provisional)_:
Instructions, and the glue code they produce, for wiring a library into a composition or wiring a pairing together.
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
