# Spacefarer

Our house rules: tabs, not spaces.

<!-- automap:begin -->
## Automap

This engine is an Automap composition. Before changing any glue, read the router skill (`.claude/skills/automap/SKILL.md`) and go through the `automap` CLI.

- Manifest: `automap.manifest.json`. Lock: `automap.lock.json`, which only the CLI writes.
- Slots: input = stub-window, tools-ui = stub-ui, windowing = stub-window.
- Build: `cmake -S . -B build -G Ninja -DCMAKE_TOOLCHAIN_FILE=$VCPKG_ROOT/scripts/buildsystems/vcpkg.cmake && cmake --build build`
- Test: `automap verify --json`
- Run: `./build/testbed`
<!-- automap:end -->
