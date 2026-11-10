# 04: Version, help and a usage that lists every command

**What to build:** `automap --version`, `automap version`, `automap help [<command>]` and `automap <command> --help` work and exit 0; the usage lists every command; the README installs the released binary. Field notes [1, 2, 3].

**Blocked by:** None

**Status:** done

- [x] `automap --version` and `automap version` print `automap <version>`; with `--json`, a CLI result whose `cli_version` carries it.
- [x] `automap help`, `automap --help` and `automap` with no arguments print the usage; only the last exits 2.
- [x] `automap help <command>` and `automap <command> --help` print that command's usage line, what it does and its flags.
- [x] The usage and each command's help are written from the one table of commands and flags the CLI parses with, so neither can leave one out; a test checks every command is listed.
- [x] The README's quick start downloads the release binary for the platform and checks its SHA-256, with `go install` as the alternative, and its status section states the released version.

## Comments

- `internal/cli/cli.go` holds one `commandList`: each command's synopsis, what it does and its flags. `Run` parses with it and `usage()` and `help()` print from it.
- `version` is a command in the cli-result schema, so `automap version --json` is a valid result.
