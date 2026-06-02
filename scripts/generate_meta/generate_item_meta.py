#!/usr/bin/env python3
"""Generate/enrich internal/registry/meta_data.json from erp-items-core + erp-core Java sources."""

from __future__ import annotations

import json
import re
import sys
from dataclasses import dataclass, field
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
ERP_ITEMS = ROOT.parent / "erp-items-core"
ERP_CORE = ROOT.parent / "erp-core"
CONTROLLER_ROOT = ERP_ITEMS / "dmj-controllers/dmj-items-stock-controller/src/main/java"
META_PATH = ROOT / "internal/registry/meta_data.json"

JAVA_SEARCH_ROOTS = [
    ERP_ITEMS / "dmj-controllers/dmj-items-stock-controller/src/main/java",
    ERP_ITEMS / "dmj-services",
    ERP_ITEMS / "dmj-domain",
    ERP_CORE / "dmj-services",
    ERP_CORE / "dmj-domain",
    ERP_CORE / "dmj-api",
]

CORE_BY_PATH_METHOD = {
    ("POST", "/item/stock/queryList"): "stock-list",
    ("POST", "/item/stock/queryCount"): "stock-count",
    ("GET", "/item/getItemDetail"): "item-detail",
    ("POST", "/item/saveItem"): "item-save",
}

CORE_OPS = frozenset(
    {"stock-list", "stock-count", "item-detail", "item-save", "item-update-title"}
)

CLASS_MAPPING = re.compile(
    r"@RequestMapping\s*\(\s*(?:value\s*=\s*)?[\"'](/item[^\"']*)[\"']"
)
METHOD_MAPPING = re.compile(
    r"@(?:(?:Get|Post|Put|Delete|Patch)Mapping|RequestMapping)\s*\(([^)]*)\)",
    re.DOTALL,
)
STRING_LIT = re.compile(r"[\"']([^\"']+)[\"']")
REQUEST_METHOD = re.compile(r"RequestMethod\.(\w+)")
METHOD_SIG = re.compile(
    r"(?:public|protected)\s+([\w.<>,\s\[\]]+?)\s+(\w+)\s*\(([^)]*)\)",
    re.DOTALL,
)
FIELD_RE = re.compile(
    r"(?:/\*\*([^*]*(?:\*+[^/][^*]*)*)\*/\s*)?"
    r"private\s+([\w.<>,\s\[\]]+?)\s+(\w+)\s*;",
    re.DOTALL,
)
SKIP_PARAM_TYPES = frozenset(
    {
        "Staff",
        "HttpServletRequest",
        "HttpServletResponse",
        "Model",
        "ModelMap",
        "BindingResult",
        "RedirectAttributes",
        "Principal",
        "Locale",
        "TimeZone",
        "OutputStream",
        "Writer",
        "PrintWriter",
        "MultipartFile",
        "MultipartHttpServletRequest",
    }
)
WRITE_KEYWORDS = (
    "save",
    "update",
    "delete",
    "remove",
    "add",
    "create",
    "insert",
    "batch",
    "bind",
    "release",
    "import",
    "export",
    "upload",
    "modify",
    "edit",
    "set",
    "clear",
    "reset",
    "enable",
    "disable",
    "callback",
    "fire",
    "sync",
    "convert",
    "merge",
    "assign",
    "cancel",
    "confirm",
    "submit",
    "send",
    "push",
    "pull",
    "move",
    "copy",
    "repair",
    "fix",
    "init",
    "generate",
    "build",
    "apply",
    "revoke",
    "lock",
    "unlock",
)


def camel_to_kebab(name: str) -> str:
    s = re.sub(r"([a-z0-9])([A-Z])", r"\1-\2", name)
    s = s.replace("_", "-").lower()
    return re.sub(r"-+", "-", s).strip("-")


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
            if lit:
                paths.append(lit.group(1))
    if not paths:
        paths = [""]
    if not methods:
        methods = {"GET", "POST"}
    return paths, methods


