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

# ---- map metadata ----
maps = {}
for p in sorted(EX.glob("*/map.json")):
    d = validate(p, "map")
    maps[p.parent.name] = d
    if (p.parent.name == "map") != (d["kind"] == "public"):
        fail(p, "only the public map has kind public")
public_map = maps.get("map", {})
map_ids = {d["id"]: name for name, d in maps.items()}

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
        if r["kind"] == "probed" and r["applies_to"] == "composition" and "seeded_violation" not in r:
            fail(p, f"probed composition rule {r['id']} needs a seeded violation (ADR-0071)")
contract_rules = {r["id"]: r for c in contracts.values() for r in c["rules"]}


def check_results(e_path, pred):
    """Results for platform-limited rules are not-applicable elsewhere (ADR-0076)."""
    for r in pred["results"]:
        plats = contract_rules.get(r["rule"], {}).get("platforms")
        if plats and pred["platform"] not in plats and r["status"] != "not-applicable":
            fail(e_path, f"{r['rule']} runs only on {plats}; on {pred['platform']} it is not-applicable")

# ---- libraries (public and private maps share one schema) ----
# A private map's file with the id of a library in a map beneath it extends
# that library (ADR-0075): library fields and existing versions must match,
# and it may only add versions, variants and platform targets. Maps layer in
# directory order: map/, then private-map*/ sorted by name.
libraries = {}
library_files = sorted(EX.glob("map/libraries/*.json")) + sorted(
    p for p in EX.glob("*/libraries/*.json") if p.parts[-3] != "map")
for p in library_files:
    d = validate(p, "library")
    nda_platforms = set(maps.get(p.relative_to(EX).parts[0], {}).get("platforms", []))
    if d["id"] in libraries:
        base = libraries[d["id"]][1]
        for k in ("name", "license", "contribution_policy", "slots", "source"):
            if d.get(k) != base.get(k):
                fail(p, f"extends {d['id']} but its {k} differs from the map it extends")
        for v in d["versions"]:
            bv = next((x for x in base["versions"] if x["version"] == v["version"]), None)
            if bv is None:
                base["versions"].append(v)
                continue
            for k in set(v) | set(bv):
                if k not in ("variants", "facts") and v.get(k) != bv.get(k):
                    fail(p, f"extends {d['id']} {v['version']} but its {k} differs")
            new_t = v["facts"].get("platform", {}).get("targets", [])
            old_t = bv["facts"].get("platform", {}).get("targets", [])
            strip = lambda f: {**f, "platform": {k: x for k, x in f.get("platform", {}).items() if k != "targets"}}
            if strip(v["facts"]) != strip(bv["facts"]) or not set(old_t) <= set(new_t):
                fail(p, f"extends {d['id']} {v['version']}: facts may only gain platform targets")
            elif set(new_t) - set(old_t) - nda_platforms:
                fail(p, f"extends {d['id']} {v['version']}: only an NDA platform's map may add targets, and only its own platforms")
            elif new_t:
                bv["facts"]["platform"]["targets"] = old_t + [t for t in new_t if t not in old_t]
            existing = {x["name"]: x for x in bv["variants"]}
            for var in v["variants"]:
                if var["name"] not in existing:
                    bv["variants"].append(var)
                elif var != existing[var["name"]]:
                    fail(p, f"variant {var['name']} differs from the map it extends")
    else:
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


def merged_facts(v, variant):
    var = next((x for x in v["variants"] if x["name"] == variant), {})
    out = {seam: dict(f) for seam, f in v["facts"].items()}
    for seam, over in var.get("facts", {}).items():
        out.setdefault(seam, {}).update(over)
    return out


def variant_of(v, variant):
    return next((x for x in v["variants"] if x["name"] == variant), None)


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
        # A binding with nothing to probe says why for every probed library rule of its contract.
        lib_rules = {r["id"] for r in contracts.get(subj.get("contract"), {}).get("rules", [])
                     if r["kind"] == "probed" and r["applies_to"] == "library"}
        if d["kind"] == "pairing" or not lib_rules or lib_rules - {n["rule"] for n in d.get("not_applicable", [])}:
            fail(p, "binding and pairing recipes ship their probes, unless every probed library rule is not-applicable")
    owners = set(subj.get("libraries", [])) | {rid}
    for o in d.get("ordering", []):
        for step in (o["before"], o["after"]):
            if step.split(":", 1)[0] not in owners:
                fail(p, f"ordering step {step} isn't owned by {sorted(owners)}")

