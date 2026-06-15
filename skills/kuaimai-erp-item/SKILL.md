---
name: kuaimai-erp-item
version: 4.0.0
description: "快麦 ERP 商品档案（erp-items-core）：商品档案列表、统计、详情、标题修改。该商品不能直接用于上货/铺货；需要铺货请使用 kuaimai-scm-item。"
metadata:
  requires:
    bins: ["kuaimai-cli"]
  cliHelp: "kuaimai-cli erp-item --help"
---

# erp-item（ERP 商品档案 · erp-items-core）

**CRITICAL — 开始前 MUST Read [`../kuaimai-shared/SKILL.md`](../kuaimai-shared/SKILL.md)**
**CRITICAL — 只有明确出现“商品档案 / ERP 商品档案 / sysItemId / ERP 档案资料维护 / ERP 库存”等强信号时才使用本 Skill。**
**CRITICAL — 用户只说“商品 / 款式编码 / outerId / 标题修改 / 上货 / 铺货”时，默认不是 ERP 商品档案，必须切换到 [`kuaimai-scm-item`](../kuaimai-scm-item/SKILL.md)。**
**CRITICAL — 禁止盲目找接口：本 Skill 找不到 shortcut 时，只允许在 ERP 商品档案域内做一次有边界的 capabilities/schema 发现；不得遍历尝试 `kuaimai-cli` 所有 commands、所有 domains、所有 API。**

本 Skill 只处理 **ERP 系统中的商品档案**。凡是用户明确说“商品档案 / ERP 商品档案”的查询、列表、统计、详情、新增、编辑、保存、改标题，才归 `kuaimai-erp-item`。这些商品档案用于 ERP 内部资料、库存、SKU、标题等管理，**不是 SCM 可铺货商品，不能直接上货到店铺**。  
如用户要“上货、铺货、发布到店铺、拼多多铺货”，必须切换到 [`kuaimai-scm-item`](../kuaimai-scm-item/SKILL.md)。

本 Skill 只做 **ERP 商品档案意图识别**、**已发布 shortcuts** 与 **registry 发现调用**。  
**禁止**凭记忆或文档里的 apiId 表调用；参数以 `schema <apiId>` 为准。

## 何时读本 Skill

用户意图涉及：ERP 商品档案查询、商品档案列表、商品档案新增、商品档案编辑、商品档案保存、ERP SKU、ERP 货号、sysItemId、ERP 库存列表、ERP 商品详情、ERP 档案改名、品牌、类目、档案……

商品 / 款式编码 / outerId / 普通标题修改 / 供应链可铺货商品 / 铺货 / 分销商品 / 上货到店铺 → [`kuaimai-scm-item`](../kuaimai-scm-item/SKILL.md)

## 标准流程（每步 MUST 执行）

## 核心场景

### 1. 明确 ERP 商品档案查询 / 修改

用户明确说“商品档案”“ERP 商品档案”或提供 `sysItemId` 时，优先使用已发布 shortcut：

```bash
kuaimai-cli erp-item +list --body '{"title":"关键词","pageNo":1,"pageSize":50}' --output json
kuaimai-cli erp-item get-detail --sys-item-id <sysItemId> --output json
kuaimai-cli erp-item update-title --sys-item-id <sysItemId> --title '<新标题>' --dry-run --output json
```

写操作先 `--dry-run --verbose`，用户确认后再提交。

### 2. 用户只给款式编码 / outerId / 普通商品标题修改

这不是 ERP 商品档案强信号。必须切换到 [`kuaimai-scm-item`](../kuaimai-scm-item/SKILL.md)，先用 `scm-item +list --style-code <款式编码>` 定位 SCM 商品。不要用 `erp-item +list` 或 ERP capabilities 兜底。

例如：

```text
修改商品款式编码为 69c1025cec66640001436846 的标题
```

默认属于 `scm-item`，不是 `erp-item update-title`。

### 3. ERP 商品档案无 shortcut 的复杂能力

只允许在 ERP 商品档案域内做 bounded discovery：

1. `capabilities --output json` 后只筛选 `domain`/`label`/`apiId` 明确属于 ERP 商品档案、erp-items-core、商品档案的候选。
2. 候选超过 3 个或没有高置信候选时，停止并告诉用户“当前 CLI 已有能力”和“缺少 ERP 商品档案 registry 能力”，不要继续试 SCM、api 或所有 capabilities。
3. 有唯一高置信候选时，必须先 `schema <apiId> --output json`，写操作先 `--dry-run --verbose`。

## 决策表

| 用户需求 | 优先方式 | 是否查 capabilities/schema |
|----------|----------|---------------------------|
| 按标题查商品档案列表 | `erp-item +list` | 否 |
| 按标题统计商品档案数量 | `erp-item count` | 否 |
| 已知 `sysItemId` 查商品档案详情 | `erp-item get-detail` | 否 |
| 修改商品档案标题 | `erp-item update-title` | 否；写前先 `--dry-run` |
| 保存/编辑完整商品档案 body | `erp-item save` + reference | 通常否；字段不确定时查 schema |
| 商品档案新增、复杂编辑、复杂筛选 | ERP 商品档案域 bounded discovery | 是 |
| 品牌、类目、状态、时间、供应商等多条件组合查询商品档案 | ERP 商品档案域 bounded discovery | 是 |
| 普通商品款式编码 / outerId / 标题修改 | 切换 [`kuaimai-scm-item`](../kuaimai-scm-item/SKILL.md) | 否 |
| 查询/发布可铺货商品 | 切换 [`kuaimai-scm-item`](../kuaimai-scm-item/SKILL.md) | 按 scm-item 规则 |

