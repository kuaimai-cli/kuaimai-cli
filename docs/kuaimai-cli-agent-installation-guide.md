# kuaimai-cli Agent 安装指南

以下步骤面向 **AI Agent**。部分步骤需要用户在浏览器中手动完成（获取 `accessToken`）。

配套文档：[本地使用实操指南](./kuaimai-cli%20本地使用实操指南.md) · [AGENTS.md](../AGENTS.md)

---

## 前置条件

安装前请确认环境具备：

- macOS / Linux / Windows（amd64 或 arm64）
- 以下任选其一：
  - **Node.js**（npm/npx，用于 `npx @kuaimai-cli/cli install`）
  - **Go 1.22+**（用于 `go install`）
  - 或从 [GitHub Releases](https://github.com/kuaimai-cli/kuaimai-cli/releases) 下载预编译二进制

---

## Step 1：安装 CLI

### 方式 A — npm 一键安装（推荐，对标飞书）

```shell
npx @kuaimai-cli/cli@latest install
```

非交互环境（Agent 调用）时，向导会输出后续需用户手动完成的步骤。

### 方式 B — go install

```shell
go install github.com/kuaimai-cli/kuaimai-cli@latest
```

确保 `$GOPATH/bin` 或 `$HOME/go/bin` 在 `PATH` 中。

### 方式 C — 下载 Release 二进制

从 [GitHub Releases](https://github.com/kuaimai-cli/kuaimai-cli/releases) 下载对应平台的 `kuaimai-cli-{version}-{os}-{arch}.tar.gz`，解压后将 `kuaimai-cli` 放入 `PATH`。

### 方式 D — 内部分发压缩包（Codex / 离线）

收到 `kuaimai-cli-bundle.zip`（含 `dist/` + `skills/`）时，按 [内部分发与使用](./kuaimai-cli-内部分发与使用.md) **第二节** 解压、加入 PATH 并安装 Skill。

---

## Step 2：安装 Skills（Agent 必读）

Skills 让 Agent 了解商品域字段与命令约定。

### 从 GitHub 安装（无需 clone 仓库）

```shell
kuaimai-cli skill install-all
```

从 GitHub 默认仓库安装 `kuaimai-shared`、`kuaimai-item` 到 `~/.agents/skills/`。

安装单个 Skill：

```shell
kuaimai-cli skill install kuaimai-item
kuaimai-cli skill install kuaimai-shared
```

指定仓库：

```shell
kuaimai-cli skill install-all --repo kuaimai-cli/kuaimai-cli
```

### 从本地仓库（已 clone 时）

```shell
kuaimai-cli skill install-all --from ./skills
kuaimai-cli skill install kuaimai-item --from ./skills
```

验证：

```shell
kuaimai-cli skill list --output json
```

---

## Step 3：配置

Agent 执行：

```shell
kuaimai-cli config init
```

默认 `api.url` 为 `https://erp1.superboss.cc/`。如需修改：

```shell
kuaimai-cli config set api.url "https://erp1.superboss.cc/"
kuaimai-cli config get --output json
```

---

## Step 4：登录（需用户手动提供 Token）

**Agent 不能代填 token。** 请引导用户从 ERP 浏览器获取 `accessToken` 后执行：

1. 打开 `https://erp1.superboss.cc/` 并登录
2. 打开浏览器 DevTools → Network
3. 刷新或发起任意 API 请求，在请求头中找到 `accessToken`
4. 复制 token 值（勿提交到 Git、勿写入配置文件）

用户执行（将 `<accessToken>` 替换为真实值）：

```shell
kuaimai-cli auth login "<accessToken>"
```

---

## Step 5：验证

```shell
kuaimai-cli auth status --output json
kuaimai-cli auth check --output json
kuaimai-cli doctor --output json

kuaimai-cli item +list \
  --body '{"title":"test","pageNo":1,"pageSize":1}' \
  --output json
```

返回 `{"ok":true,...}` 表示安装与鉴权成功。`doctor` 的 `ready: true` 表示 config、鉴权、PATH、Skill 均已就绪。

---

## Agent 常用命令速查

```shell
# 按标题查商品
kuaimai-cli item +list --body '{"title":"关键字","pageNo":1,"pageSize":50}' --output json

# 商品详情
kuaimai-cli item get-detail --sys-item-id <id> --output json

# 改标题（推荐；写操作先 dry-run）
kuaimai-cli item update-title --sys-item-id <id> --title "新标题" --dry-run --verbose

# 多账号
kuaimai-cli auth login "<token>" --profile prod
kuaimai-cli auth use prod
```

详见 `~/.agents/skills/kuaimai-item/SKILL.md` 或仓库 `skills/kuaimai-item/SKILL.md`。

---

## 非 TTY 环境提示

若 Agent 在无交互终端中运行 `npx @kuaimai-cli/cli install`，向导会跳过交互步骤并输出：

```text
kuaimai-cli config init
kuaimai-cli auth login "<accessToken>"
```

请将上述命令与 token 获取说明转发给用户完成。

---

## 安全提醒

- Token 仅通过 `auth login` 写入系统密钥链，**禁止**写入 `config.yaml` 或代码仓库
- 写操作（`item save` 等）建议先用 `--dry-run --verbose` 预览
- 日志与 dry-run 预览中的敏感字段已脱敏
