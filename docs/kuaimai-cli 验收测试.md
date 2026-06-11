# kuaimai-cli 阶段开发准入、最低可用标准、阶段验收测试规范

> 对标飞书 CLI；与 [开发文档](./kuaimai-cli%20开发文档.md)、[Registry 同步](./Registry远端同步说明.md)、[系统架构说明](./系统架构与飞书对标说明.md) 保持一致。  
> **当前业务验收基准**：**erp-items-core 商品域**（标题查改）+ **erp-scm 供应链域**（`web call` 只读查询）。

---

## 一、核心结论

| 项 | 说明 |
|----|------|
| 开发方式 | 按白皮书分阶段顺序开发，不跳步、不乱改目录 |
| 最低可用 | **阶段一**通过即可使用配置、鉴权、`api`、结构化输出 |
| 日常业务 | **阶段三** + **item shortcuts**（标题查询与修改） |
| 阶段一 | MVP ~60%：基建 + 三级命令形态 |
| 阶段二 | 企业版 ~80%：runner、schema/web call、重试、completion |
| 阶段三 | 平台 ~95%：item 域、Skill、csv/ndjson、脱敏、page-all、**标题闭环** |

---

## 二、各阶段能否使用

| 阶段 | 可使用 CLI | 可投入业务 | 飞书完成度 |
|------|------------|------------|------------|
| 一 | ✅ | ✅ 基建与 `api` | ~60% |
| 二 | ✅ | ✅ schema/web call、执行管线 | ~80% |
| 三 | ✅ | ✅ **商品标题** list/save/update-title 等 | ~95% |
| 四 | ✅ 部分 | E2E、doctor/upgrade、多账号、扩域仍规划 | ~98% |

---

## 三、阶段一验收清单

### 3.1 架构（硬性）

- [ ] 目录符合白皮书：`cmd/`、`internal/`、`shortcuts/`、`internal/registry/`、`scripts/fetch_meta/`
- [ ] `skills/` 含 `SKILL.md` 与 `references/` 工作流文档，无 Go 业务代码
- [ ] 业务逻辑在 `shortcuts/`，`cmd` 仅注册与参数
- [ ] stderr 日志 / stdout 结构化数据分离
- [ ] 输出 `{ok,data,error,hint}`
- [ ] `go build` 无报错

### 3.2 功能（逐条执行）

**1）全局与配置**

```bash
kuaimai-cli --help
kuaimai-cli --verbose config get
kuaimai-cli config init
kuaimai-cli config get
kuaimai-cli config get api.url
kuaimai-cli config get api.gateway_url
kuaimai-cli config set api.url "https://erp1.superboss.cc/"
kuaimai-cli config set api.gateway_url "https://open-cli.kuaimai.com"
kuaimai-cli config set cli.output json
```

验收要点：

- [ ] 再次 `config init` 不覆盖，提示已存在
- [ ] `config set` 为两参数：`key` `value`（无 `=`）
- [ ] 模板含 `api.url`、`api.gateway_url`、`api.retry`、`cli.output`、`cli.color`、连接池/熔断项
- [ ] 模板**不含** `json_suffix`（erp1 不追加 `.json`）

**2）鉴权**

```bash
kuaimai-cli auth login <your-accessToken>
kuaimai-cli auth status --output json
kuaimai-cli auth logout
```

- [ ] Token 在密钥链，非 config 明文
- [ ] 请求头为 `accessToken`（非 Bearer，由 `auth login` 自动附加）

**3）原始 API**

```bash
kuaimai-cli auth login <token>
kuaimai-cli api POST /item/stock/queryCount --body '{}' --output json --verbose
```

- [ ] stderr 中 URL **不含** 多余的 `.json` 后缀（erp1 约定）
- [ ] 未登录时友好提示；登录后 `ok: true` 或业务层错误（非 HTML 登录页）

**4）输出**

- [ ] `--output json` 可 `jq .ok`
- [ ] 日志仅在 stderr
- [ ] 错误无 Go 堆栈

**5）异常**

- [ ] 未登录 → `{ok:false}` + hint
- [ ] 参数缺失 → Cobra 帮助

---

## 四、阶段二验收清单

### 4.1 架构与基建

- [ ] `shortcuts/common/runner` 可用
- [ ] 远端 registry 可同步：`registry sync` 成功
- [ ] `capabilities` / `schema` / `web call` 可用
- [ ] **无** `service` 顶层命令（`kuaimai-cli -h` 中不存在）

### 4.2 Registry 与 web call

```bash
kuaimai-cli registry sync --output json
kuaimai-cli capabilities --output json
kuaimai-cli schema api.luotao.test.get --output json
kuaimai-cli schema --output json | jq '.data.total'
kuaimai-cli web call api.luotao.test.get --params '{"keyword":"测试"}' --output json
```

