# 01: Repo setup and the CLI seam

**What to build:** The build repository adopts the instruction conventions Automap generates for users, and the CLI exists as a Go binary whose behaviour is tested by running it as a process. With no composition present, `automap verify --json` prints a schema-valid CLI result naming the problem. This is the tracer bullet for every later ticket's tests (ADR-0066, ADR-0077). It starts on `main` once the exploration has merged there (ADR-0078), and switching the instruction files ends the exploration's "no product code" rule.

**Blocked by:** None (can start immediately)

**Status:** done

- [x] `AGENTS.md` is the one instruction file: build and test commands, the repo layout, and the working rules that still apply; `CLAUDE.md` is a single `@AGENTS.md` import line (ADR-0066).
- [x] The published JSON Schemas are in the repo and embedded in the binary.
- [x] `automap verify --json` in a directory with no manifest exits with status 2 and prints a CLI result with `ok: false` and failure code `NO_MANIFEST`.
- [x] A test harness builds the binary, runs it in a temporary copy of a fixture repo, compares its JSON output and written files with golden files, and validates every JSON result against the CLI result schema.
- [x] Go tests run in CI on Linux on every push.
