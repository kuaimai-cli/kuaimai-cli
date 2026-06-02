# @kuaimai-cli/cli

快麦 ERP **商品（erp-items-core）** 命令行工具 npm 分发包，对标飞书 [`@larksuite/cli`](https://www.npmjs.com/package/@larksuite/cli)。

本包为 **薄壳**：`postinstall` / `run.js` 从 [GitHub Releases](https://github.com/kuaimai-cli/kuaimai-cli/releases) 下载对应平台的 Go 二进制；失败时自动回退 [npmmirror](https://registry.npmmirror.com) `/-/binary/kuaimai-cli/`（对标 `@larksuite/cli`，须在 [cnpmcore 注册](../docs/npmmirror-二进制镜像.md) 后镜像才可用）。

## 能力快照

| 项 | 说明 |
|----|------|
| 业务域 | 商品 list / count / 详情 / 改标题（`item` shortcuts） |
| 元数据 | `meta_data.json` **v1.6.0**，**1157** 个 `/item` 接口（`schema` / `service item`） |
| 分页 | `--page-all` · `--page-limit` · `--page-confirm`（海量数据防护） |
| Agent | `skill install` / `--if-stale`；`upgrade` 默认一键升级 + stderr 新版本提示 |

## 用户安装（推荐）

```bash
npx @kuaimai-cli/cli@latest install
```

安装向导会：下载二进制、提示 `config init` / `auth login`、可选安装 Skill。

非交互环境（CI / Agent）会跳过向导，请按终端提示手动执行后续步骤。

### 安装后（5 步）

```bash
kuaimai-cli config init
kuaimai-cli auth login "<accessToken>"    # 须向 ERP 管理员申请
kuaimai-cli skill install                 # Agent 建议
kuaimai-cli auth check --output json
kuaimai-cli doctor --output json
```

### 试一条命令

```bash
kuaimai-cli item +list \
  --body '{"title":"test","pageNo":1,"pageSize":10}' \
  --output json
```

## 升级

```bash
kuaimai-cli upgrade                        # 默认：有新版则 npm 全局升级并同步 Skills
kuaimai-cli upgrade --check-only --output json
npx @kuaimai-cli/cli@latest install
# 或
npm install -g @kuaimai-cli/cli@latest
```

任意命令结束后（24h 缓存）若有新版会在 **stderr** 提示；`upgrade` 默认会执行 `npm install -g @kuaimai-cli/cli@latest` 并同步 Skills。仅检查请加 `--check-only`。禁用提示：`KUAIMAI_CLI_SKIP_UPDATE_CHECK=1`。

`npx @kuaimai-cli/cli@latest install`（**0.1.8+**）：若全局 npm 包版本低于向导包，会自动升级而非「有旧包就跳过」。强制重装：`KUAIMAI_CLI_FORCE_INSTALL=1 npx @kuaimai-cli/cli@latest install`。

Skill 有更新时（覆盖写入，无需手删各 Agent 缓存）：

```bash
kuaimai-cli skill install --if-stale
# 或强制重装
kuaimai-cli skill install
```

禁用 Skill 自动同步：`KUAIMAI_CLI_SKIP_SKILL_SYNC=1`

## 全局安装

```bash
npm install -g @kuaimai-cli/cli@latest
kuaimai-cli --version
```

## 平台支持

| OS | 架构 |
|----|------|
| macOS | `amd64`、`arm64` |
| Linux | `amd64`、`arm64` |
| Windows | `amd64`（无 `arm64` 包） |

## 包内容

| 路径 | 说明 |
|------|------|
| `scripts/run.js` | `kuaimai-cli` 入口，转发至 `bin/kuaimai-cli` |
| `scripts/install.js` | postinstall：从 GitHub Release 下载并校验 checksum |
| `scripts/install-wizard.js` | 交互式首次配置向导 |
| `checksums.txt` | 与 Release 资产一致的 SHA256（发布时由 CI 同步） |

## 环境要求

- Node.js **16+**
- 网络可访问 `github.com`（下载 Release 二进制）
- 业务请求默认 `https://erp1.superboss.cc/`

## 文档

| 文档 | 说明 |
|------|------|
| [仓库 README](../README.md) | 用户快速上手 |
| [Agent 安装指南](../docs/快麦%20CLI%20安装（Agent%20专用）.md) | Cursor / Codex 等 |
| [开发发布流程](../docs/快麦%20CLI%20开发发布流程文档.md) | 维护者发版、CI、checksum |
| [CHANGELOG](../CHANGELOG.md) | 版本变更 |

## 维护者：本地调试

```bash
cd npm
npm link
kuaimai-cli --version
```

## 维护者：发布流程

1. 更新根目录 [CHANGELOG.md](../CHANGELOG.md)，合并 `main`，CI 为绿  
2. 打 tag 并推送：`git tag v0.1.x && git push origin v0.1.x`  
3. GitHub Actions `release.yml`：GoReleaser 创建 Release → 同步 `checksums.txt` → `npm publish`  
4. 发布说明由 [.goreleaser.yaml](../.goreleaser.yaml) 的 `release.header` / `footer` 生成；手动编辑可参考 [.github/RELEASE_TEMPLATE.md](../.github/RELEASE_TEMPLATE.md)

手动发布 npm（需已有 Release 与 `checksums.txt`）：

```bash
cp /path/to/checksums.txt npm/checksums.txt
cd npm
# 同步 package.json version 与 tag 一致
npm publish --access public
```

**注意**：打 tag 后须等 Release 完成再 publish，否则 `install.js` 下载会 404。`checksums.txt` 须与 Release 资产一致，否则校验失败。

## License

MIT
