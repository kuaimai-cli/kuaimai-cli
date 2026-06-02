# 快麦 CLI 安装指南
以下步骤面向 AI Agent，登录环节需要用户向管理员申请并提供 `accessToken`。

## 环境要求
开始安装之前，请确保环境中已安装：
- Node.js 16+（含 npm/npx）

## 第 1 步 安装 CLI 本体
```shell
# 安装快麦 CLI 命令行工具
npx @kuaimai-cli/cli@latest install
```

## 第 2 步 安装必需 Skill

从 GitHub 拉取 **整目录**（`SKILL.md` + `references/` 工作流文档），安装到各 Agent 目录（`~/.agents/skills`、`~/.cursor/skills` 等）：

```shell
kuaimai-cli skill install
```

默认安装 `kuaimai-shared`（全局约定）与 `kuaimai-item` v2.0.0（商品域）。安装后目录结构示例：

```text
~/.agents/skills/kuaimai-item/
├── SKILL.md
└── references/
    ├── kuaimai-item-list.md
    ├── kuaimai-item-count.md
    ├── kuaimai-item-get-detail.md
    ├── kuaimai-item-update-title.md
    ├── kuaimai-item-save.md
    ├── kuaimai-item-meta-execution.md    # meta 驱动规则、分页防护
    ├── kuaimai-item-service.md           # service 层兜底
    └── kuaimai-item-query-list-v2.md     # 档案 V2 列表（无 shortcut）
```

**Agent 使用约定**：

- 处理商品域请求前，先 Read `kuaimai-item/SKILL.md`，并按其 **CRITICAL** 提示 Read `kuaimai-shared/SKILL.md`
- 写操作（`save`、`update-title`）须先 Read 对应 `references/` 文档
- 全量翻页时使用 `--page-all`；Agent/脚本续查加 `--page-confirm yes`；限制条数用 `--page-limit`

## 第 3 步 初始化配置
Agent 运行以下命令，完成基础配置初始化。
```shell
kuaimai-cli config init
```

## 第 4 步 账号登录
本步骤需要用户提供 `accessToken`。Agent 应先提示用户：**如尚未持有 accessToken，请联系 ERP 管理员申请分配**；待用户将 token 提供后，再执行以下命令完成授权（切勿代填或猜测 token）。
```shell
kuaimai-cli auth login "<accessToken>"
```

## 第 5 步 状态验证
```shell
kuaimai-cli auth status
kuaimai-cli doctor --output json
```

`doctor` 会检查 `kuaimai-item` Skill 是否已安装且含 `references/`；若缺失请执行 `kuaimai-cli skill install` 或 `kuaimai-cli skill install --if-stale`（覆盖写入，无需手动删除各 Agent 缓存目录）。

## 第 6 步 升级与 Skill 同步（对标飞书 lark-cli）

日常使用任意命令后，若 GitHub 有新版，CLI 会在 **stderr** 提示（24h 缓存，不污染 stdout JSON）。也可主动升级：

```shell
# 默认：有新版则 npm 全局升级并同步 Skills（与飞书一致）
kuaimai-cli upgrade

# 仅检查、不安装
kuaimai-cli upgrade --check-only --output json
```

CLI 版本变更后，会在后台尝试将 `kuaimai-shared`、`kuaimai-item` 同步到最新 Release（状态见 `~/.kuaimai-cli/skill-sync.json`）。也可手动：

```shell
kuaimai-cli skill install --if-stale
```

| 环境变量 | 作用 |
|----------|------|
| `KUAIMAI_CLI_SKIP_UPDATE_CHECK=1` | 禁用「新版本」stderr 提示 |
| `KUAIMAI_CLI_SKIP_SKILL_SYNC=1` | 禁用 CLI 版本变更后的 Skill 自动同步 |

升级完成后请 **重新打开终端** 并执行 `kuaimai-cli --version` 确认；必要时重开 Agent 会话以加载最新 Skill。

安装 Skill 后请 **重新打开 Agent 会话**。更多命令与使用说明，可查阅 [AGENTS.md](../AGENTS.md) 与 [系统架构与飞书对标说明](./系统架构与飞书对标说明.md)。
