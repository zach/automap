# Draft map, manifest, lock and evidence formats (spike)

Vocabulary follows Automap's glossary. A **slot** is a role (windowing, physics-3d). A **library** fills slots. Automap owns one **contract** per **seam**; each contract has **facts** (what a library declares), **choices** (what a composition decides), **rules** (each either *probed* or *guidance*) and optionally a **shared interface** delivered as glue source. A **binding** connects one library to one contract; a **pairing** is a direct integration between exactly two libraries. A **recipe** (for a binding or pairing) is **reference glue** plus **probes** plus **adaptation notes**. **Glue** is code in the developer's repo; the developer owns it. Success is judged by probes passing, never by resemblance to the reference glue.

## Map files

- `slots.json`: slot ids and descriptions.
- `contracts/<seam>.json`: `contract`, `version`, `facts` (what libraries must declare), `choices` (with defaults), `rules[]` (`id`, `kind`: `probed` | `guidance`, `text`, `failure_code` for probed rules), `shared_interface` (file list under `interfaces/`, or null).
- `interfaces/automap/*`: shared-interface glue source, copied verbatim into the composition's glue directory. Namespaced `am::`.
- `libraries/<id>.json`: identity, `license`, `contribution_policy`, `slots`, `versions[]` each with `pin` (git URL, tag, commit), `vcpkg` (port, version, features), `cmake` (package and imported targets, plus how to build it from source in this sandbox), `variants[]`, and `facts` keyed by contract.
- `recipes/bindings/<library>/recipe.json`: one entry per contract the library binds to: `contract`, `contract_version`, `reference_glue` files, `probes` (rule ids this binding makes executable, with the glue function that registers them), `adaptation_notes`.
- `recipes/pairings/<a>--<b>/recipe.json`: same shape for a pairing.
- `recipes/compositions/testbed/`: reference glue for the testbed's own entry point and top-level build file.

## Manifest (`automap.manifest.json`, edited by the agent)

```json
{
  "$comment": "Draft manifest format for the spike",
  "format": "automap-manifest/0.1-draft",
  "map_release": "spike-draft",
  "platforms": ["linux-x64-clang"],
  "slots": { "windowing": "sdl3", "input": "sdl3", "gpu": "sdl3", "tools-ui": "dear-imgui" },
  "variants": { "sdl3": "default", "dear-imgui": "docking" },
  "choices": {
    "allocation": { "allocator": "system" },
    "diagnostics": { "backend": "stderr" },
    "jobs": { "scheduler": "automap-thread-pool", "workers": 3 },
    "determinism": { "enabled": false }
  },
  "glue": {
    "interfaces": "glue/automap",
    "bindings": { "sdl3/allocation": "glue/sdl3_bindings.cpp" },
    "pairings": { "dear-imgui/sdl3": "glue/imgui_sdl3_gpu.cpp" }
  }
}
```

Every binding and pairing the composition uses appears under `glue` with the file where its glue lives, whether or not that file came from reference glue.

## Lock (`automap.lock.json`, written by the CLI, never by hand)

```json
{
  "format": "automap-lock/0.1-draft",
  "map_release": "spike-draft",
  "pins": {
    "sdl3": { "version": "3.4.18", "variant": "default", "registry": "git", "source": "https://github.com/libsdl-org/SDL", "commit": "<full sha>", "content_hash": "git-commit:<full sha>" }
  },
  "contracts": { "allocation": "0.1.0" },
  "lock_hash": "sha256 of this document in canonical JSON (RFC 8785) with lock_hash removed"
}
```

## Evidence (`evidence/<platform>.json`, written by the CLI)

```json
{
  "format": "automap-evidence/0.1-draft",
  "subject": { "kind": "composition", "lock_hash": "<sha256>" },
  "platform": "linux-x64-clang",
  "level": "probed",
  "signed": false,
  "produced_at": "2026-10-30T00:00:00Z",
  "results": [ { "rule": "allocation.hooks-before-first-allocation", "library": "sdl3", "ok": true, "failure_code": null } ]
}
```

Levels, lowest to highest: `builds`, `smoke`, `probed`, `behavioral`, `hardware`. `probed` requires every probed rule to pass, including a ThreadSanitizer run when the jobs contract is in use.
