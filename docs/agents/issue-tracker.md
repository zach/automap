# Issue tracker: Local Markdown

Issues and specs for this repo live as markdown files in `docs/issues/`.

## Conventions

- One feature per directory: `docs/issues/<feature-slug>/`
- The feature's spec from `/to-spec` is `docs/issues/<feature-slug>/spec.md`. The product-wide spec is separate, at `docs/spec/SPEC.md`.
- Implementation issues are one file per ticket at `docs/issues/<feature-slug>/<NN>-<slug>.md`, numbered from `01`, never a single combined tickets file
- Triage state is recorded as a `Status:` line near the top of each issue file. The `triage` skill is not installed, so the only state skills write is `ready-for-agent`; humans may also use `ready-for-human`, `needs-info`, `wontfix` or `done`.
- Comments and conversation history append to the bottom of the file under a `## Comments` heading

## When a skill says "publish to the issue tracker"

Create a new file under `docs/issues/<feature-slug>/` (creating the directory if needed).

## When a skill says "fetch the relevant ticket"

Read the file at the referenced path. The user will normally pass the path or the issue number directly.
