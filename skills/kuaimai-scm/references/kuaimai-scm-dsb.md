# 平台铺货配置（/dsb）

## 常用接口

| operation | 说明 |
|-----------|------|
| `dsb-query-distribution-config` | 查询平台铺货配置 |
| `dsb-save-or-update-distribution-config` | 保存配置（写） |
| `dsb-query-freight-template` | 运费模板 |
| `dsb-get-douyin-template-list` | 抖音模板列表 |

## queryDistributionConfig

```bash
kuaimai-cli service scm dsb-query-distribution-config \
  --body '{"shopType":"TouTiaoFXG"}' \
  --output json --no-color
```

## shopType 示例

| shopType | 平台 |
|----------|------|
| `TouTiaoFXG` | 抖音 |
| `KWaiShop` | 快手 |
| `PinShop` | 拼多多 |
| `JdShop` | 京东 |

## 写操作

`dsb-save-or-update-distribution-config` 须先 `--dry-run --verbose`。
