# SCM 商品铺货 Shortcut 设计与使用说明

## 定位

本文说明 `kuaimai-cli scm-item publish-pdd` 与 `kuaimai-cli scm-item publish-log` 的设计、使用方式和接口链路。当前 shortcut 支持的业务目标是：

```text
按 SCM 款式编码定位商品，将商品铺货到任意一个支持铺货的拼多多店铺，并可查询铺货日志和失败原因。
```

该能力用于承接 Agent/Shortcuts 场景，例如用户授权后描述“帮我将款式编码 `<款式编码>` 这个商品上货到拼多多店铺 `<店铺名>` 上”，Agent 可以先执行预检，再在用户确认后提交铺货，并查询结果。

## 命令边界

| 命令 | 用途 | 是否提交铺货 |
|---|---|---|
| `scm-item +list` | 按款式编码/标题查询 SCM 可铺货商品，确认 `canPublishPlatform` | 否 |
| `scm-item shops` | 查询某个 SCM 商品在指定平台可铺货店铺，输出 `can_publish` / `disabled_reason` | 否 |
| `scm-item publish-pdd` | 定位商品、定位 PDD 店铺、保存临时配置、校验平台资料，默认输出提交计划 | 默认否 |
| `scm-item publish-pdd --submit` | 在预检通过后提交 `/pdd/batchPublishItem` | 是 |
| `scm-item publish-pdd --submit --check-log` | 提交后查询最近铺货日志与失败原因 | 是 |
| `scm-item publish-log --detail` | 单独查询铺货日志明细、单品状态和 `errorMessage` | 否 |

`publish-pdd` 默认不会调用最终提交接口。实际铺货必须显式传入 `--submit`，这是 shortcut 的安全边界。

## 标准流程

### 1. 定位 SCM 可铺货商品

```bash
kuaimai-cli scm-item +list \
  --style-code '<款式编码>' \
  --output json
```

确认返回商品唯一，并且 `canPublishPlatform` 包含目标平台，例如 `pdd`。

### 2. 查询可铺货店铺

```bash
kuaimai-cli scm-item shops \
  --platform pdd \
  --style-code '<款式编码>' \
  --output json
```

如果店铺不可铺货，输出中的 `disabled_reason` 会说明原因。实际铺货时建议使用 `shop-id`，避免店铺名重复。

### 3. 预检铺货

```bash
kuaimai-cli scm-item publish-pdd \
  --style-code '<款式编码>' \
  --shop '<拼多多店铺名>' \
  --output json
```

预检会执行商品查询、店铺查询、临时配置保存和 PDD 平台资料校验，但不会提交最终铺货任务。输出中会包含 `dry_run:true`、商品摘要、店铺摘要、`flow_number` 和待提交的 `publish_body`。

### 4. 确认后提交

```bash
kuaimai-cli scm-item publish-pdd \
  --style-code '<款式编码>' \
  --shop '<拼多多店铺名>' \
  --submit \
  --output json
```

如果店铺名可能重复，使用店铺 ID：

```bash
kuaimai-cli scm-item publish-pdd \
  --style-code '<款式编码>' \
  --shop-id 123456 \
  --submit \
  --output json
```

### 5. 提交后查询日志

```bash
kuaimai-cli scm-item publish-pdd \
  --style-code '<款式编码>' \
  --shop '<拼多多店铺名>' \
  --submit \
  --check-log \
  --output json
```

`--check-log` 会在提交后查询最近铺货日志。由于后端铺货任务可能异步落库，如果首次没有查到结果，可以稍后单独执行 `publish-log`。

### 6. 单独查询失败原因

```bash
kuaimai-cli scm-item publish-log \
  --style-code '<款式编码>' \
  --shop '<拼多多店铺名>' \
  --detail \
  --output json
```

指定时间范围：

```bash
kuaimai-cli scm-item publish-log \
  --style-code '<款式编码>' \
  --shop-id 123456 \
  --start-time '2026-06-01 00:00:00' \
  --end-time '2026-06-12 23:59:59' \
  --detail \
  --output json
```

日志时间范围默认近 30 天。`--detail` 会读取日志明细，按款式编码过滤明细行，并汇总成功、失败、执行中、排队等状态；失败原因来自明细行 `errorMessage`。

## 铺货前置条件

- SCM 商品必须能通过 `outerId` 精确唯一匹配。
- 商品的 `canPublishPlatformList` 必须包含 `pdd`。
- PDD 店铺必须存在于 `/shop/allShop` 返回的授权店铺列表中。
- 店铺不能处于异常、授权过期、未配置运费模板、平台限制等不可铺货状态。
- `/pdd/queryBatchDetail` 返回的商品平台资料必须满足 `informationLack == 1`。

任一条件不满足时，shortcut 会阻断提交并返回明确错误。

## 接口链路

### 铺货链路

| 步骤 | 方法 | 路径 | 关键参数 | 说明 |
|---|---|---|---|---|
| 查 SCM 商品 | POST JSON | `/item/base/page` | `outerIds:[styleCode]`、`outerIdBlur:0` | `scm-item +list` 与铺货预检共用，按款式编码精确定位 SCM 商品 |
| 查可铺货店铺 | GET query | `/shop/allShop` | `source:pdd`、`baseItemId` | `scm-item shops` 与铺货预检共用，查询该商品可用的 PDD 授权店铺 |
| 保存临时配置 | POST JSON | `/pdd/saveBatchTempConf` | `shopType:pdd`、`shelfState`、`flowNumber` | 生成批量铺货临时配置 |
| 查询平台资料 | POST JSON | `/pdd/queryBatchDetail` | `flowNumber`、`baseItemIds`、`shopIds` | 校验 PDD 平台资料完整性 |
| 提交铺货 | POST JSON | `/pdd/batchPublishItem` | `shopIds`、`flowNumber`、`batchItemDetailList` | 仅 `--submit` 时调用 |

### 日志链路

| 步骤 | 方法 | 路径 | 关键参数 | 说明 |
|---|---|---|---|---|
| 查铺货日志 | POST JSON | `/logging/publishLog` | `startTime`、`endTime`、`pageNo`、`pageSize` | 查询铺货日志列表 |
| 查日志明细 | GET query | `/logging/publishLogDetail` | `id=<日志行 id>` | 明细行包含 `outerId`、`status`、`errorMessage` |
| 查日志头信息 | GET query | `/logging/publishLogById` | `id=<operationLogId>` | 补充操作日志头信息 |

## Agent 调用规范

1. 用户只描述铺货需求时，先执行不带 `--submit` 的预检命令。
2. 如需要向用户展示候选，先执行 `scm-item +list` 和 `scm-item shops`。
3. 预检通过后，向用户展示商品、店铺、即将提交的动作，并等待明确确认。
4. 用户确认后，再执行 `--submit`。
5. 提交后优先加 `--check-log`；如日志尚未落库，稍后执行 `scm-item publish-log --detail`。
6. 不要把浏览器 cookie、账号密码或 accessToken 写入仓库、文档或持久化脚本。

## 认证与环境

命令复用 `kuaimai-cli` 现有 `auth login <accessToken>` 和 API gateway 转发机制，目标 host 固定为 `https://scm3.superboss.cc/`。如果服务端 token 与浏览器 cookie 登录态权限不一致，需要在 registry/API gateway 侧补齐 SCM3 环境 token 支持。
