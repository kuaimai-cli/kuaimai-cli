#!/usr/bin/env python3
"""Generate services.scm in internal/registry/meta_data.json from erp-scm Java controllers."""

from __future__ import annotations

import json
import re
import sys
from dataclasses import dataclass
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
ERP_SCM = ROOT.parent / "erp-scm"
CONTROLLER_ROOT = ERP_SCM / "erp-scm-web/src/main/java"
META_PATH = ROOT / "internal/registry/meta_data.json"
SCM_BASE_URL = "https://scm.superboss.cc/"

JAVA_SEARCH_ROOTS = [
    ERP_SCM / "erp-scm-web/src/main/java",
    ERP_SCM / "erp-scm-domain/src/main/java",
    ERP_SCM / "erp-scm-integration/src/main/java",
    ERP_SCM / "erp-scm-common/src/main/java",
]

PATH_PREFIXES = ("/staff", "/logging", "/item", "/dsb")

CLASS_MAPPING = re.compile(
    r"@RequestMapping\s*\(\s*(?:value\s*=\s*)?(?:\{[^}]*\}|[\"'](/(?:staff|logging|item|dsb)[^\"']*)[\"'])"
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
    r"(?:@\w+(?:\([^)]*\))?\s*)*"
    r"(?:(?:private|protected|public)\s+)?"
    r"([\w.<>,\s\[\]]+?)\s+(\w+)\s*;",
    re.DOTALL,
)
API_OPERATION_RE = re.compile(
    r'@ApiOperation\s*\(\s*value\s*=\s*["\']([^"\']+)["\']',
    re.DOTALL,
)
API_MODEL_PROP_RE = re.compile(
    r'@ApiModelProperty\s*\(\s*(?:value\s*=\s*)?["\']([^"\']+)["\']',
)
NOT_NULL_RE = re.compile(r'@NotNull(?:\([^)]*message\s*=\s*["\']([^"\']+)["\'][^)]*\))?')
BAD_SUMMARY_MARKERS = ("Created by", "Copyright", "@author", "@program:", "版权所有")
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
    "publish",
    "offsale",
    "piece",
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


def path_allowed(path: str) -> bool:
    return any(path == p or path.startswith(p + "/") for p in PATH_PREFIXES)


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
        if depth > 3:
            return {}
        seen = seen or set()
        if simple_name in seen:
            return {}
        seen.add(simple_name)
        source = self.find_class(simple_name)
        if not source:
            return {}
        return parse_field_block(source, self, depth, seen)


def clean_javadoc(raw: str | None) -> str:
    if not raw:
        return ""
    lines = []
    for line in raw.splitlines():
        line = line.strip()
        line = re.sub(r"^/\*\*?|\*/$|^\* ?", "", line).strip()
        if line.startswith("@") or not line:
            continue
        if any(m in line for m in BAD_SUMMARY_MARKERS):
            continue
        lines.append(line)
    text = " ".join(lines[:2]) or ""
    text = re.sub(r"<[^>]+>", "", text).strip()
    return text


def extract_api_operation(text: str, pos: int) -> str:
    chunk = text[max(0, pos - 400) : pos]
    matches = list(API_OPERATION_RE.finditer(chunk))
    if not matches:
        return ""
    return clean_javadoc(matches[-1].group(1))


def parse_field_block(source: str, idx: JavaIndex, depth: int, seen: set[str]) -> dict:
    props: dict = {}
    extends_match = re.search(r"class\s+\w+\s+extends\s+(\w+)", source)
    if extends_match and depth < 2:
        parent = extends_match.group(1)
        if parent not in ("Object", "Serializable") and parent not in seen:
            parent_src = idx.find_class(parent)
            if parent_src:
                props.update(parse_field_block(parent_src, idx, depth + 1, seen | {parent}))

    for m in FIELD_RE.finditer(source):
        javadoc, jtype, fname = m.group(1), m.group(2).strip(), m.group(3)
        field_start = m.start()
        field_chunk = source[max(0, field_start - 200) : field_start]
        api_desc = ""
        apm = list(API_MODEL_PROP_RE.finditer(field_chunk))
        if apm:
            api_desc = clean_javadoc(apm[-1].group(1))
        desc = api_desc or (clean_javadoc(javadoc) if javadoc else "")
        if not desc or desc == fname:
            desc = fname
        prop = java_type_to_schema(jtype, desc, idx, depth)
        props[fname] = prop
        if NOT_NULL_RE.search(field_chunk):
            props[fname]["_required"] = True
    return props


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
    nested = idx.parse_fields(jtype, depth + 1)
    if nested:
        return {"type": "object", **base, "properties": nested}
    return {"type": "object", **base, "desc": desc or jtype}


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


