# SCM 商品铺货 Shortcut 设计与使用说明

## 定位

本文说明 `kuaimai-cli scm-item publish`、`publish-pdd`、底层 primitive 与 `publish-log` 的设计、使用方式和接口链路。

目标是按 SCM 款式编码定位商品，将商品铺货到指定平台店铺，并可查询铺货日志和失败原因。`publish-pdd` 保留为 PDD 兼容入口；新增平台优先使用通用命令：

```bash
kuaimai-cli scm-item publish --platform <平台key> --style-code '<款式编码>' --shop-id <店铺ID> --output json
```

## 支持平台

当前通用链路支持这些平台 key：

`pdd`、`tb`、`fxg`、`kuaishou`、`jd`、`1688`、`tm`、`tjb`、`yz`、`wxsph`、`wxxd`、`wd`、`xhs`、`xy`、`fxg_gx`、`yyjk`、`pddtemu`、`shein`、`ktt`、`jdms`。

前端源码确认的差异：

| 平台 | 前置接口 |
|---|---|
| `pdd` | 商品、店铺、PDD 视频、控价、PDD 授权、二次确认、提交 |
| `tb` / `fxg` / `jd` / `1688` | 商品、店铺、控价、二次确认、提交 |
| `kuaishou` | 商品、店铺、控价、提交 |
| 其他平台 | 商品、店铺、控价、提交；如后续 HAR 发现额外前置校验，再在平台配置中补齐 |

## 命令边界

| 命令 | 用途 | 是否提交铺货 |
|---|---|---|
| `scm-item +list` | 分页查询 SCM 可铺货商品，可按款式编码/标题过滤 | 否 |
| `scm-item shops` | 查询某商品在指定平台可铺货店铺 | 否 |
| `scm-item publish --platform <key>` | 通用平台铺货预检，默认输出提交计划 | 默认否 |
| `scm-item publish --platform <key> --submit` | 在预检通过后提交 `/taskScheduling/storageTask` | 是 |
| `scm-item publish-pdd` | PDD 兼容入口，等价于 `publish --platform pdd` | 默认否 |
| `scm-item platform-*` | 通用底层接口 primitive，用于调试和分步验证 | `platform-storage-task --submit` 才提交 |
| `scm-item pdd-*` | PDD 专用 primitive，额外包含视频和授权接口 | `pdd-storage-task --submit` 才提交 |
| `scm-item publish-log --detail` | 查询铺货日志明细、单品状态和 `errorMessage` | 否 |

实际铺货必须显式传入 `--submit`，这是 shortcut 的安全边界。

## 标准流程

### 1. 定位 SCM 可铺货商品

```bash
kuaimai-cli scm-item +list --style-code '<款式编码>' --output json
```

`--style-code` 与 `--title` 都是可选过滤条件，可以只传一个、两个都传，或都不传只分页浏览。确认返回商品唯一，并且 `canPublishPlatform` 包含目标平台 key。

### 2. 查询可铺货店铺

```bash
kuaimai-cli scm-item shops --platform fxg --style-code '<款式编码>' --output json
```

如果店铺不可铺货，输出中的 `disabled_reason` 会说明原因。实际铺货建议使用 `shop-id`，避免店铺名重复。

### 3. 预检铺货

```bash
kuaimai-cli scm-item publish \
  --platform fxg \
  --style-code '<款式编码>' \
  --shop-id 123456 \
  --output json
```

预检会执行商品查询、店铺查询、平台前置校验，并输出 `dry_run:true`、商品摘要、店铺摘要、预检结果和待提交的 `publish_body`。

PDD 也可以继续使用：

```bash
kuaimai-cli scm-item publish-pdd --style-code '<款式编码>' --shop-id 123456 --output json
```

### 4. 确认后提交

```bash
kuaimai-cli scm-item publish \
  --platform fxg \
  --style-code '<款式编码>' \
  --shop-id 123456 \
  --submit \
  --output json
```

### 5. 分步核对底层接口

通用平台：

```bash
kuaimai-cli scm-item platform-price-precheck --platform tb --style-code '<款式编码>' --shop-id 123456 --output json
kuaimai-cli scm-item platform-sec-confirm --platform tb --style-code '<款式编码>' --shop-id 123456 --output json
kuaimai-cli scm-item platform-storage-task --platform tb --style-code '<款式编码>' --shop-id 123456 --output json
kuaimai-cli scm-item platform-task-speed --batch-task-id '<batchTaskId>' --output json
```

PDD 专用：

```bash
kuaimai-cli scm-item pdd-video-info --style-code '<款式编码>' --output json
kuaimai-cli scm-item pdd-auth-status --style-code '<款式编码>' --shop-id 123456 --output json
```

### 6. 查询失败原因

```bash
kuaimai-cli scm-item publish-log \
  --style-code '<款式编码>' \
  --shop-id 123456 \
  --detail \
  --output json
```

日志时间范围默认近 30 天。`--detail` 会读取日志明细，按款式编码过滤明细行，并汇总成功、失败、执行中、排队等状态；失败原因来自明细行 `errorMessage`。

## 输出中的接口标识

所有 shortcut 的 JSON `data` 都会返回 `interfaces`，用于确认本次命令实际依赖的底层接口：

```json
{
  "name": "item.base.page",
  "method": "POST",
  "path": "/item/base/page.json"
}
```

`endpoints` 仍保留为路径数组，兼容旧解析方式；新判断建议使用 `interfaces[].name`。

## 接口链路

| 步骤 | 接口名 | 方法 | 路径 | 关键参数 |
|---|---|---|---|---|
| 查 SCM 商品 | `item.base.page` | POST JSON | `/item/base/page.json` | `outerIds:[styleCode]`、`outerIdBlur:0` |
| 查可铺货店铺 | `shop.allShop` | GET query | `/shop/allShop.json` | `source:<platform>`、`baseItemId` |
| PDD 查轮播视频 | `pdd.getCarouselVideoInfo` | POST JSON | `/pdd/getCarouselVideoInfo.json` | `baseItemIds` |
| 控价校验 | `ltsTask.preCheckControllerPrice` | POST JSON | `/ltsTask/preCheckControllerPrice.json` | `itemId`、`shopIds`、`platformType` |
| PDD 授权校验 | `pdd.authorize.authStatus` | POST JSON | `/pdd/authorize/authStatus.json` | `shopIds:[taobaoId]` |
| 二次确认 | `item.base.secConfirmationItemV2` | POST JSON | `/item/base/secConfirmationItemV2.json` | `platformType`、`itemShopList` |
| 提交任务 | `taskScheduling.storageTask` | POST JSON | `/taskScheduling/storageTask.json` | `itemIds`、`shopIds`、`taskType:PUBLISH_ITEM` |
| 查询进度 | `taskScheduling.queryTaskSpeed` | GET query | `/taskScheduling/queryTaskSpeed.json` | `taskTypeEnum:PUBLISH_ITEM`、`batchTaskId` |

## Agent 调用规范

1. 用户只描述铺货需求时，先执行不带 `--submit` 的预检命令。
2. 如需要向用户展示候选，先执行 `scm-item +list` 和 `scm-item shops`。
3. 预检通过后，向用户展示商品、店铺、即将提交的动作，并等待明确确认。
4. 用户确认后，再执行 `--submit`。
5. 提交后优先加 `--check-log`；如日志尚未落库，稍后执行 `scm-item publish-log --detail`。
6. 不要把浏览器 cookie、账号密码或 accessToken 写入仓库、文档或持久化脚本。
