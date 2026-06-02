# kuaimai-cli Agent 约定

> **首次使用**：请先按 [Agent 安装指南](./docs/快麦 CLI 安装（Agent 专用）.md) 完成 CLI、Skills 与鉴权配置。

## 架构（选命令前必读）

| 层级 | 模块 | Agent 动作 |
|------|------|------------|
| meta | `meta_data.json` v1.6.0（1157 op） | 需要发现接口时 `schema --output json` |
| Skill | `kuaimai-item` v2.0.0 + `references/` | **优先** Read SKILL.md；写操作 Read references |
| CLI | shortcuts → service → api | 有 shortcut 不查 schema |

## 输出

- **stdout**：结构化数据；默认 `{ok, data, error, hint}`
  - `--output json|table`：完整信封
  - `--output csv`：列表成功时输出 CSV（失败仍为 JSON 信封）
  - `--output ndjson`：列表成功时每行一条 JSON 记录
- **stderr**：`--verbose` 日志、友好错误；**禁止**向 stdout 打日志

## 商品域（item）

优先使用 shortcuts，勿手写 URL：

```bash
kuaimai-cli item +list --body '{"title":"关键字","pageNo":1,"pageSize":50}' --output json
kuaimai-cli item count --body '{"title":"关键字"}' --output json
kuaimai-cli item get-detail --sys-item-id <id>
kuaimai-cli item save --body '{...}' --dry-run
kuaimai-cli item update-title --sys-item-id <id> --title "新标题" --dry-run
kuaimai-cli doctor
kuaimai-cli upgrade                    # 默认一键升级；仅检查加 --check-only
```

- **升级**：任意命令后 stderr 可能提示新版（24h 缓存）；`upgrade` 默认执行 `npm install -g @kuaimai-cli/cli@latest` 并同步 Skills
- list/count：`application/x-www-form-urlencoded`（`--body` JSON 会转为 form）
- 鉴权：`accessToken` 须由用户提供（联系 ERP 管理员申请），Agent 不可代填；`auth login` 后请求头自动带 token；可用 `auth check` 探测
- 多账号：`auth login --profile <name>` · `auth use <name>` · `auth list`
- **列表翻页**（`pageable:true`，如 `stock-list`、`item-query-list-v2`）：
  - 默认单页，用户未要求「全部」时**不要**加 `--page-all`
  - 全量：`--page-all`；Agent/管道续查：`--page-confirm yes`；限条数：`--page-limit N`
  - 规则详见 `skills/kuaimai-item/references/kuaimai-item-meta-execution.md`
- 改标题：优先 `item update-title`；或 `get-detail` + jq + `item save`（见 `skills/kuaimai-item/references/`）
- **无 shortcut**（如档案 V2 列表）：`service item item-query-list-v2 --body '{...}'`
- 原子 API 兜底：`service item <operation>`（operation 名如 `stock-list`，**非** `list`；见 [meta 定义规范](./docs/kuaimai-cli%20meta_data.json%20定义规范.md)）

## Skill

```bash
kuaimai-cli skill list
kuaimai-cli skill install                  # kuaimai-shared + kuaimai-item（整目录含 references/）
kuaimai-cli skill install --if-stale       # 仅在 Release/CLI 变化或未安装时更新
kuaimai-cli skill install kuaimai-item     # 安装单个 Skill
```

安装后 Agent 优先读取 `~/.agents/skills/kuaimai-item/SKILL.md`；仓库内开发时读 `skills/kuaimai-item/SKILL.md`。

**CRITICAL**：

1. 开始前 Read `kuaimai-shared/SKILL.md`
2. 写操作（`save`、`update-title`）须先 Read 对应 `references/` 文档
3. 命令选型见 [Agent命令选型与schema流程.md](./docs/Agent命令选型与schema流程.md)

## 安全

- Token 仅通过 `auth login` 写入密钥链
- 写操作先 `--dry-run --verbose`，用户确认后再去掉 `--dry-run`
- 日志与 dry-run 预览中的敏感字段已脱敏
