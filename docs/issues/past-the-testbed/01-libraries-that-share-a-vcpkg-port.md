# 01: Libraries that share a vcpkg port

**What to build:** Two libraries in a composition that come from one vcpkg port compose: `vcpkg.json` names the port once, with the union of their features and one override. Field notes [28].

**Blocked by:** None

**Status:** ready-for-agent

- [ ] On the fixture map, a private library on the same port as a public one composes: `vcpkg.json` lists the port once with both libraries' features and one override, and the composition verifies.
- [ ] Default features are off for a shared port only when every library sharing it turns them off.
- [ ] Two libraries that pin one port at different versions fail with `PORT_CONFLICT(<port>)`, naming both libraries and versions.
- [ ] `vcpkg-configuration.json` names a shared Automap registry port once; `eject` and `vendor` copy it once.
- [ ] SPEC §5.5 says how ports merge.
