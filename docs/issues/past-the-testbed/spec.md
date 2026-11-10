# Past the testbed

**Status:** ready-for-agent

The product-wide spec is [`docs/spec/SPEC.md`](../../spec/SPEC.md); this file is the feature spec the tickets in this directory are cut from. Terms are defined in [`CONTEXT.md`](../../../CONTEXT.md). The evidence behind it is [`docs/research/08-spintacular-field-report.md`](../../research/08-spintacular-field-report.md): the product owner built a complete game, Spintacular, on Automap 0.1.0 and logged 50 findings.

## Problem Statement

Automap 0.1.0 gets a developer to a verified testbed quickly and then leaves them alone. Building Spintacular, the composition reached `probed` on two platforms at the first CI run, and most of the time spent on Automap came afterwards: three bugs stopped work when a private map shared a vcpkg port with the public map, when an in-repo library was checked out on Windows, and when a private recipe changed; nothing said how the testbed becomes a game, how to remove a library, how shaders get built or what to ship; results and search spent the agent's context without answering its question; and the map stated two facts about what vcpkg builds that weren't so.

## Solution

Fix the bugs, then make the path from testbed to shipped game one the agent can follow from the CLI and the router skill alone. The CLI merges libraries that share a port, pins in-repo libraries by content that survives a checkout, and carries a private map's changed glue. Its results name the cause of a failed build, and `find` answers "is there a library for X?" across every map layer in a few lines. The router skill covers the game's executable, release builds, shipping, shaders, private maps, changing a composition and restricted networks. Reference glue for the window and the Dear ImGui frame is usable by a game unchanged. The reference workflow follows the manifest and the CLI version that wrote the composition. Automap's CI checks every variant's stated features against vcpkg's install plan, so the map can't drift from the build.

## User Stories

### Bugs

1. As a developer's agent, I want two libraries that come from one vcpkg port to compose, so that a private map can add one more header from a port the public map already uses.
2. As a developer, I want a lock written on Linux to verify on a Windows checkout, so that CI on both platforms agrees about my in-repo libraries.
3. As a developer's agent, I want `upgrade` to bring a private recipe's changed reference glue into the repo where my copy still matches the old reference, so that private maps behave like the public one.
4. As a developer's agent, I want each problem reported once, so that I don't read a doubled failure as two problems.

### The CLI's own surface

5. As a developer's agent, I want `automap --version` and `automap help [<command>]`, so that I can check what I installed and how to call it without triggering a usage error.
6. As a developer, I want the usage text, the README and the release to agree, so that I install the released binary and see every command.

### Results and search

7. As a developer's agent, I want a failed build's result to carry the first error and, for a failed download, the URL, so that I know it is the network without opening the log.
8. As a developer's agent, I want `find <query>` to match slots and descriptions in every map layer, and to list the slots when nothing matches, so that "is there a library for text?" has an answer.
9. As a developer's agent, I want `find --json` to give one short entry per library with its slots, license class and evidence state per platform, so that looking at the map costs little context.
10. As a developer's agent, I want a seed run and a failed start to describe themselves plainly, so that I don't read a lower level into them.

### From testbed to game

11. As a developer's agent, I want the router skill to say how the testbed becomes a game: what the manifest's executable may be, the `--frames` and `--probe` protocol, where assets go, how to build a release and what to ship.
12. As a developer's agent, I want the CLI to run my executable wherever my CMake project is laid out, so that I don't learn its location rule from the source.
13. As a game developer, I want to set the window's title and size without editing glue, so that the first change every game makes doesn't diverge it.
14. As a game developer, I want the Dear ImGui pairing glue in pieces (process an event, start a frame, render into my pass), so that my frame loop uses it unchanged.
15. As a developer's agent, I want the router skill to cover writing a private map, starting from a similar public recipe, so that I don't need the spec to add a library the map lacks.
16. As a developer, I want `start` to ignore its build directory and `eject` to remove the agent files that point at the CLI, so that the repo says only what is true.
17. As a game developer, I want a build step that turns shader source into the SPIR-V the glue takes, and a game that doesn't ship an 18 MB compiler it never calls.

### CI

18. As a developer, I want the reference workflow to verify exactly the platforms my manifest targets with the CLI release that wrote my composition, to keep the build logs, and to let me keep the built executable.
19. As a developer, I want to know what a workflow of my own must do to get signed evidence.

### Map truth

20. As a developer, I want a variant's description to match what vcpkg builds in my composition, whichever other libraries are in it.
21. As Automap's CI, I want to compare every variant's features against vcpkg's install plan for each reference composition, so that a fact about a build can't drift.
22. As a developer's agent, I want a library's system packages as a list I can install and `verify` to name a missing one, instead of a failed port build.

