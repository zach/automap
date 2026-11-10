# Spintacular field report: evaluation

Evaluated 2026-11-10. Spintacular is a complete single-player Windows game the product owner built on an Automap composition (CLI `cli-v0.1.0`, map release `map-2026.11.07`) as a proof of the product. Its repo carries a running log of 50 findings and a report sorted by theme. This note checks each finding against Automap's code and map, says what causes it, and says what the plan does with it. The plan is the feature spec and tickets in [`docs/issues/past-the-testbed/`](../issues/past-the-testbed/).

## Sources

- `github.com/zach/spintacular` at commit `19c7919`, read 2026-11-10: `docs/automap/field-notes.md` (the log; entries cited as [n]) and `docs/automap/report.md`.
- Automap at commit `6b6f06a` (`main` on 2026-11-10): the CLI under `internal/`, the public map under `map/`, the Automap registry's ports under `ci/registry/ports/`, the reference workflow `.github/workflows/automap-verify.yml`.
- A vcpkg dry run on 2026-11-10 (`vcpkg install --dry-run`, tool release 2026-09-26, triplet `x64-linux`) of the day-one composition as `automap start --with sdl3 --with dear-imgui:docking` writes it.

**[Verified]** marks a finding reproduced or read in the code or the map; **[Inference]** marks a reading of the report that wasn't reproduced here.

## What the exercise shows

**[Verified]** The core held. `start` and `extend` composed ten libraries, and the composition reached `probed` on Windows and Linux at the first CI run with one line of glue changed [22, 40]. The facades, the `--json` results, the failure messages, `verify --seed` and `eject` worked as designed [11, 26, 31, 45, 48, 49]. Nothing in the 50 entries questions the model: slots, contracts, recipes, probes, evidence.

**[Verified]** The cost sits after the testbed, in four places:

1. **Three bugs.** Two libraries on one vcpkg port break `vcpkg.json` [28]; an in-repo library's pin breaks on a Windows checkout [39]; `upgrade` doesn't carry a private map's changed reference glue [29]. Each one stopped work until the developer found a workaround, and one workaround made the map say something untrue.
2. **Nothing says how a testbed becomes a game.** The executable's protocol, where it must land, release builds, what to ship, shaders, private maps and removing a library are all absent from the router skill, and two pieces of reference glue (the Dear ImGui frame, the SDL3 window) can't be used by a game as written [15, 25, 32–35, 38, 44].
3. **Results and search cost context.** `find --json` is 600 lines without the slots [6]; `find text` finds nothing though a library fills `text` [46]; a failed build's result carries no cause, so six runs in a row needed the log opened to learn it was one blocked download [19].
4. **The map says two things vcpkg doesn't build** [14, 42], and nothing in Automap's CI would catch a third.

**[Verified]** A fourth bug, found while fixing the third: the lock pins a private map at the last commit that changed its directory, and in a shallow clone, which is how CI checks a repo out, that is the clone's one commit, so a composition with a private variant fails its lock check in CI (ticket 14).

**[Inference]** The order of value follows the report's own ranking: the bugs first, then the testbed-to-game path, then results and search, then map truth. Coverage (text, data files) and libraries for development builds only matter, but they set direction, so they are the product owner's to decide ([ADR-0069](../adr/0069-human-checkpoints-only-where-direction-is-set.md)).

## The 50 entries

Disposition names the ticket in `docs/issues/past-the-testbed/`, or says why nothing changes.