- [ ] 本地缓存路径 `~/.kuaimai-cli/registry/registry.json` 存在
- [ ] `schema <apiId>` 输出 `requestSchema` / `contentType` / `write` / `pageable`
- [ ] `web call` 支持 `--params` / `--data` / `--body`
- [ ] item/scm 全量 apiId 随 api-onboard 发布逐步增多（过渡期允许仅测试 api）
- [ ] `contentType` 仅为 `get_query` / `post_form` / `post_json`
- [ ] 查询接口 `write:false`；写接口 `write:true` 且 `--dry-run` 可预览
- [ ] 分页列表 `pageable:true`（`stock-list`、`item-query-list-v2` 等）；`--page-all` 可全量翻页
- [ ] `--page-limit` 达条数上限后停止并 stderr 提示
- [ ] `--page-confirm no` 在 500 条阈值处静默停止（可用 mock 或大数据环境）
- [ ] `requestSchema` / `responseSchema` 已写入（对照 [接口 JSON 设计](./接口JSON生成与同步系统设计.md) 与 registry v2 schema）

### 4.3 企业级能力

```bash
kuaimai-cli item save --body '{"sysItemId":1,"title":"x"}' --dry-run --output json
kuaimai-cli completion zsh > /dev/null
```

- [ ] 彩色 table（默认）/`--no-color`
- [ ] `api.retry` 配置可读（`config get api.retry`）
- [ ] dry-run 不写真实数据

### 4.4 联调环境

- [ ] `api.url` = `https://erp1.superboss.cc/`
- [ ] `api.gateway_url` = `https://open-cli.kuaimai.com`（或本地 open 服务地址）
- [ ] `auth login` 使用有效 `accessToken`
- [ ] dry-run 输出含 `gateway` 与 `target_host` 字段

---

## 五、阶段三验收清单（item 域 · 商品标题）

> 以下命令需先 `auth login`。改标题请在测试商品上进行，避免误改生产数据。

### 5.1 命令注册

```bash
kuaimai-cli item --help
```

- [ ] 子命令包含：`+list`、`list`、`count`、`get-detail`、`save`、`update-title`

### 5.2 按标题查询列表

```bash
kuaimai-cli item +list \
  --body '{"title":"2026","pageNo":1,"pageSize":10}' \
  --output json
```

- [ ] `ok: true`，`data` 为列表结构（非 HTML）
- [ ] `--verbose` 时 stderr 显示 form 请求体预览
- [ ] 精简 body（仅 `title`+分页）可工作（默认值由 CLI 补齐）

```bash
kuaimai-cli item count --body '{"title":"2026"}' --output json
```

- [ ] 返回总数，筛选与 list 一致

### 5.3 商品详情

```bash
kuaimai-cli item get-detail --sys-item-id <有效sysItemId> --output json
```

- [ ] 返回商品详情 JSON
- [ ] 缺少 `--sys-item-id` 时 Cobra 报错

### 5.4 修改标题（核心验收）

**方式 A — `update-title`（推荐）**

```bash
kuaimai-cli item update-title \
  --sys-item-id <测试ID> \
  --title "CLI验收标题" \
  --dry-run --verbose --output json

kuaimai-cli item update-title \
  --sys-item-id <测试ID> \
  --title "CLI验收标题" \
  --output json
```

**方式 B — get-detail + jq + save**

```bash
# 1. dry-run（<测试ID> 换成真实 sys-item-id）
kuaimai-cli item save \
  --body "$(kuaimai-cli item get-detail --sys-item-id <测试ID> --output json | jq -c '.data[0] | .title = "CLI验收标题" | .suiteBridgeList = .itemSuiteBridgeList | del(.itemSuiteBridgeList)')" \
  --dry-run --verbose --output json

# 2. 正式保存
kuaimai-cli item save \
  --body "$(kuaimai-cli item get-detail --sys-item-id <测试ID> --output json | jq -c '.data[0] | .title = "CLI验收标题" | .suiteBridgeList = .itemSuiteBridgeList | del(.itemSuiteBridgeList)')" \
  --output json

# 3. 验证
kuaimai-cli item +list --body '{"title":"CLI验收","pageNo":1,"pageSize":10}' --output json
```

- [ ] dry-run 不发送写请求，`stderr` 有脱敏预览
- [ ] 正式 `save` 返回 `ok: true`（或业务可识别的成功结构）
- [ ] 列表或 get-detail 中 `title` 已更新

### 5.5 Skill

```bash
kuaimai-cli skill list --output json
kuaimai-cli skill install kuaimai-item
kuaimai-cli skill install kuaimai-scm
```

- [ ] 能列出 `kuaimai-item`、`kuaimai-shared`、`kuaimai-scm`
- [ ] `skill install` 写入 `~/.agents/skills/<name>/` 整目录（`SKILL.md` + `references/`）
- [ ] `kuaimai-item/references/` 含 **10** 个工作流文档
- [ ] `kuaimai-scm/references/` 含 **7** 个工作流文档（含 `kuaimai-scm-web-call.md`）
- [ ] `kuaimai-item/SKILL.md` 含 CRITICAL 读 shared、Shortcuts 表、意图路由（**无**硬编码 meta 大表）
- [ ] `kuaimai-shared/SKILL.md` 含 registry 发现流程（capabilities → schema → web call）
- [ ] `kuaimai-scm/SKILL.md` 含 scm 路由表、`web call` 说明、item/scm 分流
- [ ] `kuaimai-shared/SKILL.md` 含 `metadata.cliHelp`，路由至 item/scm 域 Skill
- [ ] `doctor` 检测 item 与 scm 的 `references/` 缺失时提示重装

