---
name: kuaimai-scm
version: 2.0.0
description: "快麦 ERP 供应链（erp-scm）：员工、铺货日志、供应链商品、平台铺货配置。用户提到供应链、scm、铺货、铺货日志、操作日志、分销商品、平台配置、staff、dsb 时使用。"
metadata:
  requires:
    bins: ["kuaimai-cli"]
  cliHelp: "kuaimai-cli web --help"
---

# scm（erp-scm）

**CRITICAL — 开始前 MUST Read [`../kuaimai-shared/SKILL.md`](../kuaimai-shared/SKILL.md)**

scm 域 **无 shortcuts**，统一 `web call scm.<operation>`。  
接口是否已发布、参数是什么 → **`capabilities` / `schema <apiId>`**，不在此维护全量接口表。

**域名**：`web call scm.*` 自动请求 `https://scm.superboss.cc/`（registry `baseUrl`），**不要**改 `api.url`。

## 意图路由

**item vs scm 分流**：必读 [`references/kuaimai-scm-domain-routing.md`](references/kuaimai-scm-domain-routing.md)。  
路径同为 `/item/*` 时，`web call item.*`（erp1）与 `web call scm.item-*`（scm）语义不同。

| 用户意图 | apiId | reference |
|----------|-------|-----------|
| 查员工列表 | `scm.staff-query` | [`kuaimai-scm-staff.md`](references/kuaimai-scm-staff.md) |
| 员工店铺权限 | `scm.staff-show-edit-staff-shop` | staff |
| 铺货日志 | `scm.logging-publish-log` | [`kuaimai-scm-logging.md`](references/kuaimai-scm-logging.md) |
| 平台商品铺货日志 | `scm.logging-platform-product-publish-log` | logging |
| 商品编辑日志 | `scm.logging-product-edit-log-page` | logging |
| 操作日志 | `scm.logging-operator-log` | logging |
| 日志途径枚举 | `scm.logging-query-channel-by-type` | logging |
| 供应链商品列表 | `scm.item-base-page` | [`kuaimai-scm-item-base.md`](references/kuaimai-scm-item-base.md) |
| 供应链商品详情 | `scm.item-base-detail` | item-base |
| 平台铺货配置 | `scm.dsb-query-distribution-config` | [`kuaimai-scm-dsb.md`](references/kuaimai-scm-dsb.md) |

不确定 apiId 时：

```bash
kuaimai-cli capabilities --output json
kuaimai-cli schema scm.<operation> --output json
```

## 前置条件

| 场景 | 要求 |
|------|------|
| 铺货/操作日志 | 多数需 `startTime`/`endTime`（`yyyy-MM-dd HH:mm:ss`）；缺省默认近 7 天并问用户 |
| 写操作 | Read 对应 reference；先 `--dry-run --verbose` |
| 分页 | 仅 `pageable:true` 可用 `--page-all`；见 [`kuaimai-scm-meta-execution.md`](references/kuaimai-scm-meta-execution.md) |

**CRITICAL — 写操作前 MUST Read 对应 references**

## 快速决策

- ERP 库存商品 / 改标题 → [`kuaimai-item`](../kuaimai-item/SKILL.md)
- 供应链 / 铺货 / 操作日志 → 本 Skill + `web call scm.*`
- 日志缺时间范围 → 向用户确认后填入
- `result=901` → 引导 `auth login`

## 典型场景

```bash
kuaimai-cli web call scm.staff-query \
  --body '{"pageNo":1,"pageSize":20}' --output json --no-color

kuaimai-cli web call scm.logging-publish-log \
  --body '{"pageNo":1,"pageSize":20,"startTime":"2026-06-01 00:00:00","endTime":"2026-06-08 23:59:59"}' \
  --output json --no-color

kuaimai-cli web call scm.item-base-page \
  --body '{"pageNo":1,"pageSize":50}' --output json --no-color

kuaimai-cli web call scm.dsb-query-distribution-config \
  --params '{"shopType":"TouTiaoFXG"}' --output json --no-color
```

通用调用规则：[`kuaimai-scm-web-call.md`](references/kuaimai-scm-web-call.md)

## 不在本 skill 范围

- 配置、登录、registry → [`kuaimai-shared`](../kuaimai-shared/SKILL.md)
- ERP 商品 shortcuts → [`kuaimai-item`](../kuaimai-item/SKILL.md)
