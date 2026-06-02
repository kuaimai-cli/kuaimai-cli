# meta 驱动执行规则（Skill 能力层核心）

> 本文档说明 Agent 如何理解 `meta_data.json` 契约，并正确调用 CLI。meta 是静态注册表，**执行逻辑由 CLI + 本 Skill 共同约束**。

## 1. 三层职责

```text
meta_data.json（契约）  →  schema 命令（自省）  →  service / shortcuts（执行）
                              ↑
                         Skill 指导 Agent 选路
```

- **meta**：定义 path / method / contentType / write / pageable / requestSchema / responseSchema
- **schema**：`kuaimai-cli schema --output json` 格式化输出 meta，供 Agent 发现接口
- **shortcuts**：高频操作的友好封装（默认 body、编排、flag）
- **service**：按 meta 零代码驱动，`service item <operation> --body '{...}'`

## 2. contentType 三种请求类型

Agent 统一用 `--body '{"key":"value"}'` 传 JSON；CLI 按 meta 的 `contentType` 自动组装请求。

| contentType | HTTP 行为 | 典型接口 | Agent 注意 |
|-------------|-----------|----------|------------|
| `get_query` | 参数拼接到 URL query | `item-detail` | shortcuts 用 `--sys-item-id` flag；service 用 `--body '{"sysItemId":123}'` |
| `post_form` | `application/x-www-form-urlencoded` | `stock-list`、`stock-count`、`item-query-list-v2` | `--body` JSON 会转为 form 字段；布尔/数字按字符串提交 |
| `post_json` | `application/json` 请求体 | `item-save`、`item-update-title` | body 须为完整 JSON 对象；save 禁止瘦身 |

### 示例对照

```bash
# get_query — item-detail
kuaimai-cli service item item-detail --body '{"sysItemId":123456}' --output json

# post_form — stock-list
kuaimai-cli service item stock-list \
  --body '{"title":"test","pageNo":1,"pageSize":50}' --output json

# post_json — item-save（写操作，先 dry-run）
kuaimai-cli service item item-save --body '{...全量 SysItemModel...}' \
  --dry-run --verbose --output json
```

## 3. pageable 与 --page-all

| pageable | `--page-all` | 行为 |
|----------|--------------|------|
| `true` | 开启 | CLI 循环递增 `pageNo`，合并各页 `data`，直到无更多数据或达硬上限 |
| `true` | 未开启 | 只请求 body 中指定的单页 |
| `false` | 任意 | **忽略** `--page-all`，stderr 提示「pageable=false，已忽略」 |

**当前 pageable:true 的 item 域接口**：

| operation | path |
|-----------|------|
| `stock-list` | `/item/stock/queryList` |
| `item-query-list-v2` | `/item/queryListV2` |

**翻页参数**：

- 识别 body 中 `pageNo`（默认 1）、`pageSize`（默认 50，来自 requestSchema）
- 终止条件：当页返回 0 条，或 `pageNo * pageSize >= total`，或达到 CLI 硬上限（1000 页），或 `--page-limit` 条数上限

**海量数据防护（CLI + Agent 双层，已对齐）**：

| 阈值 | CLI 行为 |
|------|----------|
| 已拉取 ≥ **500** 条 | 下一页前触发续查逻辑 |
| 接口 `total` > **1000** | 下一页前触发续查逻辑 |
| `--page-confirm prompt`（默认） | 交互终端 `[y/N]`；非交互环境停止并 stderr 提示 |
| `--page-confirm yes` | 跳过询问，继续翻页直至结束/硬上限 |
| `--page-confirm no` | 达阈值静默停止，返回已拉取数据 |
| `--page-limit N` | 硬限制最大条数 |

CLI 内部分片合并（每 500 条一块）以降低内存峰值。Agent 仍应：翻页前 count 评估；非交互续查用 `--page-confirm yes`。

## 4. write 与 --dry-run

| write | 接口类型 | `--dry-run` | Agent 规则 |
|-------|----------|-------------|------------|
| `false` | 查询 | ❌ 不支持（CLI 报错） | 直接执行；`item +list` 走非 dry-run 管线 |
| `true` | 写操作 | ✅ 预览 URL/body（脱敏） | **首次必须** `--dry-run --verbose`；用户确认后去掉 `--dry-run` |

**写操作清单**（item 域核心）：

- `item-save` / `item save`
- `item-update-title` / `item update-title`（shortcuts 含 get-detail 编排）

**查询接口**（不可 dry-run）：

- `stock-list`、`stock-count`、`item-detail`、`item-query-list-v2`

## 5. Schema 能力

### requestSchema

- **required 字段**：service 层会轻校验；缺必填项 CLI 报错并提示见 schema
- **default 值**：`service` 命令 `--body` 默认值来自 schema 的 `default`（如 `pageNo:1`、`pageSize:50`）
- **字段 desc**：Agent 向用户解释参数含义时参考 schema 输出

### responseSchema

- 成功时 `ok === true`，业务数据在 `data`
- 列表接口：`data` 为数组，或嵌套在 `data.list` / `data.records` 等（CLI 会 NormalizeList）
- 统计接口：`stock-count` 总数在 `data.data.total`

### 何时查 schema

| 场景 | 是否查 |
|------|--------|
| Skill 表 + references 已覆盖的 shortcut | ❌ 不查 |
| 写操作 save / update-title | ❌ 不查 schema；✅ Read references |
| 走 `service item <operation>` | ✅ 必须先查 |
| 不确定参数名/类型 | ✅ 查 `schema --output json`，过滤 `operation` 字段 |

## 6. 业务特殊参数

### 多值逗号分隔字符串

以下字段在 schema 中标注「多个使用英文逗号隔开」，传 **单个字符串** 而非数组：

```json
{"catIds": "1,2,3", "userIds": "100,200", "brandNames": "品牌A,品牌B"}
```

### JSON 字符串参数透传

部分 form 接口含前端冗余字段，值为 JSON 字符串，须原样传递：

```json
{"searchItems": "[]", "api_name": "item_stock_queryList"}
```

`item +list` shortcut 的默认 body 已包含 ARCHIVE_V2 页所需字段；用户只改筛选条件时，保留其余默认字段。

### 前端冗余参数兼容

库存列表（`stock-list`）与浏览器 ARCHIVE_V2 页对齐，含 `pageType`、`subPageType`、`suiteSearchType` 等。shortcut `item +list` 已内置默认值；走 service 时若后端报错，参考 shortcut 默认 body 补全。

## 7. shortcuts vs service 选型

| 维度 | shortcuts | service |
|------|-----------|---------|
| 默认 body | ✅ ARCHIVE_V2 等 | ❌ 仅 schema default |
| 友好 flag | ✅ `--sys-item-id` | ❌ 仅 `--body` |
| 多步编排 | ✅ update-title | ❌ 原子单接口 |
| dry-run（查询） | ❌ | ❌ |
| dry-run（写） | ✅ | ✅ |
| meta Schema 校验 | 部分 | ✅ required 校验 |
| Agent 默认 | **优先** | 无 shortcut 时 |

## 8. 执行链路（Agent 视角）

```text
用户意图
  → Read kuaimai-item/SKILL.md 选命令表
  → 写操作？Read references/
  → 无 shortcut？schema → service item <op>
  → 组装 --body（JSON）
  → pageable 且要全量？评估数据量 → 用户确认 → --page-all
  → write？--dry-run 预览 → 用户确认 → 正式提交
  → 解析 {ok, data, error, hint}
```
