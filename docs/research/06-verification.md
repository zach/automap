# 06 — Verifying that a composition of C++ libraries works

**Scope.** How Automap could prove that a library *pairing* or a whole *composition* is "known-good", keep that claim fresh, and what it costs. Covers CI build matrices and pricing (late 2026), how package ecosystems verify at scale, test types for an engine composition (including headless graphics, golden images, determinism, sanitizers and ABI/define-mismatch checks), combinatorics and pruning with a worked cost estimate, signed evidence records, staleness, and using verification as the feedback loop for coding agents. Not covered: which libraries belong in which slot (see other research notes).

Accessed: 2026-10-29

**Method note.** In this session WebFetch could reach only `github.com`, `raw.githubusercontent.com`, `anthropic.com` and `code.claude.com`. `docs.github.com`, `learn.microsoft.com`, `devblogs.microsoft.com`, `slsa.dev`, `arxiv.org`, `buildkite.com`, `cirrus-ci.org`, `*.debian.org`, `csrc.nist.gov` and `google.github.io` were blocked. Where possible I read the primary source's *source repository* instead, for example `github/docs`, `MicrosoftDocs/vcpkg-docs`, `MicrosoftDocs/cpp-docs` and `slsa-framework/slsa`. I used these on their `main` branches as of the access date. The shared web-search budget ran out partway through. Claims based only on search-result summaries say so.

---

## Key findings

- **[Verified] Compute is cheap. macOS is the expensive part, and evidence expires quickly.** Standard GitHub-hosted runners are free for public repositories [3]. For private repositories the per-minute prices are: Linux x64 2-core $0.006, Linux arm64 $0.005, Windows x64 $0.010, macOS $0.062 [2]. Hosted T4 GPU runners cost $0.052/min on Linux and $0.102/min on Windows, and GPU and larger runners are billed even in public repositories [1][7]. On the Free plan, only 5 macOS jobs can run at once [8]. GitHub's current billing docs list no self-hosted runner fee [3]. Secondary sources say the $0.002/min fee announced for March 2026 was postponed [10]. Cirrus CI shut down on 2026-06-01 [11].
- **[Verified] A link to a CI run is not durable evidence.** From 2026-10-01, checks, workflow runs, commit statuses, logs and artifacts in public repositories are deleted after at most 90 days (private repositories: at most 400) [9]. A "known-good" badge that links to an Actions run will stop working within one quarter. Evidence must be exported and signed, for example as in-toto attestations via `actions/attest` [70]. Public repositories can do this on every plan. Private repositories need GitHub Enterprise Cloud [69].
- **[Verified] Large ecosystems verify that packages *build and link*, not that libraries *work together*.** vcpkg builds every affected port on 13 triplets for each pull request. It records expected failures in `ci.baseline.txt`, which has about 1,300 entries [18][19][20]. ConanCenter calls itself "not a testing service" but "a binary building service" [26]. Its per-package check, `test_package`, only shows that a consumer can link the package and run a small program [27]. Neither ecosystem tests contracts *between* libraries (allocator, threading, ownership). Automap's "known-good pairing" claim is stronger than anything these ecosystems publish, so Automap needs its own test layer.
- **[Verified] Testing a whole ecosystem is slow, and people still have to read the results.** One Rust crater run takes 3–6 days [29]. The Rust release process expects "quite a few regressions" per run, each triaged by hand [31]. A full vcpkg rebuild takes about 1,900–2,000 minutes per Windows triplet (from a search-result summary of a Microsoft blog post) [24]. These ecosystems keep it tractable by pruning: rebuild only what an ABI-hash change touches [15], and test features as core + each feature alone + all features together instead of every combination [21].
- **[Verified] Graphics tests can run without a GPU on hosted runners, but golden images are expensive to maintain.** Options: SDL's dummy and offscreen drivers (the offscreen driver supports Vulkan) [39][40]; Mesa lavapipe on Linux and Windows [43]; WARP for D3D12 [42]; Metal through a paravirtual device on macOS arm64 runners, which Bevy uses for Metal screenshots [35] and PyTorch excludes as too limited [55]; and WebGPU on SwiftShader in headless Chrome [49]. Golden images need a separate baseline for each backend and rasterizer, and a tolerance metric. wgpu uses FLIP with mean and percentile thresholds [45]. Filament keeps goldens per platform and backend, and treats comparisons on any other host architecture as advisory [52].
- **[Verified] Several integration contracts are cheap to check by machine, but only if the library exposes a hook.** Jolt packs 12 ABI-affecting defines into `JPH_VERSION_ID`. `VerifyJoltVersionID()` and `RegisterTypes` detect a mismatch and abort [57][58][59]. MSVC's `#pragma detect_mismatch` turns define mismatches into a link error (LNK2038) [62]. Jolt also checks fixed simulation hashes across more than 20 platform and compiler combinations [56]. Define mismatches keep happening in practice: a Jolt PR in October 2026 dealt with a `JPH_DEBUG` export mismatch [60].
- **[Inference] Worked estimate.** Assume 25 libraries, 4 platforms, and the latest plus the previous version of each, testing only *declared* pairings. One full sweep is about 4,400 runner-minutes: about $85 at private-repository list prices, $0 on public standard runners. Testing *all* pairs instead is about 16,400 minutes (about $335). Steady state, triggered by changes, is about 10–11k minutes per month (about $220). macOS is about 75% of the dollar cost. The main ongoing cost is people triaging failures and flaky tests, not compute.
- **[Verified] Tooling to trigger re-verification when a C++ dependency updates is uneven.** Dependabot supports vcpkg by bumping `builtin-baseline` [76]. Renovate supports Conan but, as of today, has no vcpkg, CMake or CPM manager [74][75]. Automap will need its own watcher for upstream releases.
- **[Verified] Agents are already being driven by C++ CI signals.** Since April 2026, the vcpkg maintainers ship agent skills that sort CI failures into regressions and known baseline failures [83]. Anthropic's parallel-agent C-compiler project found that the quality of the verifier, regression gates in CI, and test output written for an agent to read decided whether the agents made progress [78]. Automap's verification harness and its agent feedback loop should be the same artifact.

