# item get-detail — 查询商品详情

## 适用场景

用户已知 **sysItemId**，或经 `+list` 定位到 ID 后要查看详情。

对应 meta operation：`item-detail`（`get_query`，`pageable:false`，`write:false`）。

## 前置检查

```bash
kuaimai-cli auth status --output json
```

## 命令

参数为 **sysItemId**（系统商品长整型 ID），**不是**货号、商家编码等短编号。

```bash
kuaimai-cli item get-detail \
  --sys-item-id <sysItemId> \
  --output json --no-color
```

## 请求说明

- **contentType**：`get_query` — 参数拼接到 URL：`/item/getItemDetail?sysItemId=…`
- **write**：`false` — 不支持 `--dry-run`
- **pageable**：`false` — `--page-all` 无效

### 等价 service 调用

```bash
kuaimai-cli service item item-detail \
  --body '{"sysItemId":<sysItemId>}' \
  --output json --no-color
```

`sysItemId` 在 requestSchema 中为 **required**；service 层缺参会报错。

## 响应解析

- 成功：`ok === true`，详情通常在 `data[0]`
- 展示给用户时提取关键字段（如 `title`、`sysItemId`、SKU 相关字段）
- save / update-title 需要本接口返回的**全量 body**，禁止瘦身

## 注意

- 只有标题没有 ID 时，先 [`kuaimai-item-list.md`](kuaimai-item-list.md) 取 `sysItemId`
- 详情中 `itemSuiteBridgeList` 保存时需映射为 `suiteBridgeList`（见 save reference）

## 失败处理

读 `error` 与 `hint`；必要时加 `--verbose`。
