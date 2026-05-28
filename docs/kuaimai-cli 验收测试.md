# kuaimai-cli 阶段开发准入、最低可用标准、阶段验收测试规范

> 对标飞书 CLI；与 [开发文档](./kuaimai-cli%20开发文档.md)、[每阶段新增能力](./每阶段新增能力.md)、[系统架构说明](./系统架构与飞书对标说明.md) 保持一致。  
> **当前业务验收基准**：**erp-items-core 商品域**，以 **商品标题查改** 为核心场景（`item +list` → `save`）。

---

## 一、核心结论

| 项 | 说明 |
|----|------|
| 开发方式 | 按白皮书分阶段顺序开发，不跳步、不乱改目录 |
| 最低可用 | **阶段一**通过即可使用配置、鉴权、`api`、结构化输出 |
| 日常业务 | **阶段三** + **item shortcuts**（标题查询与修改） |
| 阶段一 | MVP ~60%：基建 + 三级命令形态 |
| 阶段二 | 企业版 ~80%：runner、schema/service、重试、completion |
| 阶段三 | 平台 ~95%：item 域、Skill、csv/ndjson、脱敏、page-all、**标题闭环** |

---

## 二、各阶段能否使用

| 阶段 | 可使用 CLI | 可投入业务 | 飞书完成度 |
|------|------------|------------|------------|
| 一 | ✅ | ✅ 基建与 `api` | ~60% |
| 二 | ✅ | ✅ schema/service、执行管线 | ~80% |
| 三 | ✅ | ✅ **商品标题** list/save/update-title 等 | ~95% |
| 四 | ✅ 部分 | E2E、doctor/upgrade、多账号、扩域仍规划 | ~98% |

---

## 三、阶段一验收清单

### 3.1 架构（硬性）

- [ ] 目录符合白皮书：`cmd/`、`internal/`、`shortcuts/`、`internal/registry/`、`scripts/fetch_meta/`
- [ ] `skills/` 仅 `SKILL.md`，无 Go 业务代码
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
kuaimai-cli config set api.url "https://erp1.superboss.cc/"
kuaimai-cli config set cli.output json
```

验收要点：

- [ ] 再次 `config init` 不覆盖，提示已存在
- [ ] `config set` 为两参数：`key` `value`（无 `=`）
- [ ] 模板含 `api.url`、`api.retry`、`cli.output`、`cli.color`、连接池/熔断项
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
- [ ] `meta_data.json` 嵌入，`schema` / `service` 可用
- [ ] `item` 已在 meta 中注册（当前 v1.2.0，4 operations）

### 4.2 元数据与服务命令

```bash
kuaimai-cli schema --output json | jq '.data.services | length'   # 期望 1（item）
kuaimai-cli service item list --help
kuaimai-cli service item save --help
```

- [ ] `service item list` 与 shortcuts `item list` 路径一致（`/item/stock/queryList`）

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
- [ ] `auth login` 使用有效 `accessToken`

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
kuaimai-cli skill install kuaimai-item --from ./skills
```

- [ ] 能列出 `kuaimai-item`、`kuaimai-shared`
- [ ] `skill install` 写入 `~/.agents/skills/<name>/SKILL.md`

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
kuaimai-cli item +list \
  --body '{"title":"2026","pageNo":1,"pageSize":50}' \
  --page-all --output json

cat ~/.kuaimai-cli/audit.log | tail -5
```

- [ ] `--page-all` 合并多页（数据量足够时）
- [ ] `audit.log` 含命令与时间戳

---

## 六、阶段四验收

### 6.1 工程化

```bash
go test ./...
go test ./tests/cli_e2e/...
kuaimai-cli doctor --output json
kuaimai-cli upgrade --output json
```

- [ ] 单元测试与 E2E 冒烟通过
- [ ] `doctor` 输出 `ready` 与 `next` 步骤
- [ ] `upgrade` 可查询 GitHub Release（需网络）

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

### 6.4 仍规划

扩展更多 erp 业务域 · 自动替换二进制升级

---

## 七、开发节奏建议

1. 阶段一全部通过后再扩展阶段二  
2. 阶段三以 **item 标题查改** 为业务验收基准  
3. 新接口：更新 `shortcuts/item` + `meta_data.json` + Skill + 本文档  

---

## 八、一句话总结

完成**阶段一**即可获得飞书架构同款可用 CLI；**阶段三**以 **`item` 商品标题查改**（`+list` → `save`）为业务验收基准。登录后按本文档第五节逐条执行。