## Shortcuts（优先使用）

| Shortcut | 说明 |
|----------|------|
| `+list` | 查询 ERP 商品档案列表 |
| `count` | 统计 ERP 商品档案数量 |
| `get-detail` | 按 `sysItemId` 查询 ERP 商品档案详情 |
| `update-title` | 按 `sysItemId` 修改 ERP 商品档案标题 |
| `save` | 保存 ERP 商品档案完整 body |

已支持的 ERP 商品档案 shortcut：

```bash
kuaimai-cli erp-item +list --body '{"title":"关键词","pageNo":1,"pageSize":50}' --output json
kuaimai-cli erp-item count --body '{"title":"关键词"}' --output json
kuaimai-cli erp-item get-detail --sys-item-id <sysItemId> --output json
kuaimai-cli erp-item update-title --sys-item-id <sysItemId> --title '<新标题>' --dry-run --output json
```

## API Resources / Registry 发现边界

无 shortcut、复杂入参或字段不确定的 ERP 商品档案接口按以下流程，且必须限制在 ERP 商品档案域：

```bash
# 1. 发现：按 domain/label/apiId/title/description 明确属于 ERP 商品档案筛选
kuaimai-cli capabilities --output json

# 2. 自省：requestSchema / pageable / write / examples / contentType
kuaimai-cli schema <apiId> --output json

# 3. 调用（经 open-cli 网关）
kuaimai-cli web call <apiId> --params '{"k":"v"}'    # get_query
kuaimai-cli web call <apiId> --data '{"k":"v"}'     # post_json / post_form
kuaimai-cli web call <apiId> --body '{"k":"v"}'     # 按 contentType 自动路由
```

| 步骤 | 规则 |
|------|------|
| 选 apiId | 只从 `capabilities` 返回的 ERP 商品档案候选中选；最多保留 1-3 个高置信候选 |
| 填参数 | 只按 `schema` 的 `requestSchema`；可参考 `examples` |
| 写操作 | `write:true` 时先 `--dry-run --verbose`，用户确认后再提交 |
| 翻页 | 仅 `pageable:true` 可用 `--page-all`；全量前先评估数据量 |
| 找不到 | `registry sync` → 重跑 `capabilities`；仍无则停止并说明当前 CLI 支持的 `erp-item` shortcut；不要继续试 `scm-item`、`api` 或所有 capabilities |

> **重要**：使用 registry API 前必须先运行 `schema` 查看参数结构，不要猜测字段格式。找不到能力时要向用户反馈能力缺口，而不是盲目重试。

## 域分流（CRITICAL）

路径前缀同为 `/item/`，但 **erp1 与 scm 是不同后端**：

| 用户描述 | 读本 Skill | 后端 |
|----------|-----------|------|
| 明确 ERP 商品档案查询/列表/新增/编辑/保存/改标题、ERP 库存、sysItemId | ✅ `kuaimai-erp-item` | erp1（`api.url`） |
| 商品、款式编码、outerId、普通标题修改、SCM 可铺货商品、供应链商品、铺货、分销商品 | ❌ → `kuaimai-scm-item` | scm.superboss.cc |

不确定时 Read [`../kuaimai-scm-item/references/kuaimai-scm-domain-routing.md`](../kuaimai-scm-item/references/kuaimai-scm-domain-routing.md)

## 意图 → 发现策略（非接口表）

用 `capabilities` + `schema` 匹配，**不要硬编码 apiId**：

| 用户意图 | 在 capabilities/schema 中找 |
|----------|------------------------------|
| 列表 / 搜索 | title 或 description 含 list、query、stock、列表 |
| 统计 / 有多少 | 含 count、统计 |
| 详情 | 含 detail、get、详情 |
| 新建 / 修改 / 保存 | `write:true`，含 save、update、修改 |

只有标题没有 ID：优先用 `erp-item +list` 取 `sysItemId`，再调详情 / 写接口。只有筛选条件超出 shortcut 能力时，才 `capabilities` → `schema` → `web call`。

## 典型流程：按标题定位再修改

| 步骤 | 操作 |
|------|------|
| 1 | `erp-item +list` 按标题搜索 |
| 2 | 从 `data` 取 `sysItemId`（多条则让用户选） |
| 3 | `capabilities` → 选写 apiId → `schema` → `web call --dry-run --verbose` |
| 4 | 用户确认后去掉 `--dry-run` |

## 不在本 Skill 范围

- 配置、登录、registry 机制 → [`kuaimai-shared`](../kuaimai-shared/SKILL.md)
- SCM 可铺货商品 / 铺货 / 操作日志 → [`kuaimai-scm-item`](../kuaimai-scm-item/SKILL.md)
