# erp-item vs scm-item 商品选型

## 核心原则

| 维度 | ERP 商品档案（kuaimai-erp-item） | SCM 可铺货商品（kuaimai-scm-item） |
|------|---------------------|----------------------|
| 后端 | erp-items-core | erp-scm |
| 域名 | `erp1.superboss.cc`（config `api.url`） | `scm.superboss.cc`（registry `baseUrl`） |
| 商品语义 | ERP 商品档案、库存资料、SKU、标题 | SCM 供应链商品、分销商品、可铺货商品 |
| 是否可上货 | 否 | 是 |
| 来源 | ERP 手工/接口维护商品档案 | 可手工新增，也可从 ERP 商品档案导入 |
| 典型路径 | `/item/stock/queryList` | `/item/base/page` |

**路径前缀同为 `/item/` 但属于不同后端，不可混用。**

## 用户说法 → 选哪个 Skill

| 用户描述 | Skill |
|----------|-------|
| 商品档案查询、商品档案列表、商品档案新增、商品档案编辑、商品档案保存、库存、sysItemId、改标题、品牌类目档案 | `kuaimai-erp-item` |
| SCM 商品、供应链商品、可铺货商品、铺货、上货到店铺、发布到店铺、分销商品、供销、铺货日志 | `kuaimai-scm-item` |
| 操作日志、平台铺货配置、员工（scm 入口） | `kuaimai-scm-item` |

## 调用方式

两个域均走 registry 发现，**不要硬编码 apiId**：

```bash
kuaimai-cli capabilities --output json
kuaimai-cli schema <apiId> --output json
kuaimai-cli web call <apiId> --body '...' --output json
```

## 易错

- ❌ 用 `kuaimai-erp-item` 查到商品档案后直接铺货
- ❌ 用户说“商品档案新增/编辑”时使用 `kuaimai-scm-item`
- ❌ 用 erp1 商品域 apiId 查 SCM 可铺货商品列表
- ✅ 用户说“商品档案”的查询、列表、新增、编辑、保存、改标题时使用 `kuaimai-erp-item`
- ✅ 铺货前必须先定位 SCM 商品（如 `/item/base/page` 返回的 `baseItemId` / `canPublishPlatformList`）
- ✅ 在 `capabilities` 中筛选 scm 域 / 铺货相关 apiId，再 `schema` → `web call`

## staff 说明

`/staff/query` 在 scm 域名入口调用，部分数据代理自 ERP 主系统；token 须对 scm 环境有效。
