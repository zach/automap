# 32: automap adopt

**What to build:** `automap adopt` brings an existing CMake and vcpkg project under Automap: it matches the project's targets to map libraries, reports targets no map describes, and once every target is described, writes a manifest pointing at the glue that already exists and runs `verify`.

**Blocked by:** [31: Private maps, in-repo libraries and private variants](31-private-maps-and-in-repo-libraries.md)

**Status:** ready-for-agent

- [ ] On a fixture copy of Acme's engine, `adopt` reports `UNMAPPED_TARGET(Acme::Jobs)` and `UNMAPPED_TARGET(Acme::Renderer)` until the private map describes them, then writes the manifest and lock.
- [ ] The manifest's glue entries point at existing files; probes judge them (ADR-0016).
- [ ] The engine's own startup code is marked developer-owned in the manifest, no startup is generated, and the lifecycle probe judges the engine's order (ADR-0080).
- [ ] `adopt` never overwrites existing glue.
