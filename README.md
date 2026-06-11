# kuaimai-cli

快麦 ERP **商品（erp-items-core）** 与 **供应链（erp-scm）** 命令行工具：查商品、改标题、查铺货日志、供应链商品等。输出为结构化 JSON，适合脚本与 AI Agent 调用。架构与交互对标 [飞书 lark-cli](https://github.com/larksuite/lark-cli)。

**能力快照**：远端 **registry.json**（[open-cli.kuaimai.com](http://open-cli.kuaimai.com/registry/registry.json)）· 每次命令前自动同步 · `capabilities` / `schema` / `web call` · **6** 个 item shortcuts · 分页防护 `--page-all` / `--page-limit` / `--page-confirm`。

| 资源 | 链接 |
|------|------|
| 最新版本 | [GitHub Releases](https://github.com/kuaimai-cli/kuaimai-cli/releases) |
| npm 包 | [`@kuaimai-cli/cli`](https://www.npmjs.com/package/@kuaimai-cli/cli) |

---

## 你能用它做什么

- 按标题、货号等条件 **搜索商品列表**
- 查看商品 **详情**（含 SKU）
- **修改商品标题**（支持 `--dry-run` 预览）
- 查询 **供应链员工、铺货/操作日志、供应链商品、平台铺货配置**（`web call scm.<operation>`）
- 管理 **配置、鉴权、Skill**（供 Cursor 等 Agent 读取领域约定）
- 通过 **`capabilities` / `schema` / `web call`** 发现并调用远端 registry 已发布接口（无需手写 URL）

---

## 环境要求

- **操作系统**：macOS / Linux / Windows（`amd64` 或 `arm64`；Windows 暂无 `arm64` 包）
- **网络**：能访问快麦 ERP（商品默认 `https://erp1.superboss.cc/`；供应链 `https://scm.superboss.cc/`）及 GitHub（安装时下载二进制）
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

`accessToken` 须由本人向 **ERP 管理员申请分配**（勿写入 Git、勿发给他人）。拿到 token 后执行：

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

让 Agent 了解商品与供应链域命令约定与工作流（对标飞书 lark-* Skill 结构）：

```bash
kuaimai-cli skill install
```

从 npm 包安装 `kuaimai-shared` v1.1 + `kuaimai-item` v3 + `kuaimai-scm` v2 **整目录**（含 `references/` 工作流文档），同时写入 `~/.agents/skills`、`~/.cursor/skills` 等 5 个 Agent 目录。仓库内开发时可直接读 `skills/kuaimai-item/SKILL.md` 或 `skills/kuaimai-scm/SKILL.md`。

### 4. 自检（含 registry 自动同步）

```bash
kuaimai-cli auth check --output json
kuaimai-cli doctor --output json
```

首次执行任意业务命令时，CLI 会自动从 `http://open-cli.kuaimai.com/registry/registry.json` 拉取并缓存 registry（对标飞书 lark-cli 的 OpenAPI registry 刷新）。也可手动同步：

```bash
kuaimai-cli registry sync --output json
kuaimai-cli capabilities --output json
```

`doctor` 返回 `ready: true` 表示配置、鉴权、registry、PATH、Skill（含 `references/`）均已就绪。

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

# 按标题统计数量
kuaimai-cli item count --body '{"title":"关键字"}' --output json

# 自动翻页（pageable 接口；达 500/1000 条阈值可交互确认）
kuaimai-cli item +list \
  --body '{"title":"关键字","pageNo":1,"pageSize":50}' \
  --page-all --output json

# Agent/脚本：限制条数或自动续查
kuaimai-cli item +list \
  --body '{"title":"关键字","pageNo":1,"pageSize":50}' \
  --page-all --page-limit 200 --page-confirm yes --output json

# 商品档案 V2 列表（无 shortcut，走 web call）
kuaimai-cli web call item.item-query-list-v2 \
  --body '{"title":"关键字","pageNo":1,"pageSize":20}' \
  --output json

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

### 供应链（scm）

scm 域暂无 shortcuts，统一走 `web call scm.<operation>`（自动请求 `https://scm.superboss.cc/`）：

```bash
# 员工列表
kuaimai-cli web call scm.staff-query \
  --body '{"pageNo":1,"pageSize":20}' \
  --output json

# 铺货日志（多数日志接口需 startTime/endTime）
kuaimai-cli web call scm.logging-publish-log \
  --body '{"pageNo":1,"pageSize":20,"startTime":"2026-06-01 00:00:00","endTime":"2026-06-08 23:59:59"}' \
  --output json

# 供应链商品列表
kuaimai-cli web call scm.item-base-page \
  --body '{"pageNo":1,"pageSize":50}' \
  --output json

# 平台铺货配置
kuaimai-cli web call scm.dsb-query-distribution-config \
  --params '{"shopType":"TouTiaoFXG"}' \
  --output json
```

详见 `skills/kuaimai-scm/SKILL.md`。**勿与 item 混用**：路径同为 `/item/*` 时，`web call item.*`（erp1）与 `web call scm.item-*`（scm）语义不同。

### Skill

```bash
kuaimai-cli skill list --output json
kuaimai-cli skill install              # kuaimai-shared + kuaimai-item + kuaimai-scm
kuaimai-cli skill install kuaimai-scm  # 单个 Skill
```

### Registry 接口发现与调用（对标飞书 CLI）

远端 registry 由 `kuaimaierp-cli-auto` 发布至 `kuaimai-cli-open`，CLI **每次启动命令前**自动检查 version/ETag 并更新本地缓存 `~/.kuaimai-cli/registry/registry.json`。

```bash
# 列出全部已发布接口
kuaimai-cli capabilities --output json

# 单接口 schema 自省
kuaimai-cli schema api.luotao.test.get --output json

# 按 apiId 调用（推荐，与 registry 示例一致）
kuaimai-cli web call api.luotao.test.get \
  --params '{"keyword":"测试商品"}' \
  --output json

# 手动同步 / 开发时监听远端变化
kuaimai-cli registry sync --output json
kuaimai-cli registry watch --interval 30 --verbose
```

配置（`~/.kuaimai-cli/config.yaml`）：

```yaml
registry:
  source: "http://open-cli.kuaimai.com/registry/registry.json"
  auto_sync: true   # 默认开启；设 false 或 KUAIMAI_CLI_SKIP_REGISTRY_SYNC=1 可跳过
```

### 其它

```bash
kuaimai-cli --version
kuaimai-cli upgrade                 # 默认：有新版则升级并同步 Skills
kuaimai-cli upgrade --check-only --output json
kuaimai-cli doctor --output json
```

### 全局参数（分页与写操作）

| 参数 | 说明 |
|------|------|
| `--output json\|table\|csv\|ndjson` | 输出格式（Agent 推荐 `json`） |
| `--dry-run` | 写操作预览，不提交 |
| `--verbose` | stderr 调试日志 |
| `--page-all` | `pageable:true` 列表自动翻页 |
| `--page-limit N` | 与 `--page-all` 配合，最多拉取 N 条 |
| `--page-confirm` | `prompt`（默认）\| `yes` \| `no`；500/1000 条阈值续查策略 |

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

对标飞书 lark-cli：**用着就知道** + **默认一键升级**。

```bash
# 默认：对比 GitHub Release，有新版则 npm 全局安装并同步 Skills
kuaimai-cli upgrade

# 仅检查、不安装
kuaimai-cli upgrade --check-only --output json

# 等价手动路径
npx @kuaimai-cli/cli@latest install
npm install -g @kuaimai-cli/cli@latest
```

- 任意命令结束后，若有新版会在 **stderr** 提示（24h 缓存，见 `~/.kuaimai-cli/version-check.json`）
- Skill：`skill install --if-stale`；CLI 版本变更后会尝试自动同步（`~/.kuaimai-cli/skill-sync.json`）
- 禁用提示：`KUAIMAI_CLI_SKIP_UPDATE_CHECK=1`；禁用 Skill 自动同步：`KUAIMAI_CLI_SKIP_SKILL_SYNC=1`

也可从 [Releases](https://github.com/kuaimai-cli/kuaimai-cli/releases) 下载新版本二进制覆盖安装。

---

## 常见问题

| 现象 | 处理 |
|------|------|
| `auth check` 失败 / 401 | 重新 `auth login`；确认 token 未过期 |
| `本地 registry 未同步` | 执行任意命令会自动同步；或 `kuaimai-cli registry sync` |
| `doctor` 中 Skill 未就绪 | `kuaimai-cli skill install` 或 `skill install --if-stale`（覆盖写入，一般无需手删缓存） |
| `npx install` 显示「已安装跳过」但仍是旧版 | `kuaimai-cli upgrade` 或 `npm install -g @kuaimai-cli/cli@latest`（0.1.8+ 向导已自动比对版本） |
| 不知道有新版本 | 日常用 CLI 即可（stderr 提示）；或 `kuaimai-cli upgrade` |
| `permission denied: kuaimai-cli` | npm 全局入口 `run.js` 缺执行位；运行 `chmod +x $(npm root -g)/../bin/kuaimai-cli` 或重装 `@kuaimai-cli/cli@latest`（0.1.2+ 已自动修复） |
| macOS 仍无法执行 Go 二进制 | `xattr -d com.apple.quarantine ~/.npm-global/lib/node_modules/@kuaimai-cli/cli/bin/kuaimai-cli` |
| checksum 校验失败 | 重打 tag 后须重新 `npm publish`，使包内 `checksums.txt` 与 Release 一致 |
| 写操作不敢直接执行 | 所有写接口先加 `--dry-run` 预览请求体 |
| `--page-all` 中途停止 | 非交互环境达阈值会停止；加 `--page-confirm yes` 或 `--page-limit` |
| `item +list` 不支持 dry-run | 查询接口；用 `web call` 写接口（`write:true`）才支持 `--dry-run` |

---

## 给 AI Agent 使用者

- 安装与鉴权详见 [Agent 安装指南](docs/快麦 CLI 安装（Agent 专用）.md)
- Agent 行为约定见仓库根目录 [AGENTS.md](AGENTS.md)
- Skill 安装后位于 `~/.agents/skills/kuaimai-{shared,item,scm}/`（含 `SKILL.md` 与 `references/`）
- 商品域：先 Read `kuaimai-item/SKILL.md` → 写操作再 Read 对应 `references/` 文档
- 供应链域：先 Read `kuaimai-scm/SKILL.md` → `web call scm.<operation>`（无 shortcuts）
- 先 Read `kuaimai-shared/SKILL.md`（registry 发现流程）
- 无 shortcut 的 item 接口：Read `references/kuaimai-item-web-call.md` + `schema <apiId>`
- 全量翻页：非交互执行加 `--page-confirm yes`；见各域 `references/*-meta-execution.md`
- 命令选型见 [Agent 命令选型与 schema 流程](docs/Agent命令选型与schema流程.md)

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
| [Registry 远端同步说明](docs/Registry远端同步说明.md) | 自动同步、capabilities、web call |
| [接口 JSON 生成与同步系统设计](docs/接口JSON生成与同步系统设计.md) | api-onboard → open-cli 边界 |
| [meta_data.json 定义规范](docs/kuaimai-cli%20meta_data.json%20定义规范.md) | 历史 v1 字段语义；新接口见 registry v2 |
| [开发发布流程](docs/快麦%20CLI%20开发发布流程文档.md) | npm 安装/升级、CI/Release、内网 dist |
| [GitHub Release 模板](.github/RELEASE_TEMPLATE.md) | 手动编辑 Release 说明时参考 |
| [Agent 安装指南](docs/快麦 CLI 安装（Agent 专用）.md) | Cursor / Codex 等 IDE |
| [Agent 命令选型与 schema 流程](docs/Agent命令选型与schema流程.md) | shortcut / web call 选型、何时查 schema |
| [Registry 远端同步说明](docs/Registry远端同步说明.md) | 自动同步、web call、watch 开发 |
| [开发白皮书](docs/kuaimai-cli%20开发文档.md) | 架构与规划（维护者） |
| [系统架构与飞书对标](docs/系统架构与飞书对标说明.md) | 分层、命令树、分页防护 |
| [验收测试](docs/kuaimai-cli%20验收测试.md) | 分阶段验收清单 |

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
