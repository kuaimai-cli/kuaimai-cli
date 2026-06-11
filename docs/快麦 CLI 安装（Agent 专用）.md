# 快麦 CLI 安装指南

以下步骤面向 AI Agent。登录环节需要用户向管理员申请并提供 `accessToken`。

## 环境要求

- Node.js 16+（含 npm/npx）

## 第 1 步 安装 CLI 与 Skills（一条命令）

```shell
npx @kuaimai-cli/cli@latest install
```

向导会依次：全局安装 CLI → 从 npm 包内 `skills/` 复制到各 Agent 目录 → `config init` → 提示登录。

## 第 2 步 Skills 说明（通常无需手动执行）

Skills（`SKILL.md` + `references/`）**随 npm 包 `@kuaimai-cli/cli` 一起发布**，安装向导会从包内复制到 `~/.agents/skills`、`~/.cursor/skills` 等目录。**不需要 GitHub Token**。

`kuaimai-cli skill install` 同样**优先**使用 npm 包或本地仓库内的 `skills/`；仅在没有 bundled 源时才回退 GitHub。

**通常无需手动执行**：任意 `kuaimai-cli` 命令结束后会在后台自动同步 Skill（未安装 / CLI 版本变化时）。首次安装或需立即覆盖时可手动：

```shell
kuaimai-cli skill install          # 仅在需要时更新
kuaimai-cli skill install --force  # 强制覆盖重装
```

默认安装三个 Skill：

| Skill | 版本 | 职责 |
|-------|------|------|
| `kuaimai-shared` | v1.1.0 | auth、输出、**registry 发现流程**（capabilities → schema → web call） |
| `kuaimai-item` | v3.0.0 | 商品域**意图路由** + item shortcuts 工作流 |
| `kuaimai-scm` | v2.0.0 | 供应链域**意图路由** + `web call scm.*` 工作流 |

安装后目录结构示例：

```text
~/.agents/skills/kuaimai-item/
├── SKILL.md
└── references/
    ├── kuaimai-item-list.md
    ├── kuaimai-item-count.md
    ├── kuaimai-item-get-detail.md
    ├── kuaimai-item-update-title.md
    ├── kuaimai-item-save.md
    ├── kuaimai-item-meta-execution.md
    ├── kuaimai-item-web-call.md
    ├── kuaimai-item-count-dimensions.md
    ├── kuaimai-item-query-count.md
    └── kuaimai-item-query-list-v2.md

~/.agents/skills/kuaimai-scm/
├── SKILL.md
└── references/
    ├── kuaimai-scm-domain-routing.md
    ├── kuaimai-scm-meta-execution.md
    ├── kuaimai-scm-web-call.md
    ├── kuaimai-scm-staff.md
    ├── kuaimai-scm-logging.md
    ├── kuaimai-scm-item-base.md
    └── kuaimai-scm-dsb.md
```

**Agent 使用约定（对标飞书 lark-shared + lark-calendar）**：

1. **任何域请求前**先 Read `kuaimai-shared/SKILL.md`（auth、registry 流程、输出与安全）
2. **接口发现**走 registry，**不要在 Skill 里猜 apiId**：
   ```shell
   kuaimai-cli capabilities --output json
   kuaimai-cli schema <apiId> --output json
   kuaimai-cli web call <apiId> --params/--data/--body '...' --output json
   ```
3. **商品域**：Read `kuaimai-item/SKILL.md`；有 shortcut 优先 `item +list` 等；写操作须先 Read 对应 `references/`
4. **供应链域**：Read `kuaimai-scm/SKILL.md`；统一 `web call scm.<operation>`（无 shortcuts）
5. 全量翻页：`--page-all`；Agent 续查 `--page-confirm yes`；限条数 `--page-limit`

## 第 3 步 初始化配置

```shell
kuaimai-cli config init
```

默认 `api.url` 为 `https://erp1.superboss.cc/`。registry 源默认：

```yaml
registry:
  source: "http://open-cli.kuaimai.com/registry/registry.json"
  auto_sync: true
```

## 第 4 步 账号登录

本步骤需要用户提供 `accessToken`。Agent 应先提示用户：**如尚未持有 accessToken，请联系 ERP 管理员申请分配**；待用户提供后再执行：

```shell
kuaimai-cli auth login "<accessToken>"
```

## 第 5 步 状态验证

```shell
kuaimai-cli auth status
kuaimai-cli doctor --output json
kuaimai-cli registry sync --output json
kuaimai-cli capabilities --output json
```

`doctor` 检查 config、auth、PATH、**registry 缓存**、Skill（含 `references/`）。`ready: true` 表示环境就绪。

跳过 registry 自动同步（调试）：`KUAIMAI_CLI_SKIP_REGISTRY_SYNC=1`

## 第 6 步 试一条命令

```shell
# registry 测试接口（远端已发布时）
kuaimai-cli web call api.luotao.test.get \
  --params '{"keyword":"测试"}' --output json

# 商品 shortcut（需有效 token + ERP 可达）
kuaimai-cli item +list \
  --body '{"title":"关键字","pageNo":1,"pageSize":10}' --output json
```

## 第 7 步 升级与 Skill 同步（对标飞书 lark-cli）

```shell
kuaimai-cli upgrade
kuaimai-cli upgrade --check-only --output json
npx @kuaimai-cli/cli@latest install
kuaimai-cli skill install --force
```

| 环境变量 | 作用 |
|----------|------|
| `KUAIMAI_CLI_SKIP_REGISTRY_SYNC=1` | 禁用 registry 自动同步 |
| `KUAIMAI_CLI_SKIP_UPDATE_CHECK=1` | 禁用新版本 stderr 提示 |
| `KUAIMAI_CLI_SKIP_SKILL_SYNC=1` | 禁用 Skill 自动同步 |
| `KUAIMAI_CLI_FORCE_INSTALL=1` | 安装向导强制重装 |

升级完成后请 **重新打开 Agent 会话** 以加载最新 Skill。

更多说明：[AGENTS.md](../AGENTS.md) · [系统架构与飞书对标说明](./系统架构与飞书对标说明.md) · [Registry 远端同步说明](./Registry远端同步说明.md)
