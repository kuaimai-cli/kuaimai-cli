---
name: kuaimai-shared
version: 1.0.0
description: "快麦 kuaimai-cli 全局约定：首次安装、config、auth login、输出信封解析、Skill 安装、安全与排错。用户要配置 CLI、登录 ERP、看不懂 ok/data/error、或尚未安装 Skill 时使用。"
metadata:
  requires:
    bins: ["kuaimai-cli"]
  cliHelp: "kuaimai-cli --help"
---

# kuaimai-cli 共享规则

本技能指导你如何通过 kuaimai-cli 操作快麦 ERP 资源，以及有哪些注意事项。商品查询、改标题等 **item 域** 见 [`kuaimai-item`](../kuaimai-item/SKILL.md)；供应链、铺货日志等 **scm 域** 见 [`kuaimai-scm`](../kuaimai-scm/SKILL.md)。

## CLI 可执行文件

按顺序选用（第一个可用即可）：

1. 环境变量 `KUAIMAI_CLI`
2. 当前目录 `./kuaimai-cli`（在 kuaimai-cli 仓库根目录开发时）
3. `PATH` 中的 `kuaimai-cli`

下文命令均以 `kuaimai-cli` 表示上述可执行文件。Agent **必须在终端真实执行**，不要手写 curl 或拼接 ERP URL。

## 配置初始化

首次使用需运行 `kuaimai-cli config init` 完成基础配置。

```bash
kuaimai-cli config init
kuaimai-cli config get --output json
kuaimai-cli config set api.url "<url>"
```

| 项 | 说明 |
|----|------|
| 配置文件 | `~/.kuaimai-cli/config.yaml` |
| 默认 API | `api.url` = `https://erp1.superboss.cc/`（**item 域**） |
| scm 域 | `service scm *` 自动使用 `https://scm.superboss.cc/`（meta `baseUrl`），无需改 `api.url` |

