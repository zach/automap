# 31: Private maps, in-repo libraries and private variants

**What to build:** Compositions can list private maps. The CLI layers them over the public map, never uploads them, and treats their content exactly like public content. Libraries whose source is a directory in the developer's repo join the build as CMake targets and are pinned by path and content hash. Private variants, such as a patched library, have their own pin and an overlay port in the private map.

**Blocked by:** [16: Map releases, fetch and cache, and evidence states](16-map-releases-and-evidence-states.md)

**Status:** ready-for-agent

- [ ] The Acme examples from the product spec work end to end with stub sources: in-house libraries build with `add_subdirectory`, and the patched SDL3 variant builds from its overlay port.
- [ ] Every check applies to private content unchanged; no code path branches on whether a library is private (ADR-0013, ADR-0047).
- [ ] No command sends private-map content over the network.
