# item-query-count — 商品档案总数

## 适用场景

- 用户在**商品档案**维度问「有多少 / 几个 / 总数 / 统计」
- 筛选含档案字段：`brandNames`、`catIds`、`itemType`、`outerId`、`cIds`、`userIds` 等
- **无 shortcut**；必须通过 `web call item.item-query-count`

与 [`kuaimai-item-count.md`](kuaimai-item-count.md)（`stock-count` / `item count`）的区别见 [`kuaimai-item-count-dimensions.md`](kuaimai-item-count-dimensions.md)。

## 前置检查

```bash
kuaimai-cli auth status --output json
```

## 命令

```bash
kuaimai-cli web call item.item-query-count \
  --body '{"brandNames":"洛可可","pageNo":1,"pageSize":1}' \
  --output json --no-color
```

按标题统计（档案口径）：

```bash
kuaimai-cli web call item.item-query-count \
  --body '{"title":"<关键词>","pageNo":1,"pageSize":1}' \
  --output json --no-color
```

## 请求说明

- **contentType**：`post_form`（`--body` JSON → form-urlencoded）
- **write**：`false`（不支持 `--dry-run`）
- **pageable**：`true`（meta 标记；统计场景仍须传 `pageNo`、`pageSize`，二者为 **required**）
- **path**：`/item/queryCount`

### 常用筛选字段

与 `item-query-list-v2` 一致，例如：

| 字段 | 说明 |
|------|------|
| `title` | 商品名称 |
| `brandNames` | 品牌名称，逗号分隔 |
| `catIds` / `cIds` | 类目 / 分类 ID，逗号分隔 |
| `outerId` / `skuOuterId` | 主商家编码 / 规格商家编码 |
| `itemType` | 商品类型 |
| `pageNo` / `pageSize` | **必填** |

完整字段见 `kuaimai-cli schema --output json` 中 `item-query-count` 的 `requestSchema`。

## 响应解析

- 成功：`ok === true`
- 向用户报告 **`data.data.total`**（商品档案符合条件的总数）

## 与 list-v2 的配合

| 需求 | 接口 |
|------|------|
| 只要总数 | **本接口** `item-query-count` |
| 要列表行 / `sysItemId` | `item-query-list-v2` |
| 全量 list 前先评估规模 | 先用本接口 count，再决定是否 `--page-all` list |

## 禁止

- 不要用 `item count`（`stock-count`）代替本接口做档案品牌/类目统计
- 不要仅为 total 调用 `item-query-list-v2`（应用本接口）
- 不要加 `--dry-run`（查询接口）

## 失败处理

读 `error` 与 `hint`；参数不确定时查 `schema`；必要时 `--verbose`。
