# kuaimai-cli 内部分发与使用

面向**打包分享**的同事和**收到压缩包**的同事。无需 clone 仓库、无需 Go/Node 环境。

> **交给 Codex / Cursor Agent 安装？** 发送本文档即可；Agent 按**第二节**解压、选二进制、配置 PATH 并执行 `skill install-all --from ./skills`。

---

## 一、打包方：如何准备压缩包

### 1. 编译多平台二进制（在仓库根目录）

```bash
make dist
```

产物在 `dist/` 目录：

| 文件 | 适用环境 |
|------|----------|
| `kuaimai-cli-darwin-arm64` | Mac（Apple 芯片 M1/M2/M3 等） |
| `kuaimai-cli-darwin-amd64` | Mac（Intel） |
| `kuaimai-cli-linux-amd64` | Linux x86_64 |
| `kuaimai-cli-linux-arm64` | Linux ARM |
| `kuaimai-cli-windows-amd64.exe` | Windows |

### 2. 确认 Skills 目录

压缩包需包含 `skills/`（至少含 `kuaimai-shared`、`kuaimai-item`），与 `dist/` 同级，例如：

```text
kuaimai-cli-bundle/
├── dist/
│   ├── kuaimai-cli-darwin-arm64
│   ├── kuaimai-cli-darwin-amd64
│   └── ...
└── skills/
    ├── kuaimai-shared/
    └── kuaimai-item/
```

### 3. 打压缩包

在包含 `dist` 和 `skills` 的目录上一级执行，例如：

```bash
zip -r kuaimai-cli-bundle.zip dist skills
```

将 `kuaimai-cli-bundle.zip` 通过网盘、飞书等方式发给同事即可。

发给 Agent 用户时，可附提示词（路径按实际修改）：

```text
请按 docs/kuaimai-cli-内部分发与使用.md 第二节安装：
压缩包 ~/Downloads/kuaimai-cli-bundle.zip，解压到 ~/kuaimai-cli-bundle，配置 PATH 并 skill install-all --from ./skills；
验收 which kuaimai-cli、kuaimai-cli --help、kuaimai-cli skill list 通过即算安装完成。
```

---

## 二、接收方：安装（macOS）

以下以解压到 `~/kuaimai-cli-bundle` 为例，路径可按实际修改。

### 安装完成标准（做到这里即可）

在**任意新开的终端**里能直接执行：

```bash
which kuaimai-cli
kuaimai-cli --help
```

并且已执行过 **Skill 安装**（`skill install-all --from ./skills`）。  
满足以上两点即视为 **安装完成**。`config init`、`auth login`、查商品属于**首次用 ERP 时**再做（见下文第四节）。

---

### 1. 解压

```bash
cd ~
unzip ~/Downloads/kuaimai-cli-bundle.zip -d kuaimai-cli-bundle
cd ~/kuaimai-cli-bundle
```

### 2. 选对二进制并加入 PATH

先确认 Mac 芯片类型：

```bash
uname -m
```

按 `uname -m` 的结果，**只执行对应一组**命令（把二进制复制到 `~/bin/kuaimai-cli`）：

**Apple 芯片（`uname -m` 输出 `arm64`）**

```bash
mkdir -p ~/bin && cp dist/kuaimai-cli-darwin-arm64 ~/bin/kuaimai-cli && chmod +x ~/bin/kuaimai-cli
```

**Intel 芯片（`uname -m` 输出 `x86_64`）**

```bash
mkdir -p ~/bin && cp dist/kuaimai-cli-darwin-amd64 ~/bin/kuaimai-cli && chmod +x ~/bin/kuaimai-cli
```

**Apple 芯片 · 从解压目录起一行装完（含 PATH + Skill）**（在 `~/kuaimai-cli-bundle` 下执行）：

```bash
mkdir -p ~/bin && cp dist/kuaimai-cli-darwin-arm64 ~/bin/kuaimai-cli && chmod +x ~/bin/kuaimai-cli && xattr -d com.apple.quarantine ~/bin/kuaimai-cli 2>/dev/null || true && (grep -q 'export PATH="$HOME/bin:$PATH"' ~/.zshrc 2>/dev/null || echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc) && export PATH="$HOME/bin:$PATH" && kuaimai-cli skill install-all --from ./skills && which kuaimai-cli && kuaimai-cli --help
```

**Intel 芯片**：把上一行里的 `darwin-arm64` 换成 `darwin-amd64`。

> Linux 安装见本文档**第五节**；Agent 按 `uname -m` 选择 `dist/` 中对应二进制即可。

若 macOS 提示「无法验证开发者」或无法打开，可在「系统设置 → 隐私与安全性」中允许，或对刚复制的文件执行：

