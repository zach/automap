# Automap

Automap is in **product exploration**: no product code yet. The exploration ends with `docs/spec/SPEC.md`, a spec an agent team can build from. The mission, phases, required spec contents and definition of done are in `docs/EXPLORATION_BRIEF.md`; current progress is in `docs/PLAN.md`. Read both before doing anything else.

## Repo map

| Path | What lives there |
| --- | --- |
| `docs/EXPLORATION_BRIEF.md` | The product owner's brief (verbatim). Source of truth for process. |
| `docs/PLAN.md` | Phase plan and the Phase 2 decision tree, with status. |
| `CONTEXT.md` | Glossary. Every term the spec relies on. |
| `docs/adr/` | Architecture decision records. |
| `docs/research/` | Phase 1 findings, one file per topic, plus `SUMMARY.md`. |
| `docs/OPEN_QUESTIONS.md` | Open questions and parked branches. |
| `docs/handoff/` | Session handoffs, newest last. Read the newest one when resuming. |
| `docs/spec/` | `SPEC.md` (Phase 4). |
| `docs/issues/` | Vertical-slice tickets (Phase 4). |
| `spikes/` | Throwaway prototypes (Phase 3 only). |

## Working rules

These override the upstream skills where they conflict.

- **Documents state the current version only.** The current version of any document is never in dialogue with its prior revisions: no "amended", "previously", "as first written", "no longer", changelog notes or superseded markers. When a decision changes, rewrite the affected ADR and docs in place; git history is the record of what changed.
- **One question at a time.** Grill the product owner (Zach) one question per message, walking the decision tree in `docs/PLAN.md` in dependency order. Upstream `grilling` asks a whole round at once; don't.
- **Recommend an answer with every question**, with a one- or two-sentence rationale grounded in research. Word it so "yes" accepts the recommendation.
- **Say so when a question needs Zach's taste or judgment** (positioning, what feels right to game developers) instead of hiding it behind a recommendation.
- **Don't ask what you can find out.** Research, the docs, or a quick experiment first.
- **Write docs as decisions land.** New or sharpened terms go into `CONTEXT.md` immediately. Every decision with real alternatives gets an ADR in `docs/adr/` with **Context, Options, Decision, Consequences** sections (fuller than the upstream minimal format, by request).
- **Challenge Zach** when an answer conflicts with an earlier decision, the research, or what developers need, before recording it.
- **An ADR records only what Zach agreed to.** If writing it surfaces a new consequence or assumption, ask about it instead of writing it in.
- **Don't build in assumptions that agents are weaker stewards of code than humans.** Coding-agent capability moves fast and benchmark figures go stale within months; Automap assumes agents can be the primary steward of an engine's code.
- **Keep `docs/OPEN_QUESTIONS.md` current.** By the end it may contain only items explicitly deferred past v1.
- **Research** goes in `docs/research/`, one file per topic, citing sources with access dates and separating verified facts from inference.
- **No product code.** Throwaway spikes only, in `spikes/<name>/`, via `/prototype`. Record the result in an ADR, then delete or archive the code.
- **Handoffs** go in `docs/handoff/YYYY-MM-DD-<slug>.md`, not the OS temp directory: the cloud container is ephemeral.
- **Short chat messages.** Summaries go in the docs, not the chat.

## Agent skills

### Installed skills

`.claude/skills/` holds skills vendored from [mattpocock/skills](https://github.com/mattpocock/skills) at commit `a496f55` (v1.3.2, MIT, see `.claude/skills/LICENSE-mattpocock-skills`). They were copied by hand because `npx skills` could not run in this environment. Installed for the exploration: `setup-matt-pocock-skills`, `grill-with-docs` (with `grilling` and `domain-modeling`), `prototype`, `handoff`, `to-spec`, `to-tickets`, `research`. Installed for the build ([ADR-0067](docs/adr/0067-build-team-skills.md)): `tdd`, `implement`, `implement-spec`, `code-review`, `diagnosing-bugs`.

The brief names two skills differently: `/to-prd` is `to-spec`; `/to-issues` is `to-tickets`. Upstream skills call the glossary `GLOSSARY.md`; this repo uses `CONTEXT.md` at the product owner's request, so read `GLOSSARY.md` in any skill as `CONTEXT.md`.

### Issue tracker

Local markdown under `docs/issues/`. See `docs/agents/issue-tracker.md`.

### Domain docs

Single-context: `CONTEXT.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.