@dataclass
class Endpoint:
    method: str
    path: str
    java_method: str
    summary: str
    content_type: str
    write: bool
    pageable: bool
    request_schema: dict | None = None
    response_schema: dict | None = None
    return_type: str = "Object"


class JavaIndex:
    def __init__(self) -> None:
        self._by_simple: dict[str, list[Path]] = {}

    def build(self) -> None:
        for root in JAVA_SEARCH_ROOTS:
            if not root.is_dir():
                continue
            for path in root.rglob("*.java"):
                self._by_simple.setdefault(path.stem, []).append(path)

    def find_class(self, simple_name: str) -> str | None:
        paths = self._by_simple.get(simple_name)
        if not paths:
            return None
        return paths[0].read_text(encoding="utf-8", errors="ignore")

    def parse_fields(self, simple_name: str, depth: int = 0, seen: set[str] | None = None) -> dict:
        if depth > 2:
            return {}
        seen = seen or set()
        if simple_name in seen:
            return {}
        seen.add(simple_name)
        source = self.find_class(simple_name)
        if not source:
            return {}
        props: dict = {}
        extends_match = re.search(r"class\s+\w+\s+extends\s+(\w+)", source)
        if extends_match and depth < 2:
            parent = extends_match.group(1)
            if parent not in ("Object", "Serializable"):
                props.update(self.parse_fields(parent, depth + 1, seen))
        for m in FIELD_RE.finditer(source):
            javadoc, jtype, fname = m.group(1), m.group(2).strip(), m.group(3)
            desc = clean_javadoc(javadoc) if javadoc else fname
            props[fname] = java_type_to_schema(jtype, desc, self, depth)
        return props


def clean_javadoc(raw: str | None) -> str:
    if not raw:
        return ""
    lines = []
    for line in raw.splitlines():
        line = line.strip()
        line = re.sub(r"^/\*\*?|\*/$|^\* ?", "", line).strip()
        if line.startswith("@"):
            continue
        if line:
            lines.append(line)
    return " ".join(lines[:2]) or ""


def java_type_to_schema(jtype: str, desc: str, idx: JavaIndex, depth: int) -> dict:
    jtype = re.sub(r"\s+", "", jtype)
    base = {"desc": desc or jtype}
    if jtype.startswith("List<") or jtype.startswith("Set<"):
        inner = re.search(r"<(.+)>", jtype)
        inner_type = inner.group(1) if inner else "Object"
        item_schema = java_type_to_schema(inner_type, inner_type, idx, depth + 1)
        return {"type": "array", **base, "items": {"type": item_schema.get("type", "object")}}
    if jtype.startswith("Map"):
        return {"type": "object", **base}
    mapping = {
        "String": "string",
        "Integer": "number",
        "int": "number",
        "Long": "number",
        "long": "number",
        "Double": "number",
        "double": "number",
        "Float": "number",
        "float": "number",
        "Boolean": "boolean",
        "boolean": "boolean",
        "BigDecimal": "number",
        "Date": "string",
    }
    if jtype in mapping:
        return {"type": mapping[jtype], **base}
    if jtype.endswith("[]"):
        return {"type": "array", **base}
    # nested POJO — shallow properties only
    nested = idx.parse_fields(jtype, depth + 1)
    if nested:
        return {"type": "object", **base, "properties": nested}
    return {"type": "object", **base, "desc": desc or jtype}


