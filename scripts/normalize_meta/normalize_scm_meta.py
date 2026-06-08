#!/usr/bin/env python3
"""Post-normalize services.scm in meta_data.json to match kuaimai-cli meta 定义规范."""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
META_PATH = ROOT / "internal/registry/meta_data.json"

BAD_SUMMARY_MARKERS = (
    "Created by",
    "Copyright",
    "@author",
    "@program:",
    "erp-scm",
    "版权所有",
)

LOGGING_PAGE_OPS = frozenset(
    {
        "logging-publish-log",
        "logging-platform-product-publish-log",
        "logging-product-edit-log-page",
        "logging-operator-log",
        "logging-auto-publish-log",
    }
)

FIELD_DESC_MAP: dict[str, str] = {
    "pageNo": "页码",
    "pageSize": "每页条数",
    "startTime": "开始时间，格式 yyyy-MM-dd HH:mm:ss",
    "endTime": "结束时间，格式 yyyy-MM-dd HH:mm:ss",
    "shopType": "平台类型，如 TouTiaoFXG、KWaiShop、PinShop、JdShop",
    "shopId": "店铺ID",
    "shopIds": "店铺ID列表",
    "baseItemId": "供应链商品ID",
    "categoryId": "类目ID",
    "companyId": "公司ID",
    "outerId": "款式编码",
    "skuOuterId": "SKU商品编码",
    "title": "商品名称",
    "staffId": "员工ID",
    "staffIds": "操作人/员工ID列表",
    "queryStaffName": "姓名/登录账号/手机号",
    "useSupplyChain": "是否启用供应链，0否 1是",
    "publishStatus": "铺货状态：0铺货中 1全成功 2全失败 3部分成功",
    "hasRetry": "重试情况：0未重试 1已重试",
    "status": "状态：1成功 2失败",
    "operationType": "操作类型：0铺货 1添加 2删除 3编辑",
    "type": "类型：0铺货 1添加商品 2删除商品 3编辑商品",
    "userId": "店铺用户ID",
    "brandName": "品牌名称",
    "api_name": "接口名",
    "total": "总条数",
    "records": "结果列表",
}

CORE_SCM_OVERRIDES: dict[str, dict] = {
    "staff-query": {
        "summary": "查询员工列表",
        "write": False,
        "pageable": True,
    },
    "staff-show-edit-staff-shop": {
        "summary": "查询员工店铺权限",
        "write": False,
        "pageable": False,
        "requestSchema": {
            "type": "object",
            "properties": {
                "staffId": {"type": "number", "desc": "员工ID"},
                "api_name": {"type": "string", "desc": "接口名"},
            },
            "required": ["staffId"],
        },
    },
    "logging-publish-log": {
        "summary": "铺货日志分页查询",
        "write": False,
        "pageable": True,
        "requestSchema": {
            "type": "object",
            "required": ["startTime", "endTime", "pageNo", "pageSize"],
            "properties": {
                "pageNo": {"type": "number", "desc": "页码", "default": 1},
                "pageSize": {"type": "number", "desc": "每页条数", "default": 50},
                "startTime": {"type": "string", "desc": "铺货开始时间，格式 yyyy-MM-dd HH:mm:ss"},
                "endTime": {"type": "string", "desc": "铺货结束时间，格式 yyyy-MM-dd HH:mm:ss"},
                "shopId": {"type": "number", "desc": "店铺ID"},
                "shopIds": {"type": "array", "desc": "店铺ID列表", "items": {"type": "number"}},
                "outerId": {"type": "string", "desc": "款式编码"},
                "publishStatus": {"type": "number", "desc": "铺货状态：0铺货中 1全成功 2全失败 3部分成功"},
                "hasRetry": {"type": "number", "desc": "重试情况：0未重试 1已重试"},
                "staffIds": {"type": "array", "desc": "操作人ID列表", "items": {"type": "string"}},
            },
        },
    },
    "logging-platform-product-publish-log": {
        "summary": "平台商品铺货日志分页查询",
        "write": False,
        "pageable": True,
    },
    "logging-product-edit-log-page": {
        "summary": "商品编辑日志分页查询",
        "write": False,
        "pageable": True,
    },
    "logging-operator-log": {
        "summary": "操作日志分页查询",
        "write": False,
        "pageable": True,
    },
    "logging-query-channel-by-type": {
        "summary": "查询操作途径枚举",
        "write": False,
        "pageable": False,
        "requestSchema": {
            "type": "object",
            "required": ["type"],
            "properties": {
                "type": {"type": "number", "desc": "类型：0铺货 1添加商品 2删除商品 3编辑商品"},
            },
        },
    },
    "item-base-page": {
        "summary": "供应链商品分页列表",
        "write": False,
        "pageable": True,
    },
    "item-base-detail": {
        "summary": "查询供应链商品详情",
        "write": False,
        "pageable": False,
    },
    "dsb-query-distribution-config": {
        "summary": "查询平台铺货配置",
        "write": False,
        "pageable": False,
        "requestSchema": {
            "type": "object",
            "required": ["shopType"],
            "properties": {
                "shopType": {"type": "string", "desc": "平台类型，如 TouTiaoFXG"},
            },
        },
    },
    "dsb-apply-brand-config": {
        "summary": "一键应用品牌配置",
        "write": True,
        "pageable": False,
        "requestSchema": {
            "type": "object",
            "required": ["shopType"],
            "properties": {
                "shopType": {"type": "string", "desc": "平台类型"},
                "baseItemId": {"type": "string", "desc": "供应链商品ID"},
                "categoryId": {"type": "number", "desc": "类目ID"},
                "shopIds": {"type": "string", "desc": "店铺ID，多个逗号分隔"},
            },
        },
    },
    "dsb-save-or-update-distribution-config": {
        "summary": "保存或更新平台铺货配置",
        "write": True,
        "pageable": False,
    },
}


