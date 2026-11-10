# 10: The reference workflow follows the composition

**What to build:** The reference workflow verifies the platforms the manifest targets with the CLI release that wrote the composition, keeps the build logs, and can keep the built executable. Field notes [21, 50].

**Blocked by:** 04

**Status:** done

- [x] A first job reads the manifest's `platforms` into the matrix: a composition for Windows and Linux starts no macOS job.
- [x] Each job installs the CLI release named by the router skill's `automap-cli-version` (or the `automap-version` input), checking its SHA-256; a development version, or `automap-ref`, builds from source.
- [x] `build/automap/*.log` is uploaded as an artifact whether or not verification passes; with `keep-build: true`, so is the built executable's directory.
- [x] The reference-repo workflow exercises it from source, with the built executable kept, and with the released 0.1.0 CLI, and is green.
- [x] SPEC §2.4 and the skill's caller match.

## Comments

- A `plan` job reads the manifest's platforms into the matrix and the CLI version from the router skill's `automap-cli-version`; each `verify` job downloads that release's binary, checks its SHA-256, and installs the libraries' `system_packages` through `automap show library`.
- Evidence of a failed run is now attested and committed at the level it reached before the job fails; the old workflow's step stopped at the failing command and recorded nothing.
- Whether a workflow of one's own gets signed evidence is in the router skill (ticket 08).
