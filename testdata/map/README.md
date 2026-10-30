# Fixture map

The map the CLI's own tests run against (ADR-0077). It has the public map's layout and schemas, and two tiny stub C++ libraries:

- `stub-window` fills `windowing` and `input`, like SDL3. Its `wide-events` variant changes an ABI-affecting define.
- `stub-ui` fills `tools-ui`, like Dear ImGui, and reads the window `stub-window` provides.

They have facts for the allocation, diagnostics, lifecycle, build-flags and platform contracts, a binding for allocation and diagnostics each, a pairing (`stub-ui--stub-window`), and a library recipe each. Every probe has a seeded violation under its recipe's `seeded/` directory, and every probed composition rule is seeded by a recipe or its contract.

Everything here is MIT-0. Pins and hashes are placeholders.