def parse_method_params(sig_args: str, idx: JavaIndex) -> tuple[dict, list[str], str, bool]:
    props: dict = {}
    required: list[str] = []
    has_body = False
    body_type = ""
    if not sig_args.strip():
        return props, required, "post_form", False
    parts = split_params(sig_args)
    for part in parts:
        part = part.strip()
        if not part:
            continue
        body = "@RequestBody" in part
        part_clean = re.sub(r"@\w+(?:\([^)]*\))?\s*", "", part).strip()
        tokens = part_clean.split()
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
        if body or simple.endswith(
            ("Request", "Model", "DTO", "Vo", "VO", "Param", "Params", "Page")
        ):
            nested = idx.parse_fields(simple)
            req_fields = []
            for k, v in nested.items():
                props[k] = {kk: vv for kk, vv in v.items() if kk != "_required"}
                if v.get("_required"):
                    req_fields.append(k)
            required.extend(req_fields)
            has_body = body
            body_type = simple
        else:
            props[pname] = java_type_to_schema(simple, pname, idx, 0)
    content = "post_json" if has_body else "post_form"
    pageable = "pageNo" in props and "pageSize" in props
    if body_type.endswith("Page") or "Page<" in body_type:
        pageable = True
    return props, required, content, pageable


def infer_write(method: str, java_method: str, path: str) -> bool:
    if method == "GET":
        low = java_method.lower()
        path_low = path.lower()
        # GET 副作用接口仍视为写操作
        for kw in ("save", "update", "delete", "apply", "sync", "publish", "submit", "cancel", "edit"):
            if kw in low or kw in path_low:
                return True
        return False
    low = java_method.lower()
    path_low = path.lower()
    for kw in ("query", "get", "list", "count", "search", "find", "load", "fetch", "check", "detail", "view", "read", "show", "page"):
        if kw in low and not any(w in low for w in ("save", "update", "delete", "edit", "sync", "apply")):
            return False
    for kw in WRITE_KEYWORDS:
        if kw in low or kw in path_low:
            if kw == "export" and ("query" in low or "list" in low or "get" in low):
                continue
            if kw == "init" and ("check" in low or "query" in low):
                continue
            return True
    return method == "POST" and "query" not in low and "list" not in low and "page" not in low


def default_response_schema(return_type: str, pageable: bool) -> dict:
    props = {
        "result": {"type": "number", "desc": "业务结果码，1 为成功"},
        "message": {"type": "string", "desc": "提示信息"},
    }
    data_type = "array" if pageable else "object"
    data_desc = return_type if return_type != "Object" else ("分页列表" if pageable else "业务数据")
    props["data"] = {"type": data_type, "desc": data_desc}
    return {"type": "object", "properties": props}


def path_to_op_id(path: str, method: str, java_method: str) -> str:
    segments = [camel_to_kebab(seg) for seg in path.strip("/").split("/") if seg]
    if not segments:
        return f"scm-{camel_to_kebab(java_method)}"
    op_id = "-".join(segments)
    return op_id


def extract_controller_base(text: str) -> str | None:
    m = re.search(
        r"@RequestMapping\s*\(\s*(?:value\s*=\s*)?(?:\{([^}]+)\}|[\"'](/(?:staff|logging|item|dsb)[^\"']*)[\"'])",
        text,
    )
    if not m:
        return None
    if m.group(2):
        return m.group(2).rstrip("/")
    for lit in STRING_LIT.findall(m.group(1)):
        if path_allowed(lit.rstrip("/")):
            return lit.rstrip("/")
    return None


def extract_javadoc_summary(text: str, pos: int) -> str:
    before = text[:pos]
    matches = list(re.finditer(r"/\*\*([^*]*(?:\*+[^/][^*]*)*)\*/", before, re.DOTALL))
    if not matches:
        return ""
    return clean_javadoc(matches[-1].group(1))


