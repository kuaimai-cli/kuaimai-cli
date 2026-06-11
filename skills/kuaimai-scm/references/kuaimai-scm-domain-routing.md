# item vs scm 域选型

## 核心原则

| 维度 | `web call` / `item` shortcuts | `web call` |
|------|-----------------------------------|---------------|
| 后端 | erp-items-core | erp-scm |
| 域名 | `erp1.superboss.cc`（config `api.url`） | `scm.superboss.cc`（meta `baseUrl`） |
| 商品语义 | ERP 库存 / 档案商品 | 供应链分销商品 |
| 典型路径 | `/item/stock/queryList` | `/item/base/page` |

**路径前缀同为 `/item/` 但属于不同后端（erp1 vs scm），不可混用。**

## 用户说法 → 选型

| 用户描述 | 选型 |
|----------|------|
| 库存、sysItemId、改标题、品牌类目档案 | `kuaimai-item` |
| 供应链、铺货、分销商品、供销、铺货日志 | `kuaimai-scm` |
| 操作日志、平台铺货配置、员工（scm 入口） | `kuaimai-scm` |

## 易错示例

```bash
# ❌ 用 item 查供应链商品列表
kuaimai-cli web call item.stock-list ...

# ✅ 供应链商品列表
kuaimai-cli web call scm.item-base-page --body '{"pageNo":1,"pageSize":50}'
```

## staff 说明

`/staff/query` 在 scm 域名入口调用，部分数据代理自 ERP 主系统；token 须对 scm 环境有效。
