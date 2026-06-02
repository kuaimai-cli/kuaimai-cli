# item-query-list-v2 — 商品档案列表 V2

## 适用场景

- 用户要在**商品档案**维度搜索/列出商品（非库存 ARCHIVE_V2 页）
- 需要 V2 独有筛选：`outerId`、`skuOuterId`、`itemType`、`catIds`、`brandNames` 等
- **无 shortcut**；必须通过 `service item item-query-list-v2`

与 `item +list`（`stock-list` / 库存页）的区别：

| 维度 | `item +list` | `item-query-list-v2` |
|------|--------------|------------------------|
| meta operation | `stock-list` | `item-query-list-v2` |
| path | `/item/stock/queryList` | `/item/queryListV2` |
| 页面场景 | 库存 ARCHIVE_V2 | 商品档案 V2 |
| shortcut | ✅ `item +list` | ❌ 仅 service |

## 前置检查

```bash
kuaimai-cli auth status --output json
```

## 命令

```bash
kuaimai-cli service item item-query-list-v2 \
  --body '{"title":"<关键词>","pageNo":1,"pageSize":50}' \
  --output json --no-color
```

全量翻页（**先评估数据量，见下方防护规则**）：

```bash
kuaimai-cli service item item-query-list-v2 \
  --body '{"title":"<关键词>","pageNo":1,"pageSize":50}' \
  --page-all --output json --no-color
```

## 请求说明

- **contentType**：`post_form`（`--body` JSON → form-urlencoded）
- **write**：`false`（不支持 `--dry-run`）
- **pageable**：`true`（支持 `--page-all`）

### 常用筛选字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `title` | string | 商品名称 |
| `outerId` | string | 主商家编码 |
| `skuOuterId` | string | 规格商家编码 |
| `pageNo` | number | 页码，默认 1 |
| `pageSize` | number | 每页条数，默认 50 |
| `itemType` | number | 商品类型（0 全部，1 普通，3 套件…） |
| `catIds` | string | 类目 ID，**逗号分隔** `"1,2,3"` |
| `cIds` | string | 分类 ID，逗号分隔 |
| `userIds` | string | 店铺 ID，逗号分隔 |
| `brandNames` | string | 品牌名称，逗号分隔 |

完整字段见 `kuaimai-cli schema --output json` 中 `item-query-list-v2` 的 `requestSchema`。

## 响应解析

- 成功：`ok === true`，列表在 `data`（CLI NormalizeList 后）
- 从记录取 `sysItemId` 供 `get-detail` / `update-title`

## 海量数据防护

本接口 `pageable:true`，`--page-all` 可能拉取大量数据。Agent **必须**：

1. 用户未明确要求「全部」时，只用单页（`pageNo:1, pageSize:50`）
2. 需要全量时，先用相同筛选条件的 **`item-query-count`** 评估规模（档案维度）；库存页可用 `item count`
3. 已拉取 > 500 条或预估 > 1000 条时，**询问用户**是否继续
4. CLI 硬上限 1000 页；达到后返回已拉取部分并告知用户

## 选型提示

- 用户只说「搜标题含 XX 的商品」且未指定档案/库存页 → 默认 **`item +list`**（shortcut 更简单）
- 用户提到档案、商家编码、`itemType`、类目筛选 → 用 **本接口**
- 统计数量（档案维度）→ **`service item item-query-count`**，见 [`kuaimai-item-query-count.md`](kuaimai-item-query-count.md) 与 [`kuaimai-item-count-dimensions.md`](kuaimai-item-count-dimensions.md)
- 统计数量（仅标题、库存页）→ `item count`

## 禁止

- 不要用 `item +list` 代替本接口（path 不同）
- 不要加 `--dry-run`（查询接口）
- 不要手写 URL 或 curl

## 失败处理

读 `error` 与 `hint`；参数不确定时查 `schema`；必要时 `--verbose`。
