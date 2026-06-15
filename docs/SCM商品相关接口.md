# SCM 商品相关接口

本文记录 `kuaimai-cli scm-item` shortcuts 当前使用的 SCM 商品相关后端接口。旧的 PDD 铺货链路接口仍保留；新增接口主要来自 SCM 商品名称编辑和铺货日志查询。

## 接口总览

| 场景 | 接口名 | 方法 | 路径 | 关键参数 |
|---|---|---|---|---|
| 查 SCM 商品 | `item.base.page` | POST JSON | `/item/base/page.json` | `pageNo`、`pageSize`、`outerIds:[styleCode]`、`outerIdBlur:0`、`title` |
| 查 SCM 商品详情 | `item.base.detailByBaseItemId` | GET query | `/item/base/detailByBaseItemId.json` | `baseItemId` |
| 按自增 ID 查 SCM 商品详情 | `item.base.detail` | GET query | `/item/base/detail.json` | `id` |
| 保存前 ERP 同步校验 | `item.base.queryErpItems` | POST JSON | `/item/base/queryErpItems.json` | `id`、`companyId`、`outerId`、`title`、商品详情字段 |
| 编辑 SCM 商品 | `item.base.edit` | POST JSON | `/item/base/edit.json` | `item`、`checkOpenSync`、`skipAddItemToErp` |
| 查可铺货店铺 | `shop.allShop` | GET query | `/shop/allShop.json` | `source:<platform>`、`subSource`、`baseItemId`、`channel` |
| PDD 查轮播视频 | `pdd.getCarouselVideoInfo` | POST JSON | `/pdd/getCarouselVideoInfo.json` | `baseItemIds` |
| 控价校验 | `ltsTask.preCheckControllerPrice` | POST JSON | `/ltsTask/preCheckControllerPrice.json` | `itemId`、`shopIds`、`platformType` |
| PDD 授权校验 | `pdd.authorize.authStatus` | POST JSON | `/pdd/authorize/authStatus.json` | `shopIds:[taobaoId]` |
| 二次确认 | `item.base.secConfirmationItemV2` | POST JSON | `/item/base/secConfirmationItemV2.json` | `platformType`、`itemShopList` |
| 提交铺货任务 | `taskScheduling.storageTask` | POST JSON | `/taskScheduling/storageTask.json` | `itemIds`、`shopIds`、`taskType:PUBLISH_ITEM` |
| 查询铺货进度 | `taskScheduling.queryTaskSpeed` | GET query | `/taskScheduling/queryTaskSpeed.json` | `taskTypeEnum:PUBLISH_ITEM`、`batchTaskId` |
| 查询铺货日志 | `logging.publishLog` | POST JSON | `/logging/publishLog.json` | `operateType`、`startTime`、`endTime`、`pageNo`、`pageSize` |
| 查询铺货日志明细 | `logging.publishLogDetail` | GET query | `/logging/publishLogDetail.json` | `id` |
| 查询铺货日志头信息 | `logging.publishLogById` | GET query | `/logging/publishLogById.json` | `id` |

## 链路分组

### 商品查询

| 命令 | 涉及接口 |
|---|---|
| `scm-item +list` | `item.base.page` |
| `scm-item shops` | `item.base.page`、`shop.allShop` |

### 商品编辑

| 命令 | 涉及接口 | 说明 |
|---|---|---|
| `scm-item update-title` | `item.base.page`、`item.base.detailByBaseItemId`、`item.base.queryErpItems`、`item.base.edit` | 使用 `--style-code` 时先查列表定位 `baseItemId`；优先按前端编辑页一致的 `detailByBaseItemId` 读取详情。默认只输出 `save_body`，加 `--submit` 才保存。 |

编辑保存体格式：

```json
{
  "item": {
    "id": 123456,
    "title": "新商品名称"
  },
  "checkOpenSync": true,
  "skipAddItemToErp": false
}
```

实际 `item` 会来自 `/item/base/detailByBaseItemId.json` 的完整商品详情，只覆盖 `title`，并补齐前端保存时使用的 `smallShopItem:false`、`notCheckSyncFiledConf:true` 等默认字段。

### 铺货预检与提交

| 平台/命令 | 涉及接口 |
|---|---|
| `publish --platform pdd` / `publish-pdd` | `item.base.page`、`shop.allShop`、`pdd.getCarouselVideoInfo`、`ltsTask.preCheckControllerPrice`、`pdd.authorize.authStatus`、`item.base.secConfirmationItemV2`、`taskScheduling.storageTask`、`taskScheduling.queryTaskSpeed` |
| `publish --platform tb/fxg/jd/1688` | `item.base.page`、`shop.allShop`、`ltsTask.preCheckControllerPrice`、`item.base.secConfirmationItemV2`、`taskScheduling.storageTask`、`taskScheduling.queryTaskSpeed` |
| `publish --platform kuaishou` | `item.base.page`、`shop.allShop`、`ltsTask.preCheckControllerPrice`、`taskScheduling.storageTask`、`taskScheduling.queryTaskSpeed` |
| 其他已支持平台 | `item.base.page`、`shop.allShop`、`ltsTask.preCheckControllerPrice`、`taskScheduling.storageTask`、`taskScheduling.queryTaskSpeed` |

`taskScheduling.storageTask` 是实际提交铺货任务的写接口。shortcut 默认停在 dry-run/预检结果，只有显式传 `--submit` 才调用提交接口。

### 铺货日志

| 命令 | 涉及接口 |
|---|---|
| `scm-item publish-log` | `logging.publishLog` |
| `scm-item publish-log --detail` | `logging.publishLog`、`logging.publishLogDetail`、`logging.publishLogById` |
| `scm-item publish --submit --check-log` | 提交链路接口 + `logging.publishLog`、`logging.publishLogDetail`、`logging.publishLogById` |

## 与旧表对比

旧文档里的 8 个铺货链路接口仍然有效，不需要删除：

`item.base.page`、`shop.allShop`、`pdd.getCarouselVideoInfo`、`ltsTask.preCheckControllerPrice`、`pdd.authorize.authStatus`、`item.base.secConfirmationItemV2`、`taskScheduling.storageTask`、`taskScheduling.queryTaskSpeed`。

本次新增写入文档的接口：

`item.base.detailByBaseItemId`、`item.base.detail`、`item.base.queryErpItems`、`item.base.edit`、`logging.publishLog`、`logging.publishLogDetail`、`logging.publishLogById`。