def parse_method_params(sig_args: str, idx: JavaIndex) -> tuple[dict, str, bool]:
    props: dict = {}
    has_body = False
    return_type_hint = ""
    if not sig_args.strip():
        return props, "post_form", False
    parts = split_params(sig_args)
    for part in parts:
        part = part.strip()
        if not part:
            continue
        body = "@RequestBody" in part
        part = re.sub(r"@\w+(?:\([^)]*\))?\s*", "", part).strip()
        tokens = part.split()
        if len(tokens) < 2:
            continue
        ptype, pname = tokens[0], tokens[1]
        ptype = ptype.replace("final", "").strip()
        simple = ptype.split(".")[-1].replace("[]", "")
        if simple in SKIP_PARAM_TYPES:
            continue
        if simple in ("String", "Integer", "Long", "int", "long", "Double", "Boolean", "boolean"):
            props[pname] = java_type_to_schema(simple, pname, idx, 0)
            continue
        if body or simple.endswith("Request") or simple.endswith("Model") or simple.endswith("DTO") or simple.endswith("Vo") or simple.endswith("VO") or simple.endswith("Param") or simple.endswith("Params"):
            nested = idx.parse_fields(simple)
            for k, v in nested.items():
                props[k] = v
            has_body = body
        else:
            props[pname] = java_type_to_schema(simple, pname, idx, 0)
    content = "post_json" if has_body else "post_form"
    pageable = "pageNo" in props and "pageSize" in props
    return props, content, pageable


def split_params(s: str) -> list[str]:
    parts: list[str] = []
    depth = 0
    cur: list[str] = []
    for ch in s:
        if ch == "<":
            depth += 1
        elif ch == ">":
            depth -= 1
        elif ch == "," and depth == 0:
            parts.append("".join(cur))
            cur = []
            continue
        cur.append(ch)
    if cur:
        parts.append("".join(cur))
    return parts


def infer_write(java_method: str, path: str) -> bool:
    low = java_method.lower()
    path_low = path.lower()
    for kw in WRITE_KEYWORDS:
        if kw in low or kw in path_low:
            if kw == "export" and ("query" in low or "list" in low or "get" in low):
                continue
            if kw == "init" and ("check" in low or "query" in low):
                continue
            return True
    for kw in ("query", "get", "list", "count", "search", "find", "load", "fetch", "check", "detail", "view", "read"):
        if kw in low:
            return False
    return False


def default_response_schema(return_type: str, pageable: bool) -> dict:
    props = {
        "result": {"type": "number", "desc": "业务结果码，1 为成功"},
        "message": {"type": "string", "desc": "提示信息"},
    }
    data_type = "array" if pageable else "object"
    data_desc = return_type if return_type != "Object" else ("分页列表" if pageable else "业务数据")
    props["data"] = {"type": data_type, "desc": data_desc}
    if pageable:
        props["total"] = {"type": "number", "desc": "总条数"}
    return {"type": "object", "properties": props}


def path_to_op_id(path: str, method: str, java_method: str) -> str:
    key = (method.upper(), path)
    if key in CORE_BY_PATH_METHOD:
        return CORE_BY_PATH_METHOD[key]
    p = path.strip("/")
    if not p.startswith("item/"):
        if p == "item":
            return f"item-root-{method.lower()}"
        p = "item/" + p
    sub = p[5:] if p.startswith("item/") else p
    segments = sub.split("/") if sub else []
    if not segments:
        return f"item-{camel_to_kebab(java_method)}"
    return "-".join(camel_to_kebab(seg) for seg in segments)


def extract_javadoc_summary(text: str, pos: int) -> str:
    before = text[:pos]
    matches = list(re.finditer(r"/\*\*([^*]*(?:\*+[^/][^*]*)*)\*/", before, re.DOTALL))
    if not matches:
        return ""
    return clean_javadoc(matches[-1].group(1))


