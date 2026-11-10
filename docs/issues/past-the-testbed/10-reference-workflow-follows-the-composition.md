# 10: The reference workflow follows the composition

**What to build:** The reference workflow verifies the platforms the manifest targets with the CLI release that wrote the composition, keeps the build logs, and can keep the built executable. Field notes [21, 50].

**Blocked by:** 04

**Status:** ready-for-agent

- [ ] A first job reads the manifest's `platforms` into the matrix: a composition for Windows and Linux starts no macOS job.
- [ ] Each job installs the CLI release named by the router skill's `automap-cli-version` (or the `automap-version` input), checking its SHA-256; a development version, or `automap-ref`, builds from source.
- [ ] `build/automap/*.log` is uploaded as an artifact whether or not verification passes; with `keep-build: true`, so is the built executable's directory.
- [ ] The reference-repo workflow exercises it and is green.
- [ ] SPEC §2.4 and the skill's caller match.
