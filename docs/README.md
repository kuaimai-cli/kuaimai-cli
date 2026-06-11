# kuaimai-cli 文档索引

快麦 ERP 商品 + 供应链 CLI，架构对标飞书 [lark-cli](https://github.com/larksuite/lark-cli)。

**接口真相源**：远端 [registry.json](http://open-cli.kuaimai.com/registry/registry.json)（v2）→ CLI 本地缓存 → `capabilities` / `schema` / `web call`。

## 按读者

| 你是… | 从这里开始 |
|--------|------------|
| 新用户 / 运维 | 仓库根 [README.md](../README.md) |
| AI Agent / IDE | [Agent 安装指南](./快麦%20CLI%20安装（Agent%20专用）.md) · [Agent 命令选型](./Agent命令选型与schema流程.md) · [AGENTS.md](../AGENTS.md) |
| 开发者 / 发版 | [开发发布流程](./快麦%20CLI%20开发发布流程文档.md) · [开发白皮书](./kuaimai-cli%20开发文档.md) · [系统架构与飞书对标](./系统架构与飞书对标说明.md) |
| Registry 接入 | [Registry 远端同步](./Registry远端同步说明.md) · [API 网关转发](./API网关转发说明.md) · [接口 JSON 生成与同步系统设计](./接口JSON生成与同步系统设计.md) |
| 测试 / 验收 | [验收测试](./kuaimai-cli%20验收测试.md) |

## 全部文档

| 文档 | 说明 |
|------|------|
| [系统架构与飞书对标说明.md](./系统架构与飞书对标说明.md) | 分层图、registry 数据流、命令树、Skill 分工 |
| [Registry远端同步说明.md](./Registry远端同步说明.md) | 自动同步、ETag、`registry sync/watch` |
| [API网关转发说明.md](./API网关转发说明.md) | CLI → open-cli 网关 → ERP/SCM、限流与配置 |
| [接口JSON生成与同步系统设计.md](./接口JSON生成与同步系统设计.md) | api-onboard → open-cli → kuaimai-cli 边界 |
| [Agent命令选型与schema流程.md](./Agent命令选型与schema流程.md) | 自然语言 → shortcut / web call、何时查 schema |
| [快麦 CLI 安装（Agent 专用）.md](./快麦%20CLI%20安装（Agent%20专用）.md) | 安装、Skill、registry 验证步骤 |
| [快麦 CLI 开发发布流程文档.md](./快麦%20CLI%20开发发布流程文档.md) | 本地自测、PR/CI、Tag 发版、npm-skills-sync |
| [kuaimai-cli 开发文档.md](./kuaimai-cli%20开发文档.md) | 架构白皮书、目录规范、分阶段路线图 |
| [kuaimai-cli meta_data.json 定义规范.md](./kuaimai-cli%20meta_data.json%20定义规范.md) | 历史 v1 字段规范；新接口见 registry v2 与接口 JSON 设计文档 |
| [kuaimai-cli 验收测试.md](./kuaimai-cli%20验收测试.md) | 分阶段验收命令与检查项 |

## 当前能力快照（2026-06）

### 架构三层

| 层级 | 模块 | 说明 |
|------|------|------|
| 远端 | `registry.json` v2 | `kuaimaierp-cli-auto` 生成 → open-cli 发布 |
| CLI | capabilities · schema · web call · registry sync | 消费本地缓存 `~/.kuaimai-cli/registry/` |
| Agent | shared v1.1 + item v3 + scm v2 | **意图路由 + 工作流**；接口发现走 registry |

### 命令

| 类型 | 示例 |
|------|------|
| Shortcuts | `item +list`、`item update-title`（6 个，手写） |
| Registry | `web call api.luotao.test.get`、`web call scm.staff-query`（待全量发布） |
| 兜底 | `api POST /path` |

**已移除**：`service` 命令（统一 `web call <apiId>`）。

### Skill 职责（对标飞书）

- `kuaimai-shared`：auth、输出、**registry 发现流程**
- `kuaimai-item`：商品 shortcuts + references 工作流
- `kuaimai-scm`：供应链意图路由 + `web call scm.*`

变更记录：[CHANGELOG.md](../CHANGELOG.md)
