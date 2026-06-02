# Agent 命令选型与 schema 流程

> 说明 AI Agent 如何将用户自然语言路由到 **shortcuts**（`item …`）或 **service**（`service item …`），以及 **何时查询 `schema`**。  
> 配套：[meta_data.json 定义规范](./kuaimai-cli%20meta_data.json%20定义规范.md) · [系统架构与飞书对标说明](./系统架构与飞书对标说明.md) · [AGENTS.md](../AGENTS.md) · [kuaimai-item/SKILL.md](../skills/kuaimai-item/SKILL.md)

---

## 1. 三级命令回顾

| 层级 | 命令示例 | 实现 | Agent 默认 |
|------|----------|------|------------|
| **1. shortcuts** | `item +list`、`item update-title` | `shortcuts/item/` | ✅ 优先 |
| **2. service** | `service item stock-list` | `meta_data.json` + `cmd/servicecmd` | 兜底 |
| **3. api** | `api POST /item/stock/queryList` | `cmd/api` | 最后手段 |

飞书对标：`calendar +agenda` → `item +list` · `calendar events list` → `service item stock-list` · 原生 path → `api POST …`

**Agent 口诀**：

```text
有 Shortcut → 不查 schema，读 Skill / references
走 service / api → 先 schema 全量（或规划中的 schema item.<operation>），再执行
不知道有哪些接口 → schema 全量
```

---

## 2. 总览：自然语言 → 命令层级

```mermaid
flowchart TD
  NL[用户自然语言] --> RouteSkill{匹配哪个 Skill?}

  RouteSkill -->|配置/登录/输出/安装| Shared[kuaimai-shared]
  RouteSkill -->|商品/标题/SKU/列表/详情/改名| ItemSkill[kuaimai-item/SKILL.md]

  Shared --> SharedCmd[config / auth / skill install / doctor …]
  ItemSkill --> ReadShared[CRITICAL: Read kuaimai-shared]

  ReadShared --> Intent[解析意图: 选哪个命令表]
  Intent --> HasShortcut{Skill 有明确 Shortcut?}

  HasShortcut -->|是| WriteOp{写操作?}
  HasShortcut -->|否/不确定| NeedDiscover[需要发现接口能力]

  WriteOp -->|是| ReadRef[CRITICAL: Read references/]
  WriteOp -->|否| Shortcut[item 子命令 shortcuts 层]

  ReadRef --> Shortcut

  NeedDiscover --> SchemaFull[schema 全量]
  SchemaFull --> PickOp[选定 item.operation]
  PickOp --> SchemaOne[schema item.operation 或全量 schema]

  SchemaOne --> Prefer{schema.prefer / 编排?}
  Prefer -->|prefer shortcut| Shortcut
  Prefer -->|prefer service / 无 shortcut| Service[service item operation]
  Prefer -->|有 orchestration| Shortcut

  Shortcut --> Exec1["kuaimai-cli item +list / count / update-title …"]
  Service --> Exec2["kuaimai-cli service item stock-list --body …"]
  NeedDiscover --> ApiFallback["api POST /item/… 最后兜底"]
  ApiFallback --> SchemaOne
```

---

## 3. 商品域：Skill 意图 → shortcut / service

```mermaid
flowchart TD
  U[用户说: 商品相关自然语言] --> Examples

  subgraph Examples [意图示例]
    E1[有多少标题带XX的商品]
    E2[搜标题含 test 的商品]
    E3[查 sysItemId 详情]
    E4[把标题改成新名称]
    E5[调用某个未文档化接口]
  end

  Examples --> Table{kuaimai-item 选哪个命令表}

  Table -->|统计+标题| C1[item count]
  Table -->|列表/搜索+标题（库存）| C2[item +list]
  Table -->|列表/搜索（档案 V2）| C2b[service item-query-list-v2]
  Table -->|已知 ID 查详情| C3[item get-detail]
  Table -->|改标题| C4[item update-title]
  Table -->|无匹配| Unknown[未覆盖能力]

  C1 --> S1[shortcut]
  C2 --> S2[shortcut]
  C3 --> S3[shortcut]
  C4 --> S4[shortcut + references]

  Unknown --> Q1{为什么要离开 shortcut?}

  Q1 -->|只想知道有哪些 op| SF[查 schema 全量]
  Q1 -->|要用 service 原子 API| SO[查 schema item.op]
  Q1 -->|手写 path 探索| SO

  SF --> SO
  SO --> Decide{schema 结果}

  Decide -->|prefer shortcut| Back[回到 item shortcut]
  Decide -->|prefer service / 纯原子| SV[service item op]
  Decide -->|orchestration 仅 shortcut| Back
  Decide -->|仍不确定 path| API[api POST/GET …]
```

