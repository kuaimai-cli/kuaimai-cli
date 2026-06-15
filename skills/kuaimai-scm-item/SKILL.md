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
**CRITICAL — 商品默认归 SCM：用户只说“商品 / 款式编码 / outerId / 标题修改 / 上货 / 铺货”，默认使用本 Skill。只有明确说“商品档案 / ERP 商品档案 / sysItemId / ERP 档案资料维护”才切到 `kuaimai-erp-item`。**
**CRITICAL — 禁止盲目找接口：本 Skill 找不到 shortcut 时，只允许在 SCM / 供应链域内做一次有边界的 capabilities/schema 发现；不得遍历尝试 `kuaimai-cli` 所有 commands、所有 domains、所有 API。**

本 Skill 处理 **SCM 可铺货商品**：这类商品才是后续可以上货、铺货、发布到店铺的商品。  
SCM 商品来源可以是手动新增，也可以从 [`kuaimai-erp-item`](../kuaimai-erp-item/SKILL.md) 的 ERP 商品档案导入；当前 CLI 先不实现导入 shortcut。

如果用户明确说的是“商品档案 / ERP 商品档案 / sysItemId / ERP 档案资料维护”的查询、列表、新增、编辑、保存、改标题，不要使用本 Skill，必须切换到 [`kuaimai-erp-item`](../kuaimai-erp-item/SKILL.md)。

scm-item 域优先使用已发布 CLI shortcut；无 shortcut 时统一 `web call <apiId>`。  
**禁止**凭记忆或文档里的 apiId 表调用；参数以 `schema <apiId>` 为准。

## 何时读本 Skill

用户意图涉及：商品、款式编码、outerId、标题修改、SCM 商品、供应链商品、可铺货商品、分销商品、供销商品、铺货、上货到店铺、发布到店铺、铺货日志、操作日志、平台铺货配置、员工（scm 入口）、dsb……

明确 ERP 商品档案 / ERP 库存 / ERP 档案改标题 / sysItemId → [`kuaimai-erp-item`](../kuaimai-erp-item/SKILL.md)。普通“商品 / 款式编码 / outerId / 标题修改”仍属于本 Skill。

## 核心场景

### 1. 按款式编码定位普通商品 / SCM 商品

用户说“商品款式编码为 X”“outerId 为 X”“款号 X”时，默认是 SCM 商品：

```bash
kuaimai-cli scm-item +list --style-code '<款式编码>' --output json
```

如果用户要“修改这个商品标题”，先用 `+list` 定位商品；如果当前 shortcut 没有写操作，不要切换到 `erp-item`，只能在 SCM 域内做 bounded discovery。

### 2. 修改商品标题 / 字段，但没有 shortcut

这是 SCM 域需求。处理方式固定为：

1. `scm-item +list --style-code <款式编码> --output json` 定位商品。
2. `capabilities --output json` 后只筛选 `domain`/`label`/`apiId` 明确属于 SCM / 供应链 / `scm.*` 且语义包含 `item/base/title/update/save/edit` 的候选。
3. 候选超过 3 个或没有高置信候选时，停止并告诉用户“当前 CLI 已有能力”和“缺少 SCM 标题修改 shortcut/registry 能力”，不要继续试其他命令。
4. 有唯一高置信候选时，必须先 `schema <apiId> --output json`，写操作先 `--dry-run --verbose`，用户确认后再提交。

### 3. 铺货 / 上货

优先使用 `scm-item publish --platform <平台key>`；实际提交必须用户明确确认后加 `--submit`。

## 决策表

| 用户需求 | 优先方式 | 说明 |
|----------|----------|------|
| 查询 SCM 可铺货商品 | `scm-item +list [--style-code <款式编码>] [--title <关键词>]` | 分页查询；款式编码和标题均为可选过滤条件，确认 `canPublishPlatform` 是否包含目标平台 |
| 按商品款式编码修改标题，但无 shortcut | SCM 域 bounded discovery | 只找 SCM/供应链 item update/save/title 候选；找不到就停止并说明缺口 |
| 查询商品可铺货店铺 | `scm-item shops --platform <平台key> --style-code <款式编码>` | 输出 `can_publish` 与 `disabled_reason` |
| 按款式编码把商品铺货到指定平台店铺 | `scm-item publish --platform <平台key>` | 默认预检不提交；实际提交必须加 `--submit` |
| PDD 兼容入口 | `scm-item publish-pdd` | 等价于 `publish --platform pdd`，保留给旧用法 |
| 分步核对发布接口 | `scm-item platform-*` / `scm-item pdd-*` | 仅用于排查/对齐 HAR；实际铺货优先用 `publish --platform` |
| 查询铺货日志 / 失败原因 | `scm-item publish-log --detail` | shortcut 已封装日志明细 |
| 查询 SCM 可铺货商品列表、平台资料、分销配置等复杂条件 | `capabilities` → `schema <apiId>` → `web call` | 复杂条件看 schema |
| 明确 ERP 商品档案查询/列表/新增/编辑/保存/改标题 | `kuaimai-erp-item` | 必须出现“商品档案 / ERP 商品档案 / sysItemId”等强信号 |

## Shortcuts（优先使用）

| Shortcut | 说明 |
|----------|------|
| `+list` | 查询 SCM 商品列表，可按款式编码/标题过滤 |
| `shops` | 查询指定平台可铺货店铺 |
| `publish` | 通用平台铺货预检/提交 |
| `publish-pdd` | PDD 兼容铺货入口 |
| `publish-log` | 查询铺货日志和失败原因 |
| `platform-*` | 通用底层接口 primitive，仅排查用 |
| `pdd-*` | PDD 专用 primitive，仅排查用 |

