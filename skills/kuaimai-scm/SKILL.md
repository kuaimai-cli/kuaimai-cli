---
name: kuaimai-scm
version: 3.0.0
description: "快麦 ERP 供应链域：员工、铺货日志、scm 商品、平台铺货配置。用户提到 scm、供应链、铺货、操作日志、分销、dsb 时使用。接口一律经 registry 发现，不维护本地接口表。"
metadata:
  requires:
    bins: ["kuaimai-cli"]
  cliHelp: "kuaimai-cli capabilities --output json"
---

# scm（供应链域 · erp-scm）

**CRITICAL — 开始前 MUST Read [`../kuaimai-shared/SKILL.md`](../kuaimai-shared/SKILL.md)**

scm 域 **无 CLI shortcuts**，统一 `web call <apiId>`。  
**禁止**凭记忆或文档里的 apiId 表调用；参数以 `schema <apiId>` 为准。

## 何时读本 Skill

用户意图涉及：供应链、scm、铺货、铺货日志、操作日志、分销商品、平台铺货配置、员工（scm 入口）、dsb……

ERP 库存商品 / 改标题 → [`kuaimai-item`](../kuaimai-item/SKILL.md)

## 标准流程（每步 MUST 执行）

```bash
# 1. 发现：按 domain 或 title/description 关键词筛选（apiId 通常以 scm. 开头）
kuaimai-cli capabilities --output json

# 2. 自省
kuaimai-cli schema <apiId> --output json

# 3. 调用（targetHost 来自 registry baseUrl，通常 scm.superboss.cc）
kuaimai-cli web call <apiId> --body '{"pageNo":1,"pageSize":20}' --output json
```

| 步骤 | 规则 |
|------|------|
| 选 apiId | 只从 `capabilities` 返回的 apis 中选 |
| 域名 | `web call scm.*` 自动走 `scm.superboss.cc`；**不要**改 `api.url` |
| 日志类 | `schema` 若要求 `startTime`/`endTime`，缺省先问用户；格式 `yyyy-MM-dd HH:mm:ss` |
| 写操作 | `write:true` 时先 `--dry-run --verbose` |
| 翻页 | 仅 `pageable:true` 可用 `--page-all` |

## 域分流（CRITICAL）

路径同为 `/item/*` 时，erp1 与 scm **语义不同**。  
MUST Read [`references/kuaimai-scm-domain-routing.md`](references/kuaimai-scm-domain-routing.md)

## 意图 → 发现策略（非接口表）

| 用户意图 | 在 capabilities/schema 中找 |
|----------|------------------------------|
| 员工 | staff |
| 铺货 / 操作日志 | logging、publish、log |
| 供应链商品 | item、base、分销 |
| 平台铺货配置 | dsb、distribution、config |

## 不在本 Skill 范围

- 配置、登录、registry → [`kuaimai-shared`](../kuaimai-shared/SKILL.md)
- ERP 商品（erp1）→ [`kuaimai-item`](../kuaimai-item/SKILL.md)
