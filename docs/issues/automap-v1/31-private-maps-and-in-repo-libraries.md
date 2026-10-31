# 31: Private maps, in-repo libraries and private variants

**What to build:** Compositions can list private maps. The CLI layers them over the public map, never uploads them, and treats their content exactly like public content. Libraries whose source is a directory in the developer's repo join the build as CMake targets and are pinned by path and content hash. Private variants, such as a patched library, have their own pin and an overlay port in the private map.

**Blocked by:** [16: Map releases, fetch and cache, and evidence states](16-map-releases-and-evidence-states.md)

**Status:** done

- [x] The Acme examples from the product spec work end to end with stub sources: in-house libraries build with `add_subdirectory`, and the patched SDL3 variant builds from its overlay port.
- [x] Every check applies to private content unchanged; no code path branches on whether a library is private (ADR-0013, ADR-0047).
- [x] No command sends private-map content over the network.

## Comments

- `testdata/repos/acme-stubs` is the Acme example with stub sources: a private map with an in-repo library (`acme-tools`, built with `add_subdirectory`) and a patched `stub-ui` variant built from the map's overlay port. `TestAcmeStubsBuildWithPrivateContent` reaches probed and checks the installed header came from the overlay.
- A private map is pinned at the last commit that changed its directory, so commits elsewhere in the repo don't move it; a private variant's pin names that map and commit, and `vcpkg-configuration.json` lists the map's `ports/` as overlay ports.
- An in-repo library's content hash covers its directory. When the source changes, the lock check reports `LOCK_MISMATCH(<library>)` until `automap upgrade <library>` re-pins it.
- Published evidence for an edge whose recipe or pinned variants come from a private map is read from that map's directory, never fetched; `TestNoRequestCarriesPrivateContent` records every request the CLI makes.
- A variant change is applied by editing the manifest and running `automap upgrade <library>`, as `extend` refuses a library already present.
