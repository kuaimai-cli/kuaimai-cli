# item save — 全量保存商品（手动编排）

## 适用场景

需要修改 **save 载荷中多个字段**，或 `update-title` 无法覆盖的场景。

对应 meta operation：`item-save`（`post_json`，`write:true`，`pageable:false`）。

**BLOCKING REQUIREMENT：`save` 必须传 get-detail 返回的全量 body，禁止瘦身载荷。**

## 前置检查

```bash
kuaimai-cli auth status --output json
```

写操作**首次必须** `--dry-run --verbose`，用户确认后再去掉 `--dry-run`。

## 推荐：改标题仍优先 update-title

仅改标题时请用 [`kuaimai-item-update-title.md`](kuaimai-item-update-title.md)，不要手写 jq。

## 命令（示例：改标题）

**禁止**只传 `{"sysItemId":...,"title":"..."}`，会失败。

```bash
kuaimai-cli item save \
  --body "$(kuaimai-cli item get-detail --sys-item-id <sysItemId> --output json | jq -c '.data[0] | .title = "<新标题>" | .suiteBridgeList = .itemSuiteBridgeList | del(.itemSuiteBridgeList)')" \
  --dry-run --verbose --output json --no-color
```

用户确认预览无误后，**同一命令**去掉 `--dry-run --verbose` 再执行。

## 字段注意

- 详情中 `itemSuiteBridgeList` 保存时需映射为 `suiteBridgeList`
- 其它字段保持 get-detail 原样，避免覆盖丢失

## 与 web call 的区别

- `kuaimai-cli web call item.item-save` 为原子 API，**不会**自动 get-detail；需自行构造完整 `--body`
- `kuaimai-cli web call item.item-update-title` 在 meta 中登记为写操作，但**无** get-detail 编排

## 禁止

- 不要手写 URL 或 curl
- 不要在未 dry-run 预览的情况下直接提交 save

## 失败处理

读 `error` 与 `hint`；配合 `--verbose` 查看 JSON body 预览。