| # | Mark | Finding | Checked | Cause | Disposition |
| --- | --- | --- | --- | --- | --- |
| 1 | FRICTION | README says `go install`; the release ships binaries | Verified: `README.md` | README predates `cli-v0.1.0` | 04 |
| 2 | GAP | No `--version`, `help`, per-command help | Verified: `internal/cli/cli.go` | `Run` knows only the ten commands | 04 |
| 3 | GAP | Usage leaves out `upgrade` | Verified: the `usage` constant | Usage is hand-written beside the command table | 04 |
| 4 | OK | No setup | n/a | | none |
| 5 | OK | `find` for people is one line | n/a | | none |
| 6 | FRICTION | `find --json` is 600 lines, no slots | Verified: `find` in `mapcmd.go` | Each entry carries every evidence row and no slots | 06 |
| 7 | OK | `show library` is thorough | n/a | | none |
| 8 | FRICTION | System dependencies are prose; nothing checks them | Verified: `map/libraries/sdl3.json`, platform contract 1.0.0 | The fact is a string per platform | 12 |
| 9 | OK | `start` writes a coherent composition | n/a | | none |
| 10 | OK | `--platform` | n/a | | none |
| 11 | OK | Failure results are honest | n/a | | none |
| 12 | FRICTION | "the composition is none on …" | Verified: `writeAndVerify`, `runVerify` | The level is spliced into a sentence written for a level that was reached | 05 |
| 13 | ENV | GitHub archive downloads blocked; nothing in Automap helps | Verified: this repo's sessions hit the same proxy | vcpkg downloads sources from `/archive/` URLs | 05 (names the URL), 08 (`vendor` as the route) |
| 14 | GAP | SDL3's variant says IBus is off; the build has it | Verified by dry run: the plan is `sdl3[core,dbus,ibus,vulkan,wayland,x11]` | vcpkg unions features: `imgui[sdl3-binding]` depends on `sdl3` with default features | 11 (Dear ImGui from a registry port that doesn't) |
| 15 | GAP | The skill doesn't say how a testbed becomes a game | Verified: `SKILL.md.tmpl` | The skill covers start, extend, upgrade, verify | 08 |
| 16 | OK | Day-one composition reached `probed` locally | n/a | | none |
| 17 | OK | `extend` is quick and correct | n/a | | none |
| 18 | ENV | A commit-patch download was blocked too | Inference | As 13 | 05, 08 |
| 19 | FRICTION | Six failed builds, the same log, no cause in the result | Verified: `verify.Run` reports `BUILD_FAILED` with the log only | The result never reads the log | 05 |
| 20 | OK | `extend` writes the composition even when the build fails | n/a | | none |
| 21 | GAP | The reference workflow doesn't fit a game | Verified: `automap-verify.yml` | Fixed three-OS matrix, CLI from `main`, no logs or executable kept | 10 |
| 22 | OK | Eight libraries at `probed` on two platforms, first run | n/a | | none |
| 23 | OK | The released Windows binary works on a runner | n/a | | none |
| 24 | OK | A private map is easy to start | n/a | | none |
| 25 | FRICTION | No guide to writing a private map | Verified: the skill never mentions private maps | | 08 |
| 26 | OK | Validation errors are specific | n/a | | none |
| 27 | FRICTION | A private map change needs a commit before the CLI reads it | Partly: a private map with no commit is refused (`Resolve`); once it has one, the CLI reads the working tree and pins the last commit, so uncommitted changes are read but not pinned | `lastCommit` is `git log -1 -- .` | 14 |
| 28 | GAP (bug) | Two libraries on one vcpkg port break `vcpkg.json` | Verified: `VcpkgJSON` and `vcpkgOverrides` append per library | No merge by port | 01 |
| 29 | GAP | Nothing re-resolves or removes; `upgrade` does, with "0 files updated" and a doubled `LOCK_MISMATCH` | Verified: `upgrade` loads the private maps' working tree for both the old and the new map, so a changed private reference equals itself; `Check` reports a pin once per platform | | 03 (carry and dedupe), 13 (commands) |
| 30 | OK | The allocation binding and probe for a new library pass | n/a | | none |
| 31 | OK | The facades are the right seam | n/a | | none |
| 32 | OK, with two things found in the source | The executable must land in the build root and runs from there | Verified: `runCapturing` | The CLI joins the build directory and the target's name | 07 (the CLI puts it there), 08 |
| 33 | FRICTION | The Dear ImGui pairing glue owns the frame | Verified: `dear-imgui--sdl3` glue | `begin_frame` polls events; `end_frame` owns the pass | 09 |
| 34 | FRICTION | SDL3's setup hard-codes the window | Verified: `sdl3_setup.cpp` line 15 | | 09 |
| 35 | GAP | No path from shader source to SPIR-V | Verified: no recipe note or build function | The map has the run-time half only | 15 |
| 36 | OK | miniaudio's glue fits a game | n/a | | none |
| 37 | GAP | `start` doesn't ignore `build/` | Verified: no `.gitignore` among the composition's files | | 07 |
| 38 | GAP | Automap stops at a Debug build; `AGENTS.md` assumes `VCPKG_ROOT` | Verified: `build` in `verify.go`, `agentfiles.Section` | A fair boundary, undocumented | 08 |
| 39 | GAP (bug) | An in-repo library's pin fails on a Windows checkout | Verified: `SourceHash` hashes raw bytes | Git's CRLF conversion changes them | 02 |
| 40 | OK | The game verifies as the composition's executable | n/a | | none |
| 41 | OK | Windows and Linux render the same | n/a | | none |
| 42 | GAP | "No DXC at runtime" isn't true of the DLL the port builds | Verified in the port: `SDLSHADERCROSS_DXC=ON` on Windows and x64 Linux links DXC into the library | The note describes the code path, not the link | 15 |
| 43 | GAP | Tracy's one variant can't ship; choices can't differ by build type | Verified: `tracy.json` has one variant; the manifest has one set of choices | | 16 (the variant; the owner decides choices per build type) |
| 44 | GAP | No `automap remove`; orphans left behind | Verified | | 13 |
| 45 | OK | The facade made removal free for the game | n/a | | none |
| 46 | FRICTION | `find text` finds nothing though a library fills `text` | Verified: `find` matches library ids and names only | | 06 |
| 47 | OK | `verify --platform` on the wrong machine says what to do | n/a | | none |
| 48 | OK, one oddity | `verify --seed` reports `level: smoke` | Verified: `runSeed` copies the scratch run's level | | 05 |
| 49 | OK, two leftovers | After `eject` the `AGENTS.md` section and the skill still point at the CLI | Verified: `eject` touches neither | | 07 |
| 50 | OK | Evidence is CI's; nothing but the workflow's source says how to get it signed | Verified: the skill describes the reference workflow only | | 08, 10 |

## What the report asks for that the plan declines

- **A release build and packaging command.** `verify` builds Debug to probe; building and packaging a release is the game's own CMake. The plan documents the release build and what to ship (08) and removes the surprise DLL (15), and adds no command.
- **A download workaround built into the CLI** for networks that block source archives. The route that already exists is `automap vendor` where the network is open; the plan makes a blocked download say so (05) and documents the route (08).

## Follow-ups this note doesn't settle

- Whether a firewall prompt appears for Tracy's listening socket on Windows [43] was not tested by the report or here.
- Whether SDL_shadercross built without DXC on Windows still compiles SPIR-V to DXBC is checked in ticket 15, in CI, before the port changes.