**商品域默认**：「选哪个命令」表能命中的，**一律走 shortcuts**；仅表外或刻意走原子 API 时用 service / api。

---

## 4. shortcut vs service 决策

```mermaid
flowchart LR
  subgraph always [永远优先 shortcut]
    A1[Skill 选命令表有对应项]
    A2[有 + 别名 如 +list]
    A3[多步编排 如 update-title]
    A4[需要友好 flag 如 --sys-item-id]
    A5[需要默认 body ARCHIVE_V2]
  end

  subgraph serviceCase [才考虑 service]
    B1[Skill 未覆盖的 operation 如 item-query-list-v2]
    B2[脚本化 已知完整 --body]
    B3[维护/联调 验证 meta 路由]
  end

  subgraph never [不要用 service 冒充 shortcut]
    C1[update-title 改标题]
    C2[get-detail 带 sysItemId]
  end

  always --> ItemCmd[item …]
  serviceCase --> ServiceCmd[service item …]
  never --> ItemCmd
```

### 商品域对照表

| 用户意图 | 选（shortcuts） | 不选 |
|----------|-----------------|------|
| 搜/列商品（库存页） | `item +list` | `service item stock-list` |
| 搜/列商品（档案 V2） | `service item item-query-list-v2` | `item +list`（path 不同） |
| 统计数量（库存页 / 标题） | `item count` | `item +list` 再人工数 |
| 统计数量（商品档案 / 品牌类目等） | `service item item-query-count` | `item count`（库存口径）、`item-query-list-v2` 仅取 total |
| 查详情 | `item get-detail --sys-item-id` | `service item item-detail --body '{"sysItemId":…}'` |
| 改标题 | `item update-title` | `service item item-update-title`（无 get-detail 编排） |
| 复杂改字段 | `item save` + references | `service item item-save` 须全量 body |

### shortcuts 与 service 行为差异（同接口）

| 能力 | `item +list` | `service item stock-list` |
|------|--------------|---------------------------|
| 默认 body（ARCHIVE_V2） | ✅ | ❌ 需自传或 schema default |
| `--dry-run` | ❌ 查询接口 | ❌ 查询接口（`write:false` 拒绝） |
| `--page-all` | ✅ | ✅ |
| `--page-limit` / `--page-confirm` | ✅ | ✅ |

