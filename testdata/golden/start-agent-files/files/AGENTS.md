# Spacefarer

Our house rules: tabs, not spaces.

<!-- automap:begin -->
## Automap

This engine is an Automap composition. Before changing any glue, read the router skill (`.claude/skills/automap/SKILL.md`) and go through the `automap` CLI.

- Manifest: `automap.manifest.json`. Lock: `automap.lock.json`, which only the CLI writes.
- Slots: input = stub-window, tools-ui = stub-ui, windowing = stub-window.
- Test: `automap verify --json`, which builds a Debug build in `build/automap/<platform>/`, runs it headless and runs the probes.
- Build: `cmake -S . -B build -G Ninja -DCMAKE_TOOLCHAIN_FILE=$VCPKG_ROOT/scripts/buildsystems/vcpkg.cmake && cmake --build build`. Without `VCPKG_ROOT`, use the vcpkg the CLI keeps: the first line of `build/automap/build-<platform>.log` is the configure command `automap verify` ran, toolchain path included.
- Run: `./build/testbed`
<!-- automap:end -->
