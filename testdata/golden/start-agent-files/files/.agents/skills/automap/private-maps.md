# Private maps

A private map describes libraries the public map doesn't have: your own code, a proprietary library, a patched build, or a library nobody has added yet. It has the public map's layout, lives in your repo, and layers over the public map. This guide was written by automap 0.0.0-dev.

## Set one up

```
automap-private/
  map.json                       {"format": "automap-map/1", "id": "mygame", "kind": "private"}
  slots/<slot>.json              only for a slot the public map lacks
  libraries/<library>.json
  recipes/library/<library>.json            and recipes/library/<library>/<library>.cmake
  recipes/binding/<library>--<contract>.json   and its glue, probe and seeded violation
  ports/<port>/                  overlay ports, for a patched or unpackaged library
```

Name it in the manifest:

```json
"private_maps": [{ "id": "mygame", "location": "automap-private" }]
```

`automap find`, `automap show` and `automap extend` read it at once, committed or not: the lock pins its content. While it has uncommitted changes the result carries a `PRIVATE_MAP_UNCOMMITTED` warning, and evidence about the composition isn't signed.

## Start from a library like yours

Don't write entries from nothing. Find a public library with the same shape (a header-only library, a library with allocation hooks, a C library with an init call) and copy its entries:

```console
$ automap show library stb-image --json      # the library entry, under entries[0].definition
$ automap show recipe stb-image --json       # its library recipe
$ automap show recipe stb-image--allocation --json
```

The files themselves, glue and probes included, are in the CLI's cache: `<cache>/maps/<release>/`, where `<cache>` is `AUTOMAP_CACHE` or `automap` in your user cache directory (`~/.cache/automap` on Linux) and `<release>` is the lock's `map_release`. Copy the files, rename them, and change what differs. Every mistake comes back as one `MAP_INVALID(<file>)` failure that says what is missing.

## What each entry needs

- **A slot** (`slots/<id>.json`): `id` and a `description` in the words someone would search for.
- **A library** (`libraries/<id>.json`): `id`, `name`, `license`, `contribution_policy`, the `slots` it fills, its `source`, and one or more `versions`, each with its `pin`, `build_backend` (the vcpkg port and its registry: `curated` for vcpkg's own, `private` for an overlay port in `ports/`), `cmake` package and targets, `variants`, and `facts` for each contract it touches; the lifecycle, build-flags and platform contracts apply to every library.
- **A library recipe** (`recipes/library/<id>.json`) with a build fragment: a `reference_glue` file with the role `build-fragment`, the CMake that finds the package and links it to `automap_glue`. Without one the library can't join a composition (`NO_BUILD_FRAGMENT`).
- **A binding** (`recipes/binding/<library>--<contract>.json`) for each contract the library's facts say it touches: the glue that connects the library to the contract's shared interface, a probe for each probed rule, and a seeded violation per probe, a patch that breaks the glue the way the rule forbids. `automap show contract <id> --json` has the rules. A rule that can't apply goes under `not_applicable` with the reason.

A library whose source is a directory in your repo has `"source": {"kind": "path", "path": "<dir>"}` and no pin or build backend: it joins the build through `add_subdirectory`.

Two libraries can come from one vcpkg port, such as two headers of `stb`: give both the port's name and the same version, and they share its entry in `vcpkg.json`.

## Compose, and check your own probes

```console
$ automap extend <library> --json
$ automap verify --seed <rule>:<library> --json     # your probe must catch its seeded violation
$ git add automap-private && git commit -m "Private map: <library>"
$ automap resolve --json                            # records the commit; evidence can be signed again
```

`verify --seed` applies the seeded violation in a scratch copy and passes only when your probe fails there: a probe that can't fail proves nothing.

## After you change it

`automap resolve --json` re-pins the private map and updates the glue in the repo that you haven't changed since the old reference, without moving the map release or any library's version. Until then, commands that check the lock fail with `LOCK_MISMATCH(private map <id>)`. Iterate with `resolve`; commit when it works, and `resolve` once more to record the commit. Nothing in a private map is ever uploaded; the evidence `verify` writes for a composition that uses it is the composition's own, in `evidence/`.
