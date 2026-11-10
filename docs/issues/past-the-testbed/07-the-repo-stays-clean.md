# 07: The repo stays clean

**What to build:** `start` and `adopt` ignore the build directory; `eject` removes the agent files that point at the CLI; the executable is run wherever the project's CMake puts it. Field notes [32, 37, 49].

**Blocked by:** None

**Status:** ready-for-agent

- [ ] `start` writes a `.gitignore` with `/build/`, or adds the line to one that lacks it; `adopt` adds `/build/automap/`.
- [ ] `eject` removes the Automap section from `AGENTS.md` (and the file when nothing else is in it) and every router skill directory it wrote, and its summary names what stays: `evidence/` and any private map.
- [ ] The CLI configures with `CMAKE_RUNTIME_OUTPUT_DIRECTORY` at the build root, so an executable whose target is defined in a subdirectory is found and run.
- [ ] Golden tests cover the three.