def is_bad_summary(summary: str) -> bool:
    s = (summary or "").strip()
    if not s or len(s) < 2:
        return True
    if any(m in s for m in BAD_SUMMARY_MARKERS):
        return True
    if re.match(r"^[a-z][a-zA-Z0-9]+$", s):  # javaMethodName
        return True
    return False


def humanize_path_summary(path: str, method: str) -> str:
    seg = path.rstrip("/").split("/")[-1]
    name = re.sub(r"([a-z0-9])([A-Z])", r"\1-\2", seg).replace("_", "-")
    verbs = {
        "query": "查询",
        "get": "查询",
        "list": "列表查询",
        "page": "分页查询",
        "save": "保存",
        "update": "更新",
        "delete": "删除",
        "edit": "编辑",
        "apply": "应用",
        "sync": "同步",
        "publish": "铺货",
        "show": "查看",
    }
    low = seg.lower()
    for en, zh in verbs.items():
        if low.startswith(en):
            rest = seg[len(en) :]
            rest = re.sub(r"([A-Z])", r"\1", rest).strip()
            if rest:
                return f"{zh}{rest}"
            return zh
    return f"{method} {path}"


def enrich_property_desc(name: str, prop: dict) -> dict:
    prop = dict(prop)
    desc = str(prop.get("desc") or "").strip()
    if not desc or desc == name or desc.isascii() and desc.replace("_", "").isalnum():
        if name in FIELD_DESC_MAP:
            prop["desc"] = FIELD_DESC_MAP[name]
        elif not desc:
            prop["desc"] = name
    else:
        prop["desc"] = re.sub(r"<[^>]+>", "", desc).strip()
        if prop["desc"].endswith("<p>"):
            prop["desc"] = prop["desc"][:-3].strip()
    if name == "pageNo" and "default" not in prop:
        prop["default"] = 1
    if name == "pageSize" and "default" not in prop:
        prop["default"] = 50
    return prop


