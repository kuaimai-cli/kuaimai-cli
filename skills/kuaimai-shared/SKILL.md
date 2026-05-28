---
name: kuaimai-shared
version: 1.0.0
description: "快麦 kuaimai-cli 全局约定：首次安装、config、auth login、输出信封解析、Skill 安装、安全与排错。用户要配置 CLI、登录 ERP、看不懂 ok/data/error、或尚未安装 Skill 时使用。"
metadata:
  requires:
    bins: ["kuaimai-cli"]
---

# kuaimai-shared

本 Skill 覆盖 **kuaimai-cli 全局限定**。商品查询/改标题等 **item 域** 操作见 [`kuaimai-item`](../kuaimai-item/SKILL.md)。

## CLI 可执行文件

按顺序选用（第一个可用即可）：

1. 环境变量 `KUAIMAI_CLI`
2. 当前目录 `./kuaimai-cli`（在 kuaimai-cli 仓库根目录开发时）
3. `PATH` 中的 `kuaimai-cli`

下文命令均以 `kuaimai-cli` 表示上述可执行文件。Agent **必须在终端真实执行**，不要手写 curl / URL。

## 首次使用（Agent 编排）

```bash
# 1. 安装 CLI（任选其一）
npx @kuaimai/cli@latest install
# 或: go install github.com/kuaimai/kuaimai-cli@latest

# 2. 安装 Skills（商品域必读 kuaimai-item）
kuaimai-cli skill install-all
# 或单个: kuaimai-cli skill install kuaimai-shared
#         kuaimai-cli skill install kuaimai-item

# 3. 初始化配置
kuaimai-cli config init

# 4. 登录（须用户提供 accessToken，Agent 不可代填）
kuaimai-cli auth login "<accessToken>"

# 5. 验证
kuaimai-cli auth status --output json
```

安装 Skill 后 **重新打开 Agent 会话**，以便加载最新 `SKILL.md`。

本地已 clone 仓库时：

```bash
kuaimai-cli skill install-all --from ./skills
```

## 配置

| 项 | 说明 |
|----|------|
| 配置文件 | `~/.kuaimai-cli/config.yaml` |
| 默认 API | `api.url` = `https://erp1.superboss.cc/` |
| 修改 | `kuaimai-cli config set api.url "<url>"` |
| 查看 | `kuaimai-cli config get --output json` |

## 鉴权

```bash
kuaimai-cli auth login "<accessToken>"   # 写入系统密钥链
kuaimai-cli auth status --output json    # 是否已登录
kuaimai-cli auth logout                  # 退出
```

**Agent 不能代用户获取 token。** 引导用户：

1. 打开 `https://erp1.superboss.cc/` 并登录
2. 浏览器 DevTools → Network，刷新或发起任意 API 请求
3. 从请求头复制 `accessToken`（勿写入 Git、勿写入 `config.yaml`）

未登录时 CLI 会提示先执行 `auth login`；应先 `auth status`，再引导用户登录，不要反复盲试业务命令。

## 输出约定

**stdout**（结构化数据，禁止把日志打到 stdout）：

| `--output` | 行为 |
|------------|------|
| `json`（推荐 Agent 使用） | 完整信封 `{ok, data, error, hint}` |
| `table` | 默认人类可读表格 |
| `csv` / `ndjson` | 列表**成功**时为裸数据流；**失败**仍为 JSON 信封 |

解析成功响应：

- `ok === true` → 业务数据在 `data`
- `ok === false` → 读 `error` 与 `hint`，向用户说明原因

**stderr**：`--verbose` 调试日志、友好错误文案。

常用全局参数：

| 参数 | 用途 |
|------|------|
| `--output json` | Agent 解析结果 |
| `--no-color` | 避免 ANSI 干扰管道 |
| `--verbose` | 排错、配合 `--dry-run` 看请求预览 |
| `--dry-run` | 写操作试跑，不真正提交 |
| `--page-all` | 列表自动翻页（body 含 `pageNo`/`pageSize` 时） |

## Skill 命令

```bash
kuaimai-cli skill list --output json
kuaimai-cli skill install <name>              # 单个，默认 GitHub
kuaimai-cli skill install-all                 # kuaimai-shared + kuaimai-item
```

已安装路径：`~/.agents/skills/<name>/SKILL.md`。仓库开发时也可直接读 `skills/<name>/SKILL.md`。

## 安全

- **禁止**将 `accessToken` 写入配置文件、代码仓库或对话外的持久化存储
- **写操作**（如 `item save`）先 `--dry-run --verbose`，用户确认后再去掉 `--dry-run`
- 日志与 dry-run 预览中的敏感字段已脱敏；仍避免向用户重复粘贴完整 token

## 排错

| 现象 | 处理 |
|------|------|
| 提示未登录 / 401 | `auth status` → 引导 `auth login` |
| `ok: false` 且无网络错误 | 读 `error`、`hint`；必要时加 `--verbose` 重试 |
| 命令找不到 | 检查 `KUAIMAI_CLI` / PATH / `npx @kuaimai/cli` 是否已安装 |
| Skill 行为不符合预期 | `skill install-all` 或 `skill install kuaimai-item`，并重开 Agent |

## 域 Skill 路由

| 用户意图 | Skill |
|----------|-------|
| 登录、配置、输出格式、安装 CLI/Skill | **本 Skill** |
| 商品列表、按标题统计、详情、改标题 | [`kuaimai-item`](../kuaimai-item/SKILL.md) |