完整安装步骤见仓库 [快麦 CLI 安装（Agent 专用）](https://github.com/kuaimai-cli/kuaimai-cli/blob/main/docs/快麦%20CLI%20安装（Agent%20专用）.md)。

## 认证

快麦 CLI 使用用户提供的 `accessToken` 鉴权（写入系统密钥链），**无 OAuth 交互式登录**。

```bash
kuaimai-cli auth login "<accessToken>"              # 写入密钥链
kuaimai-cli auth login "<accessToken>" --profile p  # 多账号
kuaimai-cli auth use <profile>                      # 切换当前 profile
kuaimai-cli auth list --output json
kuaimai-cli auth status --output json
kuaimai-cli auth check --output json                # token + API 探针
kuaimai-cli auth logout
```

**Agent 不能代用户获取或填写 token。** 引导用户：

1. 如尚未持有 `accessToken`，**请联系 ERP 管理员申请分配**
2. 待用户提供 token 后，再执行 `auth login`（勿写入 Git、勿写入 `config.yaml`）

未登录时 CLI 会提示先执行 `auth login`；应先 `auth status`，再引导用户登录，不要反复盲试业务命令。

## 输出约定

**stdout** 输出结构化数据，**禁止**把日志打到 stdout：

| `--output` | 行为 |
|------------|------|
| `json`（推荐 Agent 使用） | 完整信封 `{ok, data, error, hint}` |
| `table` | 默认人类可读表格 |
| `csv` / `ndjson` | 列表**成功**时为裸数据流；**失败**仍为 JSON 信封 |

解析响应：

- `ok === true` → 业务数据在 `data`
- `ok === false` → 读 `error` 与 `hint`，向用户说明原因；必要时加 `--verbose` 重试

**stderr**：`--verbose` 调试日志、友好错误文案（无 Go 堆栈）。

优先级：`--output` > `config cli.output` > `table`

## 全局参数

| 参数 | 用途 |
|------|------|
| `--output json` | Agent 解析结果 |
| `--no-color` | 避免 ANSI 干扰管道 |
| `--verbose` | 排错、配合 `--dry-run` 看请求预览 |
| `--dry-run` | 写操作试跑，不真正提交 |
| `--page-all` | 列表自动翻页（仅 `pageable:true` 接口；识别 body 中 `pageNo`/`pageSize`，默认 50 条/页） |
| `--page-limit` | 与 `--page-all` 配合：最大拉取条数（0=不限条数，仍受 1000 页硬上限） |
| `--page-confirm` | 达阈值（500 条/预估>1000）时的策略：`prompt`（交互 Y/N）\|`yes`（自动继续）\|`no`（静默停止） |

## 安装与更新

```bash
# 安装 CLI + Skills（推荐；Skills 从 npm 包内置目录复制，无需 GitHub Token）
npx @kuaimai-cli/cli@latest install
# 或: go install github.com/kuaimai-cli/kuaimai-cli@latest

# 单独重装 Skills（优先 bundled，回退 GitHub）
kuaimai-cli skill install --force

# 自检与版本
kuaimai-cli doctor --output json
kuaimai-cli upgrade                # 默认：有新版则 npm 升级 + Skills 同步
kuaimai-cli upgrade --check-only --output json
kuaimai-cli skill install --if-stale
kuaimai-cli schema --output json   # 全量 meta（item + scm）
```

已安装路径：`~/.agents/skills/<name>/`（同时写入 `~/.cursor/skills` 等 Agent 目录）。`skill install` **覆盖写入**对应目录，无需手删缓存。任意命令后若有新版会在 **stderr** 提示（24h 缓存）。安装或升级 Skill 后 **重新打开 Agent 会话** 以加载最新内容。

## --page-all 安全（Agent 必读）

- 仅对 meta 中 `pageable:true` 的接口生效（如 `stock-list`、`item-query-list-v2`）
- 用户未明确要求「全部/导出/所有页」时，**不要**加 `--page-all`
- CLI 与 Agent **双层防护**（阈值已对齐）：
  - **500 条**已拉取或接口预估 **>1000 条**时，CLI 在交互终端弹出 `[y/N]` 续查确认
  - 非交互环境（Agent 管道）默认停止并返回已拉取数据；需继续时加 `--page-confirm yes`
  - `--page-limit N` 硬限制最大条数；`--page-confirm no` 在阈值处静默停止
  - 硬上限仍为 **1000 页**；大数据量用 `--output ndjson` / `csv`
- 详见 [`kuaimai-item` references/kuaimai-item-meta-execution.md](../kuaimai-item/references/kuaimai-item-meta-execution.md)

## 安全规则

- **禁止**将 `accessToken` 写入配置文件、代码仓库或对话外的持久化存储
- **写入/修改操作前必须确认用户意图**
- 写操作先 `--dry-run --verbose` 预览，用户确认后再去掉 `--dry-run`
- 日志与 dry-run 预览中的敏感字段已脱敏；仍避免向用户重复粘贴完整 token

## 错误与排错

| 现象 | 处理 |
|------|------|
| 提示未登录 / 401 | `auth status` → 提示用户向 ERP 管理员申请 accessToken → 引导 `auth login` |
| `ok: false` 且无网络错误 | 读 `error`、`hint`；必要时加 `--verbose` 重试 |
| 命令找不到 | 检查 `KUAIMAI_CLI` / PATH / `npx @kuaimai-cli/cli` 是否已安装 |
| Skill 行为不符合预期 | `skill install --force` 或 `npx @kuaimai-cli/cli@latest install` 覆盖重装，并重开 Agent |
| CLI 版本过旧 | `kuaimai-cli upgrade`（默认一键）；或 `npx @kuaimai-cli/cli@latest install` |

## 域 Skill 路由

| 用户意图 | Skill |
|----------|-------|
| 登录、配置、输出格式、安装 CLI/Skill | **本 Skill** |
| 商品列表、按标题统计、详情、改标题 | [`kuaimai-item`](../kuaimai-item/SKILL.md) |
| 供应链、铺货日志、scm 商品、平台配置 | [`kuaimai-scm`](../kuaimai-scm/SKILL.md) |
| 维护者文档 | 仓库 [`docs/README.md`](../../docs/README.md) |
