# Automap

Automap is an agent-ready map and CLI for composing custom C++ game engines from libraries. The product spec is [`docs/spec/SPEC.md`](docs/spec/SPEC.md); the work is cut into vertical-slice tickets in [`docs/issues/automap-v1/`](docs/issues/automap-v1/), with their feature spec in `spec.md` there. Terms are defined in [`CONTEXT.md`](CONTEXT.md); decisions are the ADRs in [`docs/adr/`](docs/adr/). When resuming, read the newest handoff in `docs/handoff/`.

## Build and test

The CLI is Go ([ADR-0064](docs/adr/0064-cli-is-written-in-go.md)), at the version in `go.mod`.

| Task | Command |
| --- | --- |
| Build the CLI | `go build -o automap ./cmd/automap` |
| Run every test | `go test ./...` |
| Run the build tests too | `VCPKG_ROOT=<vcpkg> go test ./...` (they need vcpkg, CMake, Ninja and clang; CI always runs them) |
| Rewrite golden files from the CLI's output | `go test ./... -update`, then review the diff |
| Format and vet | `gofmt -l .` (must print nothing), `go vet ./...` |
| Check the spec's examples against the schemas | `python3 docs/spec/check-examples.py` (needs `pip install jsonschema`) |

CI (`.github/workflows/ci.yml`) runs all of these on Linux on every push.

### How tests work

Tests attach at the CLI run as a process ([ADR-0077](docs/adr/0077-tests-attach-at-the-cli-process.md)). A test calls `clitest.Automap(t, "<fixture repo>", "<command>", ..., "--json")`, which builds the binary once, runs it in a temporary copy of `testdata/repos/<fixture repo>/`, and validates its stdout against the published CLI result schema. `ExpectGolden(t, "<case>")` compares stdout with `testdata/golden/<case>/result.json` and the repo after the run with the fixture plus `testdata/golden/<case>/files/` (every file the run adds or changes). Write the golden file from the spec before the code, red first. Go tests below the CLI cover pure functions only, such as lock canonicalization and resolution.

## Repo layout

| Path | What lives there |
| --- | --- |
| `cmd/automap/` | The CLI's entry point and its process-level tests. |
| `internal/cli/` | The commands and the CLI result. |
| `internal/mapmodel/` | Loads, validates and layers the public and private maps. |
| `internal/composition/` | The manifest and lock, resolution, startup planning, and the files a composition is made of. |
| `internal/canonical/` | RFC 8785 canonical JSON, for the lock. |
| `internal/verify/` | Builds through CMake and vcpkg, runs the smoke test and probes, writes evidence. |
| `internal/agentfiles/` | The `AGENTS.md` section, the `CLAUDE.md` import, the router skill and the four nevers. |
| `internal/clitest/` | The test harness that runs the CLI as a process. |
| `map/` | The public map (SPEC §5.3). Everything in it is MIT-0. |
| `ci/` | Automap's CI for the public map: `cmd/verifymap` plans which declared edges need verifying, verifies them and writes their evidence; `compositions.json` lists the reference compositions. The `Verify map` workflow signs the evidence and commits it into `map/evidence/`. |
| `evals/` | The eval suite: tasks and the runner, run on demand by the `Evals` workflow. |
| `schemas/` | The published JSON Schemas, embedded in the binary. |
| `testdata/map/` | The fixture map's own part: three stub libraries with recipes, probes and seeded violations. Tests add the public map's contracts and probe harness (`clitest.Repo.FixtureMap`). |
| `testdata/repos/` | Fixture repos the CLI runs in. |
| `testdata/golden/` | Golden CLI results and written files, one directory per case. |
| `docs/spec/` | `SPEC.md`, its examples and `check-examples.py`. |
| `docs/issues/automap-v1/` | The feature spec and tickets. |
| `docs/adr/`, `CONTEXT.md` | Decisions and the glossary. |
| `docs/research/` | Research notes, one file per topic, plus `SUMMARY.md`. |
| `docs/handoff/` | Session handoffs, newest last. |
| `docs/OPEN_QUESTIONS.md` | Questions deferred past v1. |
| `docs/EXPLORATION_BRIEF.md`, `docs/PLAN.md` | The product exploration's brief and plan. |

SPEC §10.1 gives the layout later tickets add: the `registry` branch.

## Working on a ticket

- Work the frontier of `docs/issues/automap-v1/`: a ticket whose blockers are all done. Branch from `main`.
- Work test-first with the `tdd` skill, at the CLI seam.
- When every acceptance criterion is met, make the last commit on the branch tick the ticket's criteria and set its `Status:` to `done`. Once CI is green, open a pull request to `main` and merge it with a merge commit.
- Where the spec fills in a detail no ADR decided (SPEC §12.3), follow it as a default; change it if the build needs to, and update the spec to match.

## Working rules

These override the upstream skills where they conflict.

- **Human checkpoints exist only where direction is set** ([ADR-0069](docs/adr/0069-human-checkpoints-only-where-direction-is-set.md)): any new or changed ADR; what the map covers and new contract major versions; the v1 release. Ask the product owner (Zach) for those and nothing else; everything else merges on its evidence and evals.
- **One question at a time,** each with a recommended answer and a one- or two-sentence rationale, worded so "yes" accepts it. Say so when a question needs the product owner's taste or judgment rather than hiding it behind a recommendation.
- **Don't ask what you can find out.** Read the docs, research, or run a quick experiment first.
- **Challenge the product owner** when an answer conflicts with an earlier decision, the research, or what developers need, before recording it.
- **An ADR records only what the product owner agreed to.** If writing one surfaces a new consequence or assumption, ask about it instead of writing it in. Every decision with real alternatives gets an ADR in `docs/adr/` with **Context, Options, Decision, Consequences** sections.
- **Documents state the current version only.** No "amended", "previously", "no longer", changelog notes or superseded markers. When something changes, rewrite the affected docs in place; git history is the record of what changed.
- **Write docs as decisions land.** New or sharpened terms go into `CONTEXT.md` immediately, and the glossary's terms are the names used in code, tests and failure codes.
- **Don't build in assumptions that agents are weaker stewards of code than humans.** Automap assumes agents can be the primary steward of an engine's code, including Automap's own.
- **Keep `docs/OPEN_QUESTIONS.md` current.** It holds only items deferred past v1.
- **Research** goes in `docs/research/`, one file per topic, citing sources with access dates and separating verified facts from inference.
- **Handoffs** go in `docs/handoff/YYYY-MM-DD-<slug>.md`, not the OS temp directory.
- **Short chat messages.** Summaries go in the docs, not the chat.

## Skills

Multi-step procedures live in skills. `.claude/skills/` holds skills vendored from [mattpocock/skills](https://github.com/mattpocock/skills) at commit `a496f55` (v1.3.2, MIT, see `.claude/skills/LICENSE-mattpocock-skills`). The build uses `tdd`, `implement`, `implement-spec`, `code-review` and `diagnosing-bugs`, with `domain-modeling`, `to-tickets` and `handoff` ([ADR-0067](docs/adr/0067-build-team-skills.md)); the exploration's `grill-with-docs`, `grilling`, `prototype`, `research`, `to-spec` and `setup-matt-pocock-skills` are also installed. Upstream skills call the glossary `GLOSSARY.md`; read it as `CONTEXT.md`.

- **Issue tracker:** local markdown under `docs/issues/`. See `docs/agents/issue-tracker.md`.
- **Domain docs:** single context, `CONTEXT.md` and `docs/adr/` at the root. See `docs/agents/domain.md`.
