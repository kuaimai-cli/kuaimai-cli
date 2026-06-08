# 商品 / 铺货操作日志（/logging）

## 接口

| operation | 说明 |
|-----------|------|
| `logging-publish-log` | 铺货日志分页 |
| `logging-platform-product-publish-log` | 平台商品铺货日志 |
| `logging-product-edit-log-page` | 商品编辑日志 |
| `logging-operator-log` | 操作日志（增删改） |
| `logging-query-channel-by-type` | 途径枚举（GET） |

## 时间范围（必填）

`logging-publish-log`、`logging-operator-log` 等通常需要：

```json
{
  "startTime": "2026-06-01 00:00:00",
  "endTime": "2026-06-08 23:59:59",
  "pageNo": 1,
  "pageSize": 20
}
```

Agent 未指定时，默认近 7 天并向用户说明。

## 铺货日志

```bash
kuaimai-cli service scm logging-publish-log \
  --body '{"pageNo":1,"pageSize":20,"startTime":"2026-06-01 00:00:00","endTime":"2026-06-08 23:59:59","outerId":"款式编码"}' \
  --output json --no-color
```

`publishStatus`：0 铺货中，1 全成功，2 全失败，3 部分成功。

## queryChannelByType

```bash
kuaimai-cli service scm logging-query-channel-by-type \
  --body '{"type":1}' --output json --no-color
```

| type | 含义 |
|------|------|
| 0 | 铺货 |
| 1 | 添加商品 |
| 2 | 删除商品 |
| 3 | 编辑商品 |

## 商品编辑日志

```bash
kuaimai-cli service scm logging-product-edit-log-page \
  --body '{"pageNo":1,"pageSize":20,"startTime":"2026-06-01 00:00:00","endTime":"2026-06-08 23:59:59","outerId":"款式编码"}' \
  --output json --no-color
```

`status`：1 成功，2 失败。
