# What the CLI would do (spike stand-in)

No CLI exists yet. When the router skill says "run `automap <job>`", perform these steps by hand, exactly and in order. Every step is deterministic: if a step can't be completed as written, stop and record why in `RESULTS.md` rather than improvising silently.

All paths below are relative to the composition's repo (`workspace/`), except `map/`, which means `spikes/testbed-agent/map/` (read-only).

## `automap start <slot>=<library> ...`

1. **Validate the request against the map.**
   - Each slot exists in `map/slots.json`; each library lists that slot in `map/libraries/<id>.json`; no slot gets two libraries.
   - **platform.targets-supported:** every library's `facts.platform.targets` includes every platform in the manifest. Record the result for the evidence.
   - **License warning:** if any library's license is LGPL, GPL or proprietary, print a warning naming it.
2. **Write the manifest** `automap.manifest.json` (format in `map/FORMAT.md`): slots, variants, choices (each contract's defaults unless the request says otherwise), and a `glue` entry for every binding and pairing.
   - The composition's bindings are every (library, contract) entry in `map/recipes/bindings/<library>/recipe.json`.
   - Its pairings are every `map/recipes/pairings/<a>--<b>/` whose two libraries are both in the composition.
3. **Write the lock** `automap.lock.json`: one pin per library from its `versions[0].pin` and chosen variant; the contract versions used; `map_release`; and `lock_hash`, the SHA-256 of the lock in RFC 8785 canonical JSON with `lock_hash` removed (for the spike, `python3 -c` with `json.dumps(obj, sort_keys=True, separators=(",", ":"))` is an acceptable stand-in for RFC 8785).
4. **Write glue.**
   - Copy each contract's `shared_interface.files` from `map/interfaces/automap/` to `glue/automap/`.
   - Copy each binding's and pairing's `reference_glue` into `glue/`.
   - Copy the testbed composition's reference `main.cpp` and `CMakeLists.txt` from `map/recipes/compositions/testbed/` when starting a testbed.
   - Apply every `adaptation_notes` entry that applies.
5. **Write agent files.** Create or update `AGENTS.md` with one Automap section between `<!-- automap:begin -->` and `<!-- automap:end -->`, holding only: the build, test and probe commands; where the manifest and lock live; the slots in use; and "go through the router skill and CLI before changing any glue". Copy `skill/SKILL.md` to `.claude/skills/automap/SKILL.md`.
6. **Verify** (see below).

## `automap extend <slot>=<library>`

1. Validate as in start, step 1, for the new library against the existing manifest.
2. Update the manifest: add the slot, its variant, any new contract choices, and `glue` entries for the new library's bindings and for any new pairings.
3. Rewrite the lock (step 3 above) for the whole composition.
4. Write glue for the new bindings and pairings only, and copy any shared interfaces not already present. Wire the new library's entry points into the composition's entry point in the phases its lifecycle facts and recipe name, and call its probe registration in probe mode. Existing glue that already does a binding's job is adapted, not replaced.
5. Update the `AGENTS.md` section (slots in use).
6. Verify.

## Verify (the evidence ladder)

Run each level in order; the composition's evidence level is the highest level that passed.

1. **builds:** configure and build with CMake and Ninja (`-DAUTOMAP_HEADLESS=ON` on Linux in this sandbox).
2. **smoke:** run the composition's entry point with `--probe` under `VK_DRIVER_FILES=/usr/share/vulkan/icd.d/lvp_icd.json`; it must run its frames, shut down and exit.
3. **probed:** the `--probe` run prints `"verdict":"pass"`; the static `platform.targets-supported` check passed; and, if the jobs contract is in use, a ThreadSanitizer build (`-fsanitize=thread`) of the same composition completes a `--probe` run with no ThreadSanitizer reports (`jobs.tsan-clean`).
4. **Write evidence** to `evidence/linux-x64-<compiler>.json` (format in `map/FORMAT.md`), unsigned, with one result per probed rule, and print a one-line verdict such as `automap: probed (12/12 rules)` or `automap: smoke; failed: ALLOC_NOT_ROUTED(jolt)`.

## The four nevers

Never edit the lock by hand except by following step 3. Never write evidence except by running Verify. Never weaken, skip or delete a probe to get a pass. Never open issues or pull requests upstream.
