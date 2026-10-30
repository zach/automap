# Definition-of-done review

The brief's definition of done (section 6), checked item by item, with the evidence for each.

| # | Item | Status | Evidence |
| --- | --- | --- | --- |
| 1 | A fresh agent could begin slice 1 without asking anything foundational | Open: four questions below | Fresh-agent test, 2026-10-30 (below) |
| 2 | Every schema has a complete, valid example | Done | Nine schemas, 43 example files; `python3 docs/spec/check-examples.py` reports all valid, and fails on seeded mistakes |
| 3 | Every known-good pairing names its verification method | Done | [SPEC §8.5](SPEC.md#85-pairings) |
| 4 | Testbed, starter engine and custom integration each have a worked journey | Done | [SPEC §2.4–2.6](SPEC.md#24-journey-1-testbed-start-then-extend), with example files |
| 5 | Private libraries supported without special cases, shown with a hypothetical in-house library | Done | Acme's `acme-jobs` and `acme-renderer` ([examples](examples/private-map/)), [SPEC §4.5, §7](SPEC.md#45-libraries); the checker applies the same rules to public and private content |
| 6 | `CONTEXT.md` defines every term the spec relies on | Done | Glossary scan against SPEC's vocabulary; terms added for journey, reference composition, entry point, NDA platform, ordering constraint, console variant, contribution policy, reference thread pool, set-aside report, fixture map |
| 7 | Every significant decision has an ADR; `OPEN_QUESTIONS.md` holds only items deferred past v1 | Open: SPEC §12.3 to confirm | ADRs 0001–0077; `OPEN_QUESTIONS.md` holds OQ-008 and OQ-011, both deferred past v1 |
| 8 | Research claims are cited and dated | Done | Every note in `docs/research/` states its access date for all sources; SPEC cites the notes |

## Fresh-agent test

A general-purpose agent with no conversation history was given only the repo and asked whether it could begin slice 1 (tickets 01–09) without asking the product owner anything foundational, and to list every question it would ask. It could start ticket 01, but not finish slice 1 without design answers.

### Its questions

| # | Question | Its rating | Resolution |
| --- | --- | --- | --- |
| 1 | Has implementation been authorised, and on which branch? `CLAUDE.md` still says "No product code", and `main` holds only `LICENSE`. | Foundational | The exploration merges to `main` with a merge commit when this review is signed off; ticket 01 starts there and switches the instruction files ([ADR-0078](../adr/0078-exploration-merges-to-main-after-review.md)) |
| 2 | Where does the CLI read the public map in slices 1–2, and what goes in `map_release` before the first release in slice 3? | Foundational | The CLI reads the map from a directory given by `--map` or `AUTOMAP_MAP`, recorded as `git:<commit>` or `unreleased`; unreleased evidence is never signed ([ADR-0079](../adr/0079-map-from-a-directory-before-releases.md)) |
| 3 | How do recipe entry points become running code: a generated startup or a hand-written `main`? What supplies arguments such as the window and GPU device? | Foundational | The CLI generates the startup from the recipes; entry points take one context whose fields recipes declare; adopted engines may keep their own startup ([ADR-0080](../adr/0080-the-cli-generates-startup.md)) |
| 4 | Which agent runtime and model run the eval suite, with what credentials and budget? | Foundational | To decide with the product owner |
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
