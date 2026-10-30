#!/usr/bin/env python3
"""Validate every example under docs/spec/examples against docs/spec/schemas.

Beyond schema validation it checks what the schemas can't express:
library facts against their contract's fact schemas, manifest choices
against contract choice schemas, references between files, lock
canonical form and lock_hash, and evidence subjects.

Usage: python3 docs/spec/check-examples.py   (needs: pip install jsonschema)
"""
import hashlib
import json
import pathlib
import sys

from jsonschema import Draft202012Validator, FormatChecker
from referencing import Registry, Resource

ROOT = pathlib.Path(__file__).resolve().parent
SCHEMAS = ROOT / "schemas"
EX = ROOT / "examples"
errors = []
checked = 0


def load(path):
    return json.loads(path.read_text())


def fail(path, msg):
    errors.append(f"{path.relative_to(ROOT)}: {msg}")


def canon(obj):
    # RFC 8785 for documents without floats: sorted keys, no whitespace, UTF-8.
    return json.dumps(obj, sort_keys=True, separators=(",", ":"), ensure_ascii=False)


schemas = {p.name.removesuffix(".schema.json"): load(p) for p in SCHEMAS.glob("*.schema.json")}
resources = []
for name, s in schemas.items():
    r = Resource.from_contents(s)
    resources += [(s["$id"], r), (f"{name}.schema.json", r)]
registry = Registry().with_resources(resources)


def validate(path, schema_name, doc=None):
    global checked
    doc = load(path) if doc is None else doc
    v = Draft202012Validator(schemas[schema_name], registry=registry, format_checker=FormatChecker())
    for e in sorted(v.iter_errors(doc), key=lambda e: list(e.path)):
        fail(path, f"{schema_name}: {'/'.join(map(str, e.path)) or '<root>'}: {e.message}")
    checked += 1
    return doc


def object_schema(fields, label):
    """A JSON Schema for an object whose properties are a contract's facts or choices."""
    return {
        "type": "object",
        "additionalProperties": False,
        "properties": {k: f["schema"] for k, f in fields.items()},
        "required": [k for k, f in fields.items() if label == "facts" and f.get("required", True)],
    }


def check_against(path, where, value, schema):
    for e in Draft202012Validator(schema).iter_errors(value):
        fail(path, f"{where}: {'/'.join(map(str, e.path)) or '<root>'}: {e.message}")


# ---- schemas themselves ----
for name, s in schemas.items():
    Draft202012Validator.check_schema(s)

# ---- slots ----
slots = {}
for p in sorted(EX.glob("*/slots/*.json")):
    d = validate(p, "slot")
    slots[d["id"]] = d
    if p.stem != d["id"]:
        fail(p, "file name must be the slot id")

# ---- contracts ----
contracts = {}
for p in sorted(EX.glob("*/contracts/*/*.json")):
    d = validate(p, "contract")
    contracts[d["contract"]] = d
    if p.parent.name != d["contract"] or p.stem != d["version"]:
        fail(p, "path must be contracts/<contract>/<version>.json")
    for cid, c in d["choices"].items():
        check_against(p, f"choices/{cid}/default", c["default"], c["schema"])
    if d["contract"] == "lifecycle" and "phases" not in d:
        fail(p, "the lifecycle contract lists its phases")
    for r in d["rules"]:
        if not r["id"].startswith(d["contract"] + "."):
            fail(p, f"rule {r['id']} must be prefixed with the contract id")
contract_rules = {r["id"]: r for c in contracts.values() for r in c["rules"]}

