# item vs scm 域选型

## 核心原则

| 维度 | 商品域（item Skill） | 供应链域（scm Skill） |
|------|---------------------|----------------------|
| 后端 | erp-items-core | erp-scm |
| 域名 | `erp1.superboss.cc`（config `api.url`） | `scm.superboss.cc`（registry `baseUrl`） |
| 商品语义 | ERP 库存 / 档案商品 | 供应链分销商品 |
| 典型路径 | `/item/stock/queryList` | `/item/base/page` |

**路径前缀同为 `/item/` 但属于不同后端，不可混用。**

## 用户说法 → 选哪个 Skill

| 用户描述 | Skill |
|----------|-------|
| 库存、sysItemId、改标题、品牌类目档案 | `kuaimai-item` |
| 供应链、铺货、分销商品、供销、铺货日志 | `kuaimai-scm` |
| 操作日志、平台铺货配置、员工（scm 入口） | `kuaimai-scm` |

## 调用方式

两个域均走 registry 发现，**不要硬编码 apiId**：

```bash
kuaimai-cli capabilities --output json
kuaimai-cli schema <apiId> --output json
kuaimai-cli web call <apiId> --body '...' --output json
```

## 易错

- ❌ 用 erp1 商品域 apiId 查供应链商品列表
- ✅ 在 `capabilities` 中筛选 scm 域 / 铺货相关 apiId，再 `schema` → `web call`

## staff 说明

`/staff/query` 在 scm 域名入口调用，部分数据代理自 ERP 主系统；token 须对 scm 环境有效。
