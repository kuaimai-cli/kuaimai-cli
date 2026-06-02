---
name: kuaimai-item
version: 2.0.1
description: "快麦 ERP 商品（erp-items-core）：按标题搜索列表、统计数量、查详情、改标题。用户提到商品、SKU、标题、货号、列表、有多少、详情、改名时使用。"
metadata:
  requires:
    bins: ["kuaimai-cli"]
  cliHelp: "kuaimai-cli item --help"
---

# item (v2.0.0)

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../kuaimai-shared/SKILL.md`](../kuaimai-shared/SKILL.md)，其中包含认证、输出信封与安全规则**

## 架构分层（Agent 必读）

kuaimai-cli 商品域自底向上三层；**meta 注册已完成，Skill 是 Agent 调用的能力指南层**：

| 层级 | 模块 | 状态 | Agent 职责 |
|------|------|------|------------|
| 底层 | `meta_data.json` | ✅ 已完成 | 不直接读文件；需要时用 `kuaimai-cli schema` 自省 |
| 中层 | **Skill（本文件 + references/）** | ✅ 当前层 | 选命令、组参数、控制分页与写操作确认 |
| 上层 | shortcuts / service / api | 已部分实现 | 按本 Skill 路由到具体 CLI 命令 |

**口诀**：有 Shortcut → 不查 schema，读本 Skill / references；无 Shortcut → 先 `schema` 再 `service item <operation>`。

meta 执行规则（contentType / pageable / write / Schema）详见 [`references/kuaimai-item-meta-execution.md`](references/kuaimai-item-meta-execution.md)。

## 选哪个命令

**统计 / 列表三接口区分**（库存 vs 商品档案）：必读 [`references/kuaimai-item-count-dimensions.md`](references/kuaimai-item-count-dimensions.md)。

| 用户意图 | 命令 | 结果字段 |
|----------|------|----------|
| 有多少 + **标题**（库存页口径） | `item count` + `title` | `data.data.total` |
| 有多少 + **品牌/类目/档案筛选** | `service item item-query-count` | `data.data.total` |
| 列出 / 搜索 / 查找 / 有哪些 + 标题（库存页） | `item +list` + `title` | `data` 列表 |
| 列出 / 搜索商品档案 V2 | `service item item-query-list-v2` | `data` 列表 |
| 某商品详情 / sysItemId | `item get-detail` | `data[0]` |
| 改标题 / 改名 | `item update-title`（或 get-detail → jq → `save`） | 先 `--dry-run` |

只有标题没有 ID 时：先 `+list`（或 V2 列表）取 `sysItemId`，再 `get-detail` 或 `update-title`。

## 已注册核心接口（meta → 命令）

以下接口已在 `meta_data.json` 完成契约注册；Agent **优先 shortcuts 左列**。

| meta operation | contentType | write | pageable | path | shortcuts / service |
|----------------|-------------|-------|----------|------|---------------------|
| `stock-list` | post_form | false | **true** | `/item/stock/queryList` | `item +list` / `item list` |
| `stock-count` | post_form | false | false | `/item/stock/queryCount` | `item count` |
| `item-query-count` | post_form | false | **true** | `/item/queryCount` | `service item item-query-count` |
| `item-query-list-v2` | post_form | false | **true** | `/item/queryListV2` | `service item item-query-list-v2` |
| `item-detail` | get_query | false | false | `/item/getItemDetail` | `item get-detail` |
| `item-save` | post_json | **true** | false | `/item/saveItem` | `item save` |
| `item-update-title` | post_json | **true** | false | `/item/saveItem` | `item update-title`（编排仅在 shortcuts） |

> 更多 operation（1000+）见 `kuaimai-cli schema --output json`；无 shortcut 的一律走 [`service 层指南`](references/kuaimai-item-service.md)。

## Shortcuts（推荐优先使用）

| 命令 | 说明 |
|------|------|
| [`+list`](references/kuaimai-item-list.md) | 库存列表（ARCHIVE_V2，`title` 筛选、`--page-all`） |
| [`count`](references/kuaimai-item-count.md) | 按标题统计数量 |
| [`get-detail`](references/kuaimai-item-get-detail.md) | 按 `sysItemId` 查详情 |
| [`update-title`](references/kuaimai-item-update-title.md) | 改标题（内部 get-detail → save 编排） |
| [`save`](references/kuaimai-item-save.md) | 全量 body 保存（复杂字段修改） |

**无 shortcut 的 meta 接口**（商品档案 V2）：

| 命令 | 说明 |
|------|------|
| [`item-query-count`](references/kuaimai-item-query-count.md) | 商品档案总数（`service item item-query-count`） |
| [`item-query-list-v2`](references/kuaimai-item-query-list-v2.md) | 商品档案列表 V2（`service item item-query-list-v2`） |

**CRITICAL — 写操作（`save`、`update-title`）执行前 MUST 先用 Read 工具读取对应 references 文档**

## meta 驱动能力速查

| 能力 | 规则 | Agent 动作 |
|------|------|------------|
| **contentType** | `get_query` → URL 参数；`post_form` → form；`post_json` → JSON body | `--body` 统一传 JSON，CLI 按 meta 自动转换 |
| **pageable** | 仅 `pageable:true` + `--page-all` 才全量翻页 | 翻页前先用 `count` 评估数据量；见[海量数据防护](#海量数据查询防护) |
| **write** | `write:true` 支持 `--dry-run`；查询接口不支持 | 写操作先 `--dry-run --verbose`，用户确认后再提交 |
| **Schema** | `requestSchema` 必填校验；`responseSchema` 解析结构 | 走 service 前用 `schema` 查字段；不要猜参数名 |

完整规则：[`kuaimai-item-meta-execution.md`](references/kuaimai-item-meta-execution.md)

## 海量数据查询防护

CLI 与 Agent 规则已对齐，四重防护：

| 层级 | 机制 |
|------|------|
| Agent 预判 | 翻页前 `count` 评估；未要求全量时不加 `--page-all` |
| CLI 阈值 | 已拉取 ≥500 条或接口 `total`>1000 时触发续查逻辑 |
| CLI 交互 | 交互终端 `[y/N]` 确认；非交互默认停止（stderr 提示） |
| 硬上限 | 最多 1000 页 + 可选 `--page-limit` 条数上限 |

**CLI 参数**：

```bash
# 交互式（默认）：达阈值时在终端询问是否继续
kuaimai-cli item +list --body '{...}' --page-all