# ---- libraries (public and private maps share one schema) ----
libraries = {}
for p in sorted(EX.glob("*/libraries/*.json")):
    d = validate(p, "library")
    libraries[d["id"]] = (p, d)
    if p.stem != d["id"]:
        fail(p, "file name must be the library id")
    for s in d["slots"]:
        if s not in slots:
            fail(p, f"slot {s} has no example slot file")
    for v in d["versions"]:
        steps = set()
        for seam, facts in v["facts"].items():
            if seam in contracts:
                check_against(p, f"{v['version']}/facts/{seam}", facts, object_schema(contracts[seam]["facts"], "facts"))
        lc = v["facts"].get("lifecycle", {})
        for kind in ("init_steps", "shutdown_steps"):
            names = {s["step"] for s in lc.get(kind, [])}
            steps |= names
            for s in lc.get(kind, []):
                if not s["step"].startswith(d["id"] + ":"):
                    fail(p, f"step {s['step']} must be prefixed with '{d['id']}:'")
                for a in s.get("after", []):
                    if a not in names and not a.startswith("automap:"):
                        fail(p, f"step {s['step']} follows unknown step {a}")
        for var in v["variants"]:
            for seam, over in var.get("facts", {}).items():
                merged = {**v["facts"].get(seam, {}), **over}
                if seam in contracts:
                    check_against(p, f"{v['version']}/{var['name']}/facts/{seam}", merged, object_schema(contracts[seam]["facts"], "facts"))
    if d.get("source", {}).get("kind") == "path" and d.get("console_variants"):
        fail(p, "console_variants belong to the public map only")


def library_version(lib, version, variant=None):
    if lib not in libraries:
        return None
    for v in libraries[lib][1]["versions"]:
        if v["version"] == version and (variant is None or variant in [x["name"] for x in v["variants"]]):
            return v
    return None


# ---- recipes ----
recipes = {}
for p in sorted(EX.glob("*/recipes/*/*.json")):
    d = validate(p, "recipe")
    if p.parent.name != d["kind"]:
        fail(p, f"a {d['kind']} recipe lives in recipes/{d['kind']}/")
    subj = d["subject"]
    if d["kind"] == "binding":
        rid = f"{subj['library']}--{subj['contract']}"
        if subj["contract"] in contracts and contracts[subj["contract"]]["version"] != subj["contract_version"]:
            fail(p, "contract_version doesn't match the example contract")
    elif d["kind"] == "pairing":
        if subj["libraries"] != sorted(subj["libraries"]):
            fail(p, "pairing libraries are listed in sorted order")
        rid = "--".join(subj["libraries"])
    else:
        rid = subj["library"]
    recipes[rid] = d
    if p.stem != rid:
        fail(p, f"file name must be {rid}")
    own = {r["id"]: r for r in d.get("rules", [])}
    for r in own:
        if not r.startswith(rid + "."):
            fail(p, f"pairing rule {r} must be prefixed with '{rid}.'")
    for lib, a in d["applies_to"].items():
        for ver in a["versions"]:
            for var in a["variants"]:
                if not library_version(lib, ver, var):
                    fail(p, f"applies_to names {lib} {ver} ({var}), which no library example has")
    for pr in d["probes"]:
        rule = own.get(pr["rule"]) or contract_rules.get(pr["rule"])
        if rule is None:
            fail(p, f"probe for unknown rule {pr['rule']}")
        elif rule["kind"] != "probed":
            fail(p, f"probe for guidance rule {pr['rule']}")
    for na in d.get("not_applicable", []):
        if na["rule"] not in contract_rules:
            fail(p, f"not_applicable names unknown rule {na['rule']}")
    if d["kind"] == "library" and not any(g["role"] == "build-fragment" for g in d["reference_glue"]):
        fail(p, "a library recipe ships a build fragment")
    if d["kind"] != "library" and not any(g["role"] == "probe" for g in d["reference_glue"]):
        fail(p, "binding and pairing recipes ship their probes")
    owners = set(subj.get("libraries", [])) | {rid}
    for o in d.get("ordering", []):
        for step in (o["before"], o["after"]):
            if step.split(":", 1)[0] not in owners:
                fail(p, f"ordering step {step} isn't owned by {sorted(owners)}")