def scan_controller(path: Path, idx: JavaIndex) -> list[Endpoint]:
    text = path.read_text(encoding="utf-8", errors="ignore")
    m = CLASS_MAPPING.search(text)
    if not m:
        return []
    base = m.group(1).rstrip("/")
    endpoints: list[Endpoint] = []
    for match in METHOD_MAPPING.finditer(text):
        ann = match.group(0)
        args = match.group(1)
        if f'"{base}"' in ann or f"'{base}'" in ann:
            if "value" not in args or args.strip() in (f'"{base}"', f"'{base}'"):
                continue
        sub_paths, methods = parse_mapping_args(args)
        if "@GetMapping" in ann:
            methods = {"GET"}
        elif "@PostMapping" in ann:
            methods = {"POST"}
        elif "@PutMapping" in ann:
            methods = {"PUT"}
        elif "@DeleteMapping" in ann:
            methods = {"DELETE"}
        # locate method signature after annotation
        after = text[match.end() : match.end() + 800]
        sig = METHOD_SIG.search(after)
        if not sig:
            continue
        ret_type, java_method, sig_args = sig.group(1).strip(), sig.group(2), sig.group(3)
        ret_simple = ret_type.split(".")[-1]
        summary_doc = extract_javadoc_summary(text, match.start())
        req_props, body_ct, pageable = parse_method_params(sig_args, idx)
        for sub in sub_paths:
            full = normalize_path(base, sub)
            for method in methods:
                if method not in ("GET", "POST"):
                    continue
                ct = "get_query" if method == "GET" else body_ct
                write = infer_write(java_method, full)
                if "query" in java_method.lower() or "count" in java_method.lower() or "list" in java_method.lower():
                    write = False
                if pageable or ("pageNo" in req_props and "pageSize" in req_props):
                    pageable = True
                    req_props.setdefault("pageNo", {"type": "number", "desc": "页码", "default": 1})
                    req_props.setdefault("pageSize", {"type": "number", "desc": "每页条数", "default": 50})
                summary = summary_doc or f"{java_method}"
                endpoints.append(
                    Endpoint(
                        method=method,
                        path=full,
                        java_method=java_method,
                        summary=summary[:120],
                        content_type=ct,
                        write=write,
                        pageable=pageable,
                        request_schema={"type": "object", "properties": req_props} if req_props else None,
                        response_schema=default_response_schema(ret_simple, pageable),
                        return_type=ret_simple,
                    )
                )
    return endpoints


def schema_richness(schema: dict | None) -> int:
    if not schema:
        return 0
    return len(json.dumps(schema, ensure_ascii=False))


def merge_schema(existing: dict | None, generated: dict | None) -> dict | None:
    if not generated and not existing:
        return None
    if not existing:
        return generated
    if not generated:
        return existing
    if schema_richness(existing) >= schema_richness(generated):
        return existing
    # merge properties
    out = dict(generated)
    ex_props = (existing.get("properties") or {}) if existing.get("type") == "object" else {}
    gen_props = (generated.get("properties") or {}) if generated.get("type") == "object" else {}
    merged = dict(gen_props)
    for k, v in ex_props.items():
        if k not in merged or (v.get("desc") and not merged[k].get("desc")):
            merged[k] = v
        elif schema_richness(v) > schema_richness(merged.get(k, {})):
            merged[k] = v
    out["properties"] = merged
    if existing.get("required"):
        out["required"] = existing["required"]
    return out