### Changing a composition

23. As a developer's agent, I want to remove a library and to re-resolve after a manifest or private-map change with commands named for those jobs.
24. As a developer, I want a private map's uncommitted changes read and marked, so that iterating on one doesn't need a commit per try, and a lock that checks in CI's shallow clone as it does on my machine.
25. As a game developer, I want a profiler variant that can ship.

### Direction, for the product owner

26. As a game developer, I want text and fonts and data files in the public map.
27. As a game developer, I want profiling in development builds and not in the release.

## Implementation Decisions

- **Ports merge by name.** `vcpkg.json` lists each port once with the union of its libraries' features; default features stay off only when every library sharing the port turns them off; two libraries that pin one port at different versions fail resolution with `PORT_CONFLICT(<port>)`.
- **In-repo pins hash normalized text.** A file with no NUL byte is hashed with CRLF read as LF. A lock written by 0.1.0 still checks: the raw hash is accepted where it matches.
- **The old side of an upgrade is the lock's.** `upgrade` reads each private map at the commit the lock pins (through `git archive`) for the old reference, and the working tree for the new.
- **Usage is generated from the command table,** so a command can't be missing from it. `--version`, `version`, `help` and `<command> --help` exit 0.
- **A failed build's failure message ends with the log's first error lines.** A vcpkg download failure is `FETCH_FAILED` with the URL, from `verify` as from `fetch`.
- **`find` summarizes; `show` details.** A `find` entry has `id`, `name`, `version`, `variants`, `slots`, `license` and `evidence` as one state and level per platform. `show library` keeps every evidence row.
- **The CLI finds the executable.** It runs the one at the build root or, failing that, the one the target built anywhere in the build tree, from the executable's own directory.
- **The router skill stays one file per topic.** `SKILL.md` routes; `game.md` and `private-maps.md` beside it hold the two longer guides, written and refreshed by the CLI with it.
- **Window settings are context fields** SDL3's setup glue reads before it creates the window: `sdl3.window_title`, `sdl3.window_width`, `sdl3.window_height`, with the testbed's values as defaults.
- **The reference workflow reads the composition.** A first job turns the manifest's `platforms` into the matrix and reads the CLI version from the router skill's metadata; each verify job downloads that release's binary and checks its SHA-256, and uploads `build/automap/*.log`.
- **A variant builds what it says.** vcpkg unions the features ports ask of a port, so a variant can't promise a feature stays off while another port in the map asks for it: Dear ImGui comes from an Automap registry port that asks for SDL3 without its default features, which keeps IBus, D-Bus and libsystemd out of every composition, and SDL3's variant states `vulkan` for every platform, as SDL_shadercross's port asks.
- **System packages are a new platform fact,** `system_packages`, a list per platform and package manager, beside the prose; the platform contract moves to 1.1.0.
- **`remove` and `resolve` are commands.** `remove <library>` is the other half of extend; `resolve` re-resolves against the lock's own release after the manifest or a private map changed; `upgrade` is `resolve` plus the move to a newer release.
- **A private map is pinned by content.** The lock records the hash of its directory, as for an in-repo library, and the commit beside it when the directory is committed and clean; an uncommitted map composes with a warning and unsigned evidence.
- **SDL_shadercross links DXC only where a variant asks for it.** The default variant compiles SPIR-V to DXBC, MSL and SPIR-V with no DXC at run time; the port keeps DXC as a build-time tool, and the build fragment offers a CMake function that compiles HLSL to SPIR-V with it.

## Out of Scope

- A release-build or packaging command: the game's own CMake does both; the skill documents them.
- A download workaround inside the CLI for networks that block source archives: `automap vendor` on an open network is the route.
- A web testbed, a second build backend and NDA platforms ([ADR-0056](../../adr/0056-v1-platforms-desktop-only.md), [ADR-0034](../../adr/0034-vcpkg-is-the-v1-build-backend.md), [ADR-0083](../../adr/0083-nda-platforms-are-out-of-scope.md)).

## Decisions the product owner made

- The public map covers text and data files ([ADR-0084](../../adr/0084-the-map-covers-text-and-data-files.md), ticket 17).
- A composition has one set of slots for every build type ([ADR-0085](../../adr/0085-a-composition-has-one-set-of-slots-for-every-build-type.md), ticket 16).

Ticket 18's eval run and release wait for the product owner ([ADR-0069](../../adr/0069-human-checkpoints-only-where-direction-is-set.md)). Everything else merges on its tests and CI: command names beyond the five jobs and how the lock records a private map are the spec's details (SPEC §12.3), and a variant of a library the map covers, or the registry its port comes from, is curation.
