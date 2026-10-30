# Spec examples

Complete, valid examples of every schema in [`../schemas/`](../schemas/). `python3 docs/spec/check-examples.py` validates each file against its schema, checks library facts and manifest choices against their contracts, checks references between files, and recomputes each lock's canonical form and `lock_hash`.

| Directory | Contents |
| --- | --- |
| `map/` | A slice of the public map: six slots, seven contracts, SDL3, Dear ImGui and Jolt, and four recipes (two bindings, one pairing, one library recipe). |
| `private-map/` | A private map holding a hypothetical in-house scheduler, `acme-jobs`, with the same schemas as the public map. |
| `composition/testbed-jolt/` | The day-one composition after extending with Jolt: manifest, lock and signed composition evidence for linux-x64-clang. |
| `composition/acme-engine/` | A hypothetical studio's engine using the private map, with `acme-jobs` in the jobs slot. |
| `composition/cli-results/` | CLI JSON output for a successful extend and a failed verify. |

Library versions, commits, release dates, vcpkg ports and source hashes are real (SDL 3.4.18, Dear ImGui 1.93.0 docking, Jolt 5.6.0, vcpkg baseline `d186940`). Everything named Acme, and the hashes in its lock, are hypothetical. Reference glue sources are named but not included; the spike's working glue is archived on the [`spike/testbed-agent`](https://github.com/zach/automap/tree/spike/testbed-agent/spikes/testbed-agent/workspace/glue) branch.