## 标准流程（每步 MUST 执行）

已支持的 shortcut：

```bash
# 定位 SCM 可铺货商品
kuaimai-cli scm-item +list --style-code '<款式编码>' --output json

# 分页浏览 SCM 可铺货商品
kuaimai-cli scm-item +list --page 1 --page-size 10 --output json

# 查询该商品可铺货的目标平台店铺
kuaimai-cli scm-item shops --platform '<平台key>' --style-code '<款式编码>' --output json

# 通用铺货预检：按 SCM 款式编码查询可铺货商品、店铺、平台资料完整性，但不会提交铺货任务
kuaimai-cli scm-item publish --platform '<平台key>' --style-code '<款式编码>' --shop '<店铺名>' --output json

# PDD 兼容入口
kuaimai-cli scm-item publish-pdd --style-code '<款式编码>' --shop '<拼多多店铺名>' --output json

# 通用发布链路 primitive；platform-storage-task 默认不提交，必须加 --submit 才会提交
kuaimai-cli scm-item platform-price-precheck --platform '<平台key>' --style-code '<款式编码>' --shop-id 123456 --output json
kuaimai-cli scm-item platform-sec-confirm --platform '<平台key>' --style-code '<款式编码>' --shop-id 123456 --output json
kuaimai-cli scm-item platform-storage-task --platform '<平台key>' --style-code '<款式编码>' --shop-id 123456 --output json
kuaimai-cli scm-item platform-task-speed --batch-task-id '<batchTaskId>' --output json

# 分步核对 PDD 发布链路 primitive；pdd-storage-task 默认不提交，必须加 --submit 才会提交
kuaimai-cli scm-item pdd-video-info --style-code '<款式编码>' --output json
kuaimai-cli scm-item pdd-price-precheck --style-code '<款式编码>' --shop-id 123456 --output json
kuaimai-cli scm-item pdd-auth-status --style-code '<款式编码>' --shop-id 123456 --output json
kuaimai-cli scm-item pdd-sec-confirm --style-code '<款式编码>' --shop-id 123456 --output json
kuaimai-cli scm-item pdd-storage-task --style-code '<款式编码>' --shop-id 123456 --output json
kuaimai-cli scm-item pdd-task-speed --batch-task-id '<batchTaskId>' --output json

# 用户明确确认后才实际提交铺货
kuaimai-cli scm-item publish --platform '<平台key>' --style-code '<款式编码>' --shop '<店铺名>' --submit --output json

# 提交后查询最近铺货日志和失败原因
kuaimai-cli scm-item publish --platform '<平台key>' --style-code '<款式编码>' --shop '<店铺名>' --submit --check-log --output json

# 单独查询铺货日志明细
kuaimai-cli scm-item publish-log --style-code '<款式编码>' --shop '<店铺名>' --detail --output json
```

`scm-item publish --platform <平台key>` 默认只停在最终 `/taskScheduling/storageTask` 前；实际提交必须带 `--submit`。提交后会用 `/taskScheduling/queryTaskSpeed` 查询任务进度。`scm-item publish-log --detail` 会读取 `/logging/publishLogDetail`，失败原因看明细行 `errorMessage`。

实际铺货优先使用高层 `publish --platform`，因为它在代码里串起商品、店铺、平台前置校验、提交任务和进度查询，并保留 `--submit` 安全边界。`platform-*` / `pdd-*` primitive 只用于调试、排查、核对浏览器 HAR 或需要逐步观察接口入参/出参时使用。

支持平台 key：`pdd`、`tb`、`fxg`、`kuaishou`、`jd`、`1688`、`tm`、`tjb`、`yz`、`wxsph`、`wxxd`、`wd`、`xhs`、`xy`、`fxg_gx`、`yyjk`、`pddtemu`、`shein`、`ktt`、`jdms`。其中 PDD 会额外调用视频与授权校验；`tb/fxg/jd/1688` 会调用二次确认；`kuaishou` 当前按 HAR 不调用二次确认。

所有 `scm-item` shortcut 的 JSON `data.interfaces[]` 会返回本次命令涉及的底层接口标识，结构为 `{name, method, path}`。常用接口名包括：`item.base.page`、`shop.allShop`、`pdd.getCarouselVideoInfo`、`ltsTask.preCheckControllerPrice`、`pdd.authorize.authStatus`、`item.base.secConfirmationItemV2`、`taskScheduling.storageTask`、`taskScheduling.queryTaskSpeed`、`logging.publishLog`、`logging.publishLogDetail`、`logging.publishLogById`。

## API Resources / Registry 发现边界

无 shortcut 的 scm-item 接口按以下流程，且必须限制在 SCM 域：

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
| 选 apiId | 只从 `capabilities` 返回的 SCM/供应链候选中选；最多保留 1-3 个高置信候选 |
| 域名 | `web call scm.*` 自动走 `scm.superboss.cc`；**不要**改 `api.url` |
| 日志类 | `schema` 若要求 `startTime`/`endTime`，缺省先问用户；格式 `yyyy-MM-dd HH:mm:ss` |
| 写操作 | `write:true` 时先 `--dry-run --verbose` |
| 翻页 | 仅 `pageable:true` 可用 `--page-all` |
| 找不到 | 停止并说明当前 CLI 支持的 `scm-item` shortcut；不要继续试 `erp-item`、`api` 或所有 capabilities |

> **重要**：使用 registry API 前必须先运行 `schema` 查看参数结构，不要猜测字段格式。找不到能力时要向用户反馈能力缺口，而不是盲目重试。

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