def normalize_request_schema(op: dict, op_name: str) -> dict | None:
    rs = op.get("requestSchema")
    if not rs:
        return rs
    rs = dict(rs)
    props = {k: enrich_property_desc(k, v) for k, v in (rs.get("properties") or {}).items()}
    rs["properties"] = props
    required = list(rs.get("required") or [])
    if op.get("pageable"):
        for f in ("pageNo", "pageSize"):
            if f in props and f not in required:
                required.append(f)
    if op_name in LOGGING_PAGE_OPS:
        for f in ("startTime", "endTime"):
            if f in props and f not in required:
                required.append(f)
    if required:
        rs["required"] = required
    return rs


def normalize_response_schema(op: dict) -> dict:
    pageable = op.get("pageable", False)
    ret = op.get("responseSchema") or {}
    data_type = "object"
    data_desc = "业务数据"
    if pageable:
        data_desc = "分页结果（含 records、total 等）"
    return {
        "type": "object",
        "properties": {
            "result": {"type": "number", "desc": "业务结果码，1 为成功"},
            "message": {"type": "string", "desc": "提示信息"},
            "data": {"type": data_type, "desc": data_desc},
        },
    }


def merge_schema(existing: dict | None, override: dict | None) -> dict | None:
    if not override:
        return existing
    if not existing:
        return override
    out = dict(existing)
    for key in ("summary", "write", "pageable", "method", "path", "contentType"):
        if key in override:
            out[key] = override[key]
    if "requestSchema" in override:
        ex_rs = existing.get("requestSchema") or {"type": "object", "properties": {}}
        ov_rs = override["requestSchema"]
        props = dict(ex_rs.get("properties") or {})
        for k, v in (ov_rs.get("properties") or {}).items():
            props[k] = v
        out["requestSchema"] = {
            "type": "object",
            "properties": props,
            "required": ov_rs.get("required") or ex_rs.get("required") or [],
        }
    return out


def normalize_scm_operation(op_name: str, op: dict) -> dict:
    op = dict(op)
    if is_bad_summary(op.get("summary", "")):
        op["summary"] = humanize_path_summary(op.get("path", ""), op.get("method", ""))
    op["summary"] = re.sub(r"<[^>]+>", "", op["summary"]).strip()[:120]

    if op.get("method") == "GET" and not any(
        k in (op.get("path") or "").lower()
        for k in ("save", "update", "delete", "edit", "sync", "apply", "publish")
    ):
        op["write"] = False

    if op_name in CORE_SCM_OVERRIDES:
        op = merge_schema(op, CORE_SCM_OVERRIDES[op_name]) or op

    op["requestSchema"] = normalize_request_schema(op, op_name)
    op["responseSchema"] = normalize_response_schema(op)
    return op


def validate_scm(meta: dict) -> list[str]:
    errors: list[str] = []
    scm = meta.get("services", {}).get("scm", {})
    if not scm.get("baseUrl"):
        errors.append("scm missing baseUrl")
    for op_name, op in scm.get("operations", {}).items():
        for field in ("summary", "method", "path", "contentType", "write", "pageable"):
            if field not in op:
                errors.append(f"missing {field}: scm/{op_name}")
        if is_bad_summary(op.get("summary", "")):
            errors.append(f"bad summary: scm/{op_name}")
        if op.get("pageable"):
            props = (op.get("requestSchema") or {}).get("properties") or {}
            if "pageNo" not in props or "pageSize" not in props:
                errors.append(f"pageable missing pagination: scm/{op_name}")
    return errors


def main() -> int:
    meta = json.loads(META_PATH.read_text(encoding="utf-8"))
    scm = meta.get("services", {}).get("scm")
    if not scm:
        print("no scm service in meta", file=sys.stderr)
        return 1

    new_ops = {}
    for name, op in scm.get("operations", {}).items():
        new_ops[name] = normalize_scm_operation(name, op)

    scm = dict(scm)
    scm["operations"] = dict(sorted(new_ops.items()))
    meta.setdefault("services", {})["scm"] = scm

    errors = validate_scm(meta)
    if errors:
        print("validation errors:", file=sys.stderr)
        for e in errors[:30]:
            print(f"  {e}", file=sys.stderr)
        return 1

    META_PATH.write_text(json.dumps(meta, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"normalized scm: {len(new_ops)} operations")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