---

## 1. CI build matrices for C++

**Hosted runners, late 2026 [Verified].**

| Image | Labels | Architecture |
|---|---|---|
| Ubuntu 22.04 / 24.04 / 26.04 | `ubuntu-*`; `-arm` variants | x64 and arm64 |
| Windows Server 2022 / 2025 | `windows-2025-vs2026` is a VS 2026 variant | x64 |
| Windows 11 | `windows-11-arm`, `windows-11-vs2026-arm` | arm64 |
| macOS 15 / 26 | `macos-26`, `macos-15`; `macos-15-intel`, `macos-26-intel` | arm64 and Intel |

Sources: [5][6]. Xcode 27 is in preview, and macOS 14 is deprecated [6]. The runner-images project keeps at most 2 GA images and 1 beta image per OS, and moving a `-latest` label takes 1–2 months [6]. That rolling toolchain is itself a source of staleness (§6).

Hardware and limits:
- Public repositories get 4 vCPU and 16 GB on Linux and Windows. Private repositories get 2 vCPU and 8 GB [5]. Expect C++ builds in private repositories to take noticeably longer (**[Inference]**).
- macOS arm64 runners are 3-core M1 with 7 GB [5].
- GPU larger runners have 4 vCPU, a Tesla T4 with 16 GB VRAM, and 28 GB RAM, on Ubuntu or Windows partner images [7]. They have been generally available since July 2024 [85].
- A job can run for at most 6 hours, and a matrix can have at most 256 jobs [8].
- Concurrent jobs: Free plan 20 total and 5 macOS; Enterprise 500 total and 50 macOS; GPU jobs up to 100 on larger-runner plans [8].

**Pricing [Verified].**
- Per-minute prices are listed in finding 1 [2]. Larger runners run from $0.006 (Linux 2-core advanced) to $0.552 (Windows 96-core) [1].
- Included usage is now expressed in dollars: Free $12/month, Pro and Team $18, Enterprise $300. GitHub's own guide: "$300 USD covers 50,000 minutes on a standard Linux runner" [3][4].
- Cache storage is 10 GB per repository [3].
- The hosted-runner price cut of up to 39% on 2026-01-01, and the postponement of the self-hosted "cloud platform charge", come only from secondary or search-summary sources [10]. GitHub's own billing source does not mention any self-hosted charge [3].

**Alternatives.**
- Self-hosted runners are free under current docs [3]. They are the only option for real GPUs from a chosen vendor, Apple GPU features beyond the paravirtual device, and console devkits under NDA (**[Inference]**).
- Cirrus CI shut down on 2026-06-01. PostgreSQL moved to GitHub Actions "because it has unlimited runner time for public repositories" [11].
- **[Unverified]** Buildkite pricing: buildkite.com was blocked and I found no pricing in the `buildkite/docs` repository. The docs tree does include `getting_started_with_coding_agents.md`.

**Caching [Verified].**
- vcpkg binary caching keys each package by an ABI hash. The hash covers every port file, the triplet, the C and C++ compiler executables, features, the toolchain file, the CMake version, environment variables listed in `VCPKG_ENV_PASSTHROUGH`, and the ABI hashes of all dependencies [15].
- The `x-gha` (GitHub Actions cache) backend has been removed [15][16]. Supported backends include `files`, `nuget`, `http`, and experimental S3, GCS and Azure Blob [15].
- sccache has a GitHub Actions backend (`SCCACHE_GHA_ENABLED`). When it hits rate limits it skips storing results without failing the build [17].
- **[Inference]** For Automap, the vcpkg ABI hash is a ready-made identity for a "built library in context". It changes whenever the compiler, triplet or any dependency changes, which is exactly when a claim should be re-checked.

**Measured build times on public 4-core runners [Verified].** All include the project's own tests.

| Library | Run | Per-job time | Total job time |
|---|---|---|---|
| SDL (main) | 9 jobs | 3m26s (Steam Linux Runtime) to 6m37s (MSYS2) | 38m50s [12] |
| Jolt (master) | 52 jobs | median 3m21s; max 27m31s (UBSan) | 3h30m; wall clock about 18–40 min [13] |
| Dear ImGui | 9 jobs | Linux 2m30s, macOS 2m20s, Windows 5m59s | about 18 min; 6m03s wall clock [14] |

---

## 2. How package ecosystems verify at scale

**vcpkg [Verified].**
- **Triplets.** Azure Pipelines jobs cover 13 triplets: x86/x64/arm64 Windows (including static, static-md and release variants), arm64-osx, x64 and arm64 Linux, and three Android triplets [18][19].
- **Pinned machines.** Each job runs on a pinned image, for example `ImageVersionOverride 2026.10.27` on Windows and a Linux Docker image tagged `2026-10-19`. Jobs have a 2-day timeout [19].
- **What a PR builds.** For pull requests, `vcpkg ci` builds whatever changed against parent hashes. Ports marked `fail` in `ci.baseline.txt` are skipped, along with their dependents [20].
- **`ci.baseline.txt` states** [18]:
  - *unspecified*: must build unless blocked by a dependency.
  - `pass`: hard error even when blocked by a dependency.
  - `fail`: skipped on PRs, but rebuilt on scheduled runs, where an unexpected success counts as a CI failure. This forces negative claims to be retired once they stop being true.
  - `skip`: not built. Should be used for flaky or conflicting ports, with a comment explaining why.
- **Feature testing.** `x-test-features` installs core only, then each feature alone, then all features together. That is a linear number of builds, not exponential. `ci.feature.baseline.txt` records `feature-fails`, `combination-fails` and `options` (mutually exclusive features) [21][22].
- **Recorded evidence.** xUnit results per triplet, failure logs, `pr-hashes.json`, and a failing-ABI log [20].
- **Published evidence.** Each registry release lists port counts per triplet with a link to the Azure DevOps build. The 2026-10-16 release has 2,877 ports in total, from 2,271 on arm-neon-android to 2,835 on x64-linux [23].
- **Cost of a full run.** About 2,009 minutes for x64-windows before a 2025 caching optimisation and about 1,900 after (search summary of [24]).

