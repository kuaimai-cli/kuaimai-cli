# service scm — meta 驱动原子 API

## 适用场景

- scm 域全部接口（暂无 shortcuts）
- 维护/联调：验证 meta 与后端一致

## 前置检查

```bash
kuaimai-cli auth status --output json
```

## 基本用法

```bash
# 1. 查 schema
kuaimai-cli schema --output json | jq '.data.operations[] | select(.service=="scm" and .operation=="item-base-page")'

# 2. 执行
kuaimai-cli service scm item-base-page \
  --body '{"pageNo":1,"pageSize":50}' \
  --output json --no-color

# 3. 写操作 dry-run
kuaimai-cli service scm item-base-edit \
  --body '{...}' \
  --dry-run --verbose --output json --no-color
```

## operation 命名

- 由 path 转 kebab-case：`/logging/publishLog` → `logging-publish-log`
- **不是** Java 方法名

## 与 item shortcuts 差异

| 能力 | item shortcuts | service scm |
|------|----------------|-------------|
| 默认 body | shortcut 内置 | 需 `--body` 或 schema default |
| 域名 | erp1 | scm（自动） |
| `--page-all` | ✅ pageable | ✅ pageable |

## 分页

```bash
kuaimai-cli service scm item-base-page \
  --body '{"pageNo":1,"pageSize":50}' \
  --page-all --page-confirm yes --output ndjson
```