详见 [系统架构与飞书对标说明 · §5](./系统架构与飞书对标说明.md#5-dry-run-与-page-all按代码)。

---

## 5. 什么时候查询 schema

```mermaid
flowchart TD
  Start[准备执行命令前] --> Q0{已有 Skill + references 明确指令?}

  Q0 -->|是 且走 shortcut| NoSchema["不查 schema，直接执行"]
  Q0 -->|否| Q1{当前目标是什么?}

  Q1 -->|不知道有哪些接口| SFull["schema 全量<br/>kuaimai-cli schema"]
  Q1 -->|确定走 service item op| SOne["schema 单 op<br/>kuaimai-cli schema item.op"]
  Q1 -->|确定走 api 原始 path| SOne
  Q1 -->|维护者验收 meta| SOne
  Q1 -->|新增 op 写 Skill references| SOne

  SFull --> SOne
  SOne --> Check{看 schema 返回}

  Check -->|prefer shortcut| GoShortcut[改走 item shortcut]
  Check -->|prefer service| GoService["service item op --body"]
  Check -->|orchestration| GoShortcut
  Check -->|只有 path 无 params| GoHelp[补查 --help 或 DevTools / shortcuts 源码]
```

### 查 / 不查 对照表

| 时机 | 查不查 | 查什么 |
|------|--------|--------|
| Skill 表 + references 已覆盖 | ❌ 不查 | 直接 `item …` |
| 写操作 save / update-title | ❌ 不查 schema | ✅ Read `references/` |
| 不确定有没有这个 API | ✅ 查 | `schema` 全量 |
| 要用 `service item *` | ✅ 必须先查 | 全量 `schema`（或规划中的 `schema item.stock-list` 等） |
| 要用 `api POST /item/…` | ✅ 必须先查 | 同上 |
| 新增/改 meta 后验收 | ✅ 查 | `schema` / 单 op |
| 飞书对标：原生 API 前先看参数 | ✅ 查 | 单 op schema（P2 待实现参数字段） |

### 不必查 schema 的场景

- 按标题 list / count
- 有 `sysItemId` 的 get-detail
- update-title 改标题
- save 且已 Read references
- config / auth / skill 等 shared 域

---

## 6. 端到端示例：按标题改标题

```mermaid
sequenceDiagram
  participant U as 用户
  participant A as Agent
  participant SK as kuaimai-item Skill
  participant RF as references
  participant CLI as kuaimai-cli

  U->>A: 把标题带 test 的商品改成新名称
  A->>SK: Read SKILL.md
  SK-->>A: 只有标题无 ID → 先 +list
  Note over A: 不查 schema（Skill 已明确）

  A->>CLI: item +list --body title:test
  CLI-->>A: sysItemId 列表
  A->>U: 多条则让用户选

  A->>RF: Read update-title reference
  A->>CLI: item update-title --dry-run …
  CLI-->>A: 预览 save body
  A->>U: 确认后去掉 --dry-run 再执行

  Note over A,CLI: 全程 shortcut，不经过 service/schema
```

若用户明确要求「用 service 层 POST save」：

```text
schema（全量）→ service item item-save --body '…'（须全量 body）
```

---

## 7. schema 现状与规划

**meta 与 schema 关系**（见[定义规范 §三](./kuaimai-cli%20meta_data.json%20定义规范.md#三schema-结构定义--核心逻辑答疑)）：`meta_data.json` 是存储层；`schema` 命令是展示/自省层，从 meta 读取结构。

### 当前（已实现）

```bash
kuaimai-cli schema --output json
```

输出全部 operation（当前 **1157** 个）的**路由元信息**：service、operation（如 `stock-list`）、method、path、contentType、pageable、write、requestSchema/responseSchema、shortcut 映射等。

`jq` 示例：

```bash
kuaimai-cli schema --output json | jq '.data.operations[] | select(.operation=="stock-list")'
kuaimai-cli schema --output json | jq '[.data.operations[] | select(.pageable==true)] | length'
```

### 规划（P2，对标飞书）

```bash
kuaimai-cli schema item.stock-list --output json   # 单 operation
kuaimai-cli schema item.item-save --output json   # 含 requestSchema 字段说明
```

单 op schema 预期额外返回：`shortcut`、`prefer`、`orchestration`、`notes` 等，供 Agent 在走 service/api 前确认参数形态。

在单 op schema 落地前：**走 service/api 时先用全量 `schema` 确认 operation 存在**（operation 名为 `stock-list` 等，**非** shortcuts 子命令名 `list`），参数细节以 Skill `references/` 与 `item --help` 为准。

---

## 8. 维护者：新增接口时的同步清单

1. 按 [meta_data.json 定义规范](./kuaimai-cli%20meta_data.json%20定义规范.md) 在 `internal/registry/meta_data.json` 登记 operation（`模块-操作` 命名，填 path/method/contentType/write/pageable/Schema）  
2. `shortcuts/item/` 若 Agent 高频则加封装  
3. `skills/kuaimai-item/SKILL.md` + `references/`  
4. 验收：`schema` 能列出、`service item <operation>` 可调，path 与 shortcuts 一致

---

## 9. 相关文档

| 文档 | 说明 |
|------|------|
| [系统架构与飞书对标说明.md](./系统架构与飞书对标说明.md) | §3 三级命令、§5 dry-run、§6 Skill |
| [kuaimai-item/SKILL.md](../skills/kuaimai-item/SKILL.md) | Agent 商品域路由与 references |
| [AGENTS.md](../AGENTS.md) | 仓库根 Agent 约定 |
| [meta_data.json 定义规范.md](./kuaimai-cli%20meta_data.json%20定义规范.md) | operation 命名、contentType、write/pageable、Schema |
| [kuaimai-cli 开发文档.md](./kuaimai-cli%20开发文档.md) | §4.5 元数据注册表、新增 shortcuts 步骤 |
