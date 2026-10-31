You are curating Automap's public map (the `map/` directory of this repository) for a new upstream release. Read `AGENTS.md`, `docs/spec/SPEC.md` §5 (the map and its formats) and the library's entry first.

The release: **$LIBRARY $VERSION**, tag `$TAG` at commit `$COMMIT` in https://github.com/$REPO, released $RELEASED. Its source archive's SHA-512 is `$SHA512` (of https://github.com/$REPO/archive/$TAG.tar.gz).

Add this version to `map/libraries/$LIBRARY.json` and make the map's recipes apply to it:

1. Add a version entry after the newest one: `version`, `released` ($RELEASED), `security_fix` ($SECURITY_FIX), the `pin` (tag, commit, `content_hash` as `sha512:<hex>`), and the `build_backend`, `cmake`, `variants` and `facts` of the newest version, updated for this release.
2. Extract facts from the release's headers where you can, rather than copying them: clone the tag (`git clone --depth 1 --branch $TAG https://github.com/$REPO /tmp/upstream`) and check every fact against its headers: hook functions and their signatures, assertion and log hooks, lifecycle functions, ABI-affecting defines, the language standard. Change a fact only where the headers show it changed; note anything you couldn't check.
3. Check the vcpkg port at the map's baseline (`build_backend.vcpkg.baseline` in `map/map.json`, port `ports/<port>` in https://github.com/microsoft/vcpkg): if the curated port doesn't have this version at that baseline, say so in the summary and stop without changing anything; don't move the baseline.
4. Add the version to `applies_to.$LIBRARY.versions` in every recipe under `map/recipes/` that applies to the library, and update a recipe's reference glue, probes, seeded violations or adaptation notes only where the release's headers show the old glue no longer fits.
5. Never add a slot, a library or a contract major version, and never edit `map/evidence/`: those aren't curation's to change.

Check your work: `go run ./cmd/automap find --map map --json` must report no `MAP_INVALID`, and `go test ./cmd/automap -run 'PublicMap|EveryPublicMapSeed'` must pass. Don't commit, push or open a pull request; the workflow does. End with a short summary of what changed and why, and anything you couldn't verify.
