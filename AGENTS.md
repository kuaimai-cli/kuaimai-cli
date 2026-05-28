# kuaimai-cli Agent 约定

> **首次使用**：请先按 [Agent 安装指南](./docs/kuaimai-cli-agent-installation-guide.md) 完成 CLI、Skills 与鉴权配置。

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
kuaimai-cli item get-detail --sys-item-id <id>
kuaimai-cli item save --body '{...}' --dry-run
kuaimai-cli item update-title --sys-item-id <id> --title "新标题" --dry-run
kuaimai-cli doctor
kuaimai-cli upgrade
```

- list/count：`application/x-www-form-urlencoded`（`--body` JSON 会转为 form）
- 鉴权：`auth login` 后请求头自动带 `accessToken`；可用 `auth check` 探测
- 多账号：`auth login --profile <name>` · `auth use <name>` · `auth list`
- 列表翻页：`--page-all`（识别 body 中 `pageNo`/`pageSize`）
- 改标题：优先 `item update-title`；或 `get-detail` + jq + `item save`（见本地使用实操指南）

## Skill

```bash
kuaimai-cli skill list
kuaimai-cli skill install kuaimai-item     # 安装单个 Skill（默认从 GitHub）
kuaimai-cli skill install-all              # 批量安装 kuaimai-shared + kuaimai-item
```

安装后 Agent 优先读取 `~/.agents/skills/kuaimai-item/SKILL.md`；仓库内开发时读 `skills/kuaimai-item/SKILL.md`。

## 安全

- Token 仅通过 `auth login` 写入密钥链
- 日志与 dry-run 预览中的敏感字段已脱敏