**ConanCenter [Verified].**
- **Configurations.** Its documented set is MSVC 2019 x86_64, GCC 11 x86_64 and apple-clang 13 armv8, all Release, with shared and static where the recipe allows. That comes to about 30 binaries per library. The page is undated and looks stale [25].
- **Purpose.** It builds packages; it is "not a testing service", and unit tests should not run in recipes [26].
- **Evidence.** Build logs appear in the PR's Checks tab, and service status at status.conan.io [26][28].
- **`test_package`.** A minimal consumer that checks that consumers "will be able to link against it and reuse it". `can_run` skips running the program when cross-compiling [27].

**Rust crater [Verified].**
- **Method.** Builds and tests every crate on crates.io (plus some on GitHub) with a "start" and an "end" toolchain, isolated in Docker. A pass→fail transition is a regression [32][29].
- **Modes and selection.** Modes: `build-and-test`, `check-only`, `build-only`, `clippy`, `rustdoc`, `fix`. Crate selection: `full`, `top-N`, `random-N`, or an explicit list. Random-N is useful sampling precedent [30].
- **Duration.** Check-only takes 3–4 days; build and build-and-test take 5–6 days [29].
- **Triage.** Runs happen each beta cycle. Logs are read by hand: "a compiler failure is genuine, and test failures are mostly spurious" [31].

**Nix and Debian.**
- **Nix [Verified].** Hydra "regularly evaluates and builds Nixpkgs, updating the official channels when their jobs succeed". PRs get labels by how many packages they force to rebuild. 500 or more should target `staging`, and 1,000 or more is a mass rebuild. `nixpkgs-review` builds all dependents of a change [33].
- **Debian [Unverified].** Debian holds a package's migration to testing until its own autopkgtests and its reverse dependencies' tests pass (debci/britney). This is from my background knowledge; debian.org, salsa and ci.debian.net were all blocked.

**Bevy and Godot.**
- **Bevy plugins [Verified via search summary].** Bevy's plugin guidelines recommend a hand-maintained README table mapping Bevy versions to plugin versions, often with a row tracking `bevy_main` [34]. These are self-declared, unverified compatibility claims, which is the baseline Automap must beat.
- **Bevy's own CI [Verified].** Examples run on `macos-14` (Metal), Linux under `xvfb-run` with Mesa Vulkan, and Windows with `WGPU_BACKEND=dx12`. Screenshots go to Pixel Eagle, which compares against `main` on the same OS and fails the job on differences or missing images [35][36].
- **Godot [Verified].** CI runs `--headless --test`, ASan/UBSan/TSan editor builds, and a godot-cpp build. A GDExtension compatibility test loads method hashes from the reference tags `4.5-stable` and `4.4-stable`. That is a "latest plus previous two" window [37][38].

**[Inference] Lesson.** Every ecosystem publishes *per-package build status* plus a list of *expected failures*. None publishes a signed claim that "A and B work together". The useful patterns to copy:
- expected-failure baselines that are re-checked automatically;
- identity based on ABI hashes;
- feature and pair testing that grows linearly;
- comparing a start and an end toolchain;
- a fixed window of previous versions.

---

## 3. Test types for a composition

The ladder below runs from cheapest to strongest. Every rung costs runner time; rungs 3 and up also cost curation effort.

1. **Configure and build only.** Catches API drift and missing symbols. Example: bgfx CI builds only, and its only run steps are `--version` checks on its tools [53].
2. **Link and run a smoke test.** The equivalent of Conan's `test_package` [27]. For an engine: initialise every slot, run N frames headless, shut down cleanly, exit 0. Order of initialisation and shutdown is itself a contract under test.
3. **Headless backends.**
   - SDL's CTest sets `SDL_VIDEO_DRIVER` and `SDL_AUDIO_DRIVER` to `dummy`, plus `SDL_ASSERT=abort`. The default test timeout is 10 s, and `testautomation` gets 120 s [39].
   - SDL3 also has an `offscreen` video driver with framebuffer, EGL/GLES and Vulkan back ends [40].
   - SDL's dummy audio driver sleeps at about real-time speed (adjustable with `SDL_HINT_AUDIO_DUMMY_TIMESCALE`) and throws the output away [41]. Audio smoke tests can therefore check timing and callback contracts without a sound device.
   - sokol's tests default to `SOKOL_DUMMY_BACKEND` [54].
   - Bevy patches audio out on macOS runners because "audio timeouts after 15 minutes" on M1 runners [35]. Null backends are a requirement, not a nicety.