def scan_controller(path: Path, idx: JavaIndex) -> list[Endpoint]:
    text = path.read_text(encoding="utf-8", errors="ignore")
    base = extract_controller_base(text)
    if not base or not path_allowed(base):
        return []
    endpoints: list[Endpoint] = []
    for match in METHOD_MAPPING.finditer(text):
        ann = match.group(0)
        args = match.group(1)
        sub_paths, methods = parse_mapping_args(args)
        if "@GetMapping" in ann:
            methods = {"GET"}
        elif "@PostMapping" in ann:
            methods = {"POST"}
        elif "@PutMapping" in ann:
            methods = {"PUT"}
        elif "@DeleteMapping" in ann:
            methods = {"DELETE"}
        after = text[match.end() : match.end() + 800]
        sig = METHOD_SIG.search(after)
        if not sig:
            continue
        ret_type, java_method, sig_args = sig.group(1).strip(), sig.group(2), sig.group(3)
        ret_simple = ret_type.split(".")[-1]
        api_op = extract_api_operation(text, match.start())
        summary_doc = extract_javadoc_summary(text, match.start())
        req_props, required_fields, body_ct, pageable = parse_method_params(sig_args, idx)
        for sub in sub_paths:
            full = normalize_path(base, sub)
            if not path_allowed(full):
                continue
            for method in methods:
                if method not in ("GET", "POST"):
                    continue
                ct = "get_query" if method == "GET" else body_ct
                write = infer_write(method, java_method, full)
                if pageable or ("pageNo" in req_props and "pageSize" in req_props):
                    pageable = True
                    req_props.setdefault("pageNo", {"type": "number", "desc": "页码", "default": 1})
                    req_props.setdefault("pageSize", {"type": "number", "desc": "每页条数", "default": 50})
                summary = api_op or summary_doc or java_method
                if any(m in summary for m in BAD_SUMMARY_MARKERS):
                    summary = api_op or java_method
                endpoints.append(
                    Endpoint(
                        method=method,
                        path=full,
                        java_method=java_method,
                        summary=summary[:120],
                        content_type=ct,
                        write=write,
                        pageable=pageable,
                        request_schema={
                            "type": "object",
                            "properties": req_props,
                            "required": required_fields,
                        }
                        if req_props
                        else None,
                        response_schema=default_response_schema(ret_simple, pageable),
                        return_type=ret_simple,
                    )
                )
    return endpoints


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
    required: list[str] = []
    if ep.request_schema:
        props = ep.request_schema.get("properties") or {}
        required = list(ep.request_schema.get("required") or [])
    rs: dict = {"type": "object", "properties": props}
    if ep.pageable and props:
        for f in ("pageNo", "pageSize"):
            if f not in required and f in props:
                required.append(f)
    if required:
        rs["required"] = required
    op["requestSchema"] = rs
    op["responseSchema"] = ep.response_schema or default_response_schema(ep.return_type, ep.pageable)
    return op


def disambiguate_op_id(op_id: str, method: str, existing: dict[str, dict]) -> str:
    if op_id not in existing:
        return op_id
    other = existing[op_id]
    if other.get("method") == method:
        return op_id
    return f"{op_id}-{method.lower()}"


def main() -> int:
    if not CONTROLLER_ROOT.is_dir():
        print(f"error: erp-scm not found at {CONTROLLER_ROOT}", file=sys.stderr)
        return 1

    idx = JavaIndex()
    print("indexing erp-scm Java sources...")
    idx.build()
    print(f"indexed {len(idx._by_simple)} class names")

    all_eps: dict[tuple[str, str], Endpoint] = {}
    files = list(CONTROLLER_ROOT.rglob("*Controller.java"))
    for path in files:
        for ep in scan_controller(path, idx):
            key = (ep.method, ep.path)
            if key not in all_eps or len(ep.summary) > len(all_eps[key].summary):
                all_eps[key] = ep
    print(f"scanned {len(files)} controllers, {len(all_eps)} scm endpoints")

    new_ops: dict[str, dict] = {}
    for key, ep in sorted(all_eps.items(), key=lambda x: x[1].path):
        op_id = path_to_op_id(ep.path, ep.method, ep.java_method)
        op_id = disambiguate_op_id(op_id, ep.method, new_ops)
        new_ops[op_id] = endpoint_to_op(ep, op_id)

    existing: dict = {}
    if META_PATH.is_file():
        existing = json.loads(META_PATH.read_text(encoding="utf-8"))

    services = existing.setdefault("services", {})
    services["scm"] = {
        "summary": "快麦ERP供应链服务",
        "description": "供应链管理：员工、商品中心、操作日志、平台铺货配置",
        "baseUrl": SCM_BASE_URL,
        "operations": dict(sorted(new_ops.items())),
    }
    existing["version"] = "1.7.0"
    existing["generated_at"] = "2026-06-08"

    META_PATH.write_text(json.dumps(existing, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"wrote {len(new_ops)} scm operations to {META_PATH}")

    # 规范化 scm 域（summary/desc/required/pageable，对齐 item 规范）
    import subprocess

    norm = ROOT / "scripts/normalize_meta/normalize_scm_meta.py"
    if norm.is_file():
        rc = subprocess.call([sys.executable, str(norm)])
        if rc != 0:
            return rc
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
