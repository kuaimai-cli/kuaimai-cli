# kuaimai-cli Agent 约定

> **首次使用**：请先按 [Agent 安装指南](./docs/快麦 CLI 安装（Agent 专用）.md) 完成 CLI、Skills 与鉴权配置。

## 架构（选命令前必读）

| 层级 | 模块 | Agent 动作 |
|------|------|------------|
| meta | `meta_data.json` v1.7.0（item + scm） | 需要发现接口时 `schema --output json` |
| Skill | `kuaimai-item` / `kuaimai-scm` + `references/` | **优先** Read SKILL.md；写操作 Read references |
| CLI | shortcuts → service → api | item 有 shortcut 不查 schema；scm 走 `service scm` |

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
kuaimai-cli item update-title --sys-item-id <id> --title "新标题" --dry-run
```

- 域名：`api.url` 默认 `https://erp1.superboss.cc/`
- 详见 `skills/kuaimai-item/SKILL.md`

## 供应链域（scm）

**无 shortcuts**，统一走 `service scm`（自动请求 `https://scm.superboss.cc/`）：

```bash
kuaimai-cli service scm staff-query --body '{"pageNo":1,"pageSize":20}' --output json
kuaimai-cli service scm logging-publish-log \
  --body '{"pageNo":1,"pageSize":20,"startTime":"2026-06-01 00:00:00","endTime":"2026-06-08 23:59:59"}' --output json
kuaimai-cli service scm item-base-page --body '{"pageNo":1,"pageSize":50}' --output json
kuaimai-cli service scm dsb-query-distribution-config --body '{"shopType":"TouTiaoFXG"}' --output json
```

- **勿与 item 混用**：路径同为 `/item/*` 时，`service item`（erp1）与 `service scm`（scm）语义不同
- 铺货/操作日志多数需 `startTime`/`endTime`
- 详见 `skills/kuaimai-scm/SKILL.md` 与 `references/kuaimai-scm-domain-routing.md`

## 公共

```bash
kuaimai-cli doctor
kuaimai-cli upgrade                    # 默认一键升级；仅检查加 --check-only
```

- **鉴权**：`accessToken` 须由用户提供；`auth login` 后请求头自动带 token
- **多账号**：`auth login --profile <name>` · `auth use <name>` · `auth list`
- **列表翻页**（`pageable:true`）：默认单页；全量 `--page-all`；Agent 续查 `--page-confirm yes`

## Skill

```bash
kuaimai-cli skill install                  # kuaimai-shared + kuaimai-item + kuaimai-scm
kuaimai-cli skill install --if-stale
kuaimai-cli skill install kuaimai-scm
```

**Agent 路由口诀**：

- 供应链 / scm / 铺货 / 操作日志 → Read `kuaimai-scm/SKILL.md` → `service scm <operation>`
- ERP 库存商品 / 改标题 → Read `kuaimai-item/SKILL.md` → item shortcuts 优先
- 配置 / 登录 / 输出 → Read `kuaimai-shared/SKILL.md`

**CRITICAL**：

1. 开始前 Read `kuaimai-shared/SKILL.md`
2. 写操作须先 Read 对应 `references/` 文档
3. 命令选型见 [Agent命令选型与schema流程.md](./docs/Agent命令选型与schema流程.md)

## 安全

- Token 仅通过 `auth login` 写入密钥链
- 写操作先 `--dry-run --verbose`，用户确认后再去掉 `--dry-run`
- 日志与 dry-run 预览中的敏感字段已脱敏
