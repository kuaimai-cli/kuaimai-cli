---
name: kuaimai-item
version: 4.0.0
description: "快麦 ERP 商品域：库存、档案、标题、SKU、列表、详情、修改。用户提到商品、货号、sysItemId、品牌、类目时使用。接口一律经 registry 发现，不维护本地接口表。"
metadata:
  requires:
    bins: ["kuaimai-cli"]
  cliHelp: "kuaimai-cli capabilities --output json"
---

# item（商品域 · erp1）

**CRITICAL — 开始前 MUST Read [`../kuaimai-shared/SKILL.md`](../kuaimai-shared/SKILL.md)**

本 Skill 只做 **商品域意图识别** 与 **registry 发现调用**。  
**禁止**凭记忆或文档里的 apiId 表调用；参数以 `schema <apiId>` 为准。

## 何时读本 Skill

用户意图涉及：商品、SKU、标题、货号、outerId、sysItemId、库存列表、商品详情、改名、品牌、类目、档案……

供应链 / 铺货 / 分销商品 → [`kuaimai-scm`](../kuaimai-scm/SKILL.md)

## 标准流程（每步 MUST 执行）

```bash
# 1. 发现：按 domain=商品 或 title/description 关键词筛选
kuaimai-cli capabilities --output json

# 2. 自省：requestSchema / pageable / write / examples / contentType
kuaimai-cli schema <apiId> --output json

# 3. 调用（经 open-cli 网关）
kuaimai-cli web call <apiId> --params '{"k":"v"}'    # get_query
kuaimai-cli web call <apiId> --data '{"k":"v"}'     # post_json / post_form
kuaimai-cli web call <apiId> --body '{"k":"v"}'     # 按 contentType 自动路由
```

| 步骤 | 规则 |
|------|------|
| 选 apiId | 只从 `capabilities` 返回的 apis 中选；读 `title`、`description`、`domain` |
| 填参数 | 只按 `schema` 的 `requestSchema`；可参考 `examples` |
| 写操作 | `write:true` 时先 `--dry-run --verbose`，用户确认后再提交 |
| 翻页 | 仅 `pageable:true` 可用 `--page-all`；全量前先评估数据量 |
| 找不到 | `registry sync` → 重跑 `capabilities`；仍无则接口未发布 |

## 域分流（CRITICAL）

路径前缀同为 `/item/`，但 **erp1 与 scm 是不同后端**：

| 用户描述 | 读本 Skill | 后端 |
|----------|-----------|------|
| ERP 库存、档案、改标题、sysItemId | ✅ item | erp1（`api.url`） |
| 供应链、铺货、分销商品 | ❌ → scm | scm.superboss.cc |

不确定时 Read [`../kuaimai-scm/references/kuaimai-scm-domain-routing.md`](../kuaimai-scm/references/kuaimai-scm-domain-routing.md)

## 意图 → 发现策略（非接口表）

用 `capabilities` + `schema` 匹配，**不要硬编码 apiId**：

| 用户意图 | 在 capabilities/schema 中找 |
|----------|------------------------------|
| 列表 / 搜索 | title 或 description 含 list、query、stock、列表 |
| 统计 / 有多少 | 含 count、统计 |
| 详情 | 含 detail、get、详情 |
| 新建 / 修改 / 保存 | `write:true`，含 save、update、修改 |

只有标题没有 ID：先调列表类 apiId 取 `sysItemId`，再调详情 / 写接口。

## 典型流程：按标题定位再修改

| 步骤 | 操作 |
|------|------|
| 1 | `capabilities` → 选列表 apiId → `schema` → `web call` 按标题搜索 |
| 2 | 从 `data` 取 `sysItemId`（多条则让用户选） |
| 3 | `capabilities` → 选写 apiId → `schema` → `web call --dry-run --verbose` |
| 4 | 用户确认后去掉 `--dry-run` |

## 不在本 Skill 范围

- 配置、登录、registry 机制 → [`kuaimai-shared`](../kuaimai-shared/SKILL.md)
- 供应链 / 铺货 / 操作日志 → [`kuaimai-scm`](../kuaimai-scm/SKILL.md)
