# 01: Libraries that share a vcpkg port

**What to build:** Two libraries in a composition that come from one vcpkg port compose: `vcpkg.json` names the port once, with the union of their features and one override. Field notes [28].

**Blocked by:** None

**Status:** done

- [x] On the fixture map, a private library on the same port as a public one composes: `vcpkg.json` lists the port once with both libraries' features and one override.
- [x] Default features are off for a shared port only when every library sharing it turns them off.
- [x] Two libraries that pin one port at different versions fail with `PORT_CONFLICT(<port>)`, naming both libraries and versions.
- [x] `vcpkg-configuration.json` names a shared Automap registry port once; `eject` and `vendor` copy it once.
- [x] SPEC §5.5 says how ports merge.

## Comments

- `composition.vcpkgPorts` merges the composition's libraries by port, and `vcpkg.json`, its overrides and the Automap registry's package list are written from the merged ports. `composition.PortConflicts` runs at the end of `Resolve`. `eject` already keyed the ports it copies by name, and `vendor` copies what vcpkg's install plan names, so both copy a shared port once.
- The tests compose with `--no-verify`: the fixture's stub port has no features to build. Ticket 17 puts two libraries of one real port, `stb`, through a build.
