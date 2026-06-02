# kuaimai-cli 文档索引

快麦 **erp-items-core** 商品 CLI，架构对标飞书 [lark-cli](https://github.com/larksuite/lark-cli)。

## 按读者

| 你是… | 从这里开始 |
|--------|------------|
| 新用户 / 运维 | 仓库根 [README.md](../README.md) · [每阶段新增能力](./每阶段新增能力.md) |
| AI Agent / IDE | [Agent 安装指南](./快麦 CLI 安装（Agent 专用）.md) · [Agent 命令选型与 schema 流程](./Agent命令选型与schema流程.md) · 仓库根 [AGENTS.md](../AGENTS.md) |
| 收到压缩包 / 离线安装 | [开发发布流程 · §6.3 / §7](./快麦%20CLI%20开发发布流程文档.md) |
| 开发者 / 发版 | [极简命令大全](./快麦%20CLI%20极简可运行命令大全.md) · [开发白皮书](./kuaimai-cli%20开发文档.md) · [开发发布流程](./快麦%20CLI%20开发发布流程文档.md) · [meta_data.json 定义规范](./kuaimai-cli%20meta_data.json%20定义规范.md) · [系统架构与飞书对标](./系统架构与飞书对标说明.md) |
| 测试 / 验收 | [验收测试](./kuaimai-cli%20验收测试.md) |

## 全部文档

| 文档 | 说明 |
|------|------|
| [kuaimai-cli 开发文档.md](./kuaimai-cli%20开发文档.md) | 架构、目录规范、分阶段路线图 |
| [kuaimai-cli meta_data.json 定义规范.md](./kuaimai-cli%20meta_data.json%20定义规范.md) | **元数据唯一规范**：operation 命名、contentType、write/pageable、Schema、分页防护 |
| [系统架构与飞书对标说明.md](./系统架构与飞书对标说明.md) | 分层图、命令树、飞书差异与后续扩展方向 |
| [Agent命令选型与schema流程.md](./Agent命令选型与schema流程.md) | Agent 自然语言 → shortcut/service 选型、何时查 schema |
| [每阶段新增能力.md](./每阶段新增能力.md) | 阶段一～四能力清单与 shortcuts / meta 对照表 |
| [快麦 CLI 安装（Agent 专用）.md](./快麦 CLI 安装（Agent 专用）.md) | npm / Release / Skill（整目录 + references）/ 鉴权步骤 |
| [快麦 CLI 极简可运行命令大全.md](./快麦%20CLI%20极简可运行命令大全.md) | **复制即用**：开发自测、推送 PR、Tag 发版、本机升级/重置、用户安装 |
| [快麦 CLI 开发发布流程文档.md](./快麦%20CLI%20开发发布流程文档.md) | 本地自测、PR/CI、发版前查版本、Tag 发版、Release 观测、升级/重置、十条红线 |
| [.github/RELEASE_TEMPLATE.md](../.github/RELEASE_TEMPLATE.md) | GitHub Release 手动编辑模板 |
| [kuaimai-cli 验收测试.md](./kuaimai-cli%20验收测试.md) | 分阶段验收命令与检查项 |

## 当前能力快照（与代码一致，2026-06-02）

### 三层架构（已定稿）

| 层级 | 模块 | 状态 |
|------|------|------|
| 底层 | `meta_data.json` | ✅ **v1.6.0**，`item` 域 **1157** 个 operation（erp-items-core `/item` Controller 全量注册） |
| 中层 | Skill（`kuaimai-item` v2.0.0 + references） | ✅ 执行规则、meta 驱动说明、分页防护、service 兜底指南 |
| 上层 | shortcuts（6 子命令）+ `service item` + `api` | ✅ 双轨命令；高频走 shortcuts，其余走 service |

### 业务与命令

- **shortcuts**：`item +list` / `list` / `count` / `get-detail` / `save` / `update-title`
- **meta 核心 6 个**（有 shortcut 映射）：`stock-list`、`stock-count`、`item-detail`、`item-save`、`item-update-title`；另有 **`item-query-list-v2`**（仅 service，Skill 已文档化）
- **双轨**：Agent **优先 shortcuts**；原子兜底 `service item <operation>`（operation 名为 `stock-list` 等，**非** `list`）
- **contentType**：`get_query` · `post_form` · `post_json`
- **分页**：`--page-all` + **`--page-limit`** + **`--page-confirm`**（500/1000 阈值交互续查，见 `internal/pagination`）

### 平台能力

- **鉴权**：`auth login|logout|status|check|list|use`，多 profile
- **输出**：`--output table|json|csv|ndjson`；`--dry-run` · `--verbose` · `--no-color`
- **元数据**：`schema` 全量自省 · `service item <op>` 零代码驱动（required 校验 + Schema 默认值）
- **Agent**：`doctor` · `upgrade`（默认一键升级 + stderr 新版本提示）· `skill install` / `--if-stale`（GitHub 整目录覆盖写入）· E2E（`tests/cli_e2e`）
- **默认 API**：`https://erp1.superboss.cc/`

### 框架状态

**meta + Skill + CLI 基础能力已闭环**。后续新增接口只需：登记 meta →（可选）补 shortcut/Skill → 验收，**无需改造底层框架**。

变更记录见仓库根 [CHANGELOG.md](../CHANGELOG.md)。
