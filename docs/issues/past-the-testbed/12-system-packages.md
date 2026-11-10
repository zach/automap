# 12: System packages as a fact, checked by verify

**What to build:** A library's system packages are a list per platform and package manager, and `verify` names a missing one before building. Field notes [8].

**Blocked by:** None

**Status:** ready-for-agent

- [x] The platform contract 1.1.0 adds the `system_packages` fact: per platform, per package manager, a list of package names, `apt` for the build and `apt_headless` for a headless run; `system_dependencies` stays as the prose.
- [x] SDL3 states its packages; `automap show library sdl3` carries them.
- [x] On a machine with `dpkg-query`, `verify` fails before building with `SYSTEM_PACKAGE_MISSING(<package>)` for each missing build package and one `apt-get install` line for all of them, and warns about a missing headless package; a package another installed package provides counts; elsewhere it doesn't check.
- [ ] The reference workflow installs the composition's packages from the fact instead of a fixed list.

## Comments

- The Vulkan software driver is only needed to run headless on a machine without a GPU driver, so it is `apt_headless` and a warning: failing `verify` on a developer's machine with a real driver would be wrong.
- The last criterion, the reference workflow installing from the fact, is done in ticket 10, which rewrites that workflow.
