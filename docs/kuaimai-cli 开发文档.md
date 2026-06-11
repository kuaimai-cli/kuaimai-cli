# kuaimai-cli 架构设计与分阶段开发白皮书（对标飞书 Lark-CLI）

> 配套：[文档索引](./README.md) · [系统架构说明](./系统架构与飞书对标说明.md) · [Registry 同步](./Registry远端同步说明.md) · [接口 JSON 设计](./接口JSON生成与同步系统设计.md) · [验收测试](./kuaimai-cli%20验收测试.md)

---

## 一、项目概述

### 1.1 项目定位

kuaimai-cli 是快麦 ERP 商品与供应链业务专属 CLI，架构对标飞书 **lark-cli**。

**核心目标**：架构对齐飞书；适配人工运维、脚本自动化、AI Agent。

**当前状态**：

| 模块 | 状态 |
|------|------|
| 基建 | config / auth / api / output / pagination / doctor / upgrade |
| Registry | 远端 v2 JSON + 自动同步 + `capabilities` / `schema` / `web call` |
| Shortcuts | item 域 6 子命令（标题查改闭环） |
| Skill | shared v1.1 + item v3 + scm v2（工作流手册，非接口目录） |
| 已移除 | `service` 命令、内嵌 `meta_data.json` |

### 1.2 设计原则

- **registry 与 CLI 解耦**：接口由 `kuaimaierp-cli-auto` 生成，CLI 只消费
- **分层解耦**：`cmd` → `internal` → `shortcuts`
- **Skill 管工作流，registry 管接口**

---

## 二、技术栈

Go 1.22+ · Cobra · Viper · go-keyring · net/http · go-pretty

---

## 三、目录架构

```text
kuaimai-cli/
├── cmd/
│   ├── root.go, bootstrap.go
│   ├── webcmd/           # web call <apiId>
│   ├── capabilitiescmd/, schemacmd/, registrycmd/
│   ├── authcmd/, doctorcmd/, upgradecmd/
│   └── ...
├── internal/
│   ├── registry/         # v2 解析、Sync、本地缓存
│   ├── apicall/          # registry 驱动 HTTP
│   ├── client/, pagination/, output/
│   └── ...
├── shortcuts/item/       # 6 个 curated 命令
├── skills/               # Agent 工作流（随 npm 发布）
├── npm/skills/           # sync-skills.js 同步副本
└── tests/cli_e2e/
```

---

## 四、统一命令体系

### 4.1 全局命令

```bash
kuaimai-cli config init | auth login | doctor | upgrade

kuaimai-cli capabilities --output json
kuaimai-cli schema [apiId] --output json
kuaimai-cli web call <apiId> [--params|--data|--body '...']
kuaimai-cli registry sync | watch

kuaimai-cli item +list | count | get-detail | save | update-title
kuaimai-cli api POST /path
kuaimai-cli skill install
```

### 4.2 三级命令（对齐飞书）

| 层级 | 飞书 | 快麦 |
|------|------|------|
| Shortcuts | `calendar +agenda` | `item +list` |
| 元数据 | `calendar events list` | `web call <apiId>` |
| 兜底 | `api GET /open-apis/...` | `api POST /path` |

### 4.3 Registry v2（接口真相源）

```text
kuaimaierp-cli-auto → open-cli.kuaimai.com/registry/registry.json
    → kuaimai-cli bootstrapRegistry
    → ~/.kuaimai-cli/registry/registry.json
    → capabilities / schema / web call
```

字段规范见 [接口 JSON 生成与同步系统设计](./接口JSON生成与同步系统设计.md)；历史 v1 字段对照见 [meta_data.json 定义规范](./kuaimai-cli%20meta_data.json%20定义规范.md)。

### 4.4 item shortcuts 索引

| 子命令 | 路径 |
|--------|------|
| `+list` / `list` | POST `/item/stock/queryList` |
| `count` | POST `/item/stock/queryCount` |
| `get-detail` | GET `/item/getItemDetail` |
| `save` / `update-title` | POST `/item/saveItem` |

对应 registry apiId（待发布）：`item.stock-list`、`item.stock-count` 等。

### 4.5 Skill 平台

| Skill | 版本 | 职责 |
|-------|------|------|
| kuaimai-shared | v1.1 | auth、registry 发现、输出、安全 |
| kuaimai-item | v3 | 商品意图路由、shortcuts、references |
| kuaimai-scm | v2 | 供应链意图路由、references |

发版前：`node npm/scripts/sync-skills.js`

**新增接口（维护者）**：

1. 在 `kuaimaierp-cli-auto` 登记并发布至 open-cli registry
2. `kuaimai-cli registry sync` 验证 `capabilities` / `schema` / `web call`
3. item 高频接口可选：补 `shortcuts/item` + 更新 Skill references
4. 验收见 [验收测试](./kuaimai-cli%20验收测试.md)

---

## 五、分阶段落地（摘要）

| 阶段 | 内容 | 状态 |
|------|------|------|
| 一 | 基建、api、输出信封 | ✅ |
| 二 | runner、schema、web call、分页 | ✅ |
| 三 | item shortcuts、标题闭环、Skill | ✅ |
| 四 | E2E、doctor、upgrade、registry v2 中心化 | ✅ 进行中 |
| 五 | 全量 apiId 发布至远端 registry | 进行中 |

---

## 六、本地调试

```bash
make build
./kuaimai-cli registry sync --output json
./kuaimai-cli capabilities --output json
./kuaimai-cli schema api.luotao.test.get --output json
./kuaimai-cli item +list --body '{"title":"测试","pageNo":1,"pageSize":10}' --output json
go test -mod=vendor ./...
```

---

## 七、与飞书差异

**已对齐**：三级命令 · registry/schema · Skill+references · 密钥链 · 结构化输出

**差异**：飞书 18 域顶层命令 vs 快麦统一 `web call`；飞书 OAuth vs accessToken；远端 registry 逐步扩充

---

## 八、开发自查

- [ ] 无 `service` 命令残留
- [ ] `capabilities` / `schema` / `web call` 可走通
- [ ] `skills/` 已 sync 到 `npm/skills/`
- [ ] 新接口走 api-onboard 发布，非手写 CLI 内嵌 meta
- [ ] `go test -mod=vendor ./...` 通过

详见 [验收测试](./kuaimai-cli%20验收测试.md)、[开发发布流程](./快麦%20CLI%20开发发布流程文档.md)。
