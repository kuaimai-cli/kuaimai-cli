# kuaimai-cli

快麦 ERP **商品（erp-items-core）** 命令行工具：查列表、看详情、改标题等。输出为结构化 JSON，适合脚本与 AI Agent 调用。架构与交互对标 [飞书 lark-cli](https://github.com/larksuite/lark-cli)。

| 资源 | 链接 |
|------|------|
| 最新版本 | [GitHub Releases](https://github.com/kuaimai-cli/kuaimai-cli/releases) |
| npm 包 | [`@kuaimai-cli/cli`](https://www.npmjs.com/package/@kuaimai-cli/cli) |

---

## 你能用它做什么

- 按标题、货号等条件 **搜索商品列表**
- 查看商品 **详情**（含 SKU）
- **修改商品标题**（支持 `--dry-run` 预览）
- 管理 **配置、鉴权、Skill**（供 Cursor 等 Agent 读取领域约定）

---

## 环境要求

- **操作系统**：macOS / Linux / Windows（`amd64` 或 `arm64`；Windows 暂无 `arm64` 包）
- **网络**：能访问快麦 ERP（默认 `https://erp1.superboss.cc/`）及 GitHub（安装时下载二进制）
- **安装方式任选其一**：
  - Node.js **16+**（推荐，支持 `npx` 一键安装）
  - 或直接从 Release 下载二进制（无需 Node）
  - 或 Go 1.22+（`go install`）

---

## 安装

### 方式一：npx 一键安装（推荐）

```bash
npx @kuaimai-cli/cli@latest install
```

安装向导会：全局安装 npm 包、下载对应平台的 Go 二进制、安装 Skill、初始化配置，并提示你完成登录。

非交互环境（CI / Agent）下会跳过交互，请按输出提示手动执行后续步骤。

### 方式二：下载 Release 二进制

1. 打开 [Releases](https://github.com/kuaimai-cli/kuaimai-cli/releases)
2. 下载 `kuaimai-cli-{version}-{os}-{arch}.tar.gz`（Windows 为 `.zip`）
3. 解压后将 `kuaimai-cli` 放入 `PATH`

```bash
# 示例（macOS arm64，版本号以 Release 页为准）
tar -xzf kuaimai-cli-0.1.0-darwin-arm64.tar.gz
sudo mv kuaimai-cli /usr/local/bin/
kuaimai-cli --version
```

### 方式三：go install

```bash
go install github.com/kuaimai-cli/kuaimai-cli@latest
```

确保 `$HOME/go/bin`（或 `$GOPATH/bin`）在 `PATH` 中。

---

## 首次使用（5 步）

按顺序执行，完成后即可调用商品接口。

### 1. 初始化配置

```bash
kuaimai-cli config init
```

默认 API 地址为 `https://erp1.superboss.cc/`。如需修改：

```bash
kuaimai-cli config set api.url "https://erp1.superboss.cc/"
kuaimai-cli config get --output json
```

### 2. 登录（需要 ERP accessToken）

Token **不会**写入配置文件，仅保存在系统密钥链。

1. 浏览器打开 ERP 并登录
2. 打开开发者工具 → **Network**
3. 刷新或发起任意 API 请求，在请求头中找到 `accessToken`
4. 复制 token 值（勿提交到 Git、勿发给他人）

```bash
kuaimai-cli auth login "<你的 accessToken>"
```

多账号（可选）：

```bash
kuaimai-cli auth login "<token>" --profile 账号A
kuaimai-cli auth use 账号A
kuaimai-cli auth list
```

### 3. 安装 Skill（使用 Cursor / Agent 时建议）

让 Agent 了解商品域字段与命令约定：

```bash
kuaimai-cli skill install-all
```

已 clone 本仓库时，也可从本地安装：

```bash
kuaimai-cli skill install-all --from ./skills
```

### 4. 自检

```bash
kuaimai-cli auth check --output json
kuaimai-cli doctor --output json
```

`doctor` 返回 `ready: true` 表示配置、鉴权、PATH、Skill 均已就绪。

### 5. 试一条商品查询

```bash
kuaimai-cli item +list \
  --body '{"title":"关键字","pageNo":1,"pageSize":10}' \
  --output json
```

返回 `{"ok":true,"data":...}` 即表示链路正常。

---

## 常用命令

### 商品

```bash
# 按标题搜索（分页）
kuaimai-cli item +list \
  --body '{"title":"关键字","pageNo":1,"pageSize":20}' \
  --output json

# 自动翻页拉全量（识别 body 中的 pageNo / pageSize）
kuaimai-cli item +list \
  --body '{"title":"关键字","pageNo":1,"pageSize":50}' \
  --page-all --output json

# 商品详情
kuaimai-cli item get-detail --sys-item-id <商品ID> --output json

# 改标题（先预览，确认后再去掉 --dry-run）
kuaimai-cli item update-title \
  --sys-item-id <商品ID> \
  --title "新标题" \
  --dry-run

kuaimai-cli item update-title \
  --sys-item-id <商品ID> \
  --title "新标题"
```

### 配置与鉴权

```bash
kuaimai-cli config get --output json
kuaimai-cli auth status --output json
kuaimai-cli auth check --output json
kuaimai-cli auth logout
```

### Skill

```bash
kuaimai-cli skill list --output json
kuaimai-cli skill install kuaimai-item
kuaimai-cli skill install-all
```

### 其它

```bash
kuaimai-cli --version
kuaimai-cli upgrade --output json    # 检查 GitHub 是否有新版本
kuaimai-cli doctor --output json
```

---

## 输出格式

- **stdout**：结构化数据，默认 JSON 信封 `{ "ok": true|false, "data": ..., "error": ..., "hint": ... }`
- **stderr**：`--verbose` 时的日志与友好错误提示（请勿把日志混入 stdout 管道）

```bash
# 表格输出
kuaimai-cli item +list --body '{"pageNo":1,"pageSize":5}' --output table

# 列表导出 CSV / NDJSON
kuaimai-cli item +list --body '{"pageNo":1,"pageSize":100}' --output csv
kuaimai-cli item +list --body '{"pageNo":1,"pageSize":100}' --output ndjson
```

---

## 升级

```bash
# 查看是否有新版本
kuaimai-cli upgrade --output json

# 升级到最新（与飞书类似：重新安装）
npx @kuaimai-cli/cli@latest install
# 或
npm install -g @kuaimai-cli/cli@latest
```

也可从 [Releases](https://github.com/kuaimai-cli/kuaimai-cli/releases) 下载新版本二进制覆盖安装。

---

## 常见问题

| 现象 | 处理 |
|------|------|
| `auth check` 失败 / 401 | 重新 `auth login`；确认 token 未过期 |
| `doctor` 中 Skill 未就绪 | 执行 `kuaimai-cli skill install-all` |
| `npm install -g` / curl 失败 | `postinstall` **仅从 GitHub Release 下载**（与飞书相同机制）；网络受限时用代理或手动下载 Release |
| checksum 校验失败 | 重打 tag 后须重新 `npm publish`，使包内 `checksums.txt` 与 Release 一致 |
| 写操作不敢直接执行 | 所有写接口先加 `--dry-run` 预览请求体 |

---

## 给 AI Agent 使用者

- 安装与鉴权详见 [Agent 安装指南](docs/kuaimai-cli-agent-installation-guide.md)
- Agent 行为约定见仓库根目录 [AGENTS.md](AGENTS.md)
- 安装后 Skill 位于 `~/.agents/skills/kuaimai-item/`（由 `skill install-all` 写入）

---

## 安全提示

- `accessToken` **仅**通过 `auth login` 写入系统密钥链，不会出现在配置文件或日志中
- 写操作（`item save`、`item update-title`）请先用 **`--dry-run`** 预览
- 在测试环境验证后再于生产环境批量改数

---

## 更多文档

| 文档 | 说明 |
|------|------|
| [文档索引](docs/README.md) | 全部文档入口 |
| [本地使用实操指南](docs/kuaimai-cli%20本地使用实操指南.md) | 配置、鉴权、改标题完整流程 |
| [npx 分发与安装](docs/kuaimai-cli-npx-分发与安装.md) | npm 发版、镜像、与飞书对标说明 |
| [Agent 安装指南](docs/kuaimai-cli-agent-installation-guide.md) | Cursor / Codex 等 IDE |
| [内部分发与使用](docs/kuaimai-cli-内部分发与使用.md) | 离线压缩包分发 |
| [开发白皮书](docs/kuaimai-cli%20开发文档.md) | 架构与规划（维护者） |

---

## 参与开发

```bash
git clone https://github.com/kuaimai-cli/kuaimai-cli.git
cd kuaimai-cli
make build
go test ./...
go test ./tests/cli_e2e/...
```

---

## License

MIT
