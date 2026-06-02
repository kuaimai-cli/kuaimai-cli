#!/usr/bin/env python3
"""Normalize internal/registry/meta_data.json to kuaimai-cli meta_data.json 定义规范."""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
META_PATH = ROOT / "internal/registry/meta_data.json"

# CLI 核心 5 个 operation id（规范 §二.2 / §六）
CORE_OPS = frozenset(
    {"stock-list", "stock-count", "item-detail", "item-save", "item-update-title"}
)

CORE_BY_PATH_METHOD = {
    ("POST", "/item/stock/queryList"): "stock-list",
    ("POST", "/item/stock/queryCount"): "stock-count",
    ("GET", "/item/getItemDetail"): "item-detail",
}

# 路径首段作为模块前缀（/item/{module}/...）
MODULE_PREFIXES = frozenset(
    {
        "stock",
        "batch",
        "backend",
        "repair",
        "dubbo",
        "config",
        "tb",
        "shopitem",
        "intelligent",
        "platform",
        "cat",
        "customer",
        "print",
        "shipper",
        "packma",
        "aoxiang",
        "brand",
        "assist",
        "base",
        "conversion",
        "wx",
        "smallpro",
        "imagesearch",
        "dict",
        "excel",
        "open",
        "yccrm",
        "virtual",
        "process",
        "multi",
        "upload",
        "delete",
        "light",
        "update",
        "generation",
        "seller",
        "category",
        "property",
        "segment",
        "classify",
        "bridge",
        "back",
        "packaging",
        "material",
        "history",
        "price",
        "import",
        "export",
        "fill",
        "copy",
        "edit",
        "column",
        "suite",
        "entity",
        "code",
        "goods",
        "section",
        "for",
        "dms",
        "magnifier",
        "assign",
        "auto",
        "api",
        "app",
        "packma",
        "intelligent",
        "platform",
        "shopitem",
        "customer",
        "print",
        "shipper",
        "assist",
        "unit",
        "formula",
        "brand",
        "cat",
        "tb",
        "repair",
        "backend",
        "batch",
        "stock",
        "dubbo",
        "config",
        "aoxiang",
        "wx",
        "smallpro",
    }
)


def camel_to_kebab(name: str) -> str:
    s = re.sub(r"([a-z0-9])([A-Z])", r"\1-\2", name)
    s = s.replace("_", "-").lower()
    return re.sub(r"-+", "-", s).strip("-")


def path_to_op_id(path: str, method: str, old_name: str) -> str:
    key = (method.upper(), path)
    if key in CORE_BY_PATH_METHOD:
        return CORE_BY_PATH_METHOD[key]

    if old_name in CORE_OPS:
        return old_name

    p = path.strip("/")
    if p == "item":
        if "importCount" in old_name or "importCount" in path:
            return "item-import-count-4-history"
        if "exportItemList" in old_name or "export" in old_name.lower():
            return "item-export-item-list"
        return f"item-root-{method.lower()}"

    if not p.startswith("item/"):
        p = "item/" + p
    sub = p[5:]
    segments = sub.split("/")

    first_kebab = camel_to_kebab(segments[0])
    first_token = first_kebab.split("-")[0]

    if len(segments) >= 2 and first_token in MODULE_PREFIXES:
        return "-".join(camel_to_kebab(seg) for seg in segments)

    return "-".join(["item"] + [camel_to_kebab(seg) for seg in segments])


def disambiguate(op_id: str, method: str, path: str, old_name: str) -> str:
    """Avoid collisions with core ids or duplicate generated ids."""
    if op_id in CORE_OPS and old_name not in CORE_OPS:
        if path == "/item/getItemDetail" and method.upper() == "POST":
            return "item-get-item-detail-post"
        if path == "/item/saveItem" and method.upper() == "POST":
            return "item-save-item-post"
        return f"{op_id}-{method.lower()}"
    return op_id


def ensure_pageable_schema(op: dict) -> bool:
    changed = False
    if not op.get("pageable"):
        return changed
    rs = op.setdefault("requestSchema", {"type": "object", "properties": {}})
    props = rs.setdefault("properties", {})
    if "pageNo" not in props:
        props["pageNo"] = {"type": "number", "desc": "页码", "default": 1}
        changed = True
    if "pageSize" not in props:
        props["pageSize"] = {"type": "number", "desc": "每页条数", "default": 50}
        changed = True
    required = rs.setdefault("required", [])
    for field in ("pageNo", "pageSize"):
        if field not in required:
            required.append(field)
            changed = True
    return changed