### 5.5.1 供应链域（scm）

```bash
kuaimai-cli web call scm.staff-query --body '{"pageNo":1,"pageSize":5}' --output json
kuaimai-cli web call scm.logging-publish-log \
  --body '{"pageNo":1,"pageSize":5,"startTime":"2026-06-01 00:00:00","endTime":"2026-06-08 23:59:59"}' \
  --output json
kuaimai-cli web call scm.item-base-page --body '{"pageNo":1,"pageSize":5}' --output json
```

- [ ] 请求发往 `scm.superboss.cc`（非 erp1）
- [ ] 返回 JSON 信封 `ok: true`（或业务可识别结构）
- [ ] 日志类接口缺 `startTime`/`endTime` 时业务报错符合预期

### 5.6 多格式输出（需 list 有数据）

```bash
kuaimai-cli item +list \
  --body '{"title":"2026","pageNo":1,"pageSize":10}' \
  --output csv

kuaimai-cli item +list \
  --body '{"title":"2026","pageNo":1,"pageSize":10}' \
  --output ndjson
```

- [ ] `csv` 输出表头 + 数据行（失败时为 JSON 信封）
- [ ] `ndjson` 每行一条记录

### 5.7 网络与安全

- [ ] `config get api.pool_max_idle` 有默认值
- [ ] `--verbose` 日志中 token 为脱敏预览

### 5.8 分页与审计

```bash
# 全量翻页（大数据环境）
kuaimai-cli item +list \
  --body '{"title":"2026","pageNo":1,"pageSize":50}' \
  --page-all --output json

# 限制条数 + 自动续查（Agent/脚本）
kuaimai-cli item +list \
  --body '{"title":"2026","pageNo":1,"pageSize":50}' \
  --page-all --page-limit 100 --page-confirm yes --output json

cat ~/.kuaimai-cli/audit.log | tail -5
```

- [ ] `--page-all` 合并多页（数据量足够时）
- [ ] `--page-limit 100` 最多返回 100 条
- [ ] 交互终端达 500 条阈值时提示 `[y/N]`（`--page-confirm prompt`）
- [ ] `audit.log` 含命令与时间戳

---

## 六、阶段四验收

### 6.1 工程化

```bash
go test ./...
go test ./tests/cli_e2e/...
kuaimai-cli doctor --output json
kuaimai-cli upgrade --check-only --output json
kuaimai-cli upgrade   # 有新版时：npm 全局安装 + skill 同步（需网络与 npm）
```

- [ ] 单元测试与 E2E 冒烟通过
- [ ] `doctor` 输出 `ready` 与 `next` 步骤
- [ ] `upgrade --check-only` 可查询 GitHub Release（需网络）
- [ ] 任意命令后 stderr 可出现新版本提示（24h 内可能因缓存不重复）
- [ ] `skill install --if-stale` 在 Release 变化时可更新 Skills

### 6.2 鉴权增强

```bash
kuaimai-cli auth login <token> --profile prod
kuaimai-cli auth list --output json
kuaimai-cli auth use prod
kuaimai-cli auth check --output json
```

- [ ] 多 profile 可登录、切换、`auth check` 探测 token

### 6.3 改标题 shortcut

```bash
kuaimai-cli item update-title --sys-item-id <id> --title "新标题" --dry-run --output json
```

- [ ] dry-run 不发送写请求；正式执行后标题已更新

### 6.4 Registry 中心化（当前主线）

```bash
kuaimai-cli -h | grep -c service    # 期望 0（无 service 命令）
kuaimai-cli capabilities --output json
kuaimai-cli schema api.luotao.test.get --output json
```

- [ ] `service` 命令已移除
- [ ] Skill v3/v2/v1.1 已安装（`doctor` 通过）
- [ ] api-onboard 发布新 apiId 后 `web call` 可调

### 6.5 仍规划

全量 item/scm apiId 发布至远端 · `--page-delay` · `--format pretty`

---

## 七、开发节奏建议

1. 阶段一全部通过后再扩展阶段二  
2. 阶段三以 **item 标题查改** 为业务验收基准  
3. 新接口：在 api-onboard 发布至 open-cli `registry.json` → `registry sync` 验收 → 更新 Skill references；item 高频可选 `shortcuts/item`  

---

## 八、一句话总结

完成**阶段一**即可获得飞书架构同款可用 CLI；**阶段三**以 **`item` 商品标题查改**（`+list` → `save`）为业务验收基准。登录后按本文档第五节逐条执行。
