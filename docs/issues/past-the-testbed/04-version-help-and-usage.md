# 04: Version, help and a usage that lists every command

**What to build:** `automap --version`, `automap version`, `automap help [<command>]` and `automap <command> --help` work and exit 0; the usage lists every command; the README installs the released binary. Field notes [1, 2, 3].

**Blocked by:** None

**Status:** ready-for-agent

- [ ] `automap --version` and `automap version` print `automap <version>`; with `--json`, a CLI result whose `cli_version` carries it.
- [ ] `automap help`, `automap --help` and `automap` with no arguments print the usage; only the last exits 2.
- [ ] `automap help <command>` and `automap <command> --help` print that command's usage line, what it does and its flags.
- [ ] A test fails when a command in the command table is missing from the usage.
- [ ] The README's quick start downloads the release binary for the platform and checks its SHA-256, with `go install` as the alternative, and its status section states the released version.
