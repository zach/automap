# 07: The repo stays clean

**What to build:** `start` and `adopt` ignore the build directory; `eject` removes the agent files that point at the CLI; the executable is run wherever the project's CMake puts it. Field notes [32, 37, 49].

**Blocked by:** None

**Status:** done

- [x] `start` writes a `.gitignore` with `/build/`, or adds the line to one that lacks it; `adopt` adds `/build/automap/`.
- [x] `eject` removes the Automap section from `AGENTS.md` (and the file when nothing else is in it) and every router skill directory it wrote, and its summary names what stays: `evidence/` and any private map.
- [x] The CLI finds the executable at the build root or, failing that, anywhere in the build tree, and runs it from its own directory, so a target defined in a subdirectory is found and run.
- [x] Process-level tests cover the three; the start and adopt goldens have the `.gitignore`.

## Comments

- Forcing `CMAKE_RUNTIME_OUTPUT_DIRECTORY` to the build root was tried first and broke a project whose target and subdirectory share a name (`game/` and `game`): the linker can't write a file where the subdirectory's build directory is. The CLI searches the build tree instead, skipping `vcpkg_installed` and `CMakeFiles`.
- `adopt` kept an existing `AGENTS.md`, `CLAUDE.md` or router skill as it found it, so an adopted engine that already had an `AGENTS.md` never got the Automap section; it now merges them as `extend` does.