def build_core_overrides(idx: JavaIndex) -> dict[str, dict]:
    stock_props = idx.parse_fields("QueryItemStockListRequest")
    stock_props["pageNo"] = {"type": "number", "desc": "页码", "default": 1}
    stock_props["pageSize"] = {"type": "number", "desc": "每页条数", "default": 50}
    stock_props.setdefault("pageType", {"type": "string", "desc": "页面类型，如 ITEM_STOCK"})
    stock_props.setdefault("subPageType", {"type": "string", "desc": "子页面类型，如 ARCHIVE_V2"})
    stock_props.setdefault("api_name", {"type": "string", "desc": "接口名"})

    save_fields = idx.parse_fields("SysItemModel")
    # keep item-save concise: merge parsed + key fields
    for key in ("sysItemId", "title", "outerId", "catId", "skus"):
        if key not in save_fields:
            save_fields[key] = {
                "type": "array" if key == "skus" else ("number" if key == "sysItemId" else "string"),
                "desc": {
                    "sysItemId": "系统商品ID，新增时可空或0",
                    "title": "商品标题",
                    "outerId": "主商家编码",
                    "catId": "商品类目ID",
                    "skus": "规格列表（SysSkuModel）",
                }[key],
            }

    return {
        "stock-list": {
            "summary": "查询商品库存列表",
            "method": "POST",
            "path": "/item/stock/queryList",
            "contentType": "post_form",
            "write": False,
            "pageable": True,
            "requestSchema": {
                "type": "object",
                "properties": stock_props,
                "required": ["pageNo", "pageSize"],
            },
            "responseSchema": default_response_schema("ItemStockDTO", True),
        },
        "stock-count": {
            "summary": "统计商品库存数量",
            "method": "POST",
            "path": "/item/stock/queryCount",
            "contentType": "post_form",
            "write": False,
            "pageable": False,
            "requestSchema": {"type": "object", "properties": stock_props},
            "responseSchema": {
                "type": "object",
                "properties": {
                    "result": {"type": "number", "desc": "业务结果码"},
                    "data": {
                        "type": "object",
                        "properties": {"total": {"type": "number", "desc": "总条数"}},
                    },
                },
            },
        },
        "item-detail": {
            "summary": "查询商品详情",
            "method": "GET",
            "path": "/item/getItemDetail",
            "contentType": "get_query",
            "write": False,
            "pageable": False,
            "requestSchema": {
                "type": "object",
                "properties": {
                    "sysItemId": {"type": "number", "desc": "系统商品ID"},
                    "api_name": {"type": "string", "desc": "接口名"},
                },
                "required": ["sysItemId"],
            },
            "responseSchema": {
                "type": "object",
                "properties": {
                    "result": {"type": "number", "desc": "业务结果码"},
                    "data": {"type": "object", "desc": "商品详情（ItemDetailVo）"},
                },
            },
        },
        "item-save": {
            "summary": "保存/修改商品",
            "method": "POST",
            "path": "/item/saveItem",
            "contentType": "post_json",
            "write": True,
            "pageable": False,
            "requestSchema": {"type": "object", "properties": save_fields},
            "responseSchema": {
                "type": "object",
                "properties": {
                    "result": {"type": "number", "desc": "业务结果码"},
                    "data": {"type": "object", "desc": "保存结果"},
                },
            },
        },
        "item-update-title": {
            "summary": "修改商品标题（原子 saveItem）",
            "method": "POST",
            "path": "/item/saveItem",
            "contentType": "post_json",
            "write": True,
            "pageable": False,
            "requestSchema": {
                "type": "object",
                "properties": {
                    "sysItemId": {"type": "number", "desc": "系统商品ID"},
                    "title": {"type": "string", "desc": "新标题"},
                },
                "required": ["sysItemId", "title"],
            },
            "responseSchema": {
                "type": "object",
                "properties": {
                    "result": {"type": "number", "desc": "业务结果码"},
                    "data": {"type": "object", "desc": "保存结果"},
                },
            },
        },
    }


def endpoint_to_op(ep: Endpoint, op_id: str) -> dict:
    op: dict = {
        "summary": ep.summary,
        "method": ep.method,
        "path": ep.path,
        "contentType": ep.content_type,
        "write": ep.write,
        "pageable": ep.pageable,
    }
    props = {}
    if ep.request_schema:
        props = ep.request_schema.get("properties") or {}
    rs: dict = {"type": "object", "properties": props}
    if ep.pageable and props:
        rs["required"] = ["pageNo", "pageSize"]
    op["requestSchema"] = rs
    op["responseSchema"] = ep.response_schema or default_response_schema(ep.return_type, ep.pageable)
    return op