def normalize_operation(old_name: str, op: dict) -> tuple[str, dict, list[str]]:
    notes: list[str] = []
    new_name = path_to_op_id(op["path"], op["method"], old_name)
    new_name = disambiguate(new_name, op["method"], op["path"], old_name)

    # POST /item/saveItem：保留 item-save / item-update-title 两个核心 op
    if op["path"] == "/item/saveItem" and op["method"].upper() == "POST":
        if old_name in ("item-save", "item-update-title"):
            new_name = old_name

    if new_name != old_name:
        notes.append(f"rename {old_name} -> {new_name}")

    if ensure_pageable_schema(op):
        notes.append(f"add pagination schema: {new_name}")

    return new_name, op, notes


def normalize_meta(meta: dict) -> tuple[dict, list[str]]:
    all_notes: list[str] = []
    new_services: dict = {}

    for svc_name, svc in meta.get("services", {}).items():
        new_ops: dict = {}
        for old_name, op in svc.get("operations", {}).items():
            new_name, op, notes = normalize_operation(old_name, op)
            all_notes.extend(notes)
            if new_name in new_ops:
                existing = new_ops[new_name]
                if existing == op:
                    all_notes.append(f"drop duplicate {old_name} (same as {new_name})")
                    continue
                # 保留 schema 更完整的条目
                existing_score = len(json.dumps(existing.get("requestSchema") or {}))
                new_score = len(json.dumps(op.get("requestSchema") or {}))
                if new_score > existing_score:
                    all_notes.append(f"replace {new_name} with richer schema from {old_name}")
                    new_ops[new_name] = op
                else:
                    all_notes.append(f"drop {old_name} (collision with {new_name})")
            else:
                new_ops[new_name] = op
        svc = dict(svc)
        svc["operations"] = dict(sorted(new_ops.items()))
        new_services[svc_name] = svc

    meta = dict(meta)
    meta["services"] = new_services
    meta["version"] = meta.get("version", "1.4.0")
    if not meta["version"].endswith(".1"):
        meta["version"] = "1.4.1"
    meta["generated_at"] = "2026-06-01"
    return meta, all_notes


def validate(meta: dict) -> list[str]:
    errors: list[str] = []
    valid_ct = {"get_query", "post_form", "post_json"}
    for svc_name, svc in meta.get("services", {}).items():
        for op_name, op in svc.get("operations", {}).items():
            for field in ("summary", "method", "path", "contentType", "write", "pageable"):
                if field not in op:
                    errors.append(f"missing {field}: {svc_name}/{op_name}")
            ct = op.get("contentType")
            if ct not in valid_ct:
                errors.append(f"bad contentType {ct}: {svc_name}/{op_name}")
            if op.get("method") == "GET" and ct != "get_query":
                errors.append(f"GET must use get_query: {svc_name}/{op_name}")
            if "-" not in op_name and op_name not in CORE_OPS:
                errors.append(f"operation id must use module-path kebab-case: {svc_name}/{op_name}")
            if op.get("pageable"):
                props = (op.get("requestSchema") or {}).get("properties") or {}
                if "pageNo" not in props or "pageSize" not in props:
                    errors.append(f"pageable missing pagination: {svc_name}/{op_name}")
    item_ops = meta.get("services", {}).get("item", {}).get("operations", {})
    for core in CORE_OPS:
        if core not in item_ops:
            errors.append(f"missing core operation: {core}")
    return errors


def main() -> int:
    raw = META_PATH.read_text(encoding="utf-8")
    meta = json.loads(raw)
    normalized, notes = normalize_meta(meta)
    errors = validate(normalized)
    if errors:
        print("validation errors after normalize:", file=sys.stderr)
        for e in errors[:20]:
            print(f"  {e}", file=sys.stderr)
        if len(errors) > 20:
            print(f"  ... and {len(errors) - 20} more", file=sys.stderr)
        return 1

    META_PATH.write_text(json.dumps(normalized, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"normalized {META_PATH}")
    print(f"operations: {sum(len(s['operations']) for s in normalized['services'].values())}")
    print(f"changes: {len(notes)}")
    for n in notes[:40]:
        print(f"  - {n}")
    if len(notes) > 40:
        print(f"  ... and {len(notes) - 40} more")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
