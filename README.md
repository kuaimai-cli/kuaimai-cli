# kuaimai-cli

快麦 **erp-items-core** 商品业务专属命令行工具，架构与交互对标飞书 [lark-cli](https://github.com/larksuite/lark-cli)。

## 快速开始

```bash
# 安装（任选其一）
npx @kuaimai/cli@latest install
# 或从 Release 下载二进制：https://github.com/kuaimai/kuaimai-cli/releases

kuaimai-cli config init
kuaimai-cli auth login "<accessToken>"   # 浏览器 DevTools 获取
kuaimai-cli auth check
kuaimai-cli doctor
kuaimai-cli skill install-all --from ./skills
```

## 常用命令

```bash
# 按标题查商品
kuaimai-cli item +list --body '{"title":"关键字","pageNo":1,"pageSize":20}' --output json

# 改标题（推荐）
kuaimai-cli item update-title --sys-item-id <id> --title "新标题" --dry-run

# 检查更新
kuaimai-cli upgrade
```

## 文档

| 文档 | 说明 |
|------|------|
| [文档索引](docs/README.md) | 全部文档入口 |
| [开发白皮书](docs/kuaimai-cli%20开发文档.md) | 架构与分阶段规划 |
| [系统架构与飞书对标](docs/系统架构与飞书对标说明.md) | 分层、命令树、对标差异与路线图 |
| [每阶段新增能力](docs/每阶段新增能力.md) | 版本能力对照 |
| [本地使用实操指南](docs/kuaimai-cli%20本地使用实操指南.md) | 配置、鉴权、商品标题流程 |
| [Agent 安装指南](docs/kuaimai-cli-agent-installation-guide.md) | Cursor / Codex 等 IDE |
| [npx 分发与安装](docs/kuaimai-cli-npx-分发与安装.md) | 对标飞书：registry、发版、一键安装与升级 |
| [内部分发与使用](docs/kuaimai-cli-内部分发与使用.md) | 压缩包分发（含 Agent 安装） |
| [验收测试](docs/kuaimai-cli%20验收测试.md) | 分阶段验收清单 |

## 安全提示

- `accessToken` 仅通过 `auth login` 写入系统密钥链，不会写入配置文件。
- 写操作（`item save`、`item update-title`）请先用 `--dry-run` 预览。
- Agent 代操作 ERP 具有真实业务影响，请在测试环境验证。

## 开发

```bash
make build
go test ./...
go test ./tests/cli_e2e/...
```

## License

MIT