def main() -> int:
    if not CONTROLLER_ROOT.is_dir():
        print(f"error: erp-items-core not found at {CONTROLLER_ROOT}", file=sys.stderr)
        return 1

    idx = JavaIndex()
    print("indexing Java sources...")
    idx.build()
    print(f"indexed {len(idx._by_simple)} class names")

    all_eps: dict[tuple[str, str], Endpoint] = {}
    files = list(CONTROLLER_ROOT.rglob("*.java"))
    for path in files:
        for ep in scan_controller(path, idx):
            key = (ep.method, ep.path)
            if key not in all_eps or len(ep.summary) > len(all_eps[key].summary):
                all_eps[key] = ep
    print(f"scanned {len(files)} controllers, {len(all_eps)} endpoints")

    existing: dict = {}
    if META_PATH.is_file():
        existing = json.loads(META_PATH.read_text(encoding="utf-8"))
    old_ops = existing.get("services", {}).get("item", {}).get("operations", {})

    # map existing by path+method
    by_key: dict[tuple[str, str], tuple[str, dict]] = {}
    for name, op in old_ops.items():
        by_key[(op["method"].upper(), op["path"])] = (name, op)

    new_ops: dict[str, dict] = {}
    core_overrides = build_core_overrides(idx)

    for key, ep in sorted(all_eps.items(), key=lambda x: x[1].path):
        op_id = path_to_op_id(ep.path, ep.method, ep.java_method)
        if key == ("POST", "/item/saveItem"):
            # handled separately as item-save + item-update-title
            pass
        old_name, old_op = by_key.get(key, (None, None))
        if old_name and old_name in CORE_OPS:
            op_id = old_name
        elif old_name and op_id not in CORE_OPS:
            # prefer stable existing id when path matches
            if (op_id != old_name) and (old_name not in new_ops):
                op_id = old_name
        gen = endpoint_to_op(ep, op_id)
        if old_op:
            gen["summary"] = old_op.get("summary") if schema_richness({"d": old_op.get("summary")}) and len(old_op.get("summary", "")) > len(gen["summary"]) else gen["summary"]
            if old_op.get("summary", "").startswith("查询") or "商品" in old_op.get("summary", ""):
                gen["summary"] = old_op["summary"]
            gen["requestSchema"] = merge_schema(old_op.get("requestSchema"), gen.get("requestSchema"))
            gen["responseSchema"] = merge_schema(old_op.get("responseSchema"), gen.get("responseSchema"))
        # collision
        if op_id in new_ops and new_ops[op_id]["path"] != gen["path"]:
            op_id = f"{op_id}-{ep.method.lower()}"
        new_ops[op_id] = gen

    for core_id, core_op in core_overrides.items():
        if core_id == "item-save":
            continue
        new_ops[core_id] = core_op
    # item-save and item-update-title share path
    new_ops["item-save"] = core_overrides["item-save"]
    new_ops["item-update-title"] = core_overrides["item-update-title"]

    meta = {
        "version": "1.6.0",
        "generated_at": "2026-06-02",
        "services": {
            "item": {
                "summary": existing.get("services", {}).get("item", {}).get("summary", "快麦ERP商品服务"),
                "description": existing.get("services", {}).get("item", {}).get(
                    "description", "商品管理、库存、查询、编辑、上下架"
                ),
                "operations": dict(sorted(new_ops.items())),
            }
        },
    }

    META_PATH.write_text(json.dumps(meta, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"wrote {META_PATH}")
    print(f"operations: {len(new_ops)}")

    no_req = sum(1 for o in new_ops.values() if "requestSchema" not in o)
    weak_resp = sum(
        1
        for o in new_ops.values()
        if not (o.get("responseSchema") or {}).get("properties")
    )
    print(f"missing requestSchema: {no_req}, weak responseSchema: {weak_resp}")
    for c in CORE_OPS:
        print(f"  core {c}: {'ok' if c in new_ops else 'MISSING'}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
