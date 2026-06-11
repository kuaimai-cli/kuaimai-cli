# registry 执行规则（item 域）

> Agent 通过 `schema <apiId>` 获取契约，通过 `web call` 或 shortcuts 执行。本文说明 CLI 行为约束。

## 职责分层

```text
远端 registry.json  →  capabilities / schema（发现）  →  web call / shortcuts（执行）
                              ↑
                         Skill 指导选路（本目录）
```

## contentType

Agent 对 `web call` 可用 `--body`（自动按类型转 params/data）或显式 `--params` / `--data`。

| contentType | 行为 | 示例 apiId |
|-------------|------|------------|
| `get_query` | URL query | `item.item-detail` |
| `post_form` | form 编码 | `item.stock-list` |
| `post_json` | JSON body | `item.item-save` |

## pageable 与 --page-all

| pageable | `--page-all` |
|----------|--------------|
| `true` | 循环 `pageNo` 合并各页 |
| `false` | 忽略 `--page-all`（stderr 提示） |

**Agent 规则**：

1. 默认不加 `--page-all`
2. 翻页前用 `count` 或 schema 评估数据量
3. 非交互续查：`--page-confirm yes`；限条数：`--page-limit N`
4. 硬上限 1000 页

```bash
kuaimai-cli item +list --body '{...}' --page-all --page-confirm yes --page-limit 200
```

## write 与 --dry-run

| write | `--dry-run` |
|-------|-------------|
| `true` | ✅ 预览 URL/body |
| `false` | ❌ CLI 拒绝 |

查询 shortcuts（`item +list`）与查询 `web call` 均不支持 `--dry-run`。

## Schema 校验

`web call` 会校验 `requestSchema.required`。缺参时读 `schema <apiId>` 补全，不要猜字段名。

## shortcuts 专属默认 body

`item +list` 内置 ARCHIVE_V2 相关默认字段；`web call item.stock-list` 无此默认，若后端报错参考 shortcut 的 `--help` 默认 body。

## 海量数据防护（与 shared 对齐）

| 层级 | 机制 |
|------|------|
| Agent | 翻页前先 count；未要求全量不加 `--page-all` |
| CLI | ≥500 条或 total>1000 触发 `--page-confirm` |
| 硬上限 | 1000 页 + `--page-limit` |