4. **Software GPU paths.**
   - Linux: Mesa lavapipe, pinned to a version (wgpu uses Mesa 26.1.3 and points `VK_DRIVER_FILES` at its own ICD) [43].
   - Windows: WARP for D3D12 (wgpu installs its own copy) plus lavapipe [42].
   - SwiftShader implements Vulkan 1.3, with OpenGL ES 3.1 available on top via ANGLE ("SwANGLE") [48].
   - Headless Chrome: WebGPU runs on SwiftShader with `--enable-unsafe-webgpu --use-webgpu-adapter=swiftshader` (search summary of Chromium's web-test config) [49]. Real GPUs need a working driver plus `--use-angle=vulkan`, and Google warns that "some combinations of flags allow detection of the GPU, but don't allow you to actually use the GPU" [50].
   - Dawn's end-to-end tests support SwiftShader as a CPU adapter, but some SwiftShader suites were skipped for flakily exhausting X11 resources (search summary) [84].
5. **Metal on hosted macOS.**
   - PyTorch found GitHub's macOS runners expose an "Apple Paravirtual device" that lacks `MTLGPUFamilyMac2`, and skips it [55].
   - Bevy nevertheless renders Metal screenshots on `macos-14` [35].
   - **[Inference]** Smoke-level Metal tests are viable on hosted runners. Feature-complete Metal coverage needs self-hosted Apple hardware.
6. **Golden images.**
   - wgpu compares against reference PNGs with NVIDIA FLIP. Each check is either a mean threshold or a percentile threshold. Suggested starting values are 0.01–0.1. On failure it writes `-actual` and `-difference` images, and a missing reference is bootstrapped from the current output [44][45].
   - FLIP is a perceptual difference metric for rendered images, validated with a user study [47]. It is BSD-3 licensed, at v1.7, with C++/CUDA, Python (`flip-evaluator`) and PyTorch implementations [46].
   - Filament's renderdiff keeps separate goldens per platform and backend (native and web never cross-compare). Goldens are rendered on Linux/aarch64; on any other host, "comparison results are advisory". New goldens are accepted through a commit-message token [52].
   - Chromium's Gold allows several approved baselines and fuzzy matching. Its parameters include `fuzzy_max_different_pixels` and `fuzzy_pixel_delta_threshold` (search summary) [51].
   - **[Inference]** Plan on one baseline per (composition, backend, rasterizer, host architecture), and tolerances tuned per test.
7. **Physics determinism.** Jolt builds with `CROSS_PLATFORM_DETERMINISTIC=ON` and checks fixed 64-bit simulation hashes for 5 scenes, in single and double precision. Coverage: Linux, Windows, macOS, ARM, RISC-V, PowerPC, LoongArch and WebAssembly, with non-native targets run under QEMU or Intel SDE [56]. Its docs warn that sources "must be compiled with the same defines" [59].
8. **Sanitizers and fuzzing.**
   - Jolt's UBSan job is its slowest at 27.5 min [13]. Godot runs ASan, UBSan and TSan builds [37].
   - MSVC supports `/fsanitize=address` since VS 2019 16.9, on x86/x64 only. MSVC lists UBSan and TSan as future work [61].
   - **[Inference]** Run TSan on Linux only, since it is the only practical way to test the threading/jobs contract. Fuzz the serialization contracts with libFuzzer on Linux.
9. **ABI and define-mismatch probes.**
   - Jolt's `JPH_VERSION_ID` packs 12 feature bits (double precision, determinism, FP exceptions, profiling, debug renderer, temp and custom allocators, object-layer bits, asserts, object stream) with the version number [57]. `VerifyJoltVersionID()` returns false on mismatch [58]. `RegisterTypes` traces mismatches and calls `std::abort` [59].
   - MSVC `#pragma detect_mismatch` produces LNK2038 for `_ITERATOR_DEBUG_LEVEL`, `RuntimeLibrary` and similar [62].
   - Real failure: in issue #2165, a Release shared Jolt could not be linked into Debug consumers, because exporting `JPH_DEBUG` made consumers reference `JPH::AssertFailed`. The October 2026 PR #2187 proposed a fix [60].
   - **[Inference]** Automap contracts should name, for each library, a probe call or a generated `static_assert`/`detect_mismatch` translation unit that compares the consumer's defines with the library's.

---

## 4. Combinatorics and pruning

**Size of the space.**
- Testing every version combination of a 25-library composition with 2 versions each is 2^25 ≈ 3.4×10^7 builds per platform. That is infeasible.
- All pairs of 25 libraries is C(25,2) = 300 pairs. With 2×2 version combinations and 4 platforms, that is 4,800 pair builds.

**Pruning techniques.**
- **All-pairs (covering arrays).** Microsoft PICT exists because "testing all pairs is an effective alternative to exhaustive testing and much less costly" [63] (**[Verified]**).
  - **[Inference, mathematical]** For 25 two-valued version choices, 8 full-composition builds cover every pairwise version interaction, since the smallest N with C(N−1, ⌈N/2⌉) ≥ 25 is 8.
  - With slots, each parameter has more values (library × version, say about 6). The lower bound is then 36 rows, and generators typically need a few more.
  - **[Unverified]** NIST's finding (Kuhn, Wallace & Gallo) that most failures come from interactions of one or two factors, and all observed ones from six or fewer. From memory; csrc.nist.gov was blocked.
- **Declared pairings only.** Test only edges along slot boundaries where an integration contract exists, for example physics↔jobs, physics↔math, renderer↔windowing and audio↔windowing.
- **Version window.** Latest plus previous, as in Godot's two reference tags [38].
- **Change-triggered re-verification.** Rebuild only what an ABI-hash change touches, as vcpkg does [15][20]. Route large blast radii to a batch lane, as nixpkgs routes changes affecting 500 or more packages to `staging` [33].
- **Sampling.** crater's `random-N` and `top-N` [30]. Anthropic's agents ran a deterministic 1%/10% test sample per agent, so that agents covered the suite together [78].
- **Linear feature testing.** core, then each feature alone, then all together [21].

**Worked cost estimate [Inference].**

Assumptions:
- 25 libraries; latest and previous version of each, so 50 library-versions.
- 4 platforms: Linux x64, Linux arm64, Windows x64, macOS arm64.
- Average job lengths, taken from §1 (3–6 min per configuration) [12][13][14], with binary caching of dependencies:
  - library build plus smoke test: 5 min;
  - pair glue build plus 300-frame headless smoke test: 3 min;
  - full composition build plus smoke test plus golden images: 15 min;
  - sanitizer run: 20 min.
- Each library declares about 4 integration edges, so 50 declared pairs.
- 5 reference compositions, each tested in 2 snapshots (all-latest and all-previous).
- Average private-repository price across the 4 platforms: (0.006 + 0.005 + 0.010 + 0.062) / 4 = $0.02075 per minute [2].

| Tier | Jobs | Runner-min | Private list $ |
|---|---|---|---|
| L0 per library (50 × 4) | 200 | 1,000 | $20.75 |
| L1 declared pairs (50 × 4 version combos × 4 platforms) | 800 | 2,400 | $49.80 |
| L1′ all pairs instead (300 × 4 × 4) | 4,800 | 14,400 | $298.80 |
| L2 compositions (5 × 2 × 4) | 40 | 600 | $12.45 |
| Sanitizers (5 × 2 × ASan/UBSan + TSan, Linux only) | 20 | 400 | $2.40 |
| **Full sweep, declared pairs** | 1,060 | **4,400** | **≈ $85** |
| **Full sweep, all pairs** | 5,060 | **16,400** | **≈ $335** |
| GPU golden tier (10 snapshots × 10 min, T4 Linux / Windows) | 10 / 10 | 100 / 100 | $5.20 / $10.20 |

Steady state per month:

| Component | Assumption | Runner-min | Cost |
|---|---|---|---|
| Upstream releases | about 12.5 per month; each re-runs L0 (20 min), its own pairs (8 combos × 4 platforms × 3 min = 96 min) and about 3 affected compositions (180 min), so about 300 min each | 3,750 | about $78 |
| Weekly L2 re-run | catches toolchain and image drift | 2,600 | about $54 |
| Monthly full sweep | declared pairs | 4,400 | about $85 |
| **Total** | | **about 10,750** | **about $220/month** |

Further notes on the estimate:
- macOS is about 75% of the dollar cost: $0.062 out of $0.083 per platform-minute.
- 2-core private runners could make jobs 1.5–2× longer, which gives about $330–440 per month.
- On public repositories with standard runners, compute costs $0. The binding limit is Free-plan concurrency: 5 macOS jobs means about 1,000 macOS minutes per sweep takes about 3.3 hours [8].
- For scale: vcpkg spends about 33 hours per *triplet* on a full rebuild [24].
- **Conclusion:** compute is not the constraint. The constraints are triage effort (crater: "quite a few regressions" every run [31]), flaky tests [84][35], and keeping golden baselines up to date.

---

## 5. Evidence records and provenance

**Building blocks [Verified].**
- **in-toto test-result predicate** (`https://in-toto.io/attestation/test-result/v0.1`). Required fields: `result` (`PASSED`, `WARNED` or `FAILED`) and `configuration` (a reference to the test configuration). Optional: `url`, `passedTests`, `warnedTests`, `failedTests` [64].
- **SLSA Verification Summary (VSA)** (`https://slsa.dev/verification_summary/v1`). Fields: `verifier.id`, `timeVerified`, `resourceUri`, `policy` (ideally with a digest), `inputAttestations`, `verificationResult`, `verifiedLevels`, `dependencyLevels`, `slsaVersion` [66].
- The in-toto predicate catalogue also has SPDX and CycloneDX SBOM types, `vulns`, `release` and `svr` [65].
- **GitHub artifact attestations** [67][69]:
  - Default attestations reach SLSA v1.0 Build Level 2; isolated reusable workflows reach Build Level 3.
  - Public repositories sign through the Sigstore public-good instance, with an immutable public transparency log.
  - Private repositories use GitHub's own Sigstore instance, which has no transparency log.
  - Available on every current plan for public repositories; private and internal repositories need Enterprise Cloud.
- **`actions/attest`** [70][71]:
  - Signs any `predicate-type` with a predicate of up to 16 MB, for up to 1,024 subjects.
  - Needs the `id-token: write` and `attestations: write` permissions.
  - `attest-build-provenance` is now a wrapper around it.
  - Consumers verify with `gh attestation verify … --predicate-type …` [68].
- **SBOMs for C++.**
  - vcpkg writes `vcpkg.spdx.json` (SPDX 2.3) per package and per triplet. The document name includes the package, triplet, version and ABI hash [72].
  - Conan provides `cyclonedx_1_4` and `cyclonedx_1_6` (experimental) [73].
  - `actions/attest` takes an `sbom-path` input [68].

**Proposed record [Inference].** An in-toto Statement with:
- **subject:** the digest of the *composition lock*, meaning the canonical manifest plus the resolved versions and source hashes. Optionally also the built binaries.
- **predicateType:** `https://automap.dev/verification/v0.1`, a superset of test-result.
- **predicate fields:**
  - `composition` (id; slots → library@version with source digest and vcpkg/Conan ABI hash);
  - `platform` (triplet, runner image and version, for example `windows-2025-vs2026@20261027`, compiler version);
  - `buildFlags` (defines that matter to contracts, such as `JPH_*`, plus `_ITERATOR_DEBUG_LEVEL` and the CRT);
  - `tier` (L0 to L5, see the ladder below) with per-test results;
  - `contracts` (which named contracts were exercised: allocator, jobs, init order and so on);
  - `goldenSet` digest and tolerance;
  - `timeVerified`, `validUntil`, `invalidatedBy` (a list of input changes that void the claim);
  - `evidenceUri` pointing to Automap's own storage, never only an Actions URL [9];
  - `verifier` (OIDC workflow identity).

A VSA-style record can then summarise "composition X passed policy P at tier T", for agents and the UI. For private registries and console platforms under NDA, sign with an organisation-held key (KMS or cosign) or GitHub's private instance. Never write NDA platform names to a public transparency log.

---

## 6. Staleness

**How "known-good" decays [Inference, with verified mechanisms].**
- **Upstream releases** of the library, its partners, or transitive dependencies. ABI-hash changes catch these [15].
- **Toolchain and image churn.** Runner images roll and keep at most 2 GA images [6]. vcpkg pins and bumps its CI images within days, for example the 2026-10-19 Linux image and the 2026.10.27 Windows image [19]. New labels such as `windows-2025-vs2026` arrive [5].
- **Evidence loss.** Public-repository run records are deleted after 90 days or less [9].
- **Negative claims go stale too.** vcpkg re-runs `fail` entries on schedule and treats an unexpected pass as an error [18]. That is a model for automatically retiring "known-bad" entries.

**Detection mechanisms seen in practice [Verified].**
- Event-triggered rebuilds: vcpkg PR CI [20], nixpkgs rebuild counts [33].
- A cadence: crater runs each beta cycle [31]; Hydra builds "regularly" [33].
- A fixed window of reference versions: Godot [38].
- Bots that bump dependencies:
  - Dependabot updates the vcpkg `builtin-baseline`. Microsoft's docs warn that baseline bumps can bring breaking changes from transitive dependencies and say to "always test your build" [76]. GitHub announced support on 2025-08-12 (search summary) [77].
  - Renovate's manager registry, checked today, includes `conan` (`conanfile.txt`, `conanfile.py`, `conan.lock`, versions via the Conan v2 API) and `git-submodules`, but no `vcpkg`, `cmake`, `cpm`, `meson` or `xmake` [74][75].

**Proposed states [Inference].**
- `verified`: all inputs unchanged since the run, and within the time-to-live, for example 30 days for L2 claims.
- `stale`: an input changed or the time-to-live passed. Shown with a warning, and queued for re-run.
- `expired`: more than 90 days, or the evidence is unreachable.
- `broken`: the last run failed. Kept as a negative claim and re-checked on schedule.
- Agents should treat `stale` as "probably fine; re-verify before relying on it".

---

## 7. Agent-in-the-loop verification

**Practice in the field [Verified].**
- **vcpkg agent skills.** Since April 2026, the vcpkg repository has shipped agent skills under `.github/skills`: `analyze-ci-failures`, `fetch-vcpkg-ci-logs`, `review-vcpkg-pr`, `fix-reviewed-vcpkg-pr` and `evaluate-new-port`.
  - `analyze-ci-failures` scans for `REGRESSION:` lines and their exact failure keywords (`BUILD_FAILED`, `POST_BUILD_CHECKS_FAILED`, `CASCADED_DUE_TO_MISSING_DEPENDENCIES` and others).
  - It checks failures against both baselines, separates regressions from known failures, and separates proven root causes from hypotheses.
  - `fix-reviewed-vcpkg-pr` validates its fixes with local Release and Debug builds before pushing [83].
- **Anthropic's C compiler.** 16 parallel agents across about 2,000 sessions, for under $20k of API cost, produced a 100k-line compiler that builds Linux 6.9. Lessons reported [78]:
  - verifier quality is decisive;
  - continuous integration was added to stop regressions;
  - test output should be short and grep-able, with details logged to files;
  - deterministic per-agent sampling makes a fast mode practical;
  - a known-good oracle (GCC) splits a big failure into bugs that agents can fix in parallel.
- **Claude Code hooks.** A `PostToolUse` hook that exits with code 2 shows its stderr to the agent. A `Stop` hook can block the agent from finishing [79]. Running build and tests after edits is the standard way to wire this up.
- **GitHub's cloud agent** can "run tests, and revise its changes in response to test failures" [80].
- **Benchmarks.** SWE-bench, and Multi-SWE-bench (1,632 instances, including C and C++), judge resolution by building in Docker and running the project's tests [81][82].
- **[Unverified]** I found no published success rates for LLM repair of C++ compile errors specifically. arXiv was blocked and the search budget was exhausted.

**[Inference] Implication.** The same ladder that backs a "known-good" claim is the agent's reward signal. Design the output for agents:
- one-line verdicts per rung;
- a stable error taxonomy, mirroring vcpkg's keywords, for example `DEFINE_MISMATCH(JPH_DOUBLE_PRECISION)`, `INIT_ORDER`, `GOLDEN_DIFF(mean=0.07>0.05)`;
- a fast lane that runs a sample.

---

## Implications for Automap (inference)

**Recommended verification ladder** (cheapest evidence first; each tier is recorded in the attestation):

| Tier | Evidence | Where | Approx. cost |
|---|---|---|---|
| L0 Declared | Manifest and contract declared; no run | — | 0 |
| L1 Builds | Configure and build per platform; ABI hash recorded | hosted | 3–6 min/library/platform |
| L2 Links + smoke | `test_package`-style consumer; N headless frames with dummy/offscreen video and dummy audio; clean shutdown | hosted | +1–3 min |
| L3 Contract probes | Define/ABI probes (`VerifyJoltVersionID`, `detect_mismatch`, generated `static_assert`s); allocator/jobs/ownership glue tests; TSan and ASan/UBSan on Linux | hosted Linux, plus MSVC ASan | +20 min |
| L4 Behavioural | Physics determinism hashes; serialization round-trip and fuzzing; software-GPU golden images (lavapipe, WARP, SwiftShader) with FLIP thresholds | hosted | +10–15 min; baseline upkeep |
| L5 Hardware | Real-GPU goldens (T4 runners); Apple GPU; console devkits | GPU larger runners or self-hosted; NDA tier private | $0.05–0.10/min, or own hardware |

**Pruning strategy.**
1. Verify only declared slot-boundary edges, using a window of latest and previous versions.
2. Re-verify when the ABI hash of an input changes, plus weekly re-runs of reference compositions to catch toolchain drift.
3. Run a pairwise covering array (PICT) over slot × library × version as a monthly sweep, to catch interactions that were never declared.
4. Keep a vcpkg-style expected-failure baseline. It is re-checked automatically, and unexpected passes retire entries.
5. Run sanitizers and goldens on one platform each (Linux and lavapipe respectively), and run every other platform at L1–L2.
6. Never claim more than the highest tier passed. Show tier, date and validity in every agent-facing answer.

**Evidence.** Export signed in-toto/VSA-style records to Automap-owned storage on every run, because Actions retention will delete the originals [9].

**Starting hypothesis.**
- Keep "known-good", but make it *graded* (L1–L5), *time-boxed*, and tied to a *specific lock and platform*.
- Contracts must ship executable probes, not just prose.

---

## Open questions

1. How should the composition lock be canonicalised, so that identical compositions from vcpkg, Conan and CPM share one subject digest?
2. Can contract probes be generated from manifest metadata (lists of defines) for libraries that have no `VerifyJoltVersionID`-style hook?
3. How should golden baselines be governed? Who approves new goldens, and should each library or each composition own them (Filament uses commit-message acceptance [52])?
4. Should community-submitted verification runs be accepted? Only with attestation by a trusted reusable workflow (Build Level 3 [67]), or never?
5. How should verification work for console platforms under NDA: self-hosted devkits, partner-run verifiers that sign VSAs, or claims that cannot be checked?
6. Is the hosted macOS paravirtual Metal device stable enough to use as a gate, or only as an advisory signal [35][55]?
7. Will GitHub reinstate a self-hosted platform fee? Today's docs show none; secondary sources say the fee was postponed [3][10].
8. What should the time-to-live be for each tier? This needs data on how often pairings actually break between upstream releases.

---

## Sources

All accessed 2026-10-29. "src" means the documentation's source repository, read on `main`/`master` as of that date.

1. GitHub Docs (src), "Actions runner pricing" — https://github.com/github/docs/blob/main/content/billing/reference/actions-runner-pricing.md
2. GitHub Docs (src), standard runner prices — https://github.com/github/docs/blob/main/data/reusables/billing/actions-standard-runner-prices.md
3. GitHub Docs (src), "GitHub Actions" product billing — https://github.com/github/docs/blob/main/content/billing/concepts/product-billing/github-actions.md
4. GitHub Docs (src), included quotas — https://github.com/github/docs/blob/main/data/reusables/billing/actions-included-quotas.md
5. GitHub Docs (src), supported GitHub-hosted runners — https://github.com/github/docs/blob/main/data/reusables/actions/supported-github-runners.md
6. actions/runner-images README — https://github.com/actions/runner-images
7. GitHub Docs (src), larger runners reference — https://github.com/github/docs/blob/main/content/actions/reference/runners/larger-runners.md
8. GitHub Docs (src), Actions limits — https://github.com/github/docs/blob/main/content/actions/reference/limits.md
9. GitHub Docs (src), artifact and log retention — https://github.com/github/docs/blob/main/data/reusables/actions/about-artifact-log-retention.md
10. Secondary (search summaries only): GitHub, "Pricing changes for GitHub Actions" — https://resources.github.com/actions/2026-pricing-changes-for-github-actions/ (fetch blocked); Socket.dev, "GitHub Actions Pricing Whiplash" — https://socket.dev/blog/github-actions-pricing-whiplash
11. PostgreSQL commit 9c12606, "ci: Add GitHub Actions based CI" (2026-06-04) — https://github.com/postgres/postgres/commit/9c12606
12. SDL Actions run usage, run 40775940397 (2026-10-28) — https://github.com/libsdl-org/SDL/actions/runs/40775940397/usage
13. JoltPhysics Actions build runs and usage, run 40411254313 (2026-10-26) — https://github.com/jrouwe/JoltPhysics/actions/runs/40411254313/usage
14. Dear ImGui Actions run 40847969306 (2026-10-29) — https://github.com/ocornut/imgui/actions/runs/40847969306
15. vcpkg docs (src), binary caching reference — https://github.com/MicrosoftDocs/vcpkg-docs/blob/main/vcpkg/reference/binarycaching.md
16. vcpkg docs (src), GitHub Actions cache tutorial (removed; vcpkg-tool PR #1662) — https://github.com/MicrosoftDocs/vcpkg-docs/blob/main/vcpkg/consume/binary-caching-github-actions-cache.md
17. sccache, GHA backend — https://github.com/mozilla/sccache/blob/main/docs/GHA.md
18. vcpkg `scripts/ci.baseline.txt` — https://github.com/microsoft/vcpkg/blob/master/scripts/ci.baseline.txt
19. vcpkg Azure Pipelines (top-level and windows) — https://github.com/microsoft/vcpkg/tree/master/scripts/azure-pipelines
20. vcpkg `test-modified-ports.ps1` — https://github.com/microsoft/vcpkg/blob/master/scripts/azure-pipelines/test-modified-ports.ps1
21. vcpkg docs (src), `x-test-features` — https://github.com/MicrosoftDocs/vcpkg-docs/blob/main/vcpkg/commands/test-features.md
22. vcpkg `scripts/ci.feature.baseline.txt` — https://github.com/microsoft/vcpkg/blob/master/scripts/ci.feature.baseline.txt
23. vcpkg releases (2026-10-16 release, port counts) — https://github.com/microsoft/vcpkg/releases
24. Microsoft C++ Team Blog, "What's new in vcpkg (February 2025)" — https://devblogs.microsoft.com/cppblog/whats-new-in-vcpkg-february-2025-package-installation-performance-new-tested-triplet-and-more/ (fetch blocked; figures from search summary)
25. conan-center-index, supported platforms and configurations (undated) — https://github.com/conan-io/conan-center-index/blob/master/docs/supported_platforms_and_configurations.md
26. conan-center-index, adding packages — https://github.com/conan-io/conan-center-index/blob/master/docs/adding_packages/README.md
27. Conan 2 docs (src), testing packages — https://github.com/conan-io/docs/blob/develop2/tutorial/creating_packages/test_conan_packages.rst
28. conan-center-index README — https://github.com/conan-io/conan-center-index
29. rustc-dev-guide, Crater — https://github.com/rust-lang/rustc-dev-guide/blob/master/src/tests/crater.md
30. crater bot usage — https://github.com/rust-lang/crater/blob/master/docs/bot-usage.md
31. Rust Forge, release/crater — https://github.com/rust-lang/rust-forge/blob/master/src/release/crater.md
32. crater README — https://github.com/rust-lang/crater
33. nixpkgs CONTRIBUTING.md — https://github.com/NixOS/nixpkgs/blob/master/CONTRIBUTING.md
34. Bevy, plugin development / plugins guidelines — https://bevy.org/learn/quick-start/plugin-development/ (search summary)
35. Bevy `example-run.yml` — https://github.com/bevyengine/bevy/blob/main/.github/workflows/example-run.yml
36. Bevy `send-screenshots-to-pixeleagle.yml` — https://github.com/bevyengine/bevy/blob/main/.github/workflows/send-screenshots-to-pixeleagle.yml
37. Godot `linux_builds.yml` — https://github.com/godotengine/godot/blob/master/.github/workflows/linux_builds.yml
38. Godot `godot-compat-test` action — https://github.com/godotengine/godot/blob/master/.github/actions/godot-compat-test/action.yml
39. SDL `test/CMakeLists.txt` — https://github.com/libsdl-org/SDL/blob/main/test/CMakeLists.txt
40. SDL `src/video/offscreen` — https://github.com/libsdl-org/SDL/tree/main/src/video/offscreen
41. SDL dummy audio driver — https://github.com/libsdl-org/SDL/blob/main/src/audio/dummy/SDL_dummyaudio.c
42. wgpu CI workflow — https://github.com/gfx-rs/wgpu/blob/trunk/.github/workflows/ci.yml
43. wgpu `install-mesa` action — https://github.com/gfx-rs/wgpu/blob/trunk/.github/actions/install-mesa/action.yml
44. wgpu `docs/testing.md` — https://github.com/gfx-rs/wgpu/blob/trunk/docs/testing.md
45. wgpu `tests/src/image.rs` — https://github.com/gfx-rs/wgpu/blob/trunk/tests/src/image.rs
46. NVlabs/flip README (v1.7) — https://github.com/NVlabs/flip
47. Andersson et al., "ꟻLIP: A Difference Evaluator for Alternating Images", ACM PACMCGIT/HPG 2020 — https://research.nvidia.com/publication/flip (search summary)
48. SwiftShader README — https://github.com/google/swiftshader
49. Chromium, SwiftShader doc and `web_tests/FlagSpecificConfig` — https://chromium.googlesource.com/chromium/src/+/HEAD/third_party/blink/web_tests/FlagSpecificConfig (search summary)
50. Chrome for Developers, "Supercharge Web AI model testing: WebGPU, WebGL, and Headless Chrome" — https://developer.chrome.com/blog/supercharge-web-ai-testing (search summary)
51. Chromium, "GPU pixel testing with Gold" — https://chromium.googlesource.com/chromium/src/+/HEAD/docs/gpu/gpu_pixel_testing_with_gold.md (search summary)
52. Filament `test/renderdiff/README.md` — https://github.com/google/filament/blob/main/test/renderdiff/README.md
53. bgfx `main.yml` — https://github.com/bkaradzic/bgfx/blob/master/.github/workflows/main.yml
54. sokol `tests/CMakeLists.txt` — https://github.com/floooh/sokol/blob/master/tests/CMakeLists.txt
55. PyTorch PR #111576, "[MPS] Skip virtualized devices" (merged 2023-10-19) — https://github.com/pytorch/pytorch/pull/111576
56. Jolt `determinism_check.yml` — https://github.com/jrouwe/JoltPhysics/blob/master/.github/workflows/determinism_check.yml
57. Jolt `Jolt/Core/Core.h` (`JPH_VERSION_ID`) — https://github.com/jrouwe/JoltPhysics/blob/master/Jolt/Core/Core.h
58. Jolt `Jolt/RegisterTypes.h` — https://github.com/jrouwe/JoltPhysics/blob/master/Jolt/RegisterTypes.h
59. Jolt `Build/README.md` and `Docs/Architecture.md` — https://github.com/jrouwe/JoltPhysics/blob/master/Build/README.md
60. Jolt PR #2187 (closed unmerged, 2026-10-28; references issue #2165) — https://github.com/jrouwe/JoltPhysics/pull/2187
61. MSVC docs (src), AddressSanitizer — https://github.com/MicrosoftDocs/cpp-docs/blob/main/docs/sanitizers/asan.md
62. MSVC docs (src), LNK2038 — https://github.com/MicrosoftDocs/cpp-docs/blob/main/docs/error-messages/tool-errors/linker-tools-error-lnk2038.md
63. Microsoft PICT README — https://github.com/microsoft/pict
64. in-toto test-result predicate v0.1 — https://github.com/in-toto/attestation/blob/main/spec/predicates/test-result.md
65. in-toto predicates index — https://github.com/in-toto/attestation/tree/main/spec/predicates
66. SLSA Verification Summary (draft on `main`; v1.2 is the latest release branch) — https://github.com/slsa-framework/slsa/blob/main/spec/verification_summary.md
67. GitHub Docs (src), artifact attestations concept — https://github.com/github/docs/blob/main/content/actions/concepts/security/artifact-attestations.md
68. GitHub Docs (src), using artifact attestations — https://github.com/github/docs/blob/main/content/actions/how-tos/secure-your-work/use-artifact-attestations/use-artifact-attestations.md
69. GitHub Docs (src), attestations availability — https://github.com/github/docs/blob/main/data/reusables/gated-features/attestations.md
70. actions/attest README — https://github.com/actions/attest
71. actions/attest-build-provenance README (v4 wrapper notice) — https://github.com/actions/attest-build-provenance
72. vcpkg docs (src), SBOM — https://github.com/MicrosoftDocs/vcpkg-docs/blob/main/vcpkg/reference/software-bill-of-materials.md
73. Conan 2 docs (src), `conan.tools.sbom` — https://github.com/conan-io/docs/blob/develop2/reference/tools/sbom.rst
74. Renovate manager registry `lib/modules/manager/api.ts` — https://github.com/renovatebot/renovate/blob/main/lib/modules/manager/api.ts
75. Renovate Conan manager readme — https://github.com/renovatebot/renovate/blob/main/lib/modules/manager/conan/readme.md
76. vcpkg docs (src), Dependabot — https://github.com/MicrosoftDocs/vcpkg-docs/blob/main/vcpkg/concepts/dependabot.md
77. GitHub Changelog, "Dependabot version updates now support vcpkg" (2025-08-12) — https://github.blog/changelog/2025-08-12-dependabot-version-updates-now-support-vcpkg/ (search summary)
78. Anthropic Engineering, "Building a C compiler with a team of parallel Claudes" (N. Carlini, 2026-02-05) — https://www.anthropic.com/engineering/building-c-compiler
79. Claude Code docs, Hooks reference — https://code.claude.com/docs/en/hooks
80. GitHub Docs (src), about GitHub Copilot agents — https://github.com/github/docs/blob/main/content/copilot/concepts/agent-capabilities/about-agents/about-github-copilot-agents.md
81. Multi-SWE-bench README — https://github.com/multi-swe-bench/multi-swe-bench
82. SWE-bench evaluation guide — https://github.com/SWE-bench/SWE-bench/blob/main/docs/guides/evaluation.md
83. vcpkg `.github/skills` (analyze-ci-failures, fix-reviewed-vcpkg-pr; first commit 2026-04-15, #51188) — https://github.com/microsoft/vcpkg/tree/master/.github/skills
84. Dawn, end2end tests and SwiftShader commits — https://dawn.googlesource.com/dawn/ (search summary of historical commits)
85. GitHub Changelog, "GitHub Actions GPU hosted runners are now generally available" (2024-07-08) — https://github.blog/changelog/2024-07-08-github-actions-gpu-hosted-runners-are-now-generally-available (search summary)
