#!/usr/bin/env python3
"""Keep only meta operations whose path exists on @RequestMapping("/item") controllers in erp-items-core."""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
ERP_ITEMS = ROOT.parent / "erp-items-core"
CONTROLLER_ROOT = ERP_ITEMS / "dmj-controllers/dmj-items-stock-controller/src/main/java"
META_PATH = ROOT / "internal/registry/meta_data.json"

CLASS_ITEM_MAPPING = re.compile(
    r"@RequestMapping\s*\(\s*(?:value\s*=\s*)?[\"'](/item[^\"']*)[\"']"
)
METHOD_MAPPING = re.compile(
    r"@(?:(?:Get|Post|Put|Delete|Patch)Mapping|RequestMapping)\s*\(([^)]*)\)",
    re.DOTALL,
)
STRING_LIT = re.compile(r"[\"']([^\"']+)[\"']")
REQUEST_METHOD = re.compile(r"RequestMethod\.(\w+)")


def normalize_path(base: str, sub: str) -> str:
    if not sub:
        return base.rstrip("/") or "/"
    if sub.startswith("/"):
        combined = base.rstrip("/") + sub
    else:
        combined = base.rstrip("/") + "/" + sub
    combined = re.sub(r"/+", "/", combined)
    if not combined.startswith("/"):
        combined = "/" + combined
    return combined.rstrip("/") if combined != "/" else combined


def parse_mapping_args(args: str) -> tuple[list[str], set[str]]:
    paths: list[str] = []
    methods: set[str] = set()
    for m in REQUEST_METHOD.finditer(args):
        methods.add(m.group(1).upper())
    if "method" in args:
        for m in re.finditer(r"method\s*=\s*\{([^}]+)\}", args):
            for rm in REQUEST_METHOD.finditer(m.group(1)):
                methods.add(rm.group(1).upper())
    value_match = re.search(r"value\s*=\s*\{([^}]+)\}", args, re.DOTALL)
    if value_match:
        paths.extend(STRING_LIT.findall(value_match.group(1)))
    else:
        value_single = re.search(r"value\s*=\s*[\"']([^\"']+)[\"']", args)
        if value_single:
            paths.append(value_single.group(1))
        elif args.strip().startswith('"') or args.strip().startswith("'"):
            lit = STRING_LIT.search(args)
            if lit and "method" not in args.split(lit.group(0))[0]:
                paths.append(lit.group(1))
    if not paths:
        paths = [""]
    if not methods:
        ann = args
        if "@GetMapping" in ann or (not methods and "GetMapping" in ann):
            methods = {"GET"}
        elif "PostMapping" in ann:
            methods = {"POST"}
        else:
            methods = {"GET", "POST"}
    return paths, methods


def scan_controller(path: Path) -> set[tuple[str, str]]:
    text = path.read_text(encoding="utf-8", errors="ignore")
    m = CLASS_ITEM_MAPPING.search(text)
    if not m:
        return set()
    base = m.group(1).rstrip("/")
    endpoints: set[tuple[str, str]] = set()
    for match in METHOD_MAPPING.finditer(text):
        ann = match.group(0)
        if "CLASS" in ann:
            continue
        args = match.group(1)
        # skip class-level @RequestMapping("/item") itself
        if CLASS_ITEM_MAPPING.match("@RequestMapping(" + args + ")"):
            continue
        sub_paths, methods = parse_mapping_args(args)
        if "@GetMapping" in match.group(0):
            methods = {"GET"}
        elif "@PostMapping" in match.group(0):
            methods = {"POST"}
        elif "@PutMapping" in match.group(0):
            methods = {"PUT"}
        elif "@DeleteMapping" in match.group(0):
            methods = {"DELETE"}
        for sub in sub_paths:
            full = normalize_path(base, sub)
            for method in methods:
                endpoints.add((method.upper(), full))
    return endpoints


def scan_all_controllers() -> set[tuple[str, str]]:
    endpoints: set[tuple[str, str]] = set()
    files: list[Path] = []
    for path in CONTROLLER_ROOT.rglob("*.java"):
        text = path.read_text(encoding="utf-8", errors="ignore")
        if CLASS_ITEM_MAPPING.search(text):
            files.append(path)
            endpoints |= scan_controller(path)
    print(f"scanned {len(files)} controllers with @RequestMapping(\"/item\")")
    print(f"found {len(endpoints)} endpoint path+method pairs")
    return endpoints


def filter_meta(meta: dict, allowed: set[tuple[str, str]]) -> tuple[dict, list[str], list[str]]:
    removed: list[str] = []
    kept: list[str] = []
    new_ops: dict = {}
    for name, op in meta["services"]["item"]["operations"].items():
        key = (op["method"].upper(), op["path"].rstrip("/") if op["path"] != "/" else op["path"])
        path_norm = op["path"].rstrip("/") if op["path"] != "/" else op["path"]
        key = (op["method"].upper(), path_norm)
        if key in allowed:
            new_ops[name] = op
            kept.append(name)
        else:
            removed.append(f"{name}: {op['method']} {op['path']}")
    meta = dict(meta)
    meta["services"] = dict(meta["services"])
    meta["services"]["item"] = dict(meta["services"]["item"])
    meta["services"]["item"]["operations"] = dict(sorted(new_ops.items()))
    meta["version"] = "1.5.0"
    meta["generated_at"] = "2026-06-01"
    return meta, kept, removed


def main() -> int:
    if not CONTROLLER_ROOT.is_dir():
        print(f"error: erp-items-core not found at {CONTROLLER_ROOT}", file=sys.stderr)
        return 1
    allowed = scan_all_controllers()
    meta = json.loads(META_PATH.read_text(encoding="utf-8"))
    before = len(meta["services"]["item"]["operations"])
    filtered, kept, removed = filter_meta(meta, allowed)
    after = len(filtered["services"]["item"]["operations"])
    META_PATH.write_text(json.dumps(filtered, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"filtered {META_PATH}")
    print(f"operations: {before} -> {after} (removed {len(removed)})")
    core = ["stock-list", "stock-count", "item-detail", "item-save", "item-update-title"]
    ops = filtered["services"]["item"]["operations"]
    print("core ops retained:")
    for c in core:
        print(f"  {c}: {'yes' if c in ops else 'NO'}")
    if removed:
        print("\nremoved sample (first 20):")
        for r in removed[:20]:
            print(f"  - {r}")
        if len(removed) > 20:
            print(f"  ... and {len(removed) - 20} more")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
