# web call — item 域原子 API

## 适用场景

- shortcuts **未覆盖**的 apiId（如 `item.item-query-list-v2`）
- 维护/联调：验证 registry 路由与后端一致
- 脚本化：已知完整 `--body` / `--data`

**不适用**：改标题（用 `item update-title`）、库存页按标题 count/list（用 shortcuts）。

## 基本流程

```bash
# 1. 发现与自省（禁止猜 apiId / 参数名）
kuaimai-cli capabilities --output json
kuaimai-cli schema item.item-query-list-v2 --output json

# 2. 调用
kuaimai-cli web call item.item-query-list-v2 \
  --body '{"title":"关键词","pageNo":1,"pageSize":50}' \
  --output json --no-color

# 3. 写操作先 dry-run
kuaimai-cli web call item.item-save \
  --body '{...}' --dry-run --verbose --output json --no-color
```

## 与 shortcuts 对照

| apiId | 优先 shortcut |
|-------|---------------|
| `item.stock-list` | `item +list` |
| `item.stock-count` | `item count` |
| `item.item-detail` | `item get-detail` |
| `item.item-save` | `item save` |
| `item.item-update-title` | `item update-title`（有编排） |
| `item.item-query-list-v2` | 无 shortcut |
| `item.item-query-count` | 无 shortcut |

## 禁止

- 不要用 `web call item.item-update-title` 代替 `item update-title`
- 不要对查询接口加 `--dry-run`
- 不要维护手写接口表；以 `schema` 为准

## 执行规则

contentType、pageable、write、翻页防护详见 [`kuaimai-item-meta-execution.md`](kuaimai-item-meta-execution.md)。
