---
name: kuaimai-item
version: 3.0.0
description: "快麦 ERP 商品（erp-items-core）：按标题搜索列表、统计数量、查详情、改标题。用户提到商品、SKU、标题、货号、列表、有多少、详情、改名时使用。"
metadata:
  requires:
    bins: ["kuaimai-cli"]
  cliHelp: "kuaimai-cli item --help"
---

# item（erp-items-core）

**CRITICAL — 开始前 MUST Read [`../kuaimai-shared/SKILL.md`](../kuaimai-shared/SKILL.md)**（auth、registry 发现、输出与安全）

本 Skill 只管 **商品域意图路由与 shortcuts 工作流**。接口参数以 `schema <apiId>` 为准，不在此重复维护接口表。

## 意图路由

**库存页 vs 商品档案**：必读 [`references/kuaimai-item-count-dimensions.md`](references/kuaimai-item-count-dimensions.md)。

| 用户意图 | 优先命令 | 说明 |
|----------|----------|------|
| 有多少 + **标题**（库存页） | `item count` | 禁止用 `+list` 人工数 |
| 有多少 + **品牌/类目/档案** | `web call item.item-query-count` | 禁止 `item count` |
| 列出 / 搜索 + **标题**（库存页） | `item +list` | 有 ARCHIVE_V2 默认 body |
| 列出 / 搜索（档案 V2） | `web call item.item-query-list-v2` | 无 shortcut |
| 查详情（有 sysItemId） | `item get-detail` | |
| 改标题 | `item update-title` | 先 `--dry-run`；禁止用原子 save 代替 |
| 其它 registry 接口 | `schema <apiId>` → `web call` | 见 shared Registry 流程 |

只有标题没有 ID：先 `+list` 或档案列表取 `sysItemId`，再 `get-detail` / `update-title`。

## Shortcuts（永远优先）

| Shortcut | 说明 |
|----------|------|
| [`+list`](references/kuaimai-item-list.md) | 库存列表（`title` 筛选、`--page-all`） |
| [`count`](references/kuaimai-item-count.md) | 按标题统计（库存页） |
| [`get-detail`](references/kuaimai-item-get-detail.md) | 按 `sysItemId` 查详情 |
| [`update-title`](references/kuaimai-item-update-title.md) | 改标题（get-detail → save 编排） |
| [`save`](references/kuaimai-item-save.md) | 全量 body 保存（复杂字段修改） |

**无 shortcut 的常用 web call**（参数见 `schema`）：

| apiId | 说明 | reference |
|-------|------|-----------|
| `item.item-query-count` | 商品档案统计 | [`kuaimai-item-query-count.md`](references/kuaimai-item-query-count.md) |
| `item.item-query-list-v2` | 商品档案列表 V2 | [`kuaimai-item-query-list-v2.md`](references/kuaimai-item-query-list-v2.md) |

**CRITICAL — 写操作（`save`、`update-title`）前 MUST Read 对应 references**

## 前置条件

| 场景 | 要求 |
|------|------|
| 改标题 / save | Read 对应 reference；先 `--dry-run --verbose` |
| 翻页全量 | 先 `count` 评估；见 [`kuaimai-item-meta-execution.md`](references/kuaimai-item-meta-execution.md) |
| 调用未知 apiId | `kuaimai-cli schema <apiId>`，不要猜字段 |

## 端到端：按标题改标题

| 步骤 | 操作 |
|------|------|
| 1 | `item +list` 按标题搜索 → 取 `sysItemId`（多条让用户选） |
| 2 | `item update-title --dry-run --verbose` 预览 |
| 3 | 用户确认后去掉 `--dry-run` |

## web call 与 shortcuts 差异

| 能力 | shortcuts | `web call item.*` |
|------|-----------|-------------------|
| 默认 ARCHIVE_V2 body | `+list` ✅ | 需自传或看 schema default |
| `update-title` 编排 | ✅ get-detail → save | ❌ 仅原子 save |
| `--dry-run` 查询 | ❌ | ❌ |
| `--dry-run` 写操作 | ✅ | ✅ |

详见 [`kuaimai-item-web-call.md`](references/kuaimai-item-web-call.md)。

## 快速决策

- 库存页标题统计 → **`item count`**，不是 `+list`
- 档案维度统计 → **`web call item.item-query-count`**
- 只改标题 → **`update-title`**，不是 `save` / `web call item.item-update-title`
- 列表多条且下一步有副作用 → 列候选让用户选
- 失败优先转述 `hint`

## 典型场景

```bash
kuaimai-cli item count --body '{"title":"春季"}' --output json --no-color

kuaimai-cli item +list --body '{"title":"test","pageNo":1,"pageSize":50}' --output json --no-color

kuaimai-cli web call item.item-query-count \
  --body '{"brandNames":"洛可可","pageNo":1,"pageSize":1}' --output json --no-color

kuaimai-cli item update-title --sys-item-id <id> --title "新名称" \
  --dry-run --verbose --output json --no-color
```

## 不在本 skill 范围

- 配置、登录、registry 同步 → [`kuaimai-shared`](../kuaimai-shared/SKILL.md)
- 供应链 / 铺货 / scm 商品 → [`kuaimai-scm`](../kuaimai-scm/SKILL.md)
- 发现其它 apiId → `capabilities` / `schema`（shared）