```bash
xattr -d com.apple.quarantine ~/bin/kuaimai-cli 2>/dev/null || true
```

**配置 PATH**（zsh，默认 shell）：

```bash
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

若未用上面「一行装完」，分步时在第 2 步之后执行：

```bash
which kuaimai-cli
kuaimai-cli --help
```

> **不想改 PATH？** 每次用绝对路径也可以，例如 `~/bin/kuaimai-cli --help`。在 Cursor Agent 里也可设置 `export KUAIMAI_CLI=~/bin/kuaimai-cli`。

### 3. 安装 Skills（安装流程最后一步）

在解压目录里执行（`skills` 与 `dist` 同级）：

```bash
cd ~/kuaimai-cli-bundle
kuaimai-cli skill install-all --from ./skills
```

**安装验收（与第二节开头标准一致）**：

```bash
which kuaimai-cli
kuaimai-cli --help
kuaimai-cli skill list --output json
```

Skill 会安装到 `~/.agents/skills/`。若使用 **Cursor / Claude 等 Agent**，安装后**重新打开 Agent 会话**，以便加载 `kuaimai-item`、`kuaimai-shared`。

> 有网络时也可 `kuaimai-cli skill install-all`（从 GitHub 拉取）；内网离线场景务必用 `--from ./skills`。

---

## 三、可选：首次调用 ERP 前

以下**不属于**压缩包安装完成条件；需要查改商品时再执行。

### 1. 初始化配置

```bash
kuaimai-cli config init
```

默认 API 地址为 `https://erp1.superboss.cc/`，一般无需修改。查看配置：

```bash
kuaimai-cli config get --output json
```

### 2. 登录 ERP

Token 需本人在浏览器获取，**不要**发给他人或写入 Git。

1. 打开 https://erp1.superboss.cc/ 并登录  
2. 浏览器开发者工具 → Network，刷新或随便点一个接口  
3. 在请求头里复制 `accessToken`  

```bash
kuaimai-cli auth login "<你的 accessToken>"
kuaimai-cli auth status --output json
kuaimai-cli auth check --output json
kuaimai-cli doctor --output json
```

### 3. 验证是否可调用商品接口

```bash
kuaimai-cli item +list \
  --body '{"title":"test","pageNo":1,"pageSize":1}' \
  --output json
```

返回 `"ok": true` 即表示 CLI、Skill、鉴权均正常。

---

## 四、常用命令速查

```bash
# 按标题查商品
kuaimai-cli item +list --body '{"title":"关键字","pageNo":1,"pageSize":50}' --output json

# 商品详情
kuaimai-cli item get-detail --sys-item-id <id> --output json

# 改标题（推荐 update-title，先 dry-run）
kuaimai-cli item update-title --sys-item-id <id> --title "新标题" --dry-run --verbose
```

商品域详细说明见：`~/.agents/skills/kuaimai-item/SKILL.md`。

---

## 五、Linux / Windows 简要说明

| 系统 | 二进制 | PATH 配置 |
|------|--------|-----------|
| Linux | `dist/kuaimai-cli-linux-amd64` 或 `linux-arm64` | 复制到 `~/bin/kuaimai-cli`，在 `~/.bashrc` 或 `~/.zshrc` 中加入 `export PATH="$HOME/bin:$PATH"` |
| Windows | `dist/kuaimai-cli-windows-amd64.exe` | 将 `dist` 加入系统「环境变量 → Path」，或在 PowerShell 中用完整路径调用 |

安装完成标准与 macOS 相同：`which kuaimai-cli`、`kuaimai-cli --help`、`skill install-all --from ./skills`。config / auth 为可选后续步骤。

---

## 六、安全提醒

- `accessToken` 仅通过 `kuaimai-cli auth login` 写入系统密钥链，**不要**写入 `config.yaml` 或聊天记录  
- 写操作（如 `item save`）建议先加 `--dry-run --verbose` 预览  
- 压缩包若经外网传输，注意公司数据安全规范  

---

## 七、排错

| 现象 | 处理 |
|------|------|
| `command not found: kuaimai-cli` | 检查 `~/bin` 是否在 PATH，`source ~/.zshrc` |
| Mac 无法执行二进制 | `chmod +x ~/bin/kuaimai-cli`，必要时去掉 quarantine（见上文） |
| Agent 不懂商品命令 | 执行 `skill install-all --from ./skills` 后重开 Agent |
| 提示未登录 | `kuaimai-cli auth status`，再 `auth login` |
| 选错架构（如 arm64 机器用了 amd64） | 按 `uname -m` 换对应 `dist/` 文件 |

更多细节见：[Agent 安装指南](./kuaimai-cli-agent-installation-guide.md)、[本地使用实操指南](./kuaimai-cli%20本地使用实操指南.md)。
