# item count — 按标题统计商品数量

## 适用场景

用户问**有多少 / 几个 / 总数 / 统计**，且条件与**标题**相关时。

对应 meta operation：`stock-count`（`post_form`，`pageable:false`，`write:false`）。

## 前置检查

```bash
kuaimai-cli auth status --output json
```

## 命令

```bash
kuaimai-cli item count \
  --body '{"title":"<关键词>"}' \
  --output json --no-color
```

## 请求说明

- **contentType**：`post_form` — `--body` JSON 转为 form 表单
- **pageable**：`false` — `--page-all` 无效
- **write**：`false` — 不支持 `--dry-run`（走 shortcuts 时）；`service item stock-count` 对查询接口同样拒绝 dry-run

筛选字段与 `+list` 相同（`title`、`outerId` 等）；按标题统计时 body **必须**含 `title`。

## 响应解析

用中文向用户报告 **`data.data.total`**。

## 与 +list 的配合

- 用户要全量 list 前，**先用 count 评估数据量**，再决定是否 `--page-all`
- count 比拉全量 list 更轻量，优先用于「有多少个」类问题

## 等价 service 调用

```bash
kuaimai-cli service item stock-count \
  --body '{"title":"<关键词>"}' \
  --output json --no-color
```

## 禁止

- **禁止**用 `item +list` 拉列表再人工数条数
- `item count` **未带** `title` 时返回的是其它默认筛选下的总数，≠「标题含某词」的数量

## 失败处理

读 `error` 与 `hint`；必要时 `--verbose`。