# ---- compositions ----
for m_path in sorted(EX.glob("composition/*/automap.manifest.json")):
    m = validate(m_path, "manifest")
    comp = m_path.parent
    for slot, lib in m["slots"].items():
        if lib not in libraries:
            fail(m_path, f"slot {slot}: unknown library {lib}")
            continue
        if slot not in libraries[lib][1]["slots"]:
            fail(m_path, f"{lib} doesn't fill slot {slot}")
    used = set(m["slots"].values())
    if set(m["variants"]) != used:
        fail(m_path, "variants must name exactly the libraries in slots")
    for seam, chosen in m["choices"].items():
        if seam in contracts:
            check_against(m_path, f"choices/{seam}", chosen, object_schema(contracts[seam]["choices"], "choices"))
    if m["choices"].get("jobs", {}).get("scheduler") == "jobs-slot" and "jobs" not in m["slots"]:
        fail(m_path, "scheduler jobs-slot needs a library in the jobs slot")
    for key in list(m["glue"]["bindings"]) + list(m["glue"]["pairings"]):
        for part in key.split("/"):
            if part not in used and part not in contracts:
                fail(m_path, f"glue entry {key} names {part}, which isn't in this composition")
    private = {pm["id"] for pm in m.get("private_maps", [])}

    lock_path = comp / "automap.lock.json"
    raw = lock_path.read_text()
    lock = validate(lock_path, "lock", json.loads(raw))
    if raw != canon(lock):
        fail(lock_path, "lock is not in canonical form")
    body = {k: v for k, v in lock.items() if k != "lock_hash"}
    if lock["lock_hash"] != "sha256:" + hashlib.sha256(canon(body).encode()).hexdigest():
        fail(lock_path, "lock_hash doesn't match the canonical lock")
    if set(lock["pins"]) != used:
        fail(lock_path, "pins must name exactly the libraries in the manifest's slots")
    if lock["map_release"] != m["map_release"]:
        fail(lock_path, "map_release differs from the manifest")
    if set(lock.get("private_maps", {})) != private:
        fail(lock_path, "private_maps must match the manifest's")
    for lib, pin in lock["pins"].items():
        v = library_version(lib, pin["version"], pin["variant"])
        if v is None:
            fail(lock_path, f"{lib} {pin['version']} ({pin['variant']}) is not in the map")
            continue
        if pin["variant"] != m["variants"].get(lib):
            fail(lock_path, f"{lib}: pinned variant differs from the manifest")
        if "pin" in v and pin.get("commit") != v["pin"]["commit"]:
            fail(lock_path, f"{lib}: pinned commit differs from the map")
        if pin["registry"]["kind"] == "private" and pin["registry"].get("id") not in private:
            fail(lock_path, f"{lib}: private registry {pin['registry'].get('id')} is not a private map of this composition")
        for plat in m["platforms"]:
            if plat not in v["facts"].get("platform", {}).get("targets", []):
                fail(m_path, f"{lib} doesn't list {plat} among its targets")
    for c, ver in lock["contracts"].items():
        if c in contracts and contracts[c]["version"] != ver:
            fail(lock_path, f"contract {c} {ver} differs from the example contract")

    for e_path in sorted(comp.glob("evidence/*.json")):
        e = validate(e_path, "evidence")
        pred = e["predicate"]
        if e_path.stem != pred["platform"]:
            fail(e_path, "file name must be the platform id")
        if pred["subject_kind"] == "composition":
            if pred["subject_detail"].get("lock_hash") != lock["lock_hash"]:
                fail(e_path, "subject_detail.lock_hash doesn't match the lock")
            if e["subject"][0]["digest"].get("sha256") != hashlib.sha256(raw.encode()).hexdigest():
                fail(e_path, "subject digest doesn't match the lock file")
        for r in pred["results"]:
            rule = contract_rules.get(r["rule"]) or next(
                (x for rc in recipes.values() for x in rc.get("rules", []) if x["id"] == r["rule"]), None)
            if rule is None:
                fail(e_path, f"result for unknown rule {r['rule']}")
            elif rule["kind"] != "probed":
                fail(e_path, f"guidance rule {r['rule']} can't appear in evidence")
            if r.get("library") and r["library"] not in used:
                fail(e_path, f"result names {r['library']}, which isn't in the composition")
        if pred["level"] in ("probed", "behavioral", "hardware"):
            bad = [r["rule"] for r in pred["results"] if r["status"] in ("fail", "unverifiable")]
            if bad:
                fail(e_path, f"level {pred['level']} with failing or unverifiable results: {bad}")

for p in sorted(EX.glob("composition/cli-results/*.json")):
    d = validate(p, "cli-result")
    if d["ok"] == bool(d["failures"]):
        fail(p, "ok must be true exactly when there are no failures")

print(f"{checked} files checked against {len(schemas)} schemas")
if errors:
    print("\n".join(errors))
    sys.exit(1)
print("all examples valid")
