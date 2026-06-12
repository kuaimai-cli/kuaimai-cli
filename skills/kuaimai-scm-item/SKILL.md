---
name: kuaimai-scm-item
version: 3.0.0
description: "快麦 SCM 可铺货商品（erp-scm）：供应链商品、分销商品、铺货到店铺、铺货日志、平台铺货配置。该商品才可用于后续上货/铺货。"
metadata:
  requires:
    bins: ["kuaimai-cli"]
  cliHelp: "kuaimai-cli scm-item --help"
---

# scm-item（SCM 可铺货商品 · erp-scm）

**CRITICAL — 开始前 MUST Read [`../kuaimai-shared/SKILL.md`](../kuaimai-shared/SKILL.md)**

本 Skill 处理 **SCM 可铺货商品**：这类商品才是后续可以上货、铺货、发布到店铺的商品。  
SCM 商品来源可以是手动新增，也可以从 [`kuaimai-erp-item`](../kuaimai-erp-item/SKILL.md) 的 ERP 商品档案导入；当前 CLI 先不实现导入 shortcut。

如果用户说的是“商品档案”的查询、列表、新增、编辑、保存、改标题，不要使用本 Skill，必须切换到 [`kuaimai-erp-item`](../kuaimai-erp-item/SKILL.md)。

scm-item 域优先使用已发布 CLI shortcut；无 shortcut 时统一 `web call <apiId>`。  
**禁止**凭记忆或文档里的 apiId 表调用；参数以 `schema <apiId>` 为准。

## 何时读本 Skill

用户意图涉及：SCM 商品、供应链商品、可铺货商品、分销商品、供销商品、铺货、上货到店铺、发布到店铺、铺货日志、操作日志、平台铺货配置、员工（scm 入口）、dsb……

ERP 商品档案 / 库存 / 改标题 / sysItemId → [`kuaimai-erp-item`](../kuaimai-erp-item/SKILL.md)

## 决策表

| 用户需求 | 优先方式 | 说明 |
|----------|----------|------|
| 按款式编码定位 SCM 可铺货商品 | `scm-item +list --style-code <款式编码>` | 确认 `canPublishPlatform` 是否包含目标平台 |
| 查询商品可铺货店铺 | `scm-item shops --platform pdd --style-code <款式编码>` | 输出 `can_publish` 与 `disabled_reason` |
| 按款式编码把商品铺货到 PDD 店铺 | `scm-item publish-pdd` | 这是 SCM 可铺货商品，不是 ERP 商品档案 |
| 查询铺货日志 / 失败原因 | `scm-item publish-log --detail` | shortcut 已封装日志明细 |
| 查询 SCM 可铺货商品列表、平台资料、分销配置等复杂条件 | `capabilities` → `schema <apiId>` → `web call` | 复杂条件看 schema |
| 商品档案查询/列表/新增/编辑/保存/改标题 | `kuaimai-erp-item` | 不属于 SCM 可铺货商品 |

## 标准流程（每步 MUST 执行）

已支持的 shortcut：

```bash
# 定位 SCM 可铺货商品
kuaimai-cli scm-item +list --style-code '<款式编码>' --output json

# 查询该商品可铺货的 PDD 店铺
kuaimai-cli scm-item shops --platform pdd --style-code '<款式编码>' --output json

# 拼多多铺货预检：按 SCM 款式编码查询可铺货商品、店铺、平台资料完整性，但不会提交铺货任务
kuaimai-cli scm-item publish-pdd --style-code '<款式编码>' --shop '<拼多多店铺名>' --output json

# 用户明确确认后才实际提交铺货
kuaimai-cli scm-item publish-pdd --style-code '<款式编码>' --shop '<拼多多店铺名>' --submit --output json

# 提交后查询最近铺货日志和失败原因
kuaimai-cli scm-item publish-pdd --style-code '<款式编码>' --shop '<拼多多店铺名>' --submit --check-log --output json

# 单独查询铺货日志明细
kuaimai-cli scm-item publish-log --style-code '<款式编码>' --shop '<拼多多店铺名>' --detail --output json
```

`scm-item publish-pdd` 默认只停在最终 `/pdd/batchPublishItem` 前；实际提交必须带 `--submit`。`scm-item publish-log --detail` 会读取 `/logging/publishLogDetail`，失败原因看明细行 `errorMessage`。

无 shortcut 的 scm-item 接口按以下流程：

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
| SCM 可铺货商品 / 供应链商品 | item、base、分销 |
| 平台铺货配置 | dsb、distribution、config |

## 不在本 Skill 范围

- 配置、登录、registry → [`kuaimai-shared`](../kuaimai-shared/SKILL.md)
- ERP 商品档案（erp1，不可直接铺货）→ [`kuaimai-erp-item`](../kuaimai-erp-item/SKILL.md)