# ---- compositions ----
locks = {}
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
    for pm in m.get("private_maps", []):
        if pm["id"] not in map_ids:
            fail(m_path, f"private map {pm['id']} has no example map.json")
        elif set(pm.get("platforms", [])) != set(maps[map_ids[pm["id"]]].get("platforms", [])):
            fail(m_path, f"private map {pm['id']}: platforms differ from its map.json")

    lock_path = comp / "automap.lock.json"
    raw = lock_path.read_text()
    lock = validate(lock_path, "lock", json.loads(raw))
    if raw != canon(lock):
        fail(lock_path, "lock is not in canonical form")
    body = {k: v for k, v in lock.items() if k != "lock_hash"}
    if lock["lock_hash"] != "sha256:" + hashlib.sha256(canon(body).encode()).hexdigest():
        fail(lock_path, "lock_hash doesn't match the canonical lock")
    locks[lock["lock_hash"]] = (raw, used)
    if set(lock["pins"]) != used:
        fail(lock_path, "pins must name exactly the libraries in the manifest's slots")
    if lock["map_release"] != m["map_release"]:
        fail(lock_path, "map_release differs from the manifest")
    if set(lock.get("private_maps", {})) != private:
        fail(lock_path, "private_maps must match the manifest's")
    def variant_on(lib, plat):
        chosen = m["variants"].get(lib)
        return chosen if isinstance(chosen, str) else chosen.get(plat, chosen["default"])

    def check_pin(lib, pin, where):
        v = library_version(lib, pin["version"], pin["variant"])
        if v is None:
            fail(lock_path, f"{where}: {lib} {pin['version']} ({pin['variant']}) is not in the map")
            return None
        map_pin = variant_of(v, pin["variant"]).get("pin") or v.get("pin") or {}
        if pin.get("commit") != map_pin.get("commit"):
            fail(lock_path, f"{where}: {lib}: pinned commit differs from the map")
        if map_pin and pin["content_hash"] != map_pin["content_hash"]:
            fail(lock_path, f"{where}: {lib}: content hash differs from the map")
        reg = pin.get("registry", {})
        if reg.get("kind") == "vcpkg-curated" and reg.get("commit") != public_map["build_backend"]["vcpkg"]["baseline"]:
            fail(lock_path, f"{where}: {lib}: curated registry commit differs from the map's vcpkg baseline")
        if pin["license"] != libraries[lib][1]["license"]["spdx"]:
            fail(lock_path, f"{where}: {lib}: license differs from the map")
        source = libraries[lib][1].get("source", {})
        if "path" in pin and (source.get("kind") != "path" or pin["path"] != source.get("path")):
            fail(lock_path, f"{where}: {lib}: only a path library is pinned by path, and the path must match the map")
        if "registry" in pin and source.get("kind") == "path":
            fail(lock_path, f"{where}: {lib}: a path library has no registry")
        if pin.get("registry", {}).get("kind") == "private" and pin["registry"].get("id") not in private:
            fail(lock_path, f"{where}: {lib}: private registry {pin['registry'].get('id')} is not a private map of this composition")
        return v

    for lib, pin in lock["pins"].items():
        chosen = m["variants"].get(lib)
        default = chosen if isinstance(chosen, str) else chosen["default"]
        if pin["variant"] != default:
            fail(lock_path, f"{lib}: pinned variant differs from the manifest's default")
        differing = {plat for plat in m["platforms"] if variant_on(lib, plat) != default}
        if set(pin.get("platforms", {})) != differing:
            fail(lock_path, f"{lib}: platform pins must be exactly the platforms whose variant differs: {sorted(differing)}")
        for plat in m["platforms"]:
            ppin = pin.get("platforms", {}).get(plat, pin)
            if ppin["variant"] != variant_on(lib, plat):
                fail(lock_path, f"{lib} on {plat}: pinned variant differs from the manifest")
            v = check_pin(lib, ppin, plat)
            if v and plat not in merged_facts(v, ppin["variant"]).get("platform", {}).get("targets", []):
                fail(m_path, f"{lib} ({ppin['variant']}) doesn't list {plat} among its targets")
    for c, ver in lock["contracts"].items():
        if c in contracts and contracts[c]["version"] != ver:
            fail(lock_path, f"contract {c} {ver} differs from the example contract")

    for e_path in sorted(comp.glob("evidence/*.json")):
        e = validate(e_path, "evidence")
        pred = e["predicate"]
        check_results(e_path, pred)
        if pred["map_release"] == "unreleased" and pred["verifier"].get("signed"):
            fail(e_path, "evidence against an unreleased map is never signed (ADR-0079)")
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

# ---- published evidence in maps ----
for e_path in sorted(p for p in EX.glob("*/evidence/**/*.json") if p.parts[-4] != "composition"):
    e = validate(e_path, "evidence")
    pred = e["predicate"]
    check_results(e_path, pred)
    in_public = e_path.relative_to(EX).parts[0] == "map"
    if pred["verifier"]["kind"] != ("automap-ci" if in_public else "nda-holder"):
        fail(e_path, "the public map's evidence comes from Automap's CI; a private map's from the NDA holder")
    if pred["subject_kind"] == "composition":
        raw, used = locks.get(pred["subject_detail"]["lock_hash"], (None, set()))
        if raw is None:
            fail(e_path, "subject_detail.lock_hash matches no example lock")
        elif e["subject"][0]["digest"].get("sha256") != hashlib.sha256(raw.encode()).hexdigest():
            fail(e_path, "subject digest doesn't match the lock file")
        for r in pred["results"]:
            if r.get("library") and r["library"] not in used:
                fail(e_path, f"result names {r['library']}, which isn't in the composition")
    if pred["subject_kind"] == "binding":
        pin, c = pred["subject_detail"]["pin"], pred["subject_detail"]["contract"]
        if not library_version(pin["library"], pin["version"], pin["variant"]):
            fail(e_path, f"no library example has {pin['library']} {pin['version']} ({pin['variant']})")
        for r in pred["results"]:
            if not r["rule"].startswith(c + "."):
                fail(e_path, f"binding evidence for {c} reports rule {r['rule']}")
    if pred["map_release"] == "unreleased" and pred["verifier"].get("signed"):
        fail(e_path, "evidence against an unreleased map is never signed (ADR-0079)")
    if pred["level"] in ("probed", "behavioral", "hardware") and any(
            r["status"] in ("fail", "unverifiable") for r in pred["results"]):
        fail(e_path, "known-good level with failing or unverifiable results")

for p in sorted(EX.glob("composition/cli-results/*.json")):
    d = validate(p, "cli-result")
    if d["ok"] == bool(d["failures"]):
        fail(p, "ok must be true exactly when there are no failures")

print(f"{checked} files checked against {len(schemas)} schemas")
if errors:
    print("\n".join(errors))
    sys.exit(1)
print("all examples valid")
