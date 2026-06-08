# 供应链商品中心（/item/base）

## 常用接口

| operation | 说明 |
|-----------|------|
| `item-base-page` | 商品分页列表 |
| `item-base-detail` | 商品详情 |
| `item-base-edit` | 编辑商品（写） |
| `item-base-publish-item` | 铺货（写） |
| `item-base-batch-update-title` | 批量改标题（写） |

## 商品列表

```bash
kuaimai-cli service scm item-base-page \
  --body '{"pageNo":1,"pageSize":50,"title":"关键字","outerIdBlur":1}' \
  --output json --no-color
```

常用筛选：

| 字段 | 说明 |
|------|------|
| `title` | 商品名称 |
| `outerIds` | 款式编码列表 |
| `outerIdBlur` | 0 精确，1 模糊 |
| `distributionState` | 0 不完善，1 可分销，2 不可分销 |
| `onPublishShopIds` | 已铺货店铺 |
| `supplierName` | 供应商名称 |

## 商品详情

```bash
kuaimai-cli service scm item-base-detail \
  --body '{"baseItemId":123456}' \
  --output json --no-color
```

参数名以 `schema` 为准（可能为 `baseItemId` 或相关 ID 字段）。

## 写操作

编辑、铺货、批量更新等 `write:true` 接口：

1. 先 `item-base-detail` 拉全量
2. `--dry-run --verbose` 预览
3. 用户确认后提交
