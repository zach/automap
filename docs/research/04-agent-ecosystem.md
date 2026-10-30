# 04 — The coding-agent ecosystem Automap must plug into

**Scope.** This note covers the agent-facing surface Automap has to target. It looks at six areas: the Agent Skills format and how skills are distributed; router-skill patterns and what is known about triggering accuracy; MCP (the current spec, the registry, CLI-vs-MCP); AGENTS.md and CLAUDE.md conventions, plus whether the C/C++ libraries Automap would map ship agent files; existing game-dev skill packs and engine MCP servers; and evidence on agents writing C/C++ and build files. The goal is to test the starting hypothesis, not to assume it.

**Method and limits.** Many primary hosts were blocked from this session: agentskills.io, modelcontextprotocol.io, developers.openai.com, docs.github.com, arxiv.org, huggingface.co, vendor engine sites and most library sites. Where the docs are published from a public Git repository, I read the canonical source in that repository at a stated commit date (for example `agentskills/agentskills/docs/specification.mdx` instead of agentskills.io). Claude Code docs, Anthropic posts, the MCP blog and the MCP registry were reachable directly. The web-search quota ran out partway through. Claims that rest only on a search-result snippet or on secondary coverage are labelled as such.

Accessed: 2026-10-29

---

## Key findings

- **[Verified]** Agent Skills (`SKILL.md`) is an open spec. Its repository was started 2025-12-16, and since then only clarifications have landed (latest 2026-08-04). It has six frontmatter fields, **no version or dependency field**, and three-tier progressive disclosure: about 100 tokens of metadata, a body under 5,000 tokens and 500 lines, and resources loaded on demand [1][2]. The project lists 46 compatible clients, including Claude Code, ChatGPT/Codex, GitHub Copilot, VS Code, Cursor, Gemini CLI, Junie and Kiro [4].
- **[Verified]** Versioning and updates live in the *packaging* layer, which is fragmenting:
  - Claude Code plugins and marketplaces: semver dependency ranges resolved against `<plugin>--v<version>` git tags, `ref`/`sha` pinning, auto-update off by default [6][7].
  - **Agent Plugins 1.0** (Aug 2026): a vendor-neutral package format for skills plus MCP config. Its maintainers come from Amazon, Cursor, Microsoft, OpenAI and Vercel; Anthropic is not listed, and the format has no dependency mechanism yet [21].
  - `gh skill` writes provenance (repo, ref, tree SHA) into frontmatter [16]; `npx skills` keeps a content-hashed `skills-lock.json` [13].
- **[Verified]** Skill triggering is description-only and has a budget. In Claude Code the skill listing gets **1% of the context window**. Each description is capped at 1,536 characters, and when the listing overflows, the least-used skills lose their descriptions first [5]. Anthropic says Claude tends to *under*-trigger skills [23]. Codex has a separate catalog token budget and is running deterministic skill selectors (BM25, character n-grams, LRU, rank fusion) in **shadow mode** [19].
- **[Verified/Inference]** Neither vendor docs nor the search results I saw contained a controlled study of triggering accuracy versus the number of installed skills. The nearest evidence comes from search snippets of two arXiv papers:
  - Skill benefits fall toward the no-skill baseline when agents must pick from a realistic 34k-skill pool [49].
  - In SWE-Skills-Bench, 39 of 49 public skills gave zero improvement, and 3 made results worse (by up to 10%) because of **version-mismatched guidance** [48].

  The mature game-dev packs both use explicit routing rather than description triggering: a router skill with a project-fingerprint table [25], and a deterministic BM25 index with a CI-gated routing eval [26].
- **[Verified]** The current MCP spec is **2026-07-28**. It removes sessions and the initialize handshake, adds `server/discover`, and replaces server-initiated requests (including elicitation) with Multi Round-Trip Requests. Tasks become an extension, list results become cacheable, and Roots, Sampling and Logging are deprecated [29]. A **Skills extension** (SEP-2640, Final) serves `SKILL.md` over MCP, but host support is thin: ChatGPT and fast-agent are listed as partial, Claude Code is not listed [30].
- **[Verified]** First-party guidance now favours a CLI when one exists:
  - Claude Code docs: "If that system has a CLI, that's usually the easiest way in… For systems without a CLI, there's MCP" [10].
  - GitHub's gh-aw dropped built-in Playwright MCP for a CLI, citing schema token cost and attack surface [37].
  - Context7 and Summer Engine both ship "CLI plus skill, no MCP required" modes alongside MCP [39][26].
- **[Verified]** AGENTS.md is a founding project of the Linux Foundation's Agentic AI Foundation, alongside MCP [31]. Claude Code now reads `AGENTS.md` natively when no `CLAUDE.md` exists (v2.1.277+) [8]; Codex concatenates AGENTS.md files from the project root down to the working directory [19]; Copilot uses the nearest one [17].
- **[Verified]** Of the C/C++ libraries checked:
  - **SDL** ships `AGENTS.md`/`CLAUDE.md`, but they say *generative AI must not be used to contribute* [43].
  - **vcpkg** ships `AGENTS.md`, a one-line `CLAUDE.md` (`@AGENTS.md`), and seven maintainer-facing skills, including a catalogue of CI and build failure patterns [44].
  - **assimp** ships a contributor AGENTS.md [45].
  - Dear ImGui, Jolt, raylib, flecs, bgfx, GLFW, GLM, Magnum and Conan have none in their repo trees; no llms.txt was found in their repos or in their site repositories [46]. Conan has an official MCP server, `conan-mcp` 0.1.0 [40].
- **[Verified/Inference]** The game-dev skill packs found all target an existing engine: awesome-gamedev-agent-skills (77 skills plus a router across 12 engines) [25], Summer Engine (137 skills for its Godot-based engine) [26], and GameForge and game-dev-studio [28]. **No tooling for composing an engine from libraries was found** (the searches are listed in §5).
- **[Verified]** Agent C++ evidence is thin and not about composition. On Multi-SWE-bench's C++ subset (129 issue-fix tasks from five library repos), the best submission resolves 47.3% (unverified submission, 2026-07-14) [47]. vcpkg's own agent-facing failure catalogue covers CMake config-package names, linker and CRT mismatches, and C++26/libc++ include breakage [44].

