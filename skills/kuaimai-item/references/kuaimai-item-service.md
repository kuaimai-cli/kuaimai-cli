# service 层 — meta 驱动原子 API

## 适用场景

- Skill shortcuts **未覆盖**的 meta operation（如 `item-query-list-v2`）
- 维护/联调：验证 meta 路由与后端一致
- 脚本化调用：已知完整 `--body`，不需要 shortcut 默认字段

**不适用**：改标题（用 `item update-title`）、按标题 count/list（用 shortcuts）。

## 前置检查

```bash
kuaimai-cli auth status --output json
```

## 基本用法

```bash
# 1. 查 schema 确认 operation 存在及参数
kuaimai-cli schema --output json | jq '.data.operations[] | select(.operation=="item-query-list-v2")'

# 2. 执行（post_form 示例）
kuaimai-cli service item item-query-list-v2 \
  --body '{"title":"关键词","pageNo":1,"pageSize":50}' \
  --output json --no-color

# 3. 写操作必须先 dry-run
kuaimai-cli service item item-save \
  --body '{...全量 body...}' \
  --dry-run --verbose --output json --no-color
```

## item 域核心 operation 速查

| operation | method | contentType | write | pageable | 等价 shortcut |
|-----------|--------|-------------|-------|----------|---------------|
| `stock-list` | POST | post_form | false | true | `item +list` |
| `stock-count` | POST | post_form | false | false | `item count` |
| `item-query-list-v2` | POST | post_form | false | true | 无 |
| `item-detail` | GET | get_query | false | false | `item get-detail` |
| `item-save` | POST | post_json | true | false | `item save` |
| `item-update-title` | POST | post_json | true | false | `item update-title`（无编排） |

> operation 名为 meta id（如 `stock-list`），**不是** shortcut 子命令名（如 `list`）。

## 与 shortcuts 的行为差异

| 能力 | shortcuts | service |
|------|-----------|---------|
| 默认 ARCHIVE_V2 body | `item +list` ✅ | `stock-list` ❌ 需自传或依赖 schema default |
| `--dry-run` 查询接口 | ❌ | ❌（write=false 时报错） |
| `--dry-run` 写接口 | ✅ | ✅ |
| `--page-all` | ✅（pageable 接口） | ✅（pageable 接口） |
| required 校验 | 部分 | ✅ requestSchema.required |

## --page-all 用法

仅 `pageable:true` 的 operation 生效：

```bash
kuaimai-cli service item stock-list \
  --body '{"title":"test","pageNo":1,"pageSize":50}' \
  --page-all --output json --no-color
```

**Agent 规则**：全量翻页前读 [`kuaimai-item-meta-execution.md`](kuaimai-item-meta-execution.md) §海量数据防护；先 count、超阈值问用户。

## get_query 类型

```bash
kuaimai-cli service item item-detail \
  --body '{"sysItemId":123456789}' \
  --output json --no-color
```

CLI 将 body 转为 URL query：`/item/getItemDetail?sysItemId=123456789`

## 禁止

- 不要用 `service item item-update-title` 代替 `item update-title`（缺少 get-detail 编排）
- 不要对查询接口加 `--dry-run`（CLI 会拒绝）
- 不要猜测 operation 名；以 `schema` 输出为准

## 失败处理

- `缺少必填参数 xxx（见 schema xxx）` → 补全 `--body` 后重试
- `操作 xxx 为查询接口，不支持 --dry-run` → 去掉 `--dry-run`
- `pageable=false，已忽略 --page-all` → 去掉 `--page-all` 或换 pageable 接口
- 其它：读 `error` / `hint`，加 `--verbose`