# Agent/脚本：自动继续全量翻页
kuaimai-cli item +list --body '{...}' --page-all --page-confirm yes

# 限制最多 200 条
kuaimai-cli item +list --body '{...}' --page-all --page-limit 200

# 达阈值静默停止（不询问）
kuaimai-cli item +list --body '{...}' --page-all --page-confirm no
```

**Agent 规则**：

1. 默认不加 `--page-all`；用户明确要求全量时再加
2. 非交互执行（管道/Agent）达阈值会停止并返回部分数据 → 需继续时加 `--page-confirm yes` 或先征得用户同意
3. 大数据量优先 `--output ndjson` / `csv`

## 端到端：按标题改标题

| 步骤 | 操作 |
|------|------|
| 1 | `item +list` 按标题搜索 → 取 `sysItemId`（多条时让用户确认） |
| 2 | `item update-title --dry-run --verbose` 预览 |
| 3 | 用户确认后去掉 `--dry-run` 提交 |

## API Resources（兜底）

```bash
kuaimai-cli schema --output json                      # 全量 meta 自省
kuaimai-cli service item <operation> --body '{...}'   # 原子 API（见 references/kuaimai-item-service.md）
kuaimai-cli api POST /item/stock/queryList            # 原始 HTTP 兜底
```

> **重要**：`service item *` 为原子 API，无 `update-title` 的 get-detail 编排；优先用 shortcuts。调用前可用 `schema` 查看 operation 定义，不要猜测字段格式。

## 快速决策

- 用户问「标题带 XX 有多少个」（库存页）→ **`item count`** + `title`，禁止 `+list` 人工数
- 用户问「品牌/类目/档案条件下有多少个」→ **`service item item-query-count`**，禁止 `item count` 或仅靠 list 的 `total`
- 只有标题无 ID → 先 **`+list`**（库存页）或 **`service item item-query-list-v2`**（档案页）
- 只改标题 → **`update-title`**，不要瘦身 `save`
- **`item +list` 不支持 `--dry-run`**（查询接口）；`service item stock-list` 同样不支持
- **`--page-all` 仅对 `pageable:true` 生效**（`stock-list`、`item-query-list-v2`）
- 列表命中多条且下一步有副作用（改标题）→ 列候选让用户选，不要擅自选第一条
- 失败时优先转述 `hint`；hint 为空时读 `error`

## 典型场景

```bash
# 统计标题含「春季」的商品数（库存页）
kuaimai-cli item count --body '{"title":"春季"}' --output json --no-color

# 统计品牌为「洛可可」的商品档案总数
kuaimai-cli service item item-query-count \
  --body '{"brandNames":"洛可可","pageNo":1,"pageSize":1}' \
  --output json --no-color

# 搜索标题含 test 的商品（单页）
kuaimai-cli item +list --body '{"title":"test","pageNo":1,"pageSize":50}' --output json --no-color

# 商品档案 V2 列表（无 shortcut，走 service）
kuaimai-cli service item item-query-list-v2 \
  --body '{"title":"test","pageNo":1,"pageSize":50}' \
  --output json --no-color

# 改标题（先 dry-run）
kuaimai-cli item update-title --sys-item-id <id> --title "新名称" --dry-run --verbose --output json --no-color
```

## 不在本 skill 范围

- 登录、配置、输出格式、安装 CLI/Skill → [`kuaimai-shared`](../kuaimai-shared/SKILL.md)
- 原始 HTTP 探索 → `kuaimai-cli api` + `schema`
- meta 文件维护、代码实现 → 仓库 `docs/kuaimai-cli meta_data.json 定义规范.md`
