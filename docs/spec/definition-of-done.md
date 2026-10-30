# Definition-of-done review

The brief's definition of done (section 6), checked item by item, with the evidence for each.

| # | Item | Status | Evidence |
| --- | --- | --- | --- |
| 1 | A fresh agent could begin slice 1 without asking anything foundational | Done | First test: four foundational questions, resolved by ADRs 0078–0081. Re-test: no foundational design questions; its remaining questions are the sign-off gate, a prerequisite for ticket 09, and minor defaults, resolved below | Fresh-agent test, 2026-10-30 (below) |
| 2 | Every schema has a complete, valid example | Done | Nine schemas, 43 example files; `python3 docs/spec/check-examples.py` reports all valid, and fails on seeded mistakes |
| 3 | Every known-good pairing names its verification method | Done | [SPEC §8.5](SPEC.md#85-pairings) |
| 4 | Testbed, starter engine and custom integration each have a worked journey | Done | [SPEC §2.4–2.6](SPEC.md#24-journey-1-testbed-start-then-extend), with example files |
| 5 | Private libraries supported without special cases, shown with a hypothetical in-house library | Done | Acme's `acme-jobs` and `acme-renderer` ([examples](examples/private-map/)), [SPEC §4.5, §7](SPEC.md#45-libraries); the checker applies the same rules to public and private content |
| 6 | `CONTEXT.md` defines every term the spec relies on | Done | Glossary scan against SPEC's vocabulary; terms added for journey, reference composition, entry point, NDA platform, ordering constraint, console variant, contribution policy, reference thread pool, set-aside report, fixture map |
| 7 | Every significant decision has an ADR; `OPEN_QUESTIONS.md` holds only items deferred past v1 | Done | ADRs 0001–0082; the product owner confirmed SPEC §12.3's details as spec defaults, with evidence lookup recorded as [ADR-0082](../adr/0082-evidence-is-read-live-for-exact-pins.md); `OPEN_QUESTIONS.md` holds OQ-008 and OQ-011, both deferred past v1 |
| 8 | Research claims are cited and dated | Done | Every note in `docs/research/` states its access date for all sources; SPEC cites the notes |

## Fresh-agent test (first run)

A general-purpose agent with no conversation history was given only the repo and asked whether it could begin slice 1 (tickets 01–09) without asking the product owner anything foundational, and to list every question it would ask. It could start ticket 01, but not finish slice 1 without design answers.

### Its questions

| # | Question | Its rating | Resolution |
| --- | --- | --- | --- |
| 1 | Has implementation been authorised, and on which branch? `CLAUDE.md` still says "No product code", and `main` holds only `LICENSE`. | Foundational | The exploration merges to `main` with a merge commit when this review is signed off; ticket 01 starts there and switches the instruction files ([ADR-0078](../adr/0078-exploration-merges-to-main-after-review.md)) |
| 2 | Where does the CLI read the public map in slices 1–2, and what goes in `map_release` before the first release in slice 3? | Foundational | The CLI reads the map from a directory given by `--map` or `AUTOMAP_MAP`, recorded as `git:<commit>` or `unreleased`; unreleased evidence is never signed ([ADR-0079](../adr/0079-map-from-a-directory-before-releases.md)) |
| 3 | How do recipe entry points become running code: a generated startup or a hand-written `main`? What supplies arguments such as the window and GPU device? | Foundational | The CLI generates the startup from the recipes; entry points take one context whose fields recipes declare; adopted engines may keep their own startup ([ADR-0080](../adr/0080-the-cli-generates-startup.md)) |
| 4 | Which agent runtime and model run the eval suite, with what credentials and budget? | Foundational | Claude Code headless on GitHub-hosted runners, newest Claude model named in each report, API key as a repository secret, once per map release and on demand, under a monthly cap the product owner sets ([ADR-0081](../adr/0081-eval-suite-runs-claude-code-headless.md)) |
| 5 | Where do `content_hash` and the vcpkg baseline come from, if resolution is pure? | Minor | Library pins carry the content hash; the public map's new `map.json` carries the vcpkg baseline (SPEC §4.2, §4.5) |
| 6 | How is vcpkg obtained? | Minor | `VCPKG_ROOT` if set, else a pinned vcpkg bootstrapped into the CLI's cache (SPEC §5.5) |
| 7 | CLI surface: `show <id>` is ambiguous; how `--slot` attaches in `start`; no default variant; "unless pinned" has no field | Minor | `show <kind> <id>`; `start --slot <slot>=<lib>`; a version's first variant is its default; "unless pinned" removed (SPEC §4.5, §5.4) |
| 8 | Are SDL's headless build options used always, only for `verify`, or only in CI? | Minor | One build for windowed and headless runs; `verify` sets the headless environment; the fixed build options were removed (SPEC §5.5) |
| 9 | What do seeded-violation patches apply to, given map paths differ from composition paths? | Minor | Patches are written against reference glue's map paths; the CLI maps them to the composition's files through the manifest (SPEC §6.1) |

### Contradictions and gaps it found

| Finding | Resolution |
| --- | --- |
| Process docs stale: `CLAUDE.md` "No product code", `PLAN.md` status, newest handoff, a "(provisional)" note in `CONTEXT.md` | `PLAN.md` and `CONTEXT.md` updated; the handoff is rewritten at the end of the review; `CLAUDE.md` changes when implementation starts (ADR-0066, ticket 01) |
| "Entry point" meant both a recipe's call and the composition's executable | The executable is now "the composition's executable" everywhere; "entry point" means only a recipe's call |
| `extend-jolt.json` listed `CMakeLists.txt` and missed the build fragments, jobs interface and `AGENTS.md` | Example corrected |
| `find` and `show` in slice 2 in SPEC, slice 1 in ticket 02 | SPEC puts them in slice 1 |
| `map_release` required before any release exists | Resolved with question 2 |
| Resolver called pure but needs data the map lacked | Resolved with question 5 |
| Fixture overlay ports have no home in SPEC | The Automap registry's location is configurable; tests use a filesystem registry in the fixture map (SPEC §5.5) |
| A not-applicable binding's reason had no source in the map | A binding recipe may ship no probes when every probed library rule of its contract is not-applicable, with reasons ([example](examples/map/recipes/binding/dear-imgui--diagnostics.json)) |
| Choices header naming not mechanical (`"C++20"` → `20`) | One naming rule (SPEC §4.10); `cxx_standard` is an integer |

## Fresh-agent re-test

A second agent with no conversation history ran the same test after ADRs 0078–0082. It found nothing foundational in the design: it could carry slice 1 through on the spec's defaults. Procedurally, ticket 01 waits for this review's sign-off and the merge to `main` (ADR-0078).

### Its questions

| # | Question | Its rating | Resolution |
| --- | --- | --- | --- |
| 1 | Is the definition of done signed off, so `exploration` can merge to `main`? | Gate | The product owner's sign-off of this review |
| 2 | Will the API key secret and monthly cap exist before ticket 09? | Minor | Ticket 09 lists them as a prerequisite the product owner provides (ADR-0081) |
| 3 | Seeded violations for composition rules would patch generated files, and slice 1's testbed can't exercise some of them | Minor | Seeds are owned by whoever owns the file the mistake belongs in, never a generated file; manifest seeds are JSON Patches; `verify --seed` applies them; slice 1 proves composition seeds against the fixture composition (SPEC §6.1, §11.1) |
| 4 | Where do the executable and top-level CMake file come from? | Minor | The manifest names the executable; `start` writes `main.cpp` and the top-level CMake file from the map's testbed template when the repo has none (SPEC §2.3, §4.7) |
| 5 | Where do `find` and `show` put their results? | Minor | The CLI result's `entries` ([example](examples/composition/cli-results/find-physics-3d.json)) |
| 6 | How does the startup order entry points within a phase? | Minor | Entry points declare the lifecycle steps they perform; an entry point runs after those performing the steps in its `after` list (SPEC §5.6) |
| 7 | Which skill directories get the router skill? | Minor | `.claude/skills/automap/` and any other agent's skill directory already present (SPEC §5.7) |
| 8 | Smaller defaults | Minor | All three platforms by default; 120 smoke frames; the probe results file; probe builds in their own directory; verifier id `local`; the vcpkg tool release and Automap registry location in `map.json`; system dependencies as platform facts (SPEC §5.4–§6.3) |

### Contradictions and gaps it found

| Finding | Resolution |
| --- | --- |
| SDL3's example data gave SDL_GPU no Vulkan backend on Linux: at vcpkg baseline `d186940`, the sdl3 port builds it only with its non-default `vulkan` feature | SDL3's default variant enables `vulkan`; ticket 08 checks it (verified against the port and SDL's CMake, 2026-10-30) |
| Ticket 08 asked for headless build options the spec had removed | Ticket 08 asks for the `vulkan` feature, the headless environment and system dependencies |
| Ticket 07 cited ADR-0027's table, which lists fewer rules than SPEC §8.2 | Ticket 07 cites SPEC §8.2 |
| The fixture registry was described as overlay ports and as a filesystem registry, neither with a commit | Tests create a git registry from a directory in the fixture map, so pins keep a commit |
| The `none` level appeared only in schemas | Defined in SPEC §6.2 and `CONTEXT.md` |
| `not-applicable` contribution policy missing from `CONTEXT.md` | Added |
| `platform.targets-supported` couldn't be checked as worded from one platform's evidence | Reworded per platform; the CLI also refuses to resolve unsupported platforms |
| No CLI surface for applying seeds | `verify --seed` |
| Slot descriptions differed between SPEC and examples | SPEC quotes the example files |
| The newest handoff was stale | A new handoff closes this review |
| This container can't download from GitHub archives or releases, so vcpkg can't bootstrap or fetch here | Environment, not design: real-build tickets run on GitHub-hosted runners or with the network allowed |
