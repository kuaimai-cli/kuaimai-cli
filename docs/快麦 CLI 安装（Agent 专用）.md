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

从 GitHub 拉取 **整目录**（`SKILL.md` + `references/` 工作流文档），安装到各 Agent 目录（`~/.agents/skills`、`~/.cursor/skills` 等）。

**通常无需手动执行**：任意 `kuaimai-cli` 命令结束后会在后台自动同步（未安装 / Release 更新 / CLI 版本变化时，24h 缓存）。首次安装或需立即拉取时可手动：

```shell
kuaimai-cli skill install
```

（无参数时仅在需要时更新；强制覆盖请加 `--force`）

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
# 推荐：一键升级（npm 全局包 + Go 二进制 + Skills）
kuaimai-cli upgrade

# 或重新走安装向导（0.1.8+ 会在全局版本偏旧时自动升级，不再「有包就跳过」）
npx @kuaimai-cli/cli@latest install

# 仅检查、不安装
kuaimai-cli upgrade --check-only --output json
```

**说明**：`~/.kuaimai-cli/` 仅存配置与缓存，**不会**阻止 CLI 升级。若 `npx install` 曾提示「已安装 (v0.1.0)，跳过」而版本未变，请用上面 `upgrade` / 重装；`0.1.8` 起向导会对比 npm 包版本并自动 `npm install -g @kuaimai-cli/cli@<目标版本>`。

CLI 会在后台自动将 `kuaimai-shared`、`kuaimai-item` 同步到最新 Release（状态见 `~/.kuaimai-cli/skill-sync.json`）。手动检查/触发：

```shell
kuaimai-cli skill install          # 仅在需要时更新（默认）
kuaimai-cli skill install --force  # 强制覆盖重装
```

| 环境变量 | 作用 |
|----------|------|
| `KUAIMAI_CLI_SKIP_UPDATE_CHECK=1` | 禁用「新版本」stderr 提示 |
| `KUAIMAI_CLI_SKIP_SKILL_SYNC=1` | 禁用 CLI 版本变更后的 Skill 自动同步 |
| `KUAIMAI_CLI_FORCE_INSTALL=1` | 安装向导强制重装全局包与 Skills（忽略「已安装」跳过） |

升级完成后请 **重新打开终端** 并执行 `kuaimai-cli --version` 确认；必要时重开 Agent 会话以加载最新 Skill。

### 升级仍异常时的兜底（慎用）

仅当 `upgrade` / `npx install` 后 `kuaimai-cli --version` 仍不对，或 PATH 指向旧二进制时：

```shell
npm install -g @kuaimai-cli/cli@latest
hash -r   # zsh：刷新命令缓存
which kuaimai-cli
kuaimai-cli --version
kuaimai-cli skill install
```

**不要**删除 `~/.kuaimai-cli/`（会丢失 `config.yaml` 与 token）。无需 `skill install all`（无此参数；无参数即安装默认 Skills）。

国内若 `npm install` 因 GitHub 超时失败，发版后 `install.js` 会尝试 npmmirror 镜像；镜像须维护者在 [npmmirror 二进制镜像](./npmmirror-二进制镜像.md) 完成 cnpmcore 注册。临时可用代理或本地 `make build`。

安装 Skill 后请 **重新打开 Agent 会话**。更多命令与使用说明，可查阅 [AGENTS.md](../AGENTS.md) 与 [系统架构与飞书对标说明](./系统架构与飞书对标说明.md)。
