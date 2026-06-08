---
name: kuaimai-scm
version: 1.0.0
description: "快麦 ERP 供应链（erp-scm）：员工管理、铺货日志、供应链商品、平台铺货配置。用户提到供应链、scm、铺货、铺货日志、操作日志、分销商品、平台配置、staff、dsb 时使用。"
metadata:
  requires:
    bins: ["kuaimai-cli"]
  cliHelp: "kuaimai-cli service scm --help"
---

# scm (v1.0.0)

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../kuaimai-shared/SKILL.md`](../kuaimai-shared/SKILL.md)，其中包含认证、输出信封与安全规则**

## 架构分层（Agent 必读）

| 层级 | 模块 | 状态 | Agent 职责 |
|------|------|------|------------|
| 底层 | `meta_data.json` → `services.scm` | ✅ 已完成 | 需要发现接口时 `kuaimai-cli schema --output json` |
| 中层 | **Skill（本文件 + references/）** | ✅ 当前层 | 选命令、组参数、控制分页与写操作确认 |
| 上层 | `service scm` / `api` | 已部分实现 | 按本 Skill 路由到 `service scm <operation>` |

**口诀**：scm 域暂无 shortcuts → 读本 Skill / references；不确定 operation 名时先 `schema`。

meta 执行规则详见 [`references/kuaimai-scm-meta-execution.md`](references/kuaimai-scm-meta-execution.md)。

**域名**：`service scm *` 自动请求 `https://scm.superboss.cc/`（meta `baseUrl`），**不要**改全局 `api.url` 或手写 URL。

## 选哪个命令

**item vs scm 分流**：必读 [`references/kuaimai-scm-domain-routing.md`](references/kuaimai-scm-domain-routing.md)。

| 用户意图 | meta operation | CLI 命令 |
|----------|----------------|----------|
| 查员工 / 用户列表 | `staff-query` | `service scm staff-query` |
| 员工店铺权限 | `staff-show-edit-staff-shop` | `service scm staff-show-edit-staff-shop` |
| 铺货日志 | `logging-publish-log` | `service scm logging-publish-log` |
| 平台商品铺货日志 | `logging-platform-product-publish-log` | `service scm logging-platform-product-publish-log` |
| 商品编辑日志 | `logging-product-edit-log-page` | `service scm logging-product-edit-log-page` |
| 操作日志（增删改） | `logging-operator-log` | `service scm logging-operator-log` |
| 日志途径枚举 | `logging-query-channel-by-type` | `service scm logging-query-channel-by-type` |
| 供应链商品列表 | `item-base-page` | `service scm item-base-page` |
| 供应链商品详情 | `item-base-detail` | `service scm item-base-detail` |
| 平台铺货配置 | `dsb-query-distribution-config` | `service scm dsb-query-distribution-config` |

模块工作流详见 `references/`：`staff`、`logging`、`item-base`、`dsb`。

## 已注册核心接口（meta → 命令）

| meta operation | contentType | write | pageable | path |
|----------------|-------------|-------|----------|------|
| `staff-query` | post_form | false | **true** | `/staff/query` |
| `staff-show-edit-staff-shop` | get_query | false | false | `/staff/showEditStaffShop` |
| `logging-publish-log` | post_json | false | **true** | `/logging/publishLog` |
| `logging-platform-product-publish-log` | post_json | false | **true** | `/logging/platformProductPublishLog` |
| `logging-product-edit-log-page` | post_json | false | **true** | `/logging/productEditLogPage` |
| `logging-operator-log` | post_json | false | **true** | `/logging/operatorLog` |
| `logging-query-channel-by-type` | get_query | false | false | `/logging/queryChannelByType` |
| `item-base-page` | post_json | false | **true** | `/item/base/page` |
| `item-base-detail` | get_query | false | false | `/item/base/detail` |
| `dsb-query-distribution-config` | get_query | false | false | `/dsb/queryDistributionConfig` |

> 全量 ~170 个 scm operation 见 `kuaimai-cli schema --output json`；通用调用见 [`kuaimai-scm-service.md`](references/kuaimai-scm-service.md)。

**CRITICAL — 写操作执行前 MUST 先 Read 对应 references 文档**

## meta 驱动能力速查

| 能力 | 规则 | Agent 动作 |
|------|------|------------|
| **baseUrl** | scm 固定 `scm.superboss.cc` | 用 `service scm`，勿改 `api.url` |
| **contentType** | `get_query` / `post_form` / `post_json` | `--body` 统一传 JSON，CLI 按 meta 转换 |
| **pageable** | 仅 `pageable:true` + `--page-all` 全量翻页 | 默认单页；见 meta-execution |
| **write** | `write:true` 支持 `--dry-run` | 写操作先 `--dry-run --verbose` |
| **日志时间** | 多数日志接口需 `startTime`/`endTime` | 格式 `yyyy-MM-dd HH:mm:ss` |

## 快速决策

- 供应链 / 铺货 / 操作日志 → **本 Skill** + `service scm`
- ERP 库存商品 / 改标题 → **`kuaimai-item`** + `item` shortcuts
- 路径同为 `/item/*` 时 **必须** 按业务域选 service（见 domain-routing）
- 日志查询缺时间范围 → 默认近 7 天，向用户确认后填入
- 失败时优先转述 `hint`；`result=901` → 引导 `auth login`

## 典型场景

```bash
# 员工列表
kuaimai-cli service scm staff-query \
  --body '{"pageNo":1,"pageSize":20}' --output json --no-color

# 铺货日志（startTime/endTime 必填）
kuaimai-cli service scm logging-publish-log \
  --body '{"pageNo":1,"pageSize":20,"startTime":"2026-06-01 00:00:00","endTime":"2026-06-08 23:59:59"}' \
  --output json --no-color

# 供应链商品列表
kuaimai-cli service scm item-base-page \
  --body '{"pageNo":1,"pageSize":50,"title":"关键字"}' --output json --no-color

# 平台铺货配置（抖音）
kuaimai-cli service scm dsb-query-distribution-config \
  --body '{"shopType":"TouTiaoFXG"}' --output json --no-color

# 日志途径：type=1 添加商品
kuaimai-cli service scm logging-query-channel-by-type \
  --body '{"type":1}' --output json --no-color
```

## API Resources（兜底）

```bash
kuaimai-cli schema --output json | jq '.data.operations[] | select(.service=="scm")'
kuaimai-cli service scm <operation> --body '{...}'
kuaimai-cli api POST /logging/publishLog   # 需手动设 api.url 为 scm，不推荐
```

## 不在本 skill 范围

- 登录、配置、输出格式 → [`kuaimai-shared`](../kuaimai-shared/SKILL.md)
- ERP 库存商品域 → [`kuaimai-item`](../kuaimai-item/SKILL.md)
- meta 维护 → 仓库 `docs/kuaimai-cli meta_data.json 定义规范.md`
