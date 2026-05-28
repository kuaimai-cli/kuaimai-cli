# kuaimai-cli 文档索引

快麦 **erp-items-core** 商品 CLI，架构对标飞书 [lark-cli](https://github.com/larksuite/lark-cli)。

## 按读者

| 你是… | 从这里开始 |
|--------|------------|
| 新用户 / 运维 | [本地使用实操指南](./kuaimai-cli%20本地使用实操指南.md) |
| AI Agent / IDE | [Agent 安装指南](./kuaimai-cli-agent-installation-guide.md) · 仓库根 [AGENTS.md](../AGENTS.md) |
| 收到压缩包 / Agent 离线安装 | [内部分发与使用](./kuaimai-cli-内部分发与使用.md) |
| 开发者 / 发版 | [开发白皮书](./kuaimai-cli%20开发文档.md) · [npx 分发与安装](./kuaimai-cli-npx-分发与安装.md) · [系统架构与飞书对标](./系统架构与飞书对标说明.md) |
| 测试 / 验收 | [验收测试](./kuaimai-cli%20验收测试.md) |

## 全部文档

| 文档 | 说明 |
|------|------|
| [kuaimai-cli 开发文档.md](./kuaimai-cli%20开发文档.md) | 架构、目录规范、分阶段路线图 |
| [系统架构与飞书对标说明.md](./系统架构与飞书对标说明.md) | 分层图、命令树、飞书差异与待办优先级 |
| [每阶段新增能力.md](./每阶段新增能力.md) | 阶段一～四能力清单与 API 对照表 |
| [kuaimai-cli 本地使用实操指南.md](./kuaimai-cli%20本地使用实操指南.md) | 编译、配置、鉴权、商品命令示例 |
| [kuaimai-cli-agent-installation-guide.md](./kuaimai-cli-agent-installation-guide.md) | npm / Release / Skill / 鉴权步骤 |
| [kuaimai-cli-npx-分发与安装.md](./kuaimai-cli-npx-分发与安装.md) | 对标飞书 npx 安装/升级：原理、发版、registry、排错 |
| [kuaimai-cli-内部分发与使用.md](./kuaimai-cli-内部分发与使用.md) | 打包与安装 `dist`+`skills` 压缩包（含 Agent） |
| [kuaimai-cli 验收测试.md](./kuaimai-cli%20验收测试.md) | 分阶段验收命令与检查项 |

## 当前能力快照（与代码一致）

- **业务域**：`item`（6 子命令：`+list`、`list`、`count`、`get-detail`、`save`、`update-title`）
- **鉴权**：`auth login|logout|status|check|list|use`，多 profile
- **平台**：`doctor`、`upgrade`、`skill install-all`、E2E（`tests/cli_e2e`）
- **默认 API**：`https://erp1.superboss.cc/`

变更记录见仓库根 [CHANGELOG.md](../CHANGELOG.md)。
