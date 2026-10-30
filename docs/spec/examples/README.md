# Spec examples

Complete, valid examples of every schema in [`../schemas/`](../schemas/). `python3 docs/spec/check-examples.py` validates each file against its schema, checks library facts and manifest choices against their contracts, checks references between files, and recomputes each lock's canonical form and `lock_hash`.

| Directory | Contents |
| --- | --- |
| `map/` | A slice of the public map: six slots, seven contracts, SDL3, Dear ImGui and Jolt, four recipes (two bindings, one pairing, one library recipe) and one published evidence record. |
| `private-map/` | A hypothetical studio's private map, with the same schemas as the public map: an in-house scheduler (`acme-jobs`) and renderer (`acme-renderer`), a patched SDL3 variant, and their recipes. |
| `private-map-nebula/` | The same studio's private map for a hypothetical NDA console, Nebula: SDL's console variant, Acme's Nebula renderer variant, Nebula added to Jolt's and `acme-jobs`'s targets, and composition evidence signed by the studio. |
| `composition/testbed-jolt/` | The day-one composition after extending with Jolt: manifest, lock and signed composition evidence for linux-x64-clang. |
| `composition/acme-engine/` | A hypothetical studio's engine using the private map, with its in-house scheduler and renderer, the patched SDL3, and per-platform variants for Nebula. |
| `composition/cli-results/` | CLI JSON output for a successful extend and a failed verify. |

Library versions, commits, release dates, vcpkg ports and source hashes are real (SDL 3.4.18, Dear ImGui 1.93.0 docking, Jolt 5.6.0, vcpkg baseline `d186940`). Everything named Acme or Nebula, and the hashes in their files, are hypothetical. Maps layer in directory order: `map/`, then `private-map/`, then `private-map-nebula/`. Reference glue sources are named but not included; the spike's working glue is archived on the [`spike/testbed-agent`](https://github.com/zach/automap/tree/spike/testbed-agent/spikes/testbed-agent/workspace/glue) branch.
