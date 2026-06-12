---
name: kuaimai-shared
version: 1.2.0
description: "快麦 kuaimai-cli 全局约定：安装、config、auth、registry 发现（capabilities/schema/web call）、API 网关、输出信封、安全与排错。用户要配置 CLI、登录、发现接口、或看不懂 ok/data/error 时使用。"
metadata:
  requires:
    bins: ["kuaimai-cli"]
  cliHelp: "kuaimai-cli --help"
---

# kuaimai-cli 共享规则

本技能指导你如何通过 kuaimai-cli 操作快麦 ERP。**接口清单来自远端 registry**，不要手写 URL 或猜测参数。

| 用户意图 | 下一步 Skill |
|----------|--------------|
| ERP 商品档案：查询、列表、统计、详情、新增、编辑、改标题 | [`kuaimai-erp-item`](../kuaimai-erp-item/SKILL.md) |
| SCM 商品 / 可铺货商品：供应链商品、分销商品、上货、铺货、铺货日志 | [`kuaimai-scm-item`](../kuaimai-scm-item/SKILL.md) |
| 配置、登录、registry 发现、输出格式 | **本 Skill** |

## CLI 可执行文件

按顺序选用（第一个可用即可）：

1. 环境变量 `KUAIMAI_CLI`
2. 当前目录 `./kuaimai-cli`（仓库根目录开发时）
3. `PATH` 中的 `kuaimai-cli`

下文命令均以 `kuaimai-cli` 表示。Agent **必须在终端真实执行**，不要手写 curl。

## 三级命令（对标飞书 lark-cli）

| 层级 | 示例 | 何时用 |
|------|------|--------|
| **1. Shortcuts** | `erp-item +list`、`erp-item update-title` | 有 curated shortcut 时 **优先** |
| **2. web call** | `web call api.luotao.test.get` | registry 已发布接口（**主路径**） |
| **3. api** | `api POST /item/stock/queryList` | registry 未覆盖时的最后手段 |

**Agent 口诀**：

```text
有 shortcut → 读域 Skill，不查 schema
无 shortcut → capabilities → schema <apiId> → web call <apiId>
不知道有哪些接口 → capabilities 或 schema 全量
```

**商品口径必须先分清**：

- 用户说“商品档案”的查询、列表、详情、新增、编辑、改标题 → `kuaimai-erp-item`。
- 用户说“上货/铺货/发布到店铺/供应链商品/SCM 商品/分销商品” → `kuaimai-scm-item`。
- 只说“商品”但目标是铺货到店铺 → 视为 `kuaimai-scm-item`；只说“商品”且目标是档案资料维护 → 视为 `kuaimai-erp-item`。

## Registry 接口发现（核心）

远端源默认：`http://open-cli.kuaimai.com/registry/registry.json`  
CLI 每次命令前自动同步到 `~/.kuaimai-cli/registry/registry.json`。

```bash
# 1. 列出全部 apiId（按 domain 分组）
kuaimai-cli capabilities --output json

# 2. 单接口自省（requestSchema / responseSchema / pageable / write）
kuaimai-cli schema <apiId> --output json

# 3. 调用
kuaimai-cli web call <apiId> --params '{"k":"v"}'   # get_query
kuaimai-cli web call <apiId> --data '{"k":"v"}'    # post_json / post_form
kuaimai-cli web call <apiId> --body '{"k":"v"}'    # 按 contentType 自动路由

# 手动同步 / 开发监听
kuaimai-cli registry sync --output json
kuaimai-cli registry watch --interval 30 --verbose
```

跳过自动同步：`KUAIMAI_CLI_SKIP_REGISTRY_SYNC=1`

**禁止**：在 Skill 或对话中维护一份手写接口表代替 `capabilities` / `schema`。

## 配置

```bash
kuaimai-cli config init
kuaimai-cli config get --output json
kuaimai-cli config set api.url "https://erp1.superboss.cc/"
kuaimai-cli config set shortcuts.erp-item.api_url "https://erp1.superboss.cc/"
kuaimai-cli config set shortcuts.scm-item.api_url "https://scm3.superboss.cc/"
kuaimai-cli config set api.gateway_url "https://open-cli.kuaimai.com"
```

| 项 | 说明 |
|----|------|
| 配置文件 | `~/.kuaimai-cli/config.yaml` |
| 默认 API | `api.url` = `https://erp1.superboss.cc/`（未声明目标域时的兜底 `targetHost`） |
| API 网关 | `api.gateway_url` = `https://open-cli.kuaimai.com`（**所有业务 HTTP 实际请求地址**） |
| erp-item shortcut | `shortcuts.erp-item.api_url` = `https://erp1.superboss.cc/` |
| scm-item shortcut | `shortcuts.scm-item.api_url` = `https://scm3.superboss.cc/` |
| registry 域 | `web call <apiId>` 的 `targetHost` 优先来自单个 API 条目的 registry `baseUrl`；业务域 `erp系统` 为 erp1，`供应链` 为 scm3，仍经同一网关 |
| 超时 | `api.timeout` 默认 60 秒（与网关上游超时一致） |

业务请求路径：`CLI → POST {gateway_url}/api/forward → 真实后端`。Registry 同步（`registry.source`）不经网关。

## 认证

快麦使用用户提供的 `accessToken`（密钥链存储），**无 OAuth 交互登录**。

```bash
kuaimai-cli auth login "<accessToken>"
kuaimai-cli auth login "<token>" --profile <name>
kuaimai-cli auth use <profile>
kuaimai-cli auth status --output json
kuaimai-cli auth check --output json
```

**Agent 不能代用户获取 token。** 引导用户向 ERP 管理员申请，待用户提供后再 `auth login`。

## 输出约定

| `--output` | 行为 |
|------------|------|
| `json`（Agent 推荐） | 信封 `{ok, data, error, hint}` |
| `table` | 默认人类可读 |
| `csv` / `ndjson` | 列表成功时为裸流；失败仍为 JSON 信封 |

- `ok === true` → 数据在 `data`
- `ok === false` → 读 `error`、`hint`；必要时 `--verbose`

**stdout** = 数据；**stderr** = 日志与友好错误。

## 全局参数

| 参数 | 用途 |
|------|------|
| `--output json` | Agent 解析 |
| `--verbose` | 排错、配合 `--dry-run` |
| `--dry-run` | 写操作试跑（仅 `write:true`） |
| `--page-all` | 列表翻页（仅 `pageable:true`） |
| `--page-limit` / `--page-confirm` | 翻页条数上限与续查策略 |

## 安装与更新

```bash
npx @kuaimai-cli/cli@latest install
kuaimai-cli skill install --force
kuaimai-cli doctor --output json
kuaimai-cli upgrade
```

安装或升级 Skill 后 **重新打开 Agent 会话**。

## 安全规则

- **禁止**将 `accessToken` 写入 Git、`config.yaml` 或对话外的持久化存储
- 写操作先 `--dry-run --verbose`，用户确认后再提交
- 用户未明确要求全量时，**不要**加 `--page-all`

## 错误与排错

| 现象 | 处理 |
|------|------|
| 未登录 / 401 | `auth status` → 引导 `auth login` |
| 429 限流 | 提示「请求过于频繁」；降低调用频率，**不要**自动重试 |
| 未找到 apiId | `registry sync` → `capabilities` 确认是否已发布 |
| `ok: false` | 读 `error`、`hint`；`--verbose` 重试 |
| Skill 过时 | `skill install --force` 并重开 Agent |
