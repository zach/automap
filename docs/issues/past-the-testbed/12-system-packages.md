# 12: System packages as a fact, checked by verify

**What to build:** A library's system packages are a list per platform and package manager, and `verify` names a missing one before building. Field notes [8].

**Blocked by:** None

**Status:** ready-for-agent

- [ ] The platform contract 1.1.0 adds the `system_packages` fact: per platform, per package manager (`apt`), a list of package names; `system_dependencies` stays as the prose.
- [ ] SDL3 states its packages; `automap show library sdl3` carries them.
- [ ] On a machine with `dpkg-query`, `verify` fails before building with `SYSTEM_PACKAGE_MISSING(<package>)` for each missing package and one `apt-get install` line for all of them; elsewhere it doesn't check.
- [ ] The reference workflow installs the composition's packages from the fact instead of a fixed list.