---

## 1. The Agent Skills format, its adoption and distribution

**Format.** A skill is a directory named after the skill, holding a `SKILL.md` and optional `scripts/`, `references/` and `assets/` [1]. The frontmatter fields are:

| Field | Required | Constraints |
|---|---|---|
| `name` | yes | 1–64 characters, `a-z0-9-`, no leading/trailing or doubled hyphens, must match the directory |
| `description` | yes | 1–1024 characters; says what the skill does and when to use it |
| `license` | no | |
| `compatibility` | no | up to 500 characters; environment requirements |
| `metadata` | no | string→string map |
| `allowed-tools` | no | experimental |

Other conventions from the spec [1]:

- **Progressive disclosure:** about 100 tokens of metadata are loaded at startup, the body (under 5,000 tokens recommended) on activation, and resources only as needed. The guidance is to keep `SKILL.md` under 500 lines.
- **File references:** keep them one level deep and relative to the skill root.
- **Validation:** `skills-ref validate`.

The spec has **no version field**. Versions can only go in `metadata` (the spec's own example uses `version: "1.0"`). Git history shows the spec has only received clarifications since its 2025-12-16 start, the latest on 2026-08-04 [1]. Contribution rules [4]:

- The bar for additions is high ("when in doubt, leave it out").
- The project is not accepting skill submissions; it keeps no community skill directory.
- Client listings are reviewed by Anthropic.

**Client conventions** [2]:

- Clients should scan both their native directory and the cross-client `.agents/skills/` path, at project and user scope.
- Project skills override user skills.
- Project skills from untrusted repositories should be gated behind a trust check.
- Skill content should be protected from context compaction.
- Each catalog entry costs about 50–100 tokens.

**Adoption** [4][5][13][14][18]:

- The agentskills client list names 46 products, including Claude Code, Claude, ChatGPT & Codex, GitHub Copilot, VS Code, Cursor, Gemini CLI, Junie, Kiro, Amp, Goose, OpenCode, OpenHands, Roo Code, Factory, TRAE, Mistral Vibe, Databricks, Snowflake and Tabnine [4].
- `npx skills` claims support for 79 agents [13].
- Copilot reads `.github/skills`, `.claude/skills` and `.agents/skills`, and supports skills in the cloud agent, CLI, code review and IDE agent modes [14].
- Gemini CLI reads `.gemini/skills` or `.agents/skills` (the latter wins), activates skills through an `activate_skill` tool, and asks the user to consent first [18].
- Cursor's own docs could not be fetched; Cursor appears in the clients list [4] and installs to `.agents/skills/` [13].

**Claude Code extensions** [5]:

- Extra frontmatter: `when_to_use`, `disable-model-invocation`, `user-invocable`, `disallowed-tools`, `model`, `effort`, `context: fork` (run in a subagent), `paths` (glob-gated auto-activation), `hooks`, `arguments`.
- Frontmatter that sticks to the spec's six fields stays portable to the API and claude.ai.
- The listing budget is 1% of context, adjustable with `skillListingBudgetFraction`. `skillOverrides` can set low-priority skills to `"name-only"`.
- After compaction, invoked skills are re-attached: the first 5,000 tokens of each, up to 25,000 tokens in total, most recent first.

**Distribution and versioning**

- **Claude Code plugins and marketplaces** [6][7]:
  - A `marketplace.json` lists plugins from `github`, `git-subdir`, `url` or `archive` sources; archives are pinned by `sha256`.
  - `ref`/`sha` pinning holds users on a version. Users get an update only when the computed `version` changes; if `version` is omitted, they track commits.
  - Background auto-update is off for third-party marketplaces until a user or admin enables it.
  - Private marketplaces rely on the user's git credentials. `headersHelper` mints tokens for private archive registries.
  - Plugin `dependencies` accept semver ranges (`~2.1.0`) resolved against `<plugin>--v<version>` tags, and a plugin can be a pure "bundle" of dependencies. Stable and early-access channels are done with two marketplaces; there is no channel concept.
- **Codex** [19][20]:
  - Plugins are `.codex-plugin/plugin.json` plus skills, MCP and app config, listed in `.agents/plugins/marketplace.json`.
  - The protocol has `MarketplaceAdd/Upgrade` and `PluginInstall` methods.
  - `openai/skills` is deprecated in favour of `openai/plugins`.
- **Copilot** [15][16]:
  - Plugins come in two formats, Agent Plugins 1.0 or legacy. The default marketplace is `github/awesome-copilot`, and installs use `copilot plugin install NAME@MARKETPLACE`.
  - `gh skill` (public preview, gh ≥ 2.90.0) supports `search`, `preview`, `install OWNER/REPO SKILL@TAG|SHA`, `--pin` and `update`. It writes provenance (source repo, ref, tree SHA) into the skill's frontmatter, and its docs warn that skills are not verified and may contain prompt injection.
- **`npx skills`** (npm `skills` 1.7.4) [13]:
  - `add`, `check` (exits 1 when updates exist, so it can gate CI) and `update`.
  - Installs as a symlink to one canonical copy, records a `skills-lock.json` with content hashes, and supports `--frozen-lockfile`.
  - Experimental `sync` links skills shipped inside npm dependencies, so a skill updates together with its package.
- **Agent Plugins 1.0.0** [21][22]:
  - A plugin is a directory with `plugin.json` (`$schema` pins the spec version), `skills/` and `mcp.json`, plus reverse-DNS client-extension directories.
  - Technical steering committee: Amazon, Cursor, Microsoft, OpenAI, Vercel. Plugin-to-plugin dependencies are listed as future work.
  - Launch clients per the Vercel and GitHub changelogs (via search only): ChatGPT/Codex, Cursor, Copilot, Kiro and VS Code [22].
  - **[Unverified]** Whether Claude Code supports Agent Plugins: none of the Claude Code docs pages I fetched mention it.

---

## 2. Router-skill patterns and triggering limits

**Why routing matters.**

- The description "carries the entire burden of triggering."
- Agents tend to skip skills for tasks they think they can do unaided [3], and Anthropic recommends "pushy" descriptions because Claude under-triggers [23].
- When the Claude Code listing overflows, the least-used skills lose their descriptions and so the keywords that would match them [5].
- Codex budgets the catalog in tokens and is testing harness-side "cheap skill selectors" in shadow mode, without changing what the model sees [19]. This is evidence that vendors see pure description matching as a scaling problem.

**Concrete patterns.**

1. **Router skill plus fingerprint table** — awesome-gamedev-agent-skills `router/SKILL.md` (266 lines) [25]:
   - Detect exactly one engine from project files (`project.godot`, `*.uproject`, `Cargo.toml` with `bevy`, and so on) and read its version.
   - Classify the task into disciplines, at most one genre, and workflows; resolve the *minimal* skill set.
   - Read only those `SKILL.md` bodies and compose them in order: engine → discipline → genre → workflow.
   - Fall back explicitly for unknown or custom engines and "state the gap" rather than invent a skill.
   - The repo is also published as per-engine Claude Code plugins, so users can install the router plus one engine.
2. **Typed registry plus deterministic search** — Summer Engine [26]:
   - Every skill has a `resource.yaml` sidecar with `id`, `version`, `use_when`, `facets`, `compatibility.engine: ">=4.6"`, `related` skills and templates, and `status`.
   - These feed a 296-entry index (137 skills, 131 tools, 19 templates, 9 references), searched with BM25 through `summer ops search` or `summer_docs_search`.
   - A routing eval (154 queries plus a held-out set) reports recall@5 = 1.0, recall@1 = 0.84 and 15 "hijacked" queries. It gates CI, and unanswerable asks are tracked as an authoring backlog.
3. **Session-start bootstrap** — obra/superpowers injects a `using-superpowers` skill through a SessionStart hook: "if there is even a 1% chance a skill might apply… you MUST invoke" [24].
4. **Tool-side analogue** — Anthropic's Tool Search Tool defers tool definitions and searches them with regex or BM25. Anthropic reports an 85% token cut and accuracy gains from 49% to 74% (Opus 4) and from 79.5% to 88.1% (Opus 4.5) on large tool libraries [34].

**Published eval guidance** [3][23]:

- About 20 trigger queries, half should-trigger and half near-miss should-not-trigger, with realistic phrasing.
- At least 3 runs per query and a trigger-rate threshold of 0.5.
- A fixed 60/40 train/validation split, about 5 iterations, and selection by validation score.
- `skill-creator` automates the loop.

**Evidence on scale.**

- **[Unverified]** I did not find a controlled measurement of accuracy against installed-skill count. The search for one returned only practitioner commentary.
- From arXiv search snippets (the papers could not be fetched):
  - "Agentic Skills in the Wild" (2604.04323): with a 34k-skill pool, benefits approach no-skill baselines in the hardest settings. Retrieval plus query-specific refinement raised Claude Opus 4.6 on Terminal-Bench 2.0 from 57.7% to 65.5% [49].
  - SWE-Skills-Bench (2603.15401): mean gain +1.2%, with the version-mismatch regressions noted above [48].
  - SkillsBench: curated skills +16.2 points, self-generated skills about none (secondary coverage) [50].
  - A smart-contract-auditing study: at most 27 of 83 skills were ever invoked, and invocation was strongly model-dependent [51].

---

## 3. MCP: current spec, registry, and CLI vs MCP

**Spec history.** The versioning page names **2026-07-28** as current [29]. Features by release:

- **2025-06-18:** structured tool output (`outputSchema` / `structuredContent`), elicitation, resource links, OAuth resource-server classification.
- **2025-11-25:** URL-mode elicitation, icons, tools in sampling, Client ID Metadata Documents (CIMD), experimental tasks.
- **2026-07-28:**
  - Stateless protocol: no `initialize` handshake and no `Mcp-Session-Id`; version and capabilities travel in each request's `_meta`.
  - New mandatory `server/discover` RPC; `subscriptions/listen` replaces resource subscribe and the GET stream.
  - **Multi Round-Trip Requests** (`resultType: "input_required"`) replace server-initiated sampling, elicitation and roots requests.
  - Tasks move to the `io.modelcontextprotocol/tasks` extension.
  - Cacheable list results (`ttlMs`, `cacheScope`); deterministic `tools/list` order; any JSON Schema 2020-12 in input and output schemas.
  - Deprecated: Roots, Sampling, Logging and Dynamic Client Registration (CIMD preferred); a 12-month deprecation policy now applies.

**Governance and roadmap.**

- MCP was donated to the **Agentic AI Foundation**, a directed fund under the Linux Foundation co-founded by Anthropic, Block and OpenAI. The founding projects are MCP, goose and AGENTS.md (2025-12-09) [31].
- The 2026-08-22 roadmap prioritises agentic messaging (webhooks and channels), HTTP-native transport (including Streamable HTTP over stdio), agent identity, and "Server Cards" `.well-known` metadata so a server can be described without connecting to it [31].
- `.mcpb` bundles package local servers (2025-11-21) [31].

**Extensions** [30]:

- Apps, Tasks, OAuth client credentials, Enterprise-Managed Auth, and **Skills**.
- The Skills extension exposes `skills/list` and `skills/get` over Resources, with `skill://` URIs and per-file `sha256` digests. Reading a `SKILL.md` does not activate it; the host's own loading path does.
- Host matrix: ChatGPT partial, fast-agent partial; Claude Code is listed only for Tasks.

**Registry** [32]:

- `registry.modelcontextprotocol.io` is live and serves `server.json` (schema 2025-12-11) entries with `status`, `publishedAt` and `isLatest`.
- Its README still says "preview", with a v0.1 API freeze since 2025-10-24.
- GitHub documents how enterprises point Copilot at an MCP registry [15].

**Client behaviour.**

- In Claude Code, MCP **tool search is on by default**, and a v2 client runtime speaks the 2026-07-28 revision and the tasks extension [9].
- Connected servers "add their tool names and instructions to the context even when you're not using them" [10].

**Patterns for wrapping a CLI or a knowledge base.**

- *A thin MCP server over a CLI:* `conan-mcp` exposes seven tools (`get_conan_profile`, `list_conan_profiles`, `create_conan_project`, `list_conan_packages`, `install_conan_packages`, `get_conan_licenses`, `scan_conan_dependencies`) and takes the binary path from an environment variable [40].
- *One package, three surfaces:* `summer-engine` ships an MCP server, a CLI and skills. Its CLI-only mode installs skills without an MCP entry, and "each skill ends with a table of the command for every tool it names." It also exposes `summer_search` / `summer_call` meta-tools, plus a hosted MCP for chat apps with no shell [26].
- *A knowledge base over two tools:* Context7's `resolve-library-id` and `query-docs`, or the equivalent `ctx7 library` / `ctx7 docs` CLI with an auto-installed skill [39].
- *Code execution over MCP:* tools are presented as files the agent's code calls; Anthropic's example drops from 150,000 to 2,000 tokens [33].
- *Tool-design rules:* consolidate chained operations, namespace tools, offer a `response_format` of concise or detailed, paginate and truncate (Claude Code caps responses at 25k tokens) [35].

**CLI vs MCP: published positions.**

- Anthropic: prefer an existing CLI and use MCP for systems without one [10]. MCP is for external data and actions, skills for knowledge and workflows [11]. Skills were described as complementing MCP "by teaching agents more complex workflows that involve external tools" [36].
- GitHub gh-aw removed built-in Playwright MCP in favour of `@playwright/cli` (tokens and attack surface), and kept MCP opt-in for stateful exploratory automation [37].
- Mario Zechner (search snippet): many MCP servers merely reimplement CLIs, such as the GitHub MCP vs `gh` [38].
- **[Unverified]** OpenAI's own Codex guidance on skills vs MCP; developers.openai.com was blocked. What is verified is that Codex plugins bundle skills, MCP and apps together [20].
- The consistent remaining case for MCP is hosts with no shell, OAuth or multi-tenant auth, and long-lived stateful sessions [26][37].

---

## 4. AGENTS.md and CLAUDE.md conventions, and what C/C++ libraries ship

**AGENTS.md** [41][31]:

- Plain Markdown, "a README for agents". Nested files are allowed and the closest one wins; explicit user prompts override it.
- The site claims 60k+ open-source projects use it and lists Codex, Cursor, Gemini CLI, Jules, Amp, Factory, VS Code, Zed, Warp, Devin, Windsurf and Junie among supporters.
- Governance: a founding Agentic AI Foundation project under the Linux Foundation, contributed by OpenAI.

**Tool behaviour.**

- **Codex** concatenates every `AGENTS.md` from the project root (found via `.git` markers) down to the working directory. It supports `AGENTS.override.md` and fallback filenames, within a `project_doc_max_bytes` budget [19].
- **Copilot** allows AGENTS.md files anywhere and uses the nearest [17].
- **Gemini CLI** defaults to `GEMINI.md`; `context.fileName` can list `AGENTS.md` [18].
- **Claude Code** [8]:
  - Reads `AGENTS.md` when no `CLAUDE.md`/`CLAUDE.local.md` exists (v2.1.277+); otherwise you import it.
  - `@path` imports resolve relative to the importing file, recurse up to 4 hops, need approval when they point outside the project, and are skipped inside code spans.
  - Subdirectory `CLAUDE.md` files load on demand. `.claude/rules/*.md` can be gated by `paths` globs.
  - **Target under 200 lines per file.** Imports do not reduce context cost. Move procedures into skills and rules that must always hold into hooks [8][10][11].

**What good generated files contain** (Claude Code guidance [8][11] and the assimp example [45]):

- Exact build and test commands with key CMake options.
- Layout and conventions, and always/never rules.
- A methodology line such as "Never answer from memory. First try it and then show the results."
- *Not* multi-step procedures or reference material; those go in skills.

**C/C++ libraries** (repository snapshots dated 2026-10-26 to 2026-10-29):

| Project | Agent-facing files | Notes |
|---|---|---|
| SDL (3.5.0 dev) | `AGENTS.md`, `CLAUDE.md` | Both say generative AI "should not be used in any way when contributing to SDL" [43]. SDL also ships a 2,305-line `docs/README-migration.md` and a Coccinelle `SDL_migration.cocci` for SDL2→SDL3 [43]. |
| vcpkg | `AGENTS.md`, `CLAUDE.md` = `@AGENTS.md`, `.github/skills/` ×7 | Maintainer and contributor-facing: PR review, CI-failure analysis, new-port audit [44]. |
| assimp | `AGENTS.md` | Contributor build, test and methodology rules [45]. |
| Dear ImGui, Jolt, raylib, flecs, bgfx, GLFW, GLM, Magnum, Conan, conan-center-index | none | No AGENTS/CLAUDE/GEMINI.md, `SKILL.md`, llms.txt, copilot-instructions or `.cursor` rules in the repo tree [46]. |
| Library site repos (raylib.com, sdlwiki, Jolt `gh-pages`, flecs docs) | no `llms.txt` | Sites that generate llms.txt dynamically could not be checked (hosts blocked) **[Unverified]** [46]. |

So upstream agent files, where they exist, are aimed at **contributors**, not **consumers**. None tells a consuming agent how to integrate the library. Conan offers consumer-side tooling through MCP [40]. Single-library skills do exist in personal repos: GitHub code search returned 330 `SKILL.md` hits mentioning raylib, 68 for flecs and 9 for JoltPhysics, almost all project-local [56].

**Doc-freshness services.**

- **llms.txt v2** (proposal date-modified 2026-08-10) adds per-path files and `.md` page alternates. It claims "thousands of sites" publish one and that Mintlify generates and Lighthouse audits them [42].
- **Context7** [39]:
  - Community-contributed; the README disclaims accuracy.
  - Refreshes only libraries that were *recently requested*, on popularity-tiered thresholds: top 100 daily, top 1,000 every 15 days, top 5,000 every 30 days, everything else 45 days.
  - Owners can claim a library, configure `context7.json`, manage versions, and trigger refresh from a GitHub Action.
  - **[Inference]** Long-tail C++ game libraries probably fall in the 45-day, demand-driven tier.

---

## 5. Existing game-dev skill packs, engine MCP servers, and composition tooling

| Pack | Structure | Engines | Freshness mechanism | Gap relevant to Automap |
|---|---|---|---|---|
| **awesome-gamedev-agent-skills** (Apache-2.0; HEAD 2026-10-28) [25] | 77 skills + router; also published as per-engine Claude Code plugins; installs via `npx skills` | Godot, Unity, Unreal, Ebitengine, Phaser, PixiJS, three.js, Bevy, pygame, LÖVE, Defold, Roblox | Version baselines re-checked 2026-10-24 (e.g. Godot 4.7, Unity 6.3 LTS, UE 5.8, Bevy 0.20); a quarterly review rule; frontmatter and link validator plus project-check scripts in CI; rule: "never combine snippets from different engine majors" | Assumes an existing engine. `router/references/custom-engines.md` has an "engine contract" table (build/run, lifecycle, time/input, state/assets) for *existing* custom engines, the closest analogue to Automap's integration contracts. |
| **Summer Engine** (`SummerEngine/summer`, MIT; npm `summer-engine` 4.9.1) [26][27] | 137 skills plus a typed registry; MCP (local and hosted), CLI and Claude Code plugin; `doctor` checks include `skills-version-stale` | Summer, an AI-native editor built on Godot 4 (vendor blog [27]; projects use `project.godot` and GDScript) | Per-skill `version` and engine compatibility ranges; routing, tool-conformance and outcome evals in CI; a capability lint forbids library entries that touch the network, credentials or package managers | Tied to one engine. Its docs' "60+ guides in 20 categories" undercounts the repo. |
| **AlterLab GameForge** (v2.x; last commit 2026-03-30) [28] | 34 skills: studio-role "agents", workflows, 3 engine specialists | Godot 4, Unity 6, UE5 | none observed | Process and role skills, not code integration. |
| **game-dev-studio** (2026-09-21) [28] | 10 skills, 10 agents | Godot, Unity, UE | none observed | `setup-engine` configures Godot 4.6 / "Unity 7" / UE 5.6; those baselines differ from [25]'s, an example of version drift between packs. |

**Engine MCP servers.**

- **Godot:** Coding-Solo/godot-mcp (last commit 2026-04-16): launch the editor, run projects, capture debug output, manage scenes [52].
- **Unity:**
  - CoplayDev "MCP for Unity" v10.3.0 (2026-10-04): 50 tools, Unity 2021.3–6.x, MIT [52].
  - Unity's own MCP server (in `com.unity.ai.assistant` and/or the Unity CLI) and an official Claude Code plugin with 29 skills are reported only by secondary sources, which disagree on its status **[Unverified]** [54].
- **Unreal:**
  - An experimental first-party MCP plugin in UE 5.8 is reported only by secondary sources **[Unverified]** [53].
  - chongdashu/unreal-mcp is experimental and was last updated 2025-04-22 [52].
- **Blender:** ahujasid's server, renamed `mcp-for-blender`, is "not made by Blender" [52]. An official Blender Lab MCP server (v1.0.x) is reported only by secondary sources **[Unverified]** [55].
- **flecs:** `1mamute/flecs-mcp` inspects a *running* flecs app through its REST API and is read-only by default [52].

All of these drive a running editor or app. None reasons about which libraries to combine.

**Composition tooling: none found.**

- Web searches run: "awesome-gamedev-agent-skills router…", "Summer Engine game engine AI skill library", Godot/Unity/Unreal/Blender MCP queries, and an agent-skills-for-game-dev survey (results in §2).
- GitHub code searches over `SKILL.md` files [56]:
  - `"custom game engine" CMake` → **0** results.
  - `imgui SDL3 integration` → 9 results, all project-local.
  - `SDL3 vcpkg` → 6, none about composition.
  - `"game engine" "from scratch" C++` → 204, mostly tutorials, game hacking and generic packs.
- The web-search quota ran out before I could run broader queries, such as composition tools outside the skills format. **[Inference]** The gap looks real, but a final sweep is advised.

---

## 6. Evidence on agents writing C/C++ and build systems

**Benchmarks.**

- Multi-SWE-bench covers 7 languages and 1,632 instances [47].
- Its C++ verified subset has 129 tasks from only five repos: nlohmann/json (55), fmt (41), simdjson (20), Catch2 (12) and cpp-httplib (1). The C subset has 128 tasks from ponyc, zstd and jq. These are single-repo issue fixes with existing build systems [47].
- Best C++ results in the experiments repo:

  | Submission | Date | Resolved |
  |---|---|---|
  | CodeArts-Agent + CodeArts-GLM-5.1 (`verified: false`) | 2026-07-14 | 61/129 = 47.3% |
  | InfCode + GPT-5 | 2025-11-10 | 25.6% |
  | MopenHands + Claude-3.7-Sonnet | 2025-03 | 14.7% |

- Frontier-lab agents largely haven't submitted, so these numbers understate current capability **[Inference]**.
- **No benchmark found** measures cross-library integration, CMake/vcpkg/Conan manifest authoring, or engine composition **[Unverified: search budget exhausted]**.

**Failure modes, from primary maintainer material.**

- vcpkg's agent skill reference `vcpkg-failure-patterns.md` [44] catalogues:
  - dependency-chain failures and CMake configuration failures;
  - compiler errors per toolchain;
  - **linker errors, including debug/release CRT mismatch**;
  - SHA512 mismatches;
  - post-build checks, notably **`vcpkg_cmake_config_fixup` with a `PACKAGE_NAME` that must match the upstream CMake export name, "not necessarily the vcpkg port name"**;
  - C++26/libc++ transitive-include breakage, case-sensitive paths, Android NDK macro collisions, version validation, and missing build tools.
- vcpkg's AGENTS.md also warns agents to create patches with `git diff --output` so CRLF line endings survive [44].
- **Version mismatch** is the regression mechanism SWE-Skills-Bench names [48].
- The libraries move quickly. Snapshots on 2026-10-26 to 2026-10-28 showed SDL at 3.5.0 dev with a large SDL2→3 migration surface, Dear ImGui at 1.93.1 WIP, flecs at 4.1.6, and Jolt at 5.6.0 with unreleased changes [43][46]. **[Inference]** Model training data will often lag the API actually in use.

**Mitigations seen in the wild.**

- Read the project's pinned version before proposing APIs, and keep its lockfile [25].
- Machine-checkable compatibility ranges on each skill [26].
- Upstream migration guides and semantic patches [43].
- Live docs retrieval [39].
- LSP diagnostics: Claude Code's official `clangd-lsp` plugin gives Claude compiler-grade diagnostics after each edit [12]. **[Inference]** It needs a `compile_commands.json`.
- Build-and-run verification loops: "run the engine/compiler/browser build that matches the project" [25]; playing the game and taking screenshots [26].
- CI that installs and builds real packages [44].

---

## Leads checked

| Lead | Verdict | Evidence |
|---|---|---|
| "2026 agent skill packs for game dev all assume an existing engine; none cover composing one." | **Confirmed, with a caveat** | All four packs examined target existing engines [25][26][28]. Code and web searches found no composition tooling [56]. Summer *is* an engine (Godot-based) rather than a composition kit. awesome-gamedev's custom-engine path supports existing in-house engines only [25]. The search budget ran out before a final sweep. |
| "awesome-gamedev-agent-skills: 66 skills plus a router across 10 engines." | **Corrected (stale)** | True at commit `b0f6705` (2026-06-26: "66 skills and a router across 10 engines"). As of 2026-10-28: **77 skills plus a router across 12 engines** (Ebitengine and Defold added in October) [25]. |
| "Summer Engine's skill library" | **Identified** | `SummerEngine/summer`: an AI-native engine built on Godot 4 [27]. Its MIT repo ships an MCP server, CLI and **137 skills** in a 296-entry typed registry with routing evals [26]. |

---

## Implications for Automap (inference)

1. **Make the CLI the source of truth and skills the agent's interface; MCP is optional.** Composition is deterministic work: resolve slots, check integration contracts, emit CMake and vcpkg/Conan manifests, build the known-good pairings. That belongs in a CLI with `--json` output, which every shell-capable agent can call [10][37]. Skills teach *when and how* to call it, and what the contracts mean. Wrap the CLI in a thin MCP server (in the style of conan-mcp or Summer) only for no-shell hosts and authenticated private registries [26][40]. Track the MCP Skills extension, but don't depend on it yet [30].
2. **Don't rely on description triggering for hundreds of library skills.** Keep **one** router skill always discoverable, plus a handful of contract skills (allocators, jobs, math types and so on). Keep per-library skills out of the global listing (`disable-model-invocation`, or `name-only` overrides) and have the router load them by path after a deterministic `automap find` index query. This combines the fingerprint router of [25] with the BM25 registry of [26], and matches the direction Codex is testing [19]. Ship routing evals (recall@k and hijack rate on held-out asks) and description trigger evals [3] as CI gates.
3. **Carry the version key in the map, not the skill format.** Agent Skills has no version field [1]. Put library version ranges, slot, contract and pairing IDs in `metadata` or a sidecar, as Summer's `resource.yaml` does [26]. Publish through **both** Claude Code marketplaces (semver dependencies, archive and `sha256` sources, `headersHelper` for private and NDA registries) [6][7] **and** Agent Plugins 1.0 for Codex, Copilot, Cursor and Kiro [21]. Install into `.agents/skills/` for everyone else [2][13]. Record provenance and hashes the way `gh skill` and `skills-lock.json` do [13][16].
4. **The generated project files should be small and point elsewhere.**
   - Each composed engine gets a generated `AGENTS.md` under 200 lines: build, test and run commands, the composition lockfile path, the slots in use, and "consult the Automap router before changing integrations".
   - Add a `CLAUDE.md` that just says `@AGENTS.md` (vcpkg's pattern) [8][44].
   - Use nested AGENTS.md files per subsystem directory, and Claude Code `paths` rules or skills scoped to a library's glue directory [5][8].
   - Emit `compile_commands.json` so clangd diagnostics work [12].
5. **Freshness strategy.**
   - Tie each library skill to the package version it describes, ideally shipped next to the port or recipe, like `npx skills` sync [13].
   - Rebuild every known-good pairing in CI when an upstream version bumps (vcpkg's install-and-check model) [44], and expose an `automap doctor` that flags stale skills [26].
   - Pull primary upstream material (migration guides, headers, release notes) into `references/` [43]. Context7 is a fallback, not the base [39].
   - Version-mismatched guidance is the regression risk the data names [48].
6. **Policy awareness.** SDL forbids AI-generated contributions [43]. Automap's glue lives in users' projects, but any "upstream a fix" recipe must check each library's policy first.
7. **Expect the hypothesis to need tightening.** Generic skills help little on average [48][49]. Automap's value has to come from version-specific, *verified* recipes and compositions plus executable checks, not prose packs.

---

## Open questions

- Will Claude Code adopt Agent Plugins 1.0, and will that format add dependencies? This decides whether one package can serve every agent [21].
- What is the measured triggering accuracy at 50, 200 and 1,000 installed skills on each client? No public data was found. Automap may need to produce it.
- Is consumer-facing integration knowledge welcome upstream? SDL's anti-AI stance suggests Automap must host it independently [43].
- How should NDA console content be distributed: private Claude Code marketplaces with `headersHelper`, a private MCP Skills server, or a private Context7-style index? Each has different audit and caching behaviour [6][30][39].
- Do official Unity, Unreal and Blender MCP servers exist as reported? These could not be verified from primary sources [53][54][55].
- Is there any composition-oriented tooling outside the skills ecosystem, such as engine generators or template systems? A final search sweep is needed.

---

## Sources

All accessed 2026-10-29. "Repo snapshot" means the date of the latest commit in a shallow clone made on that date.

1. Agent Skills Specification — `agentskills/agentskills` `docs/specification.mdx`, commit 69ef37e (2026-08-09); git history 2025-12-16→2026-08-04. https://github.com/agentskills/agentskills (canonical site https://agentskills.io/specification was blocked)
2. Agent Skills — "Adding skills support" (client implementation). https://github.com/agentskills/agentskills/blob/main/docs/client-implementation/adding-skills-support.mdx
3. Agent Skills — "Optimizing skill descriptions". https://github.com/agentskills/agentskills/blob/main/docs/skill-creation/optimizing-descriptions.mdx
4. Agent Skills — README, CONTRIBUTING.md, `docs/snippets/clients.jsx` (46 clients). https://github.com/agentskills/agentskills
5. Claude Code docs — "Extend Claude with skills". https://code.claude.com/docs/en/skills
6. Claude Code docs — "Create a plugin marketplace" and "Host and maintain a marketplace". https://code.claude.com/docs/en/plugin-marketplaces ; https://code.claude.com/docs/en/plugins/host-marketplace
7. Claude Code docs — "Plugin dependencies". https://code.claude.com/docs/en/plugins/dependencies
8. Claude Code docs — Memory (CLAUDE.md, imports, rules, AGENTS.md). https://code.claude.com/docs/en/memory
9. Claude Code docs — MCP (tool search, client runtimes, tasks). https://code.claude.com/docs/en/mcp
10. Claude Code docs — Best practices. https://code.claude.com/docs/en/best-practices
11. Claude Code docs — Features overview. https://code.claude.com/docs/en/features-overview
12. Claude Code docs — Code intelligence plugins (clangd-lsp). https://code.claude.com/docs/en/plugins/code-intelligence
13. `vercel-labs/skills` README, `docs/commands.md`, `docs/agents.md` (repo snapshot 2026-10-29); npm `skills` latest 1.7.4. https://github.com/vercel-labs/skills
14. GitHub Docs source — "Agent skills" concepts. https://github.com/github/docs/blob/main/content/copilot/concepts/customizations/agent-skills.md (repo snapshot 2026-10-29)
15. GitHub Docs source — "Plugins for GitHub Copilot", plugin-and-marketplace reference, CLI plugin install, configure-mcp-registry. https://github.com/github/docs/tree/main/content/copilot
16. GitHub Docs source — Cloud agent "add skills" (`gh skill`). https://github.com/github/docs/blob/main/content/copilot/how-tos/copilot-on-github/customize-copilot/customize-cloud-agent/add-skills.md
17. GitHub Docs reusable — `custom-instructions-agents.md` (AGENTS.md nearest wins). https://github.com/github/docs/blob/main/data/reusables/copilot/custom-instructions-agents.md
18. Gemini CLI docs — `docs/cli/skills.md`, `docs/cli/gemini-md.md` (repo snapshot 2026-10-28). https://github.com/google-gemini/gemini-cli
19. `openai/codex` — `docs/skills.md`, app-server protocol schemas (Marketplace*/Plugin*), `codex-rs/ext/skills/src/{dynamic_skill_selector.rs,config.rs}`, `codex-rs/core/src/agents_md.rs` (repo snapshot 2026-10-29). https://github.com/openai/codex
20. `openai/skills` README (deprecation notice) and `openai/plugins` README. https://github.com/openai/skills ; https://github.com/openai/plugins
21. Agent Plugins Specification 1.0.0 — `agentplugins/agent-plugins-spec` (spec/1.0.0.md, README, GOVERNANCE.md, MAINTAINERS.md, FUTURE_CONSIDERATIONS.md; snapshot 2026-08-19). https://github.com/agentplugins/agent-plugins-spec
22. Vercel changelog "Introducing Agent Plugins 1.0.0" (2026-08-06) and GitHub changelog "Agent Plugins 1.0 in VS Code, Copilot CLI, and the Copilot app" (2026-08-12) — via search results only. https://vercel.com/changelog/introducing-agent-plugins-1-0-0 ; https://github.blog/changelog/2026-08-12-agent-plugins-1-0-in-vs-code-copilot-cli-and-the-copilot-app
23. `anthropics/skills` — `skills/skill-creator/SKILL.md` (snapshot 2026-10-28). https://github.com/anthropics/skills
24. `obra/superpowers` — `skills/using-superpowers/SKILL.md`, `hooks/session-start` (snapshot 2026-10-21). https://github.com/obra/superpowers
25. `gamedev-skills/awesome-gamedev-agent-skills` — README, `router/SKILL.md`, `router/references/custom-engines.md`, `docs/VERSION-SUPPORT.md` (checked 2026-10-24), CI workflow; HEAD 2026-10-28; commit b0f6705 (2026-06-26). https://github.com/gamedev-skills/awesome-gamedev-agent-skills
26. `SummerEngine/summer` — README, AGENTS.md, `evals/README.md`, `evals/routing/baseline.json` (2026-10-25), `library/skills/*/resource.yaml`, `registry/generated/index.json`; package 4.9.1; HEAD 2026-10-28. https://github.com/SummerEngine/summer
27. Summer Engine blog — "Summer Engine vs Quasar: AI-Native Editors Built on Godot (2026)" (vendor; via search result). https://www.summerengine.com/blog/summer-vs-quasar
28. `AlterLab-IEU/AlterLab_GameForge` (snapshot 2026-03-30); `toilanguyen2910/game-dev-studio` (snapshot 2026-09-21). https://github.com/AlterLab-IEU/AlterLab_GameForge ; https://github.com/toilanguyen2910/game-dev-studio
29. MCP specification — changelogs for 2025-06-18, 2025-11-25, 2026-07-28; versioning page (current = 2026-07-28). https://github.com/modelcontextprotocol/modelcontextprotocol/tree/main/docs/specification (snapshot 2026-10-28)
30. MCP — Skills extension overview, extensions client matrix, Skills Over MCP WG charter. https://github.com/modelcontextprotocol/modelcontextprotocol/tree/main/docs/extensions ; https://github.com/modelcontextprotocol/ext-skills
31. MCP blog — "The New MCP Roadmap" (2026-08-22); "The 2026-07-28 Specification" (2026-07-28); "MCP joins the Agentic AI Foundation" (2025-12-09); "Adopting the MCP Bundle format (.mcpb)" (2025-11-21). https://blog.modelcontextprotocol.io/
32. MCP Registry — README (preview; API freeze v0.1, 2025-10-24) and live API `/v0.1/servers` (server.json schema 2025-12-11). https://github.com/modelcontextprotocol/registry ; https://registry.modelcontextprotocol.io
33. Anthropic Engineering — "Code execution with MCP" (2025-11-04). https://www.anthropic.com/engineering/code-execution-with-mcp
34. Anthropic Engineering — "Introducing advanced tool use" (2025-11-24). https://www.anthropic.com/engineering/advanced-tool-use
35. Anthropic Engineering — "Writing effective tools for agents". https://www.anthropic.com/engineering/writing-tools-for-agents
36. Anthropic Engineering — "Equipping agents for the real world with Agent Skills" (2025-10-16). https://www.anthropic.com/engineering/equipping-agents-for-the-real-world-with-agent-skills
37. GitHub gh-aw blog — "Why the Built-In Playwright Tool Is Now CLI-Only" (2026-09-01), source in `github/gh-aw` docs. https://github.com/github/gh-aw/blob/main/docs/src/content/docs/blog/2026-09-01-why-playwright-cli-replaces-playwright-mcp.md
38. Mario Zechner — "MCP vs CLI: Benchmarking Tools for Coding Agents" (2025-08-15) — via search snippet. https://mariozechner.at/posts/2025-08-15-mcp-vs-cli/
39. `upstash/context7` — README, `docs/library-updates.mdx`, `docs/adding-libraries.mdx` (snapshot 2026-10-28). https://github.com/upstash/context7
40. `conan-io/conan-mcp` README (snapshot 2025-12-03); PyPI `conan-mcp` 0.1.0 (2025-11-12). https://github.com/conan-io/conan-mcp
41. `agentsmd/agents.md` — README and site components (snapshot 2026-09-10). https://github.com/agentsmd/agents.md
42. Jeremy Howard — "The /llms.txt file, v2" (date-modified 2026-08-10). https://github.com/AnswerDotAI/llms-txt/blob/main/nbs/index.qmd
43. `libsdl-org/SDL` — AGENTS.md, CLAUDE.md, `docs/README-migration.md`, `build-scripts/SDL_migration.cocci`, `SDL_version.h` (snapshot 2026-10-28). https://github.com/libsdl-org/SDL
44. `microsoft/vcpkg` — AGENTS.md, CLAUDE.md, `.github/skills/*` incl. `analyze-ci-failures/references/vcpkg-failure-patterns.md` and `shared/review-vcpkg-pr-guide.md` (snapshot 2026-10-28). https://github.com/microsoft/vcpkg
45. `assimp/assimp` — AGENTS.md (snapshot 2026-10-29). https://github.com/assimp/assimp
46. Repo-tree scans (no agent files found): ocornut/imgui (2026-10-28; `IMGUI_VERSION` 1.93.1 WIP), jrouwe/JoltPhysics (2026-10-26; v5.6.0) and its gh-pages, raysan5/raylib (2026-10-29) and raysan5/raylib.com, SanderMertens/flecs (2026-10-28; 4.1.6), bkaradzic/bgfx, glfw/glfw, g-truc/glm, mosra/magnum, conan-io/conan, conan-io/conan-center-index, libsdl-org/sdlwiki. https://github.com/
47. Multi-SWE-bench — README (snapshot 2025-12-18) and `multi-swe-bench/experiments` `evaluation/{c,c++}/verified/*/results/results.json` (snapshot 2026-07-21); arXiv:2504.02605. https://github.com/multi-swe-bench/multi-swe-bench ; https://github.com/multi-swe-bench/experiments
48. Han et al., "SWE-Skills-Bench: Do Agent Skills Actually Help in Real-World Software Engineering?" arXiv:2603.15401 (2026-03-16) — via search snippet. https://arxiv.org/abs/2603.15401
49. "How Well Do Agentic Skills Work in the Wild: Benchmarking LLM Skill Usage in Realistic Settings," arXiv:2604.04323 — via search snippet. https://arxiv.org/abs/2604.04323
50. "SkillsBench: Benchmarking How Well Agent Skills Work Across Diverse Tasks," arXiv:2602.12670 — via secondary coverage. https://arxiv.org/abs/2602.12670
51. "Demystifying Agent Skills for Smart Contract Auditing," arXiv:2609.29454 — via search snippet. https://arxiv.org/abs/2609.29454
52. Engine MCP repos: Coding-Solo/godot-mcp (2026-04-16); CoplayDev/unity-mcp (v10.3.0, 2026-10-04); ahujasid/blender-mcp → mcp-for-blender (2026-10-29); chongdashu/unreal-mcp (2025-04-22); 1mamute/flecs-mcp (2026-09-30). https://github.com/
53. Secondary (Unreal 5.8 MCP): StraySpark, "Epic's Official MCP Plugin Is Here (UE 5.8)". https://www.strayspark.studio/blog/epic-official-mcp-plugin-ue5-8-vs-third-party
54. Secondary (Unity MCP): StraySpark, "Unity's Official MCP Server vs Unreal, Godot, and Blender: The 2026 Comparison"; Unity blog "MCP servers and game development". https://www.strayspark.studio/blog/unity-mcp-server-vs-unreal-godot-blender-2026 ; https://unity.com/blog/mcp-servers-game-development
55. Secondary (Blender Lab MCP): StraySpark, "Official Blender MCP server comparison 2026". https://www.strayspark.studio/blog/official-blender-mcp-server-comparison-2026
56. GitHub code search (REST, 2026-10-29): `filename:SKILL.md` with `JoltPhysics` (9), `flecs ecs` (68), `raylib` (330), `SDL3 vcpkg` (6), `imgui SDL3 integration` (9), `"custom game engine" CMake` (0), `"game engine" "from scratch" C++` (204).
